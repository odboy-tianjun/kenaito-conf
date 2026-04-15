/**
 * Kenaito Config Center C Client SDK
 * 
 * 完整的配置中心客户端，支持：
 *   - JSON 解析（cJSON）
 *   - 配置缓存到本地文件
 *   - 定时同步配置变更
 *   - 配置变更回调
 *   - 全量/增量拉取
 * 
 * 依赖：
 *   - libcurl
 *   - cJSON
 *   - pthread
 * 
 * 编译：
 *   gcc -c main.c -o kenaito_conf.o -lcurl -lcjson
 * 
 * @author odboy
 * @version 1.0.0
 * @date 2026-10-16
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>
#include <cjson/cJSON.h>
#include <pthread.h>
#include <unistd.h>
#include <time.h>

#define MAX_CONFIGS 256
#define MAX_KEY_LEN 256
#define MAX_VAL_LEN 4096
#define CACHE_FILE ".kenaito-config-cache.json"

/* ============ 配置结构 ============ */
typedef struct {
    char baseUrl[512];
    char appName[128];
    char env[32];
    char version[32];
    char token[256];
    int syncInterval;
    int ignoreGlobal;
    char cacheDir[256];
} Config;

/* ============ 配置条目 ============ */
typedef struct {
    char key[MAX_KEY_LEN];
    char value[MAX_VAL_LEN];
} ConfigEntry;

/* ============ 配置映射表 ============ */
typedef struct {
    ConfigEntry entries[MAX_CONFIGS];
    int count;
} ConfigMap;

/* ============ 客户端实例 ============ */
typedef struct {
    Config config;
    ConfigMap configs;
    long lastTime;
    void (*onChange)(const char *key, const char *value);
    CURL *curl;
    int running;
    pthread_t syncThread;
} ConfigClient;

/* ============ HTTP 响应缓冲区 ============ */
typedef struct {
    char *data;
    size_t size;
} ResponseBuffer;

/* 全局客户端实例 */
static ConfigClient *g_client = NULL;

/* ============ 工具函数 ============ */

/* HTTP 响应回调 */
static size_t write_callback(void *contents, size_t size, size_t nmemb, void *userp) {
    size_t totalSize = size * nmemb;
    ResponseBuffer *buf = (ResponseBuffer *)userp;
    char *ptr = realloc(buf->data, buf->size + totalSize + 1);
    if (!ptr) return 0;
    buf->data = ptr;
    memcpy(&(buf->data[buf->size]), contents, totalSize);
    buf->size += totalSize;
    buf->data[buf->size] = 0;
    return totalSize;
}

/* HTTP POST 请求 */
static char *http_post(ConfigClient *client, const char *url, const char *jsonBody) {
    ResponseBuffer buf = {0};
    struct curl_slist *headers = NULL;
    
    headers = curl_slist_append(headers, "Content-Type: application/json");
    headers = curl_slist_append(headers, "Expect:");
    
    curl_easy_setopt(client->curl, CURLOPT_URL, url);
    curl_easy_setopt(client->curl, CURLOPT_POSTFIELDS, jsonBody);
    curl_easy_setopt(client->curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(client->curl, CURLOPT_WRITEFUNCTION, write_callback);
    curl_easy_setopt(client->curl, CURLOPT_WRITEDATA, &buf);
    curl_easy_setopt(client->curl, CURLOPT_TIMEOUT, 10L);
    curl_easy_setopt(client->curl, CURLOPT_FOLLOWLOCATION, 1L);
    
    CURLcode res = curl_easy_perform(client->curl);
    curl_slist_free_all(headers);
    
    if (res != CURLE_OK) {
        fprintf(stderr, "HTTP request failed: %s\n", curl_easy_strerror(res));
        free(buf.data);
        return NULL;
    }
    
    long httpCode = 0;
    curl_easy_getinfo(client->curl, CURLINFO_RESPONSE_CODE, &httpCode);
    if (httpCode != 200) {
        fprintf(stderr, "HTTP %ld: %s\n", httpCode, buf.data ? buf.data : "empty response");
        free(buf.data);
        return NULL;
    }
    
    return buf.data;
}

/* 从缓存加载配置 */
static int load_from_cache(ConfigClient *client) {
    FILE *f = fopen(CACHE_FILE, "r");
    if (!f) return 0;
    
    fseek(f, 0, SEEK_END);
    long fsize = ftell(f);
    fseek(f, 0, SEEK_SET);
    
    char *jsonStr = malloc(fsize + 1);
    fread(jsonStr, 1, fsize, f);
    fclose(f);
    jsonStr[fsize] = 0;
    
    cJSON *root = cJSON_Parse(jsonStr);
    free(jsonStr);
    
    if (!root) return 0;
    
    cJSON *lastTimeObj = cJSON_GetObjectItem(root, "lastTime");
    if (lastTimeObj) {
        client->lastTime = lastTimeObj->valueint;
    }
    
    cJSON *configsObj = cJSON_GetObjectItem(root, "configs");
    if (configsObj && cJSON_IsArray(configsObj)) {
        client->configs.count = 0;
        int arrSize = cJSON_GetArraySize(configsObj);
        for (int i = 0; i < arrSize && client->configs.count < MAX_CONFIGS; i++) {
            cJSON *item = cJSON_GetArrayItem(configsObj, i);
            cJSON *keyObj = cJSON_GetObjectItem(item, "key");
            cJSON *valObj = cJSON_GetObjectItem(item, "value");
            
            if (keyObj && valObj && cJSON_IsString(keyObj) && cJSON_IsString(valObj)) {
                strncpy(client->configs.entries[client->configs.count].key, 
                    keyObj->valuestring, MAX_KEY_LEN - 1);
                strncpy(client->configs.entries[client->configs.count].value, 
                    valObj->valuestring, MAX_VAL_LEN - 1);
                client->configs.count++;
            }
        }
    }
    
    cJSON_Delete(root);
    printf("[Cache] Loaded %d configs from %s\n", client->configs.count, CACHE_FILE);
    return 1;
}

/* 保存配置到缓存 */
static void save_to_cache(ConfigClient *client) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddNumberToObject(root, "lastTime", client->lastTime);
    
    cJSON *configsArr = cJSON_CreateArray();
    for (int i = 0; i < client->configs.count; i++) {
        cJSON *item = cJSON_CreateObject();
        cJSON_AddStringToObject(item, "key", client->configs.entries[i].key);
        cJSON_AddStringToObject(item, "value", client->configs.entries[i].value);
        cJSON_AddItemToArray(configsArr, item);
    }
    cJSON_AddItemToObject(root, "configs", configsArr);
    
    char *jsonStr = cJSON_PrintUnformatted(root);
    FILE *f = fopen(CACHE_FILE, "w");
    if (f) {
        fprintf(f, "%s", jsonStr);
        fclose(f);
        printf("[Cache] Saved %d configs to %s\n", client->configs.count, CACHE_FILE);
    }
    
    free(jsonStr);
    cJSON_Delete(root);
}

/* 获取配置值 */
const char *config_client_get(ConfigClient *client, const char *key) {
    for (int i = 0; i < client->configs.count; i++) {
        if (strcmp(client->configs.entries[i].key, key) == 0)
            return client->configs.entries[i].value;
    }
    return NULL;
}

/* 解析 JSON 响应 */
static int parse_json_response(const char *jsonStr, ConfigClient *client) {
    cJSON *root = cJSON_Parse(jsonStr);
    if (!root) {
        fprintf(stderr, "Failed to parse JSON: %s\n", cJSON_GetErrorPtr());
        return -1;
    }
    
    cJSON *successObj = cJSON_GetObjectItem(root, "success");
    if (successObj && cJSON_IsBool(successObj) && !successObj->valueint) {
        cJSON *msgObj = cJSON_GetObjectItem(root, "message");
        fprintf(stderr, "Server error: %s\n", msgObj ? msgObj->valuestring : "unknown");
        cJSON_Delete(root);
        return -1;
    }
    
    cJSON *dataObj = cJSON_GetObjectItem(root, "data");
    if (!dataObj) {
        fprintf(stderr, "No data in response\n");
        cJSON_Delete(root);
        return -1;
    }
    
    /* 全量配置 */
    cJSON *configsObj = cJSON_GetObjectItem(dataObj, "configs");
    if (configsObj && cJSON_IsArray(configsObj)) {
        client->configs.count = 0;
        int arrSize = cJSON_GetArraySize(configsObj);
        for (int i = 0; i < arrSize && client->configs.count < MAX_CONFIGS; i++) {
            cJSON *item = cJSON_GetArrayItem(configsObj, i);
            cJSON *keyObj = cJSON_GetObjectItem(item, "key");
            cJSON *valObj = cJSON_GetObjectItem(item, "value");
            
            if (keyObj && valObj && cJSON_IsString(keyObj) && cJSON_IsString(valObj)) {
                strncpy(client->configs.entries[client->configs.count].key, 
                    keyObj->valuestring, MAX_KEY_LEN - 1);
                
                const char *oldVal = config_client_get(client, keyObj->valuestring);
                
                strncpy(client->configs.entries[client->configs.count].value, 
                    valObj->valuestring, MAX_VAL_LEN - 1);
                
                /* 检测变更并触发回调 */
                if (client->onChange && oldVal && strcmp(oldVal, valObj->valuestring) != 0) {
                    printf("[Change] %s = %s -> %s\n", 
                        keyObj->valuestring, oldVal, valObj->valuestring);
                    client->onChange(keyObj->valuestring, valObj->valuestring);
                }
                
                client->configs.count++;
            }
        }
    }
    
    /* 增量变更 */
    cJSON *changesObj = cJSON_GetObjectItem(dataObj, "changes");
    if (changesObj && cJSON_IsArray(changesObj)) {
        int arrSize = cJSON_GetArraySize(changesObj);
        for (int i = 0; i < arrSize; i++) {
            cJSON *item = cJSON_GetArrayItem(changesObj, i);
            cJSON *keyObj = cJSON_GetObjectItem(item, "key");
            cJSON *valObj = cJSON_GetObjectItem(item, "value");
            
            if (keyObj && valObj && cJSON_IsString(keyObj) && cJSON_IsString(valObj)) {
                /* 更新或添加配置 */
                int found = 0;
                for (int j = 0; j < client->configs.count; j++) {
                    if (strcmp(client->configs.entries[j].key, keyObj->valuestring) == 0) {
                        const char *oldVal = client->configs.entries[j].value;
                        strncpy(client->configs.entries[j].value, 
                            valObj->valuestring, MAX_VAL_LEN - 1);
                        
                        if (client->onChange && strcmp(oldVal, valObj->valuestring) != 0) {
                            printf("[Change] %s = %s -> %s\n", 
                                keyObj->valuestring, oldVal, valObj->valuestring);
                            client->onChange(keyObj->valuestring, valObj->valuestring);
                        }
                        found = 1;
                        break;
                    }
                }
                
                if (!found && client->configs.count < MAX_CONFIGS) {
                    strncpy(client->configs.entries[client->configs.count].key, 
                        keyObj->valuestring, MAX_KEY_LEN - 1);
                    strncpy(client->configs.entries[client->configs.count].value, 
                        valObj->valuestring, MAX_VAL_LEN - 1);
                    client->configs.count++;
                    
                    if (client->onChange) {
                        printf("[New] %s = %s\n", keyObj->valuestring, valObj->valuestring);
                        client->onChange(keyObj->valuestring, valObj->valuestring);
                    }
                }
            }
        }
    }
    
    /* 更新 lastTime */
    cJSON *lastTimeObj = cJSON_GetObjectItem(dataObj, "lastTime");
    if (lastTimeObj && cJSON_IsNumber(lastTimeObj)) {
        client->lastTime = lastTimeObj->valueint;
    }
    
    cJSON_Delete(root);
    return 0;
}

/* ============ 客户端 API ============ */

/**
 * 初始化客户端
 * @param client 客户端实例
 * @param config 配置参数
 */
void config_client_init(ConfigClient *client, Config *config) {
    client->config = *config;
    client->configs.count = 0;
    client->lastTime = 0;
    client->onChange = NULL;
    client->running = 0;
    client->curl = curl_easy_init();
    
    g_client = client;
    
    /* 尝试从缓存加载 */
    if (!load_from_cache(client)) {
        printf("[Cache] No cache found, will fetch from server\n");
    }
}

/**
 * 拉取配置（全量或增量）
 * @param client 客户端实例
 * @param incremental 是否增量拉取
 * @return 0 成功，-1 失败
 */
static int fetch_configs(ConfigClient *client, int incremental) {
    char url[1024];
    char payload[1024];
    
    if (incremental && client->lastTime > 0) {
        snprintf(url, sizeof(url), "%s/api/fetchChangedConfigs", client->config.baseUrl);
        snprintf(payload, sizeof(payload),
            "{\"appName\":\"%s\",\"env\":\"%s\",\"version\":\"%s\","
            "\"ignoreGlobal\":%s,\"token\":\"%s\",\"lastTime\":%ld}",
            client->config.appName, client->config.env, client->config.version,
            client->config.ignoreGlobal ? "true" : "false",
            client->config.token, client->lastTime);
    } else {
        snprintf(url, sizeof(url), "%s/api/fetchAllConfigs", client->config.baseUrl);
        snprintf(payload, sizeof(payload),
            "{\"appName\":\"%s\",\"env\":\"%s\",\"version\":\"%s\","
            "\"ignoreGlobal\":%s,\"token\":\"%s\"}",
            client->config.appName, client->config.env, client->config.version,
            client->config.ignoreGlobal ? "true" : "false",
            client->config.token);
    }
    
    printf("[%s] Fetching from %s...\n", incremental ? "Incremental" : "Full", url);
    char *response = http_post(client, url, payload);
    
    if (!response) {
        fprintf(stderr, "Failed to fetch configs\n");
        return -1;
    }
    
    int ret = parse_json_response(response, client);
    free(response);
    
    if (ret == 0) {
        save_to_cache(client);
    }
    
    return ret;
}

/* 定时同步线程 */
static void *sync_thread_func(void *arg) {
    ConfigClient *client = (ConfigClient *)arg;
    
    while (client->running) {
        sleep(client->config.syncInterval);
        
        if (!client->running) break;
        
        fetch_configs(client, 1);
    }
    
    return NULL;
}

/**
 * 启动客户端
 * @param client 客户端实例
 */
void config_client_start(ConfigClient *client) {
    /* 首次全量拉取 */
    fetch_configs(client, 0);
    
    /* 启动定时同步线程 */
    client->running = 1;
    pthread_create(&client->syncThread, NULL, sync_thread_func, client);
    printf("[Sync] Started background sync thread (interval: %ds)\n", 
        client->config.syncInterval);
}

/**
 * 停止客户端
 * @param client 客户端实例
 */
void config_client_stop(ConfigClient *client) {
    client->running = 0;
    
    if (client->syncThread) {
        pthread_join(client->syncThread, NULL);
    }
    
    save_to_cache(client);
    
    if (client->curl) {
        curl_easy_cleanup(client->curl);
    }
    
    curl_global_cleanup();
    printf("[Client] Stopped\n");
}
