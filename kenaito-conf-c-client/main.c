#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>
#include <openssl/evp.h>
#include <openssl/md5.h>
#include <openssl/aes.h>
#include <time.h>

#define MAX_CONFIGS 256
#define MAX_KEY_LEN 256
#define MAX_VAL_LEN 4096
#define CRYPT_PREFIX "{Cipher}"

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

typedef struct {
    char key[MAX_KEY_LEN];
    char value[MAX_VAL_LEN];
} ConfigEntry;

typedef struct {
    ConfigEntry entries[MAX_CONFIGS];
    int count;
} ConfigMap;

typedef struct {
    Config config;
    ConfigMap configs;
    char secret[256];
    char globalSecret[256];
    long lastTime;
    void (*onChange)(const char *key, const char *value);
} ConfigClient;

typedef struct {
    char *data;
    size_t size;
} ResponseBuffer;

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

static void md5_hex(const char *input, char *output) {
    unsigned char digest[MD5_DIGEST_LENGTH];
    MD5((unsigned char *)input, strlen(input), digest);
    for (int i = 0; i < MD5_DIGEST_LENGTH; i++)
        sprintf(output + i * 2, "%02x", digest[i]);
    output[32] = 0;
}

static int base64_decode(const char *src, unsigned char *dst) {
    BIO *b64 = BIO_new(BIO_f_base64());
    BIO *mem = BIO_new_mem_buf((void *)src, strlen(src));
    mem = BIO_push(b64, mem);
    BIO_set_flags(mem, BIO_FLAGS_BASE64_NO_NL);
    int len = BIO_read(mem, dst, strlen(src));
    BIO_free_all(mem);
    return len;
}

static void decrypt_with_aes(const char *key, const char *data, char *output) {
    char md5Hash[33];
    md5_hex(key, md5Hash);
    unsigned char aesKey[16];
    memcpy(aesKey, md5Hash + 16, 16);

    unsigned char *cipherText = malloc(strlen(data) * 2);
    int cipherLen = base64_decode(data, cipherText);

    unsigned char iv[16] = {0};
    unsigned char plainText[4096];

    EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
    EVP_DecryptInit_ex(ctx, EVP_aes_128_cbc(), NULL, aesKey, iv);
    EVP_CIPHER_CTX_set_padding(ctx, 0);
    int len = 0, totalLen = 0;
    EVP_DecryptUpdate(ctx, plainText, &len, cipherText, cipherLen);
    totalLen = len;
    EVP_DecryptFinal_ex(ctx, plainText + len, &len);
    totalLen += len;
    EVP_CIPHER_CTX_free(ctx);

    int padding = plainText[totalLen - 1];
    totalLen -= padding;
    memcpy(output, plainText, totalLen);
    output[totalLen] = 0;

    free(cipherText);
}

static void decrypt_config(const char *secret, const char *data, char *output) {
    if (strncmp(data, CRYPT_PREFIX, strlen(CRYPT_PREFIX)) != 0) {
        strcpy(output, data);
        return;
    }
    decrypt_with_aes(secret, data + strlen(CRYPT_PREFIX), output);
}

static char *http_post_json(const char *url, const char *jsonBody) {
    CURL *curl = curl_easy_init();
    if (!curl) return NULL;

    ResponseBuffer buf = {0};
    struct curl_slist *headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");

    curl_easy_setopt(curl, CURLOPT_URL, url);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonBody);
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &buf);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);

    CURLcode res = curl_easy_perform(curl);
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);

    if (res != CURLE_OK) {
        free(buf.data);
        return NULL;
    }
    return buf.data;
}

void config_client_init(ConfigClient *client, Config *config) {
    client->config = *config;
    client->configs.count = 0;
    client->secret[0] = 0;
    client->globalSecret[0] = 0;
    client->lastTime = 0;
    client->onChange = NULL;
    curl_global_init(CURL_GLOBAL_DEFAULT);
}

const char *config_client_get(ConfigClient *client, const char *key) {
    for (int i = 0; i < client->configs.count; i++) {
        if (strcmp(client->configs.entries[i].key, key) == 0)
            return client->configs.entries[i].value;
    }
    return NULL;
}

void config_client_stop(ConfigClient *client) {
    curl_global_cleanup();
}

int main() {
    printf("Kenaito Config Center C Client Demo\n");
    printf("AppName: kenaito-springboot-demo, Env: daily\n");

    Config config = {
        .baseUrl = "http://localhost:8080",
        .appName = "kenaito-springboot-demo",
        .env = "daily",
        .version = "latest",
        .token = "",
        .syncInterval = 30,
        .ignoreGlobal = 0,
        .cacheDir = "."
    };

    ConfigClient client;
    config_client_init(&client, &config);

    /* Build fetchAllConfigs request */
    char url[1024];
    snprintf(url, sizeof(url), "%s/api/fetchAllConfigs", config.baseUrl);

    char payload[512];
    snprintf(payload, sizeof(payload),
        "{\"appName\":\"%s\",\"env\":\"%s\",\"version\":\"%s\",\"ignoreGlobal\":false,\"token\":\"%s\"}",
        config.appName, config.env, config.version, config.token);

    printf("Fetching configs from %s ...\n", url);
    char *response = http_post_json(url, payload);
    if (response) {
        printf("Response received (%zu bytes)\n", strlen(response));
        /* In production: parse JSON, decrypt, store configs */
        printf("First 200 chars: %.200s\n", response);
        free(response);
    } else {
        printf("Failed to fetch configs (server may not be running)\n");
    }

    config_client_stop(&client);
    return 0;
}
