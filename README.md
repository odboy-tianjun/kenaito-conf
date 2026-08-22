# Kenaito Config Center

[English](README_EN.md) | 中文

基于 Go 重构的集中配置管理平台，支持多应用、多环境、多业务组的配置管理与分发。

## 技术栈

- **后端**: Go + Gin + xorm + MySQL
- **前端**: Vue 2 + Vite 5 + Element UI 2
- **客户端**: Go / Java（Spring Boot 2.x / 3.x）/ Python / Node.js / Perl / Lua / PHP / .NET / C++ / Rust / Ruby / Bash / Dart

## 项目结构

```
kenaito-conf/
├── server/                              # Go 管理端（Gin + xorm）
├── web/                                 # Vue 2 前端
├── kenaito-conf-go-client/              # Go 客户端 SDK
├── kenaito-conf-go-client-demo/         # Go 客户端示例
├── kenaito-conf-springboot2-client/     # Java Spring Boot 2.x 客户端 SDK（JDK 8+）
├── kenaito-conf-springboot2-client-demo/# Spring Boot 2.x 客户端示例（端口 8081）
├── kenaito-conf-springboot3-client/     # Java Spring Boot 3.x 客户端 SDK（JDK 17+）
├── kenaito-conf-springboot3-client-demo/# Spring Boot 3.x 客户端示例（端口 8084）
├── kenaito-conf-python-client/          # Python 客户端 SDK（Python 3.10+）
├── kenaito-conf-python-client-demo/     # Python 客户端示例（端口 8082）
├── kenaito-conf-nodejs-client/          # Node.js 客户端 SDK（Node.js 18+）
├── kenaito-conf-nodejs-client-demo/     # Node.js 客户端示例（端口 8083）
├── kenaito-conf-perl-client/            # Perl 客户端 SDK（Perl 5.16+）
├── kenaito-conf-perl-client-demo/       # Perl 客户端示例
├── kenaito-conf-php-client/             # PHP 客户端 SDK（PHP 7.4+）
├── kenaito-conf-php-client-demo/        # PHP 客户端示例
├── kenaito-conf-dotnet-client/          # .NET 客户端 SDK（.NET 6.0+）
├── kenaito-conf-dotnet-client-demo/     # .NET 客户端示例
├── kenaito-conf-lua-client/             # Lua 客户端 SDK（Lua 5.1+）
├── kenaito-conf-lua-client-demo/        # Lua 客户端示例
├── kenaito-conf-cpp-client/             # C++ 客户端 SDK（C++11+）
├── kenaito-conf-cpp-client-demo/        # C++ 客户端示例
├── kenaito-conf-rust-client/            # Rust 客户端 SDK（Rust 2021+）
├── kenaito-conf-rust-client-demo/       # Rust 客户端示例
├── kenaito-conf-ruby-client/            # Ruby 客户端 SDK（Ruby 2.5+）
├── kenaito-conf-ruby-client-demo/       # Ruby 客户端示例
├── kenaito-conf-bash-client/            # Bash 客户端 SDK（Bash 4.0+）
├── kenaito-conf-bash-client-demo/       # Bash 客户端示例
├── kenaito-conf-dart-client/            # Dart 客户端 SDK（Dart 2.12+）
└── kenaito-conf-dart-client-demo/       # Dart 客户端示例
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
import client "kenaito-conf/go-client"

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
kenaito.config.baseUrl=http://localhost:8080
# 应用名称（必填）
kenaito.config.appName=myapp
# 环境（dev/daily/prod）
kenaito.config.profile=daily
# 配置版本
kenaito.config.version=latest
# 认证令牌（可选）
kenaito.config.token=app-token
# 同步间隔（秒）
kenaito.config.sync-interval-seconds=30
```

**3. 使用配置**

```java
@Value("${database.host:localhost}")
private String databaseHost;

@Value("${database.port:3306}")
private Integer databasePort;
```

配置变更时会自动刷新 `@Value`
字段，无需重启应用。详见 [Spring Boot 2 客户端文档](kenaito-conf-springboot2-client/README.md)。

### Java Spring Boot 3.x 客户端

适用于 JDK 17+、Spring Boot 3.x 项目，支持配置自动刷新、`@Value` / `@ConfigurationProperties` 动态更新和本地缓存容灾。

**1. 引入依赖**

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot3-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

**2. 添加配置**（`application.properties`）

```properties
# 配置中心地址（多个地址用逗号分隔）
kenaito.config.base-url=http://localhost:8080
# 应用名称（必填）
kenaito.config.app-name=myapp
# 环境（dev/daily/prod）
kenaito.config.profile=daily
# 配置版本
kenaito.config.version=latest
# 认证令牌（可选）
kenaito.config.token=app-token
# 同步间隔（秒）
kenaito.config.sync-interval-seconds=30
```

**3. 使用配置**

```java
@Value("${database.host:localhost}")
private String databaseHost;

@Value("${database.port:3306}")
private Integer databasePort;
```

配置变更时会自动刷新 `@Value` 字段和 `@ConfigurationProperties` Bean，无需重启应用。详见 [Spring Boot 3 客户端文档](kenaito-conf-springboot3-client/README.md)。

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

### Perl 客户端

适用于 Perl 5.16+ 项目。

**1. 安装**

```bash
cd kenaito-conf-perl-client
cpanm .
```

或在项目中直接引入 lib 路径：

```perl
use lib '/path/to/kenaito-conf-perl-client/lib';
```

**2. 使用**

```perl
use Kenaito::Conf::Config;
use Kenaito::Conf::Client;

my $cfg = Kenaito::Conf::Config->new(
    base_url      => 'http://localhost:8080',
    app_name      => 'my-app',
    env           => 'daily',
    sync_interval => 10,
);

my $client = Kenaito::Conf::Client->new(config => $cfg);

# 监听配置变更
$client->on_change(sub {
    my ($changed) = @_;
    for my $k (keys %$changed) {
        print "config changed: $k = $changed->{$k}\n";
    }
});

# 启动客户端（拉取配置）
$client->start();

# 获取配置
my $value = $client->get('database.host');
my $all   = $client->get_all();

# 手动同步一次
$client->sync_once();

# 停止
$client->stop();
```

### Lua 客户端

适用于 Lua 5.1+ 项目。

**1. 安装**

```bash
cd kenaito-conf-lua-client
luarocks make
```

或手动添加模块路径：

```lua
package.path = package.path .. ";/path/to/kenaito-conf-lua-client/?.lua"
```

**2. 使用**

```lua
local kenaito = require("kenaito_conf")

local config = kenaito.Config.new({
    base_url = "http://localhost:8080",
    app_name = "my-app",
    env = "daily",
    sync_interval = 10,
})

local client = kenaito.Client.new({ config = config })

-- 监听配置变更
client:on_change(function(changed)
    for k, v in pairs(changed) do
        print("config changed: " .. k .. " = " .. v)
    end
end)

-- 启动客户端（拉取配置）
client:start()

-- 获取配置
local value = client:get("database.host")
local all = client:get_all()

-- 手动同步一次
client:sync_once()

-- 停止
client:stop()
```

### PHP 客户端

适用于 PHP 7.4+ 项目。

**1. 安装**

```bash
cd kenaito-conf-php-client
composer install
```

或在项目中引入：

```bash
composer require kenaito/conf-client
```

**2. 使用**

```php
require_once 'vendor/autoload.php';

use Kenaito\Conf\Config;
use Kenaito\Conf\Client;

$config = new Config([
    'baseUrl' => 'http://localhost:8080',
    'appName' => 'my-app',
    'env' => 'daily',
    'syncInterval' => 10,
]);

$client = new Client($config);

// 监听配置变更
$client->onChange(function(array $changed) {
    foreach ($changed as $key => $value) {
        echo "config changed: $key = $value\n";
    }
});

// 启动客户端（拉取配置）
$client->start();

// 获取配置
$value = $client->get('database.host');
$all = $client->getAll();

// 手动同步一次
$client->syncOnce();

// 停止
$client->stop();
```

### Ruby 客户端

适用于 Ruby 2.5+ 项目。

**1. 安装**

```bash
cd kenaito-conf-ruby-client
gem build kenaito-conf.gemspec
gem install kenaito-conf-1.0.3.gem
```

或在项目 `Gemfile` 中引入：

```ruby
gem 'kenaito-conf', path: '/path/to/kenaito-conf-ruby-client'
```

**2. 使用**

```ruby
require 'kenaito_conf'

config = Kenaito::Conf::Config.new(
  base_url: 'http://localhost:8080',
  app_name: 'my-app',
  env: 'daily',
  sync_interval: 10,
)

client = Kenaito::Conf::Client.new(config)

# 监听配置变更
client.on_change do |changed|
  changed.each do |key, value|
    puts "config changed: #{key} = #{value}"
  end
end

# 启动客户端（拉取配置）
client.start

# 获取配置
value = client.get('database.host')
all_configs = client.get_all

# 手动同步一次
client.sync_once

# 停止
client.stop_client
```

### Bash 客户端

适用于 Bash 4.0+ 脚本、CI/CD 流水线、容器启动脚本等场景。依赖 `curl` 和 `jq`。

**1. 引入**

```bash
source /path/to/kenaito-conf-bash-client/lib/kenaito_conf.sh
```

**2. 使用**

```bash
#!/usr/bin/env bash
source /path/to/kenaito-conf-bash-client/lib/kenaito_conf.sh

export KENAITO_BASE_URL="http://localhost:8080"
export KENAITO_APP_NAME="my-app"
export KENAITO_ENV="daily"
export KENAITO_SYNC_INTERVAL=10

# 变更回调函数（可选）
kenaito_on_change() {
    local changed_keys="$1"
    echo "Config changed: $changed_keys"
}

# 启动客户端
kenaito_start

# 获取配置
value=$(kenaito_get "database.host")

# 获取所有配置
kenaito_get_all

# 同步循环
while true; do
    kenaito_sync_once
    sleep "$KENAITO_SYNC_INTERVAL"
done
```

### Dart 客户端

适用于 Dart 2.12+ 项目（支持空安全）。

**1. 安装**

在项目 `pubspec.yaml` 中添加：

```yaml
dependencies:
  kenaito_conf:
    path: /path/to/kenaito-conf-dart-client
```

**2. 使用**

```dart
import 'package:kenaito_conf/kenaito_conf.dart';

void main() async {
  final config = Config(
    baseUrl: 'http://localhost:8080',
    appName: 'my-app',
    env: 'daily',
    syncInterval: 10,
  );

  final client = Client(config);

  // 监听配置变更
  client.onChange((Map<String, String> changed) {
    changed.forEach((key, value) {
      print('config changed: $key = $value');
    });
  });

  // 启动客户端（拉取配置）
  client.start();
  await Future.delayed(Duration(seconds: 2));

  // 获取配置
  final value = client.get('database.host');
  final allConfigs = client.getAll();

  // 同步循环
  while (true) {
    client.syncOnce();
    await Future.delayed(Duration(seconds: config.syncInterval));
  }
}
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

