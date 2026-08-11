# Kenaito Config Center C++ Client

C++ 客户端 SDK，适用于 C++11 及以上项目。

## 功能特性

- 从配置中心拉取应用配置
- 定时同步配置变更
- 支持配置变更回调监听
- 本地文件缓存（配置中心不可用时降级）

## 依赖

- C++11 或更高版本
- libcurl
- nlohmann/json（已包含在 third_party 目录）

## 编译

```bash
cd kenaito-conf-cpp-client
mkdir build && cd build
cmake ..
make
```

## 使用

```cpp
#include <kenaito_conf/client.h>
#include <iostream>

int main() {
    kenaito::Config config;
    config.base_url = "http://localhost:8080";
    config.app_name = "my-app";
    config.env = "daily";
    config.sync_interval = 10;
    config.cache_dir = ".";

    kenaito::Client client(config);

    // 监听配置变更
    client.on_change([](const std::map<std::string, std::string>& changed) {
        std::cout << "Config changed:" << std::endl;
        for (const auto& [k, v] : changed) {
            std::cout << "  " << k << " = " << v << std::endl;
        }
    });

    // 启动客户端
    if (!client.start()) {
        std::cerr << "Failed to start config client" << std::endl;
        return 1;
    }

    // 获取配置
    std::string db_host = client.get("database.host");
    std::cout << "database.host = " << db_host << std::endl;

    // 获取所有配置
    auto all = client.get_all();
    for (const auto& [k, v] : all) {
        std::cout << k << " = " << v << std::endl;
    }

    // 手动同步
    client.sync_once();

    // 停止
    client.stop();

    return 0;
}
```

## 配置项

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `base_url` | string | 配置中心地址 | `http://localhost:8080` |
| `app_name` | string | 应用名称 | （必填） |
| `env` | string | 环境标识 | `dev` |
| `version` | string | 配置版本 | `latest` |
| `token` | string | 应用 Token | 空 |
| `sync_interval` | int | 同步间隔（秒） | `60` |
| `ignore_global` | bool | 忽略全局配置 | `false` |
| `cache_dir` | string | 缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-cpp-client-demo
mkdir build && cd build
cmake ..
make
./demo
```

## 环境变量

Demo 支持以下环境变量：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `KENAITO_BASE_URL` | 配置中心地址 | `http://localhost:8080` |
| `KENAITO_APP_NAME` | 应用名称 | `kenaito-springboot-demo` |
| `KENAITO_ENV` | 环境标识 | `daily` |
| `KENAITO_VERSION` | 配置版本 | `latest` |
| `KENAITO_TOKEN` | 应用 Token | 空 |
| `KENAITO_SYNC_INTERVAL` | 同步间隔（秒） | `10` |
| `KENAITO_CACHE_DIR` | 缓存目录 | `.` |
