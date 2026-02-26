import express from 'express';
import { Config, ConfigClient } from 'kenaito-conf';

const configValues = {
  'database.host': process.env.DATABASE_HOST || '192.168.1.1',
  'database.port': process.env.DATABASE_PORT || '8888',
  'database.name': process.env.DATABASE_NAME || 'xxxx',
  'database.user': process.env.DATABASE_USER || 'xxxx',
  'database.password': process.env.DATABASE_PASSWORD || 'xxxxxxxx',
};

let client = null;

function onConfigChanged(changed) {
  for (const [k, v] of Object.entries(changed)) {
    if (k in configValues) {
      configValues[k] = v;
      console.log(`config updated: ${k} = ${v}`);
    }
  }
}

async function createClient() {
  const cfg = new Config({
    baseUrl: process.env.KENAITO_BASE_URL || 'http://localhost:8080',
    appName: process.env.KENAITO_APP_NAME || 'kenaito-springboot-demo',
    env: process.env.KENAITO_PROFILE || 'daily',
    version: process.env.KENAITO_VERSION || 'latest',
    token: process.env.KENAITO_TOKEN || '',
    syncInterval: parseInt(process.env.KENAITO_SYNC_INTERVAL || '30', 10),
    cacheDir: process.env.KENAITO_CACHE_DIR || '.',
  });
  const c = new ConfigClient(cfg);
  c.onChange(onConfigChanged);
  await c.start();
  return c;
}

const app = express();

app.get('/', (req, res) => {
  res.type('application/json').send(JSON.stringify(configValues, null, 2));
});

app.get('/health', (req, res) => {
  res.send('OK');
});

const port = parseInt(process.env.SERVER_PORT || '8083', 10);

async function main() {
  try {
    client = await createClient();
    const remote = client.getAll();
    for (const k of Object.keys(configValues)) {
      if (k in remote) {
        configValues[k] = remote[k];
      }
    }
    console.info(`loaded ${Object.keys(remote).length} configs from remote`);
  } catch (err) {
    console.error('failed to init config client, using local defaults:', err);
  }

  app.listen(port, '0.0.0.0', () => {
    console.info(`server started on port ${port}`);
  });
}

main();
