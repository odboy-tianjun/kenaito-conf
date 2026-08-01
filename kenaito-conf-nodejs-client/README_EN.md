# Kenaito Config Center Node.js Client

Node.js client SDK for the Kenaito Config Center.

## Installation

### Install to Local Repository (Recommended)

Register the package as a global symlink, then reference it from any project via `npm link kenaito-conf`:

```bash
cd kenaito-conf-nodejs-client
npm link
```

In your project:

```bash
npm link kenaito-conf
```

### Direct Reference

```bash
# In your project's package.json
"dependencies": {
  "kenaito-conf": "file:../kenaito-conf-nodejs-client"
}
```

```bash
npm install
```

## Quick Start

```javascript
import { Config, ConfigClient } from 'kenaito-conf';

const config = new Config({
  baseUrl: 'http://localhost:8080',
  appName: 'my-app',
  env: 'daily',
  syncInterval: 30,
});

const client = new ConfigClient(config);

// Listen for config changes
client.onChange((changed) => {
  for (const [k, v] of Object.entries(changed)) {
    console.log(`config changed: ${k} = ${v}`);
  }
});

// Start client (fetch configs + enable periodic sync)
await client.start();

// Get config
const value = client.get('database.host');
const allConfigs = client.getAll();

// Stop
client.stop();
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `baseUrl` | Config center URL | Required |
| `appName` | Application name | Required |
| `env` | Environment (dev/daily/prod) | `dev` |
| `version` | Config version | `latest` |
| `token` | Auth token | Empty |
| `syncInterval` | Sync interval (seconds) | `60` |
| `ignoreGlobal` | Whether to ignore global configs | `false` |
| `cacheDir` | Local cache directory | `.` |

## Demo

```bash
cd kenaito-conf-nodejs-client-demo
npm install
node app.js
```

The demo listens on port `8083` by default. Visit `http://localhost:8083/` to view configs.

Environment variable overrides:

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `SERVER_PORT` | Server port | `8083` |
| `KENAITO_BASE_URL` | Config center URL | `http://localhost:8080` |
| `KENAITO_APP_NAME` | Application name | `kenaito-nodejs-demo` |
| `KENAITO_PROFILE` | Environment | `daily` |
| `KENAITO_VERSION` | Config version | `latest` |
| `KENAITO_TOKEN` | Auth token | Empty |
| `KENAITO_SYNC_INTERVAL` | Sync interval (seconds) | `30` |
| `KENAITO_CACHE_DIR` | Cache directory | `.` |
