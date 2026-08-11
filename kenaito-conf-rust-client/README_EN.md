# Kenaito Config Center Rust Client

Rust client SDK for the Kenaito Config Center, suitable for Rust 2021 edition projects.

## Features

- Fetch application configs from the config center
- Periodic config sync
- Config change callback listener
- Local file cache (fallback when config center is unavailable)

## Dependencies

- Rust 2021 edition
- reqwest (HTTP client)
- serde / serde_json (JSON processing)

## Usage

Add to your `Cargo.toml`:

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

    // Listen for config changes
    client.on_change(|changed| {
        println!("Config changed:");
        for (k, v) in changed {
            println!("  {} = {}", k, v);
        }
    });

    // Start client
    if let Err(e) = client.start() {
        eprintln!("Failed to start: {}", e);
        return;
    }

    // Get config
    let db_host = client.get("database.host");
    println!("database.host = {}", db_host);

    // Get all configs
    let all = client.get_all();
    for (k, v) in &all {
        println!("{} = {}", k, v);
    }

    // Manual sync
    client.sync_once();

    // Stop
    client.stop();
}
```

## Configuration

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `base_url` | String | Config center URL | `http://localhost:8080` |
| `app_name` | String | Application name | (Required) |
| `env` | String | Environment | `dev` |
| `version` | String | Config version | `latest` |
| `token` | String | App token | Empty |
| `sync_interval` | u64 | Sync interval (seconds) | `60` |
| `ignore_global` | bool | Whether to ignore global configs | `false` |
| `cache_dir` | String | Cache directory | `.` |

## Demo

```bash
cd kenaito-conf-rust-client-demo
cargo run
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
