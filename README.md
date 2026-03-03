# Kenaito Config Center

基于 Go 重构的集中配置管理平台，支持多应用、多环境、多业务组的配置管理与分发。

## 技术栈

- **后端**: Go + Gin + xorm + MySQL
- **前端**: Vue 2 + Vite 5 + Element UI 2
- **客户端**: Go / Java（Spring Boot 2.x & 3.x）/ Python / Node.js

## 项目结构

```
kenaito-conf/
├── server/                              # Go 管理端（Gin + xorm）
├── web/                                 # Vue 2 前端
├── kenaito-conf-go-client/              # Go 客户端 SDK
├── kenaito-conf-go-client-demo/         # Go 客户端示例
├── kenaito-conf-springboot2-client/     # Java Spring Boot 2.x 客户端 SDK（JDK 8+）
├── kenaito-conf-springboot2-client-demo/# Spring Boot 2.x 客户端示例（端口 8081）
├── kenaito-conf-springboot3-client/     # Java Spring Boot 3 客户端 SDK（JDK 11+，预留升级路径）
├── kenaito-conf-springboot3-client-demo/# Spring Boot 3 客户端示例（端口 8082，需 JDK 17+）
├── kenaito-conf-python-client/          # Python 客户端 SDK（Python 3.10+）
├── kenaito-conf-python-client-demo/     # Python 客户端示例（端口 8082）
├── kenaito-conf-nodejs-client/          # Node.js 客户端 SDK（Node.js 18+）
└── kenaito-conf-nodejs-client-demo/     # Node.js 客户端示例（端口 8083）
```

## 快速开始

### 1. 启动后端

```bash
cd server
```

编辑 `application.properties` 设置 `active` 环境，编辑 `application-{active}.properties` 配置数据库连接等信息：

```properties
# application.properties
active=dev

# application-dev.properties
server.port=8080
database.host=127.0.0.1
database.port=3306
database.name=kenaito_conf
database.user=root
database.password=your_password
jwt.secret=your-jwt-secret
jwt.expireHours=24
crypto.secretKey=your-crypto-key
```

启动服务（首次启动自动建表并初始化默认数据）：

```bash
go run cmd/main.go -config application.properties
# 或编译后运行
go build -o bin/server cmd/main.go
./bin/server -config application.properties
```

后端默认监听 `:8080`。

### 2. 启动前端

```bash
cd web
npm install
npm run dev
```

前端默认监听 `:3000`，自动代理 API 到后端。

### 3. 访问

打开 http://localhost:3000

默认管理员账号：`admin` / 密码：`admin123`

## 客户端接入

### Go 客户端

```go
import "kenaito-conf/kenaito-conf-go-client"

cfg := &client.Config{
    BaseURL:      "http://localhost:8080",
    AppName:      "myapp",
    Env:          "dev",
    Version:      "latest",
    Token:        "app-token",
    SyncInterval: 30,
    CacheDir:     ".",
}

c := client.New(cfg)

// 注册变更回调
c.OnChange(func(changed map[string]string) {
    for k, v := range changed {
        fmt.Printf("%s = %s\n", k, v)
    }
})

// 启动（拉取全量配置 + 开启定时同步）
c.Start()
defer c.Stop()

// 读取配置
dbHost := c.Get("db.host")
```

### Java Spring Boot 2.x 客户端

适用于 JDK 8+、Spring Boot 2.x 项目。

**1. 引入依赖**

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot2-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

**2. 添加配置**（`application.properties`）

```properties
kenaito.config.baseUrl=http://localhost:8080
kenaito.config.appName=myapp
kenaito.config.profile=dev
kenaito.config.version=latest
kenaito.config.token=app-token
kenaito.config.sync-interval-seconds=30
```

**3. 使用配置**

```java
@Value("${your.config.key}")
private String configValue;
```

远程配置会自动注入到 Spring Environment 中，通过 `@Value` 即可读取。

### Java Spring Boot 3 客户端

适用于 JDK 11+ 项目（当前编译目标 Spring Boot 2.7.18，预留 Spring Boot 3 升级路径）。

**1. 引入依赖**

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot3-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

**2. 配置与使用方式同 Spring Boot 2.x 客户端。**

> **注意：** 升级到 Spring Boot 3 需先将编译目标改为 JDK 17，并更新 `spring-boot.version` 为 3.x、Spring Framework 为 6.x。

### Python 客户端

适用于 Python 3.10+ 项目。

**1. 安装**

```bash
cd kenaito-conf-python-client
pip3 install -e .
```

**2. 使用**

```python
from kenaito_conf import Config, ConfigClient

config = Config(
    base_url="http://localhost:8080",
    app_name="my-app",
    env="daily",
    sync_interval=30,
)

client = ConfigClient(config)

# 监听配置变更
def on_change(changed):
    for k, v in changed.items():
        print(f"config changed: {k} = {v}")

client.on_change(on_change)

# 启动客户端（拉取配置 + 开启定时同步）
client.start()

# 获取配置
value = client.get("database.host")
all_configs = client.get_all()

# 停止
client.stop()
```

### Node.js 客户端

适用于 Node.js 18+ 项目。

**1. 安装**

```bash
cd kenaito-conf-nodejs-client
npm link
```

然后在你的项目中链接：

```bash
npm link kenaito-conf
```

**2. 使用**

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

## API 概览

### 管理端 API（需 JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/user/login` | 登录 |
| GET | `/user/loginUserInfo` | 当前用户信息 |
| POST | `/user/logout` | 登出 |
| GET | `/admin/menus` | 菜单 |
| POST | `/admin/config/list` | 配置列表 |
| POST | `/admin/config/add` | 新增配置 |
| POST | `/admin/config/update` | 更新配置 |
| GET | `/admin/config/:id` | 配置详情 |
| POST | `/admin/config/delete/:id` | 删除配置 |
| GET | `/admin/config/configHistories/:id` | 配置历史 |
| POST | `/admin/config/rollback/:id` | 回滚配置 |
| POST | `/admin/app/list` | 应用列表 |
| POST | `/admin/app/add` | 新增应用 |
| POST | `/admin/app/update` | 更新应用 |
| POST | `/admin/app/delete/:id` | 删除应用 |
| POST | `/admin/app/switch/:id` | 切换应用状态 |
| GET | `/admin/app/tree` | 应用树 |
| GET | `/admin/app/options` | 应用选项 |
| GET | `/admin/user/list` | 用户列表 |
| POST | `/admin/user/add` | 新增用户 |
| POST | `/admin/user/update` | 更新用户 |
| POST | `/admin/user/delete/:id` | 删除用户 |
| POST | `/admin/user/resetPassword/:id` | 重置密码 |
| POST | `/admin/user/switch/:id` | 切换用户状态 |
| GET | `/admin/group/list` | 业务组列表 |
| POST | `/admin/group/add` | 新增业务组 |
| POST | `/admin/group/update` | 更新业务组 |
| GET | `/admin/group/options` | 业务组选项 |
| GET | `/admin/profile/list` | 环境列表 |
| POST | `/admin/profile/add` | 新增环境 |
| POST | `/admin/profile/update` | 更新环境 |
| POST | `/admin/profile/delete/:id` | 删除环境 |
| POST | `/admin/profile/switch/:id` | 切换环境状态 |
| GET | `/admin/profile/options` | 环境选项 |

### 客户端 API（无需 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/fetchAllConfigs` | 全量拉取配置 |
| POST | `/api/fetchChangedConfigs` | 增量拉取变更 |
| GET | `/api/ping` | 健康检查 |

## 加密说明

配置值支持 AES 加密存储（`{Cipher}xxx` 格式），客户端自动解密：

- **算法**: AES-128/CBC/ZeroIV + PKCS7 填充
- **密钥派生**: `MD5(secretKey + appId_hex + env)` 取后 16 字节
- Go 客户端和 Java 客户端使用相同的加密方案，确保跨语言兼容

## 与原 Java 版的差异

- **去掉了 ZooKeeper**：配置变更通知改为客户端定时轮询
- **认证方式**：管理端用 JWT，客户端 API 通过 Token 鉴权
- **前端**：从 layui 静态页面改为 Vue 2 + Element UI 前后端分离
- **加密算法**：保持 AES 兼容，Java 客户端可继续使用
- **自动建表**：服务端启动时通过 xorm Sync2 自动创建表结构，无需手动执行 SQL
