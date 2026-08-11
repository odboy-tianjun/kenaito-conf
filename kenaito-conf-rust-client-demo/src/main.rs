use kenaito_conf_client::{Client, Config};
use std::collections::HashMap;
use std::env;
use std::sync::Arc;
use std::thread;
use std::time::Duration;

fn get_env(key: &str, default: &str) -> String {
    env::var(key).unwrap_or_else(|_| default.to_string())
}

fn get_env_int(key: &str, default: u64) -> u64 {
    env::var(key)
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(default)
}

fn main() {
    let config_values: HashMap<String, String> = [
        ("database.host".to_string(), "192.168.1.1".to_string()),
        ("database.port".to_string(), "8888".to_string()),
        ("database.name".to_string(), "xxxx".to_string()),
        ("database.user".to_string(), "xxxx".to_string()),
        ("database.password".to_string(), "xxxxxxxx".to_string()),
    ]
    .iter()
    .cloned()
    .collect();

    let config = Config {
        base_url: get_env("KENAITO_BASE_URL", "http://localhost:8080"),
        app_name: get_env("KENAITO_APP_NAME", "kenaito-springboot-demo"),
        env: get_env("KENAITO_ENV", "daily"),
        version: get_env("KENAITO_VERSION", "latest"),
        token: get_env("KENAITO_TOKEN", ""),
        sync_interval: get_env_int("KENAITO_SYNC_INTERVAL", 10),
        cache_dir: get_env("KENAITO_CACHE_DIR", "."),
        ..Default::default()
    };

    let client = Arc::new(Client::new(config.clone()));

    client.on_change(|changed| {
        println!("Config changed:");
        for (k, v) in changed {
            println!("  {} = {}", k, v);
        }
    });

    if let Err(e) = client.start() {
        eprintln!("Failed to init config client, using local defaults: {}", e);
    }

    let mut config_values = config_values;
    let remote = client.get_all();
    for (k, v) in &remote {
        if config_values.contains_key(k) {
            config_values.insert(k.clone(), v.clone());
        }
    }
    println!("loaded {} configs from remote", remote.len());

    println!("\nCurrent configs:");
    let mut keys: Vec<&String> = config_values.keys().collect();
    keys.sort();
    for k in keys {
        println!("  {} = {}", k, config_values[k]);
    }

    println!("\nListening for config changes... (press Ctrl+C to stop)");

    let client_clone = client.clone();
    ctrlc::set_handler(move || {
        println!("\nShutting down...");
        client_clone.stop();
        std::process::exit(0);
    })
    .expect("Error setting Ctrl-C handler");

    loop {
        client.sync_once();
        thread::sleep(Duration::from_secs(config.sync_interval));
    }
}
