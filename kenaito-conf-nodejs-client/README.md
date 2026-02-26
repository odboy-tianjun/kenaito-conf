# Kenaito Config Center Node.js Client

## 安装

### 安装到本地仓库（推荐）

将包注册为全局符号链接，之后任意项目都可以通过 `npm link kenaito-conf` 引用：

```bash
cd kenaito-conf-nodejs-client
npm link
```

在你的项目中使用：

```bash
npm link kenaito-conf
```

### 直接引用

```bash
# 在你的项目 package.json 中
"dependencies": {
  "kenaito-conf": "file:../kenaito-conf-nodejs-client"
}
```

```bash
npm install
```

## 快速开始

```javascript
import { Config, ConfigClient } from 'kenaito-conf';

const config = new Config({
  baseUrl: 'http://localhost:8080',
  appName: 'my-app',
  env: 'daily',
  syncInterval: 30,
});

const client = new ConfigClient(config);

// 监听配置变更
client.onChange((changed) => {
  for (const [k, v] of Object.entries(changed)) {
    console.log(`config changed: ${k} = ${v}`);
  }
});

// 启动客户端（拉取配置 + 开启定时同步）
await client.start();

// 获取配置
const value = client.get('database.host');
const allConfigs = client.getAll();

// 停止
client.stop();
```

## 配置项

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `baseUrl` | 配置中心地址 | 必填 |
| `appName` | 应用名称 | 必填 |
| `env` | 环境（dev/daily/prod） | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 认证令牌 | 空 |
| `syncInterval` | 同步间隔（秒） | `60` |
| `ignoreGlobal` | 是否忽略全局配置 | `false` |
| `cacheDir` | 本地缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-nodejs-client-demo
npm install
node app.js
```

Demo 默认监听 `8083` 端口，访问 `http://localhost:8083/` 查看配置。

环境变量覆盖：

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `SERVER_PORT` | 服务端口 | `8083` |
| `KENAITO_BASE_URL` | 配置中心地址 | `http://localhost:8080` |
| `KENAITO_APP_NAME` | 应用名称 | `kenaito-nodejs-demo` |
| `KENAITO_PROFILE` | 环境 | `daily` |
| `KENAITO_VERSION` | 配置版本 | `latest` |
| `KENAITO_TOKEN` | 认证令牌 | 空 |
| `KENAITO_SYNC_INTERVAL` | 同步间隔（秒） | `30` |
| `KENAITO_CACHE_DIR` | 缓存目录 | `.` |
