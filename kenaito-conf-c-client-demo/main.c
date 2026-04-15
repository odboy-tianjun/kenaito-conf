/**
 * Kenaito Config Center C Client Demo
 * 
 * 演示如何使用 C 客户端 SDK：
 *   - 初始化客户端
 *   - 注册配置变更回调
 *   - 启动/停止客户端
 *   - 读取配置
 * 
 * 编译：
 *   gcc -o demo main.c -I../kenaito-conf-c-client -L../kenaito-conf-c-client -lkenaito_conf -lcurl -lcjson -lpthread
 * 
 * 或直接编译所有源文件：
 *   gcc -o demo main.c ../kenaito-conf-c-client/main.c -lcurl -lcjson -lpthread
 * 
 * 运行：
 *   ./demo
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include "../kenaito-conf-c-client/main.c"

/* 全局客户端实例，用于信号处理 */
static ConfigClient *g_demo_client = NULL;

/**
 * 配置变更回调示例
 */
void on_config_change(const char *key, const char *value) {
    printf("\n[Callback] Config changed: %s = %s\n", key, value);
    
    /* 示例：根据配置键执行不同操作 */
    if (strcmp(key, "log.level") == 0) {
        printf("  -> Log level changed to: %s\n", value);
    } else if (strcmp(key, "database.host") == 0) {
        printf("  -> Database host changed to: %s\n", value);
        /* 实际应用中可在此重新初始化数据库连接 */
    } else if (strcmp(key, "database.port") == 0) {
        printf("  -> Database port changed to: %s\n", value);
    } else if (strcmp(key, "feature.enabled") == 0) {
        printf("  -> Feature flag changed: %s\n", value);
    }
}

/**
 * 信号处理 - 优雅退出
 */
static void signal_handler(int sig) {
    printf("\n[Signal] Received signal %d, shutting down...\n", sig);
    if (g_demo_client) {
        g_demo_client->running = 0;
    }
}

int main() {
    printf("========================================\n");
    printf("  Kenaito Config Center C Client Demo\n");
    printf("========================================\n\n");
    
    /* 注册信号处理 */
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    /* 初始化 CURL */
    curl_global_init(CURL_GLOBAL_DEFAULT);
    
    /* 配置客户端 */
    Config config = {
        .baseUrl = "http://localhost:8080",
        .appName = "kenaito-springboot-demo",
        .env = "daily",
        .version = "latest",
        .token = "",
        .syncInterval = 10,
        .ignoreGlobal = 0,
        .cacheDir = "."
    };
    
    ConfigClient client;
    config_client_init(&client, &config);
    client.onChange = on_config_change;
    
    g_demo_client = &client;
    
    /* 启动客户端 */
    config_client_start(&client);
    
    /* 主循环：演示配置读取 */
    printf("\n[Demo] Running... Press Ctrl+C to stop\n\n");
    
    int iterations = 0;
    while (client.running && iterations < 10) {
        sleep(5);
        
        /* 演示：定期读取并打印配置 */
        const char *dbHost = config_client_get(&client, "database.host");
        const char *dbPort = config_client_get(&client, "database.port");
        const char *logLevel = config_client_get(&client, "log.level");
        const char *featureEnabled = config_client_get(&client, "feature.enabled");
        
        printf("\n[Current Configs @ %ld]\n", time(NULL));
        printf("  database.host    = %s\n", dbHost ? dbHost : "(not set)");
        printf("  database.port    = %s\n", dbPort ? dbPort : "(not set)");
        printf("  log.level        = %s\n", logLevel ? logLevel : "(not set)");
        printf("  feature.enabled  = %s\n", featureEnabled ? featureEnabled : "(not set)");
        printf("  Total configs: %d\n\n", client.configs.count);
        
        iterations++;
    }
    
    /* 停止客户端 */
    config_client_stop(&client);
    
    printf("\n[Demo] Finished\n");
    return 0;
}
