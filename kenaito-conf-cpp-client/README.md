# Kenaito Config Center C++ Client

C++ 语言的配置中心客户端 SDK，使用 libcurl + OpenSSL。

## 依赖

- libcurl
- OpenSSL
- C++17

### macOS

```bash
brew install curl openssl
```

### Ubuntu/Debian

```bash
apt-get install libcurl4-openssl-dev libssl-dev
```

## 编译

```bash
cd kenaito-conf-cpp-client-demo
g++ -std=c++17 -o demo ../kenaito-conf-cpp-client/main.cpp -lcurl -lssl -lcrypto -lpthread
./demo
```

## 功能

- AES-128-CBC 解密
- HTTP POST 拉取配置
- 定时同步
- 变更回调
- 线程安全

## 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `baseUrl` | 配置中心地址 | 必填 |
| `appName` | 应用名称 | 必填 |
| `env` | 环境 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 认证令牌 | 空 |
| `syncInterval` | 同步间隔（秒） | `60` |
| `ignoreGlobal` | 是否忽略全局配置 | `false` |
| `cacheDir` | 本地缓存目录 | `.` |
