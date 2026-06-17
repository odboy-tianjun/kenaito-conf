import { Config } from './config.js';
import { Cache } from './cache.js';
import { Listener } from './listener.js';

export class ConfigClient {
  constructor(config) {
    this.config = config;
    this.cache = new Cache(config.cacheDir);
    this.listener = new Listener();
    this.configs = {};
    this.lastTime = 0;
    this.started = false;
    this.timer = null;
  }

  onChange(handler) {
    this.listener.onChange(handler);
  }

  get(key) {
    return this.configs[key] ?? null;
  }

  getAll() {
    return { ...this.configs };
  }

  async start() {
    if (this.started) {
      return;
    }
    this.started = true;

    try {
      await this.fetchAll();
    } catch (err) {
      console.warn('[ConfigClient] fetch all configs error:', err.message, ', trying cache');
      const cached = this.cache.read();
      if (cached) {
        this.applyConfigs(cached);
        console.info('[ConfigClient] loaded', Object.keys(this.configs).length, 'configs from cache');
      } else {
        throw new Error(`fetch configs failed and no cache: ${err.message}`);
      }
    }

    this.timer = setInterval(() => {
      this.fetchChanged().catch((err) => {
        console.error('[ConfigClient] sync error:', err);
      });
    }, this.config.syncInterval * 1000);
  }

  stop() {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  async fetchAll() {
    const url = `${this.config.baseUrl}/api/fetchAllConfigs`;
    const payload = {
      appName: this.config.appName,
      env: this.config.env,
      version: this.config.version,
      ignoreGlobal: this.config.ignoreGlobal,
      token: this.config.token,
    };
    const resp = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
      signal: AbortSignal.timeout(10000),
    });
    if (!resp.ok) {
      throw new Error(`HTTP ${resp.status}`);
    }
    const data = await resp.json();
    if (data.code != null && data.code !== 0) {
      throw new Error(`server error: ${data.msg || 'unknown'}`);
    }

    const ts = data['latest.fetch.timestamp'];
    if (ts) {
      this.lastTime = parseInt(ts, 10);
    }

    this.cache.write(data);
    this.applyConfigs(data);
    console.info('[ConfigClient] fetched', Object.keys(this.configs).length, 'configs');
  }

  async fetchChanged() {
    const url = `${this.config.baseUrl}/api/fetchChangedConfigs`;
    const payload = {
      appName: this.config.appName,
      env: this.config.env,
      version: this.config.version,
      lastTime: this.lastTime,
      ignoreGlobal: this.config.ignoreGlobal,
      token: this.config.token,
    };
    const resp = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(this.config.token ? { Authorization: `Bearer ${this.config.token}` } : {}),
      },
      body: JSON.stringify(payload),
      signal: AbortSignal.timeout(10000),
    });
    if (!resp.ok) {
      throw new Error(`HTTP ${resp.status}`);
    }
    const result = await resp.json();
    if (result.code !== 0) {
      throw new Error(`server error code: ${result.code}`);
    }

    const changedData = result.data || {};
    const ts = changedData['latest.fetch.timestamp'];
    if (ts) {
      this.lastTime = parseInt(ts, 10);
    }
    delete changedData['latest.fetch.timestamp'];

    // 处理删除的配置
    const deletedKeysStr = changedData['deletedKeys'];
    if (deletedKeysStr) {
      const deletedKeys = deletedKeysStr.split(',');
      for (const key of deletedKeys) {
        delete this.configs[key];
      }
      console.info('[ConfigClient]', deletedKeys.length, 'configs deleted');
      delete changedData['deletedKeys'];
    }

    const changed = {};
    for (const [k, v] of Object.entries(changedData)) {
      changed[k] = v;
    }

    Object.assign(this.configs, changed);
    this.listener.notify(changed);
    console.info('[ConfigClient]', Object.keys(changed).length, 'configs changed');
  }

  applyConfigs(data) {
    for (const [k, v] of Object.entries(data)) {
      if (k.startsWith('kenaito.') || k === 'latest.fetch.timestamp') {
        continue;
      }
      if (typeof v !== 'string') {
        continue;
      }
      this.configs[k] = v;
    }
  }
}

export { Config } from './config.js';
