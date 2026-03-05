use kenaito_conf::{Config, ConfigClient};
use std::collections::HashMap;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut config_values: HashMap<String, String> = HashMap::new();
    config_values.insert("database.host".to_string(), std::env::var("DATABASE_HOST").unwrap_or_else(|_| "192.168.1.1".to_string()));
    config_values.insert("database.port".to_string(), std::env::var("DATABASE_PORT").unwrap_or_else(|_| "8888".to_string()));
    config_values.insert("database.name".to_string(), std::env::var("DATABASE_NAME").unwrap_or_else(|_| "xxxx".to_string()));
    config_values.insert("database.user".to_string(), std::env::var("DATABASE_USER").unwrap_or_else(|_| "xxxx".to_string()));
    config_values.insert("database.password".to_string(), std::env::var("DATABASE_PASSWORD").unwrap_or_else(|_| "xxxxxxxx".to_string()));

    let config = Config {
        base_url: std::env::var("KENAITO_BASE_URL").unwrap_or_else(|_| "http://localhost:8080".to_string()),
        app_name: std::env::var("KENAITO_APP_NAME").unwrap_or_else(|_| "kenaito-springboot-demo".to_string()),
        env: std::env::var("KENAITO_PROFILE").unwrap_or_else(|_| "daily".to_string()),
        version: std::env::var("KENAITO_VERSION").unwrap_or_else(|_| "latest".to_string()),
        token: std::env::var("KENAITO_TOKEN").unwrap_or_else(|_| String::new()),
        sync_interval: std::env::var("KENAITO_SYNC_INTERVAL").unwrap_or_else(|_| "30".to_string()).parse().unwrap_or(30),
        cache_dir: std::env::var("KENAITO_CACHE_DIR").unwrap_or_else(|_| ".".to_string()),
        ..Default::default()
    };

    let client = ConfigClient::new(config);

    client.on_change(move |changed| {
        for (k, v) in changed.iter() {
            eprintln!("config changed: {} = {}", k, v);
        }
    });

    match client.start().await {
        Ok(_) => {
            let remote = client.get_all();
            eprintln!("loaded {} configs from remote", remote.len());
        }
        Err(e) => {
            eprintln!("failed to init config client: {}", e);
        }
    }

    println!("Current configs:");
    for (k, v) in &config_values {
        println!("  {} = {}", k, v);
    }

    println!("\nListening for config changes... (press Ctrl+C to stop)");
    tokio::signal::ctrl_c().await?;
    println!("Shutting down...");

    Ok(())
}
