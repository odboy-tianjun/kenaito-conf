import fs from 'node:fs';
import path from 'node:path';

const CACHE_FILE = 'config-cache.json';

export class Cache {
  constructor(cacheDir = '.') {
    this.cacheDir = cacheDir;
  }

  write(data) {
    const filePath = path.join(this.cacheDir, CACHE_FILE);
    fs.writeFileSync(filePath, JSON.stringify(data, null, 2), 'utf8');
  }

  read() {
    const filePath = path.join(this.cacheDir, CACHE_FILE);
    try {
      const content = fs.readFileSync(filePath, 'utf8');
      return JSON.parse(content);
    } catch {
      return null;
    }
  }
}
