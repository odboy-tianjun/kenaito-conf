# Kenaito Config Center C++ Client

C++ client SDK for the Kenaito Config Center, suitable for C++11 and above.

## Features

- Fetch application configs from the config center
- Periodic config sync
- Config change callback listener
- Local file cache (fallback when config center is unavailable)

## Dependencies

- C++11 or higher
- libcurl
- nlohmann/json (included in third_party directory)

## Build

```bash
cd kenaito-conf-cpp-client
mkdir build && cd build
cmake ..
make
```

## Usage

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

    // Listen for config changes
    client.on_change([](const std::map<std::string, std::string>& changed) {
        std::cout << "Config changed:" << std::endl;
        for (const auto& [k, v] : changed) {
            std::cout << "  " << k << " = " << v << std::endl;
        }
    });

    // Start client
    if (!client.start()) {
        std::cerr << "Failed to start config client" << std::endl;
        return 1;
    }

    // Get config
    std::string db_host = client.get("database.host");
    std::cout << "database.host = " << db_host << std::endl;

    // Get all configs
    auto all = client.get_all();
    for (const auto& [k, v] : all) {
        std::cout << k << " = " << v << std::endl;
    }

    // Manual sync
    client.sync_once();

    // Stop
    client.stop();

    return 0;
}
```

## Configuration

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `base_url` | string | Config center URL | `http://localhost:8080` |
| `app_name` | string | Application name | (Required) |
| `env` | string | Environment | `dev` |
| `version` | string | Config version | `latest` |
| `token` | string | App token | Empty |
| `sync_interval` | int | Sync interval (seconds) | `60` |
| `ignore_global` | bool | Whether to ignore global configs | `false` |
| `cache_dir` | string | Cache directory | `.` |

## Demo

```bash
cd kenaito-conf-cpp-client-demo
mkdir build && cd build
cmake ..
make
./demo
```

## Environment Variables

The demo supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `KENAITO_BASE_URL` | Config center URL | `http://localhost:8080` |
| `KENAITO_APP_NAME` | Application name | `kenaito-springboot-demo` |
| `KENAITO_ENV` | Environment | `daily` |
| `KENAITO_VERSION` | Config version | `latest` |
| `KENAITO_TOKEN` | App token | Empty |
| `KENAITO_SYNC_INTERVAL` | Sync interval (seconds) | `10` |
| `KENAITO_CACHE_DIR` | Cache directory | `.` |
