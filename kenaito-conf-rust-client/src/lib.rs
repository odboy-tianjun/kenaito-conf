use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::io::{Read, Write};
use std::path::PathBuf;
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;

const CACHE_FILE: &str = "config-cache.json";

#[derive(Debug, Clone)]
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
            base_url: "http://localhost:8080".to_string(),
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

type ChangeHandler = Box<dyn Fn(&HashMap<String, String>) + Send + Sync + 'static>;

struct Cache {
    dir: String,
}

impl Cache {
    fn new(dir: &str) -> Self {
        Self { dir: dir.to_string() }
    }

    fn write(&self, data: &str) -> Result<(), std::io::Error> {
        let path = PathBuf::from(&self.dir).join(CACHE_FILE);
        let mut file = fs::File::create(path)?;
        file.write_all(data.as_bytes())?;
        Ok(())
    }

    fn read(&self) -> Option<String> {
        let path = PathBuf::from(&self.dir).join(CACHE_FILE);
        let mut file = fs::File::open(path).ok()?;
        let mut contents = String::new();
        file.read_to_string(&mut contents).ok()?;
        Some(contents)
    }
}

pub struct Client {
    config: Config,
    cache: Cache,
    handlers: Arc<Mutex<Vec<ChangeHandler>>>,
    configs: Arc<Mutex<HashMap<String, String>>>,
    last_time: Arc<Mutex<i64>>,
    stop_flag: Arc<Mutex<bool>>,
}

impl Client {
    pub fn new(config: Config) -> Self {
        let cache_dir = if config.cache_dir.is_empty() { ".".to_string() } else { config.cache_dir.clone() };
        let version = if config.version.is_empty() { "latest".to_string() } else { config.version.clone() };
        let sync_interval = if config.sync_interval == 0 { 60 } else { config.sync_interval };

        let mut cfg = config;
        cfg.cache_dir = cache_dir;
        cfg.version = version;
        cfg.sync_interval = sync_interval;

        Self {
            cache: Cache::new(&cfg.cache_dir),
            config: cfg,
            handlers: Arc::new(Mutex::new(Vec::new())),
            configs: Arc::new(Mutex::new(HashMap::new())),
            last_time: Arc::new(Mutex::new(0)),
            stop_flag: Arc::new(Mutex::new(false)),
        }
    }

    pub fn on_change<F>(&self, handler: F)
    where
        F: Fn(&HashMap<String, String>) + Send + Sync + 'static,
    {
        let mut handlers = self.handlers.lock().unwrap();
        handlers.push(Box::new(handler));
    }

    pub fn get(&self, key: &str) -> String {
        let configs = self.configs.lock().unwrap();
        configs.get(key).cloned().unwrap_or_default()
    }

    pub fn get_all(&self) -> HashMap<String, String> {
        self.configs.lock().unwrap().clone()
    }

    pub fn start(&self) -> Result<(), String> {
        match self.fetch_all() {
            Ok(_) => {}
            Err(e) => {
                eprintln!("[ConfigClient] fetch all configs error: {}, trying cache", e);
                match self.cache.read() {
                    Some(cached) => {
                        match serde_json::from_str::<serde_json::Value>(&cached) {
                            Ok(val) => {
                                let mut data = HashMap::new();
                                if let Some(obj) = val.as_object() {
                                    for (k, v) in obj {
                                        if let Some(s) = v.as_str() {
                                            data.insert(k.clone(), s.to_string());
                                        }
                                    }
                                }
                                self.apply_configs(&data);
                                eprintln!("[ConfigClient] loaded {} configs from cache", self.configs.lock().unwrap().len());
                            }
                            Err(_) => return Err("fetch configs failed and no cache".to_string()),
                        }
                    }
                    None => return Err("fetch configs failed and no cache".to_string()),
                }
            }
        }

        Ok(())
    }

    pub fn sync_loop(&self) {
        let interval = Duration::from_secs(self.config.sync_interval);
        loop {
            thread::sleep(interval);
            if *self.stop_flag.lock().unwrap() {
                break;
            }
            if let Err(e) = self.fetch_changed() {
                eprintln!("[ConfigClient] sync error: {}", e);
            }
        }
    }

    pub fn sync_once(&self) {
        if let Err(e) = self.fetch_changed() {
            eprintln!("[ConfigClient] sync error: {}", e);
        }
    }

    pub fn stop(&self) {
        *self.stop_flag.lock().unwrap() = true;
    }

    fn http_post(&self, url: &str, body: &str) -> Result<String, String> {
        let client = reqwest::blocking::Client::builder()
            .timeout(Duration::from_secs(10))
            .build()
            .map_err(|e| e.to_string())?;

        let resp = client
            .post(url)
            .header("Content-Type", "application/json")
            .body(body.to_string())
            .send()
            .map_err(|e| e.to_string())?;

        if !resp.status().is_success() {
            return Err(format!("HTTP {}", resp.status()));
        }

        resp.text().map_err(|e| e.to_string())
    }

    fn fetch_all(&self) -> Result<(), String> {
        let url = format!("{}/api/fetchAllConfigs", self.config.base_url);

        let payload = serde_json::json!({
            "appName": self.config.app_name,
            "env": self.config.env,
            "version": self.config.version,
            "ignoreGlobal": self.config.ignore_global,
            "token": self.config.token,
        });

        let resp = self.http_post(&url, &payload.to_string())?;
        let val: serde_json::Value = serde_json::from_str(&resp).map_err(|e| e.to_string())?;

        if let Some(code) = val.get("code").and_then(|c| c.as_i64()) {
            if code != 0 {
                let msg = val.get("msg").and_then(|m| m.as_str()).unwrap_or("unknown");
                return Err(format!("server error: {}", msg));
            }
        }

        if let Some(ts) = val.get("latest.fetch.timestamp").and_then(|t| t.as_str()) {
            if let Ok(t) = ts.parse::<i64>() {
                *self.last_time.lock().unwrap() = t;
            }
        }

        // cache raw response
        let _ = self.cache.write(&resp);

        // apply configs
        let mut data = HashMap::new();
        if let Some(obj) = val.as_object() {
            for (k, v) in obj {
                if let Some(s) = v.as_str() {
                    data.insert(k.clone(), s.to_string());
                }
            }
        }
        self.apply_configs(&data);

        eprintln!("[ConfigClient] fetched {} configs", self.configs.lock().unwrap().len());
        Ok(())
    }

    fn fetch_changed(&self) -> Result<(), String> {
        let url = format!("{}/api/fetchChangedConfigs", self.config.base_url);

        let payload = serde_json::json!({
            "appName": self.config.app_name,
            "env": self.config.env,
            "version": self.config.version,
            "lastTime": *self.last_time.lock().unwrap(),
            "ignoreGlobal": self.config.ignore_global,
            "token": self.config.token,
        });

        let resp = self.http_post(&url, &payload.to_string())?;
        let val: serde_json::Value = serde_json::from_str(&resp).map_err(|e| e.to_string())?;

        if let Some(code) = val.get("code").and_then(|c| c.as_i64()) {
            if code != 0 {
                return Err(format!("server error code: {}", code));
            }
        }

        let data = val.get("data").and_then(|d| d.as_object()).cloned().unwrap_or_default();

        if let Some(ts) = data.get("latest.fetch.timestamp").and_then(|t| t.as_str()) {
            if let Ok(t) = ts.parse::<i64>() {
                *self.last_time.lock().unwrap() = t;
            }
        }

        // handle deleted keys
        if let Some(deleted_str) = data.get("deletedKeys").and_then(|d| d.as_str()) {
            if !deleted_str.is_empty() {
                let deleted_keys: Vec<&str> = deleted_str.split(',').collect();
                let mut configs = self.configs.lock().unwrap();
                for key in &deleted_keys {
                    configs.remove(*key);
                }
                eprintln!("[ConfigClient] {} configs deleted: {}", deleted_keys.len(), deleted_keys.join(", "));
            }
        }

        // handle changed configs
        let mut changed = HashMap::new();
        for (k, v) in &data {
            if k == "latest.fetch.timestamp" || k == "deletedKeys" {
                continue;
            }
            if let Some(s) = v.as_str() {
                changed.insert(k.clone(), s.to_string());
            }
        }

        if !changed.is_empty() {
            {
                let mut configs = self.configs.lock().unwrap();
                for (k, v) in &changed {
                    configs.insert(k.clone(), v.clone());
                }
            }

            let handlers = self.handlers.lock().unwrap();
            for h in handlers.iter() {
                h(&changed);
            }

            let keys: Vec<&str> = changed.keys().map(|s| s.as_str()).collect();
            eprintln!("[ConfigClient] {} configs changed: {}", changed.len(), keys.join(", "));
        }

        Ok(())
    }

    fn apply_configs(&self, data: &HashMap<String, String>) {
        let mut configs = self.configs.lock().unwrap();
        for (k, v) in data {
            if k.starts_with("kenaito.") || k == "latest.fetch.timestamp" {
                continue;
            }
            configs.insert(k.clone(), v.clone());
        }
    }
}
