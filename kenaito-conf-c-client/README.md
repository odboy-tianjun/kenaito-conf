# Kenaito Config Center C Client

C 语言的配置中心客户端 SDK，使用 libcurl。

## 依赖

- libcurl

### macOS

```bash
brew install curl
```

### Ubuntu/Debian

```bash
apt-get install libcurl4-openssl-dev
```

## 编译

```bash
cd kenaito-conf-c-client-demo
gcc -o demo ../kenaito-conf-c-client/main.c -lcurl
./demo
```

## 功能

- HTTP POST 拉取配置
- 配置缓存

## API

```c
ConfigClient client;
config_client_init(&client, &config);
const char *value = config_client_get(&client, "database.host");
config_client_stop(&client);
```

## 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `baseUrl` | 配置中心地址 | 必填 |
| `appName` | 应用名称 | 必填 |
| `env` | 环境 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 认证令牌 | 空 |
| `syncInterval` | 同步间隔（秒） | `60` |
| `ignoreGlobal` | 是否忽略全局配置 | `0` |
| `cacheDir` | 本地缓存目录 | `.` |
