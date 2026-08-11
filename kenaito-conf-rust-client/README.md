# Kenaito Config Center Rust Client

Rust 客户端 SDK，适用于 Rust 2021 edition 项目。

## 功能特性

- 从配置中心拉取应用配置
- 定时同步配置变更
- 支持配置变更回调监听
- 本地文件缓存（配置中心不可用时降级）

## 依赖

- Rust 2021 edition
- reqwest (HTTP 客户端)
- serde / serde_json (JSON 处理)

## 使用

在你的 `Cargo.toml` 中添加：

```toml
[dependencies]
kenaito-conf-client = { path = "../kenaito-conf-rust-client" }
```

```rust
use kenaito_conf_client::{Config, Client};

fn main() {
    let config = Config {
        base_url: "http://localhost:8080".to_string(),
        app_name: "my-app".to_string(),
        env: "daily".to_string(),
        sync_interval: 10,
        cache_dir: ".".to_string(),
        ..Default::default()
    };

    let client = Client::new(config);

    // 监听配置变更
    client.on_change(|changed| {
        println!("Config changed:");
        for (k, v) in changed {
            println!("  {} = {}", k, v);
        }
    });

    // 启动客户端
    if let Err(e) = client.start() {
        eprintln!("Failed to start: {}", e);
        return;
    }

    // 获取配置
    let db_host = client.get("database.host");
    println!("database.host = {}", db_host);

    // 获取所有配置
    let all = client.get_all();
    for (k, v) in &all {
        println!("{} = {}", k, v);
    }

    // 手动同步
    client.sync_once();

    // 停止
    client.stop();
}
```

## 配置项

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `base_url` | String | 配置中心地址 | `http://localhost:8080` |
| `app_name` | String | 应用名称 | （必填） |
| `env` | String | 环境标识 | `dev` |
| `version` | String | 配置版本 | `latest` |
| `token` | String | 应用 Token | 空 |
| `sync_interval` | u64 | 同步间隔（秒） | `60` |
| `ignore_global` | bool | 忽略全局配置 | `false` |
| `cache_dir` | String | 缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-rust-client-demo
cargo run
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
