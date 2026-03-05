# Kenaito Config Center Rust Client

Rust 语言的配置中心客户端 SDK。

## 依赖

- Rust 1.70+
- reqwest (HTTP 客户端)
- tokio (异步运行时)
- aes + cbc (AES 解密)

## 安装

在 `Cargo.toml` 中添加：

```toml
[dependencies]
kenaito-conf = { path = "../kenaito-conf-rust-client" }
```

## 使用

```rust
use kenaito_conf::{Config, ConfigClient};

#[tokio::main]
async fn main() {
    let config = Config {
        base_url: "http://localhost:8080".to_string(),
        app_name: "kenaito-springboot-demo".to_string(),
        env: "daily".to_string(),
        version: "latest".to_string(),
        sync_interval: 30,
        ..Default::default()
    };

    let client = ConfigClient::new(config);

    client.on_change(|changed| {
        for (k, v) in changed.iter() {
            println!("config changed: {} = {}", k, v);
        }
    });

    client.start().await.unwrap();

    let value = client.get("database.host");
    let all = client.get_all();
}
```

## 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `base_url` | 配置中心地址 | 必填 |
| `app_name` | 应用名称 | 必填 |
| `env` | 环境 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 认证令牌 | 空 |
| `sync_interval` | 同步间隔（秒） | `60` |
| `ignore_global` | 是否忽略全局配置 | `false` |
| `cache_dir` | 本地缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-rust-client-demo
cargo run
```
