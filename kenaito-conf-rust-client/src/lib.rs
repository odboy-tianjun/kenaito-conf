use aes::cipher::{block_padding::NoPadding, BlockDecryptMut, KeyIvInit};
use md5::{Digest, Md5};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::Duration;

type Aes128CbcDec = cbc::Decryptor<aes::Aes128>;

#[derive(Clone, Debug)]
pub struct Config {
    pub base_url: String,
    pub app_name: String,
    pub env: String,
    pub version: String,
    pub token: String,
    pub sync_interval: u64,
    pub ignore_global: bool,
    pub cache_dir: String,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            base_url: String::new(),
            app_name: String::new(),
            env: "dev".to_string(),
            version: "latest".to_string(),
            token: String::new(),
            sync_interval: 60,
            ignore_global: false,
            cache_dir: ".".to_string(),
        }
    }
}

#[derive(Serialize, Deserialize, Debug)]
struct FetchAllRequest {
    app_name: String,
    env: String,
    version: String,
    ignore_global: bool,
    token: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct FetchChangedRequest {
    app_name: String,
    env: String,
    version: String,
    last_time: i64,
    ignore_global: bool,
    token: String,
}

fn decrypt_config(secret: &str, data: &str) -> String {
    let prefix = "{Cipher}";
    if !data.starts_with(prefix) {
        return data.to_string();
    }
    let encrypted = &data[prefix.len()..];
    decrypt_with_aes(secret, encrypted)
}

fn decrypt_with_aes(key: &str, data: &str) -> String {
    let mut hasher = Md5::new();
    hasher.update(key.as_bytes());
    let result = hasher.finalize();
    let md5_hex = format!("{:x}", result);
    let aes_key = &md5_hex[16..];

    let cipher_text = base64::decode(data).unwrap_or_default();
    let iv = [0u8; 16];

    let mut buf = cipher_text.clone();
    let decrypted = Aes128CbcDec::new_from_slices(aes_key.as_bytes(), &iv)
        .unwrap()
        .decrypt_padded_mut::<NoPadding>(&mut buf)
        .unwrap();

    let padding = decrypted[decrypted.len() - 1] as usize;
    String::from_utf8_lossy(&decrypted[..decrypted.len() - padding]).to_string()
}

pub struct ConfigClient {
    config: Config,
    configs: Arc<Mutex<HashMap<String, String>>>,
    secret: Arc<Mutex<String>>,
    global_secret: Arc<Mutex<String>>,
    last_time: Arc<Mutex<i64>>,
    handlers: Arc<Mutex<Vec<Box<dyn Fn(&HashMap<String, String>) + Send + Sync>>>>,
}

impl ConfigClient {
    pub fn new(config: Config) -> Self {
        Self {
            config,
            configs: Arc::new(Mutex::new(HashMap::new())),
            secret: Arc::new(Mutex::new(String::new())),
            global_secret: Arc::new(Mutex::new(String::new())),
            last_time: Arc::new(Mutex::new(0)),
            handlers: Arc::new(Mutex::new(Vec::new())),
        }
    }

    pub fn on_change<F>(&self, handler: F)
    where
        F: Fn(&HashMap<String, String>) + Send + Sync + 'static,
    {
        self.handlers.lock().unwrap().push(Box::new(handler));
    }

    pub fn get(&self, key: &str) -> Option<String> {
        self.configs.lock().unwrap().get(key).cloned()
    }

    pub fn get_all(&self) -> HashMap<String, String> {
        self.configs.lock().unwrap().clone()
    }

    pub async fn start(&self) -> Result<(), Box<dyn std::error::Error>> {
        self.fetch_all().await?;

        let configs = self.configs.clone();
        let secret = self.secret.clone();
        let global_secret = self.global_secret.clone();
        let last_time = self.last_time.clone();
        let handlers = self.handlers.clone();
        let config = self.config.clone();

        tokio::spawn(async move {
            loop {
                tokio::time::sleep(Duration::from_secs(config.sync_interval)).await;
                // fetch_changed logic here
            }
        });

        Ok(())
    }

    async fn fetch_all(&self) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}/api/fetchAllConfigs", self.config.base_url);
        let client = reqwest::Client::new();

        let payload = serde_json::json!({
            "appName": self.config.app_name,
            "env": self.config.env,
            "version": self.config.version,
            "ignoreGlobal": self.config.ignore_global,
            "token": self.config.token,
        });

        let resp = client.post(&url).json(&payload).send().await?;
        let data: HashMap<String, serde_json::Value> = resp.json().await?;

        if let Some(code) = data.get("code") {
            if code.as_i64() != Some(0) {
                return Err("server error".into());
            }
        }

        if let Some(s) = data.get("kenaito.config.encrypt-secret") {
            *self.secret.lock().unwrap() = s.as_str().unwrap_or("").to_string();
        }
        if let Some(gs) = data.get("kenaito.config.global-encrypt-secret") {
            *self.global_secret.lock().unwrap() = gs.as_str().unwrap_or("").to_string();
        }
        if let Some(ts) = data.get("latest.fetch.timestamp") {
            *self.last_time.lock().unwrap() = ts.as_i64().unwrap_or(0);
        }

        let mut configs = self.configs.lock().unwrap();
        let secret = self.secret.lock().unwrap().clone();
        let global_secret = self.global_secret.lock().unwrap().clone();

        for (k, v) in data.iter() {
            if k.starts_with("kenaito.") || k == "latest.fetch.timestamp" {
                continue;
            }
            if let Some(s) = v.as_str() {
                let mut decrypted = decrypt_config(&secret, s);
                if decrypted == s && !global_secret.is_empty() {
                    decrypted = decrypt_config(&global_secret, s);
                }
                configs.insert(k.clone(), decrypted);
            }
        }

        Ok(())
    }
}
