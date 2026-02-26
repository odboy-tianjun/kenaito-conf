export class Config {
  constructor(options = {}) {
    this.baseUrl = options.baseUrl || '';
    this.appName = options.appName || '';
    this.env = options.env || 'dev';
    this.version = options.version || 'latest';
    this.token = options.token || '';
    this.syncInterval = options.syncInterval || 60;
    this.ignoreGlobal = options.ignoreGlobal || false;
    this.cacheDir = options.cacheDir || '.';
  }
}
