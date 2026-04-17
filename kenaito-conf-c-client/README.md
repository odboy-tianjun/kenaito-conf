# Kenaito Config Center C Client

C 语言的配置中心客户端 SDK，适用于嵌入式设备或高性能场景，依赖 libcurl。

## 依赖

- libcurl

### Ubuntu/Debian

```bash
apt-get install libcurl4-openssl-dev
```

### CentOS/RHEL

```bash
yum install libcurl-devel
```

> **macOS**：系统自带 libcurl，无需额外安装。

## 快速开始

### 1. 安装依赖

#### Ubuntu/Debian

```bash
apt-get install libcurl4-openssl-dev libcjson-dev
```

#### CentOS/RHEL

```bash
yum install libcurl-devel libcjson-devel
```

#### macOS

```bash
brew install cJSON
```

> **macOS**：系统自带 libcurl，无需额外安装。

### 2. 编译 SDK

**macOS：**

```bash
cd kenaito-conf-c-client
gcc -c main.c -o kenaito_conf.o -I/usr/local/opt/cjson/include -lcurl
ar rcs libkenaito_conf.a kenaito_conf.o
```

**Ubuntu/Debian：**

```bash
cd kenaito-conf-c-client
gcc -c main.c -o kenaito_conf.o $(pkg-config --cflags libcjson) -lcurl
ar rcs libkenaito_conf.a kenaito_conf.o
```

### 3. 编译 Demo

**macOS：**

```bash
cd kenaito-conf-c-client-demo
gcc -o demo main.c ../kenaito-conf-c-client/main.c -I/usr/local/opt/cjson/include -lcurl -lpthread
./demo
```

**Ubuntu/Debian：**

```bash
cd kenaito-conf-c-client-demo
gcc -o demo main.c ../kenaito-conf-c-client/main.c $(pkg-config --cflags libcjson) -lcurl -lpthread
./demo
```

### 4. 使用 SDK

在你的项目中：

```c
#include "kenaito-conf-c-client/main.c"

// 或者链接静态库
// gcc -o myapp myapp.c -lkenaito_conf -lcurl -lcjson -lpthread
```

### 5. 完整示例

```c
#include <stdio.h>
#include "../kenaito-conf-c-client/main.c"

void on_change(const char *key, const char *value) {
    printf("config changed: %s = %s\n", key, value);
}

int main() {
    Config config = {
        .baseUrl = "http://localhost:8080",
        .appName = "my-app",
        .env = "daily",
        .version = "latest",
        .token = "",
        .syncInterval = 10,
        .ignoreGlobal = 0,
        .cacheDir = "."
    };

    ConfigClient client;
    config_client_init(&client, &config);
    client.onChange = on_change;

    // 启动客户端（拉取全量配置 + 开启定时同步）
    config_client_start(&client);

    // 获取配置
    const char *dbHost = config_client_get(&client, "database.host");
    const char *dbPort = config_client_get(&client, "database.port");

    printf("database.host = %s\n", dbHost ? dbHost : "not found");
    printf("database.port = %s\n", dbPort ? dbPort : "not found");

    // 等待配置变更
    sleep(30);

    // 停止客户端
    config_client_stop(&client);
    return 0;
}
```

### 3. 配置变更监听

```c
void on_config_change(const char *key, const char *value) {
    printf("[ConfigChange] %s = %s\n", key, value);
    // 在此处处理配置更新逻辑
    // 例如：重新初始化数据库连接、更新日志级别等
}

// 注册回调
client.onChange = on_config_change;
```

## 配置参数

| 参数 | 说明 | 默认值 | 必填 |
|------|------|--------|------|
| `baseUrl` | 配置中心地址 | - | 是 |
| `appName` | 应用名称 | - | 是 |
| `env` | 环境（dev/daily/prod） | `dev` | 否 |
| `version` | 配置版本 | `latest` | 否 |
| `token` | 认证令牌 | 空 | 否 |
| `syncInterval` | 同步间隔（秒） | `10` | 否 |
| `ignoreGlobal` | 是否忽略全局配置 | `0` | 否 |
| `cacheDir` | 本地缓存目录 | `.` | 否 |

## API 参考

### 初始化客户端

```c
void config_client_init(ConfigClient *client, Config *config);
```

### 获取配置值

```c
const char *config_client_get(ConfigClient *client, const char *key);
```

**参数：**
- `client`：客户端实例
- `key`：配置键名

**返回值：**
- 成功：配置值字符串
- 失败：`NULL`

### 停止客户端

```c
void config_client_stop(ConfigClient *client);
```

**说明：** 释放 libcurl 资源，停止定时同步。

### 注册变更回调

```c
client.onChange = on_config_change;
```

**回调签名：**
```c
void on_config_change(const char *key, const char *value);
```

## 完整示例代码

```c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>
#include <time.h>

// 此处引入 kenaito-conf-c-client/main.c 的完整实现
// 或使用编译好的库文件

void on_change(const char *key, const char *value) {
    printf("[ConfigChange] %s = %s\n", key, value);
}

int main() {
    printf("=== Kenaito Config Center C Client Demo ===\n");

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
    client.onChange = on_change;

    // 获取配置
    const char *value = config_client_get(&client, "database.host");
    if (value) {
        printf("database.host = %s\n", value);
    } else {
        printf("database.host not found\n");
    }

    // 等待配置变更（实际应用中应在此处运行业务逻辑）
    printf("Listening for config changes...\n");
    sleep(30);

    config_client_stop(&client);
    printf("Client stopped.\n");
    return 0;
}
```

## Demo

运行示例项目：

```bash
cd kenaito-conf-c-client-demo
gcc -o demo ../kenaito-conf-c-client/main.c -lcurl
./demo
```

Demo 默认配置：
- 应用名称：`kenaito-springboot-demo`
- 环境：`daily`
- 配置中心地址：`http://localhost:8080`
- 同步间隔：10 秒

## 注意事项

1. **JSON 解析**：当前示例代码仅演示 HTTP 请求，实际使用需集成 JSON 解析库（如 cJSON）
2. **配置缓存**：建议实现本地文件缓存，避免频繁请求配置中心
3. **线程安全**：多线程环境下需自行添加锁机制保护配置读取
4. **错误处理**：生产环境应完善网络异常、超时、重试等错误处理逻辑
5. **内存管理**：使用 `config_client_stop()` 释放资源前不要访问配置

## 与其他客户端对比

| 特性 | C 客户端 | Go 客户端 | Java 客户端 |
|------|---------|----------|------------|
| 编译方式 | 手动 gcc | go build | Maven |
| 依赖管理 | 手动链接 | go mod | Maven |
| JSON 解析 | 需自行集成 | 标准库 | FastJSON2 |
| 定时同步 | 需自行实现 | ✅ | ✅ |
| 配置缓存 | 需自行实现 | ✅ | ✅ |
| 变更回调 | ✅ | ✅ | ✅ |
| 线程安全 | 需自行实现 | ✅ | ✅ |

## 版本历史

- **1.0.0** - 初始版本
  - 基础 HTTP 拉取配置功能
  - 配置变更回调接口
  - 简化示例代码

## 后续开发计划

- [ ] 集成 cJSON 实现 JSON 解析
- [ ] 实现定时轮询同步机制
- [ ] 添加本地文件缓存支持
- [ ] 完善错误处理和重试机制
- [ ] 添加线程安全保护
