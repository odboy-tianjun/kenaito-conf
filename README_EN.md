# Kenaito Config Center

A centralized configuration management platform rebuilt with Go, supporting multi-application, multi-environment, and multi-business-group configuration management and distribution.

## Tech Stack

- **Backend**: Go + Gin + xorm + MySQL
- **Frontend**: Vue 2 + Vite 5 + Element UI 2
- **Clients**: Go / Java (Spring Boot 2.x / 3.x) / Python / Node.js / Perl / Lua / PHP / .NET / C++ / Rust / Ruby / Bash / Dart

## Project Structure

```
kenaito-conf/
├── server/                              # Go admin backend (Gin + xorm)
├── web/                                 # Vue 2 frontend
├── kenaito-conf-go-client/              # Go client SDK
├── kenaito-conf-go-client-demo/         # Go client demo
├── kenaito-conf-springboot2-client/     # Java Spring Boot 2.x client SDK (JDK 8+)
├── kenaito-conf-springboot2-client-demo/# Spring Boot 2.x client demo (port 8081)
├── kenaito-conf-springboot3-client/     # Java Spring Boot 3.x client SDK (JDK 17+)
├── kenaito-conf-springboot3-client-demo/# Spring Boot 3.x client demo (port 8084)
├── kenaito-conf-python-client/          # Python client SDK (Python 3.10+)
├── kenaito-conf-python-client-demo/     # Python client demo (port 8082)
├── kenaito-conf-nodejs-client/          # Node.js client SDK (Node.js 18+)
├── kenaito-conf-nodejs-client-demo/     # Node.js client demo (port 8083)
├── kenaito-conf-perl-client/            # Perl client SDK (Perl 5.16+)
├── kenaito-conf-perl-client-demo/       # Perl client demo
├── kenaito-conf-php-client/             # PHP client SDK (PHP 7.4+)
├── kenaito-conf-php-client-demo/        # PHP client demo
├── kenaito-conf-dotnet-client/          # .NET client SDK (.NET 6.0+)
├── kenaito-conf-dotnet-client-demo/     # .NET client demo
├── kenaito-conf-lua-client/             # Lua client SDK (Lua 5.1+)
├── kenaito-conf-lua-client-demo/        # Lua client demo
├── kenaito-conf-cpp-client/             # C++ client SDK (C++11+)
├── kenaito-conf-cpp-client-demo/        # C++ client demo
├── kenaito-conf-rust-client/            # Rust client SDK (Rust 2021+)
├── kenaito-conf-rust-client-demo/       # Rust client demo
├── kenaito-conf-ruby-client/            # Ruby client SDK (Ruby 2.5+)
├── kenaito-conf-ruby-client-demo/       # Ruby client demo
├── kenaito-conf-bash-client/            # Bash client SDK (Bash 4.0+)
├── kenaito-conf-bash-client-demo/       # Bash client demo
├── kenaito-conf-dart-client/            # Dart client SDK (Dart 2.12+)
└── kenaito-conf-dart-client-demo/       # Dart client demo
```

## Quick Start

### 1. Start Backend

```bash
cd server
```

Edit `application.properties` to set the `active` environment, then edit `application-{active}.properties` to configure database connection and other settings:

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

Start the service (tables and default data will be initialized automatically on first startup):

```bash
go run cmd/main.go -config application.properties
# Or build and run
go build -o bin/server cmd/main.go
./bin/server -config application.properties
```

The backend listens on `:8080` by default.

### 2. Start Frontend

```bash
cd web
npm install
npm run dev
```

The frontend listens on `:3000` by default and automatically proxies API requests to the backend.

### 3. Access

Open http://localhost:3000

Default admin credentials: `admin` / Password: `admin123`

## Client Integration

### Go Client

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

// Register change callback
c.OnChange(func (changed map[string]string) {
    for k, v := range changed {
        fmt.Printf("%s = %s\n", k, v)
    }
})

// Start (fetch all configs + enable periodic sync)
c.Start()
defer c.Stop()

// Read configuration
dbHost := c.Get("db.host")
```

### Java Spring Boot 2.x Client

Suitable for JDK 8+ and Spring Boot 2.x projects, supports automatic config refresh and dynamic `@Value` updates.

**1. Add Dependency**

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot2-client</artifactId>
    <version>1.0.3</version>
</dependency>
```

**2. Add Configuration** (`application.properties`)

```properties
# Config center settings
kenaito.config.baseUrl=http://localhost:8080
# Application name (required)
kenaito.config.appName=myapp
# Environment (dev/daily/prod)
kenaito.config.profile=daily
# Config version
kenaito.config.version=latest
# Auth token (optional)
kenaito.config.token=app-token
# Sync interval (seconds)
kenaito.config.sync-interval-seconds=30
```

**3. Use Configuration**

```java
@Value("${database.host:localhost}")
private String databaseHost;

@Value("${database.port:3306}")
private Integer databasePort;
```

`@Value` fields are automatically refreshed when configuration changes, no application restart required. See [Spring Boot 2 Client Documentation](kenaito-conf-springboot2-client/README.md) for details.

### Java Spring Boot 3.x Client

Suitable for JDK 17+ and Spring Boot 3.x projects, supports automatic config refresh, dynamic `@Value` / `@ConfigurationProperties` updates, and local cache fallback.

**1. Add Dependency**

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot3-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

**2. Add Configuration** (`application.properties`)

```properties
# Config center settings (multiple addresses separated by commas)
kenaito.config.base-url=http://localhost:8080
# Application name (required)
kenaito.config.app-name=myapp
# Environment (dev/daily/prod)
kenaito.config.profile=daily
# Config version
kenaito.config.version=latest
# Auth token (optional)
kenaito.config.token=app-token
# Sync interval (seconds)
kenaito.config.sync-interval-seconds=30
```

**3. Use Configuration**

```java
@Value("${database.host:localhost}")
private String databaseHost;

@Value("${database.port:3306}")
private Integer databasePort;
```

`@Value` fields and `@ConfigurationProperties` beans are automatically refreshed when configuration changes, no application restart required. See [Spring Boot 3 Client Documentation](kenaito-conf-springboot3-client/README_EN.md) for details.

### Python Client

Suitable for Python 3.10+ projects.

**1. Install**

```bash
cd kenaito-conf-python-client
pip3 install -e .
```

**2. Usage**

```python
from kenaito_conf import Config, ConfigClient

config = Config(
    base_url="http://localhost:8080",
    app_name="my-app",
    env="daily",
    sync_interval=30,
)

client = ConfigClient(config)

# Listen for config changes
def on_change(changed):
    for k, v in changed.items():
        print(f"config changed: {k} = {v}")

client.on_change(on_change)

# Start client (fetch configs + enable periodic sync)
client.start()

# Get configuration
value = client.get("database.host")
all_configs = client.get_all()

# Stop
client.stop()
```

### Node.js Client

Suitable for Node.js 18+ projects.

**1. Install**

```bash
cd kenaito-conf-nodejs-client
npm link
```

Then link in your project:

```bash
npm link kenaito-conf
```

**2. Usage**

```javascript
import {Config, ConfigClient} from 'kenaito-conf';

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

// Get configuration
const value = client.get('database.host');
const allConfigs = client.getAll();

// Stop
client.stop();
```

### Perl Client

Suitable for Perl 5.16+ projects.

**1. Install**

```bash
cd kenaito-conf-perl-client
cpanm .
```

Or directly include the lib path in your project:

```perl
use lib '/path/to/kenaito-conf-perl-client/lib';
```

**2. Usage**

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

# Listen for config changes
$client->on_change(sub {
    my ($changed) = @_;
    for my $k (keys %$changed) {
        print "config changed: $k = $changed->{$k}\n";
    }
});

# Start client (fetch configs)
$client->start();

# Get configuration
my $value = $client->get('database.host');
my $all   = $client->get_all();

# Manual sync
$client->sync_once();

# Stop
$client->stop();
```

### Lua Client

Suitable for Lua 5.1+ projects.

**1. Install**

```bash
cd kenaito-conf-lua-client
luarocks make
```

Or manually add the module path:

```lua
package.path = package.path .. ";/path/to/kenaito-conf-lua-client/?.lua"
```

**2. Usage**

```lua
local kenaito = require("kenaito_conf")

local config = kenaito.Config.new({
    base_url = "http://localhost:8080",
    app_name = "my-app",
    env = "daily",
    sync_interval = 10,
})

local client = kenaito.Client.new({ config = config })

-- Listen for config changes
client:on_change(function(changed)
    for k, v in pairs(changed) do
        print("config changed: " .. k .. " = " .. v)
    end
end)

-- Start client (fetch configs)
client:start()

-- Get configuration
local value = client:get("database.host")
local all = client:get_all()

-- Manual sync
client:sync_once()

-- Stop
client:stop()
```

### PHP Client

Suitable for PHP 7.4+ projects.

**1. Install**

```bash
cd kenaito-conf-php-client
composer install
```

Or include in your project:

```bash
composer require kenaito/conf-client
```

**2. Usage**

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

// Listen for config changes
$client->onChange(function(array $changed) {
    foreach ($changed as $key => $value) {
        echo "config changed: $key = $value\n";
    }
});

// Start client (fetch configs)
$client->start();

// Get configuration
$value = $client->get('database.host');
$all = $client->getAll();

// Manual sync
$client->syncOnce();

// Stop
$client->stop();
```

### Ruby Client

For Ruby 2.5+ projects.

**1. Install**

```bash
cd kenaito-conf-ruby-client
gem build kenaito-conf.gemspec
gem install kenaito-conf-1.0.3.gem
```

Or add to your project `Gemfile`:

```ruby
gem 'kenaito-conf', path: '/path/to/kenaito-conf-ruby-client'
```

**2. Usage**

```ruby
require 'kenaito_conf'

config = Kenaito::Conf::Config.new(
  base_url: 'http://localhost:8080',
  app_name: 'my-app',
  env: 'daily',
  sync_interval: 10,
)

client = Kenaito::Conf::Client.new(config)

# Listen for config changes
client.on_change do |changed|
  changed.each do |key, value|
    puts "config changed: #{key} = #{value}"
  end
end

# Start client (fetch configs)
client.start

# Get config
value = client.get('database.host')
all_configs = client.get_all

# Manual sync once
client.sync_once

# Stop
client.stop_client
```

### Bash Client

For Bash 4.0+ scripts, CI/CD pipelines, container startup scripts, etc. Requires `curl` and `jq`.

**1. Source**

```bash
source /path/to/kenaito-conf-bash-client/lib/kenaito_conf.sh
```

**2. Usage**

```bash
#!/usr/bin/env bash
source /path/to/kenaito-conf-bash-client/lib/kenaito_conf.sh

export KENAITO_BASE_URL="http://localhost:8080"
export KENAITO_APP_NAME="my-app"
export KENAITO_ENV="daily"
export KENAITO_SYNC_INTERVAL=10

# Change callback function (optional)
kenaito_on_change() {
    local changed_keys="$1"
    echo "Config changed: $changed_keys"
}

# Start client
kenaito_start

# Get config
value=$(kenaito_get "database.host")

# Get all configs
kenaito_get_all

# Sync loop
while true; do
    kenaito_sync_once
    sleep "$KENAITO_SYNC_INTERVAL"
done
```

### Dart Client

For Dart 2.12+ projects (null-safe).

**1. Install**

Add to your project `pubspec.yaml`:

```yaml
dependencies:
  kenaito_conf:
    path: /path/to/kenaito-conf-dart-client
```

**2. Usage**

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

  // Listen for config changes
  client.onChange((Map<String, String> changed) {
    changed.forEach((key, value) {
      print('config changed: $key = $value');
    });
  });

  // Start client (fetch configs)
  client.start();
  await Future.delayed(Duration(seconds: 2));

  // Get config
  final value = client.get('database.host');
  final allConfigs = client.getAll();

  // Sync loop
  while (true) {
    client.syncOnce();
    await Future.delayed(Duration(seconds: config.syncInterval));
  }
}
```

## API Overview

### Admin API (JWT Authentication Required)

| Method | Path                                  | Description                                |
|--------|---------------------------------------|--------------------------------------------|
| POST   | `/user/login`                         | User login                                 |
| GET    | `/user/loginUserInfo`                 | Current user info                          |
| POST   | `/user/logout`                        | User logout                                |
| GET    | `/admin/menus`                        | Menu list                                  |
| POST   | `/admin/config/list`                  | Config list (filter by env, app)           |
| POST   | `/admin/config/add`                   | Add config                                 |
| POST   | `/admin/config/update`                | Update config                              |
| GET    | `/admin/config/:id`                   | Config details                             |
| POST   | `/admin/config/delete/:id`            | Delete config                              |
| GET    | `/admin/config/configHistories/:id`   | Config history                             |
| POST   | `/admin/config/rollback/:id`          | Rollback config                            |
| POST   | `/admin/app/list`                     | App list (paginated)                       |
| GET    | `/admin/app/:id`                      | App details                                |
| POST   | `/admin/app/add`                      | Add app                                    |
| POST   | `/admin/app/update`                   | Update app                                 |
| POST   | `/admin/app/delete/:id`               | Delete app (soft delete)                   |
| POST   | `/admin/app/switch/:id`               | Toggle app status                          |
| POST   | `/admin/app/generateToken/:id`        | Generate app token                         |
| POST   | `/admin/app/deleteToken/:id`          | Delete app token                           |
| GET    | `/admin/app/options`                  | App options (dropdown)                     |
| GET    | `/admin/user/list`                    | User list (filter by username, type, group, status) |
| POST   | `/admin/user/add`                     | Add user                                   |
| POST   | `/admin/user/update`                  | Update user                                |
| POST   | `/admin/user/delete/:id`              | Delete user (soft delete)                  |
| POST   | `/admin/user/resetPassword/:id`       | Reset password                             |
| POST   | `/admin/user/switch/:id`              | Toggle user status                         |
| GET    | `/admin/user/options`                 | User options (dropdown)                    |
| GET    | `/admin/group/list`                   | Business group list                        |
| GET    | `/admin/group/get/:id`                | Business group details                     |
| POST   | `/admin/group/add`                    | Add business group                         |
| POST   | `/admin/group/update`                 | Update business group                      |
| POST   | `/admin/group/switch/:id`             | Toggle business group status               |
| GET    | `/admin/group/options`                | Business group options (dropdown)          |
| GET    | `/admin/profile/list`                 | Environment list                           |
| GET    | `/admin/profile/:id`                  | Environment details                        |
| POST   | `/admin/profile/add`                  | Add environment                            |
| POST   | `/admin/profile/update`               | Update environment                         |
| POST   | `/admin/profile/delete/:id`           | Delete environment                         |
| POST   | `/admin/profile/switch/:id`           | Toggle environment status                  |
| GET    | `/admin/profile/options`              | Environment options (dropdown)             |

### Client API (No JWT, Authenticated via App Token)

| Method | Path                         | Description              |
|--------|------------------------------|--------------------------|
| POST   | `/api/fetchAllConfigs`       | Fetch all configs        |
| POST   | `/api/fetchChangedConfigs`   | Fetch incremental changes|
