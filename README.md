# Kenaito Config Center

基于 Go 重构的集中配置管理平台，支持多应用、多环境、多业务组的配置管理与分发。

## 技术栈

- **后端**: Go + Gin + xorm + MySQL
- **前端**: Vue 2 + Vite 5 + Element UI 2
- **客户端**: Go / Java（Spring Boot 2.x & 3.x）/ Python / Node.js / PHP / .NET / C / C++ / Lua / Kotlin / Rust / Ruby

## 项目结构

```
kenaito-conf/
├── server/                              # Go 管理端（Gin + xorm）
├── web/                                 # Vue 2 前端
├── kenaito-conf-go-client/              # （测试通过）Go 客户端 SDK
├── kenaito-conf-go-client-demo/         # （测试通过）Go 客户端示例
├── kenaito-conf-springboot2-client/     # （测试通过）Java Spring Boot 2.x 客户端 SDK（JDK 8+）
├── kenaito-conf-springboot2-client-demo/# （测试通过）Spring Boot 2.x 客户端示例（端口 8081）
├── kenaito-conf-python-client/          # （测试通过）Python 客户端 SDK（Python 3.10+）
├── kenaito-conf-python-client-demo/     # （测试通过）Python 客户端示例（端口 8082）
├── kenaito-conf-nodejs-client/          # （测试通过）Node.js 客户端 SDK（Node.js 18+）
├── kenaito-conf-nodejs-client-demo/     # （测试通过）Node.js 客户端示例（端口 8083）
├── kenaito-conf-php-client/             # （待测试）PHP 客户端 SDK（PHP 8.0+）
├── kenaito-conf-php-client-demo/        # （待测试）PHP 客户端示例
├── kenaito-conf-dotnet-client/          # （待测试）.NET 客户端 SDK（.NET 6.0+）
├── kenaito-conf-dotnet-client-demo/     # （待测试）.NET 客户端示例
├── kenaito-conf-c-client/               # （待测试）C 客户端 SDK（libcurl + OpenSSL）
├── kenaito-conf-c-client-demo/          # （待测试）C 客户端示例
├── kenaito-conf-cpp-client/             # （待测试）C++ 客户端 SDK（libcurl + OpenSSL）
├── kenaito-conf-cpp-client-demo/        # （待测试）C++ 客户端示例
├── kenaito-conf-lua-client/             # （待测试）Lua 客户端 SDK（Lua 5.3+）
├── kenaito-conf-lua-client-demo/        # （待测试）Lua 客户端示例
├── kenaito-conf-kotlin-client/          # （待测试）Kotlin 客户端 SDK（Kotlin 1.9+）
├── kenaito-conf-kotlin-client-demo/     # （待测试）Kotlin 客户端示例
├── kenaito-conf-rust-client/            # （待测试）Rust 客户端 SDK（Rust 1.70+）
├── kenaito-conf-rust-client-demo/       # （待测试）Rust 客户端示例
├── kenaito-conf-ruby-client/            # （待测试）Ruby 客户端 SDK（Ruby 2.7+）
└── kenaito-conf-ruby-client-demo/       # （待测试）Ruby 客户端示例
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
c.OnChange(func (changed map[string]string) {
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

适用于 JDK 8+、Spring Boot 2.x 项目，支持配置自动刷新和 `@Value` 动态更新。

**1. 引入依赖**

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot2-client</artifactId>
    <version>1.0.3</version>
</dependency>
```

**2. 添加配置**（`application.properties`）

```properties
# 配置中心地址
kenaito.configcenter.base-url=http://localhost:8080
# 应用名称（必填）
kenaito.configcenter.app-name=myapp
# 环境（dev/daily/prod）
kenaito.configcenter.env=daily
# 配置版本
kenaito.configcenter.version=latest
# 认证令牌（可选）
kenaito.configcenter.token=app-token
# 同步间隔（秒）
kenaito.configcenter.sync-interval=30
```

**3. 使用配置**

```java
@Value("${database.host:localhost}")
private String databaseHost;

@Value("${database.port:3306}")
private Integer databasePort;
```

配置变更时会自动刷新 `@Value` 字段，无需重启应用。详见 [Spring Boot 2 客户端文档](kenaito-conf-springboot2-client/README.md)。

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
import {Config, ConfigClient} from 'kenaito-conf';

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

### PHP 客户端

适用于 PHP 8.0+ 项目。

**1. 安装**

```bash
cd kenaito-conf-php-client
composer install
```

**2. 使用**

```php
use KenaitoConf\Config;
use KenaitoConf\ConfigClient;

$config = new Config([
    'baseUrl' => 'http://localhost:8080',
    'appName' => 'kenaito-springboot-demo',
    'env' => 'daily',
    'syncInterval' => 30,
]);

$client = new ConfigClient($config);

$client->onChange(function ($changed) {
    foreach ($changed as $k => $v) {
        echo "config changed: $k = $v\n";
    }
});

$client->start();
$value = $client->get('database.host');
$all = $client->getAll();
```

### .NET 客户端

适用于 .NET 6.0+ 项目。

**1. 引入依赖**

```xml

<ProjectReference Include="..\kenaito-conf-dotnet-client\KenaitoConf.csproj"/>
```

**2. 使用**

```csharp
using KenaitoConf;

var config = new Config
{
    BaseUrl = "http://localhost:8080",
    AppName = "kenaito-springboot-demo",
    Env = "daily",
    SyncInterval = 30,
};

var client = new ConfigClient(config);
client.OnChange(changed => {
    foreach (var kv in changed)
        Console.WriteLine($"config changed: {kv.Key} = {kv.Value}");
});

await client.StartAsync();
var value = client.Get("database.host");
```

### C 客户端

适用于嵌入式或高性能场景，依赖 libcurl + OpenSSL。

**编译**

```bash
gcc -o demo kenaito-conf-c-client/main.c -lcurl -lssl -lcrypto
```

### C++ 客户端

依赖 libcurl + OpenSSL，C++17。

**编译**

```bash
g++ -std=c++17 -o demo kenaito-conf-cpp-client/main.cpp -lcurl -lssl -lcrypto -lpthread
```

### Lua 客户端

适用于 Lua 5.3+ 项目（如 OpenResty/Nginx 网关）。

**1. 安装**

```bash
cd kenaito-conf-lua-client
luarocks install kenaito-conf-client-1.0.0-1.rockspec
```

**2. 使用**

```lua
local ConfigClient = require("kenaito-conf")

local config = {
    baseUrl = "http://localhost:8080",
    appName = "kenaito-springboot-demo",
    env = "daily",
    syncInterval = 30,
}

local client = ConfigClient.new(config)
client:on_change(function(changed)
    for k, v in pairs(changed) do
        print("config changed: " .. k .. " = " .. v)
    end
end)
client:start()
```

### Kotlin 客户端

适用于 Kotlin 1.9+ 项目。

**1. 编译**

```bash
cd kenaito-conf-kotlin-client
./gradlew build
```

**2. 使用**

```kotlin
import kenait.conf.Config
import kenait.conf.ConfigClient

val config = Config(
    baseUrl = "http://localhost:8080",
    appName = "kenaito-springboot-demo",
    env = "daily",
    syncInterval = 30,
)

val client = ConfigClient(config)
client.onChange { changed ->
    changed.forEach { (k, v) -> println("config changed: $k = $v") }
}
client.start()
val value = client.get("database.host")
```

### Rust 客户端

适用于 Rust 1.70+ 项目。

**1. 在 Cargo.toml 中添加**

```toml
[dependencies]
kenaito-conf = { path = "../kenaito-conf-rust-client" }
```

**2. 使用**

```rust
use kenaito_conf::{Config, ConfigClient};

let config = Config {
    base_url: "http://localhost:8080".to_string(),
    app_name: "kenaito-springboot-demo".to_string(),
    env: "daily".to_string(),
    sync_interval: 30,
    ..Default::default()
};

let client = ConfigClient::new(config);
client.on_change(|changed| {
    for (k, v) in changed { println!("config changed: {} = {}", k, v); }
});
client.start().await.unwrap();
let value = client.get("database.host");
```

### Ruby 客户端

适用于 Ruby 2.7+ 项目。

**1. 安装**

```bash
cd kenaito-conf-ruby-client
gem build kenaito-conf.gemspec
gem install kenaito-conf-1.0.0.gem
```

**2. 使用**

```ruby
require 'kenaito_conf'

config = KenaitoConf::Config.new(
  base_url: 'http://localhost:8080',
  app_name: 'kenaito-springboot-demo',
  env: 'daily',
  sync_interval: 30
)

client = KenaitoConf::ConfigClient.new(config)
client.on_change do |changed|
  changed.each { |k, v| puts "config changed: #{k} = #{v}" }
end
client.start
value = client.get('database.host')
```

## API 概览

### 管理端 API（需 JWT 认证）

| 方法   | 路径                                  | 说明                       |
|------|-------------------------------------|--------------------------|
| POST | `/user/login`                       | 登录                       |
| GET  | `/user/loginUserInfo`               | 当前用户信息                   |
| POST | `/user/logout`                      | 登出                       |
| GET  | `/admin/menus`                      | 菜单                       |
| POST | `/admin/config/list`                | 配置列表（支持按环境、应用筛选）         |
| POST | `/admin/config/add`                 | 新增配置                     |
| POST | `/admin/config/update`              | 更新配置                     |
| GET  | `/admin/config/:id`                 | 配置详情                     |
| POST | `/admin/config/delete/:id`          | 删除配置                     |
| GET  | `/admin/config/configHistories/:id` | 配置历史                     |
| POST | `/admin/config/rollback/:id`        | 回滚配置                     |
| POST | `/admin/app/list`                   | 应用列表（分页）                 |
| GET  | `/admin/app/:id`                    | 应用详情                     |
| POST | `/admin/app/add`                    | 新增应用                     |
| POST | `/admin/app/update`                 | 更新应用                     |
| POST | `/admin/app/delete/:id`             | 删除应用（软删除）                |
| POST | `/admin/app/switch/:id`             | 切换应用状态                   |
| POST | `/admin/app/generateToken/:id`      | 生成应用 Token               |
| POST | `/admin/app/deleteToken/:id`        | 删除应用 Token               |
| GET  | `/admin/app/options`                | 应用选项（下拉框）                |
| GET  | `/admin/user/list`                  | 用户列表（支持按用户名、类型、所属组、状态筛选） |
| POST | `/admin/user/add`                   | 新增用户                     |
| POST | `/admin/user/update`                | 更新用户                     |
| POST | `/admin/user/delete/:id`            | 删除用户（软删除）                |
| POST | `/admin/user/resetPassword/:id`     | 重置密码                     |
| POST | `/admin/user/switch/:id`            | 切换用户状态                   |
| GET  | `/admin/user/options`               | 用户选项（下拉框）                |
| GET  | `/admin/group/list`                 | 业务组列表                    |
| GET  | `/admin/group/get/:id`              | 业务组详情                    |
| POST | `/admin/group/add`                  | 新增业务组                    |
| POST | `/admin/group/update`               | 更新业务组                    |
| POST | `/admin/group/switch/:id`           | 切换业务组状态                  |
| GET  | `/admin/group/options`              | 业务组选项（下拉框）               |
| GET  | `/admin/profile/list`               | 环境列表                     |
| GET  | `/admin/profile/:id`                | 环境详情                     |
| POST | `/admin/profile/add`                | 新增环境                     |
| POST | `/admin/profile/update`             | 更新环境                     |
| POST | `/admin/profile/delete/:id`         | 删除环境                     |
| POST | `/admin/profile/switch/:id`         | 切换环境状态                   |
| GET  | `/admin/profile/options`            | 环境选项（下拉框）                |

### 客户端 API（无需 JWT，通过应用 Token 鉴权）

| 方法   | 路径                         | 说明     |
|------|----------------------------|--------|
| POST | `/api/fetchAllConfigs`     | 全量拉取配置 |
| POST | `/api/fetchChangedConfigs` | 增量拉取变更 |
| GET  | `/api/ping`                | 健康检查   |

