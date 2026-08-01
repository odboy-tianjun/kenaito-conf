# Kenaito Config Center PHP Client

[English](README_EN.md) | 中文

PHP 客户端 SDK，适用于 PHP 7.4+ 项目。

## 安装

### 使用 Composer（推荐）

```bash
cd kenaito-conf-php-client
composer install
```

在你的项目中引入：

```bash
composer require kenaito/conf-client
```

### 手动引入

```php
require_once '/path/to/kenaito-conf-php-client/autoload.php';
```

## 使用

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

## 配置项

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `baseUrl` | 配置中心地址 | `http://localhost:8080` |
| `appName` | 应用名称 | （必填） |
| `env` | 环境标识 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 应用 Token | 空 |
| `syncInterval` | 同步间隔（秒） | `10` |
| `ignoreGlobal` | 忽略全局配置 | `false` |
| `cacheDir` | 缓存目录 | `.` |

## 依赖

- PHP 7.4+
- ext-curl
- ext-json

## Demo

```bash
cd kenaito-conf-php-client-demo
php app.php
```

## 环境变量

Demo 支持以下环境变量：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `KENAITO_BASE_URL` | 配置中心地址 | `http://localhost:8080` |
| `KENAITO_APP_NAME` | 应用名称 | `kenaito-springboot-demo` |
| `KENAITO_ENV` | 环境标识 | `daily` |
| `KENAITO_VERSION` | 配置版本 | `latest` |
| `KENAITO_TOKEN` | 应用 Token | 空 |
| `KENAITO_SYNC_INTERVAL` | 同步间隔（秒） | `10` |
| `KENAITO_CACHE_DIR` | 缓存目录 | `.` |
