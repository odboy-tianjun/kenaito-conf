# Kenaito Config Center PHP Client

PHP 语言的配置中心客户端 SDK。

## 依赖

- PHP 8.0+
- ext-curl
- ext-openssl
- ext-pcntl

## 安装

### Composer

```bash
cd kenaito-conf-php-client
composer install
```

或在项目中引用：

```json
{
  "require": {
    "kenaito/conf-client": "*"
  },
  "repositories": [
    {
      "type": "path",
      "url": "../kenaito-conf-php-client"
    }
  ]
}
```

### 手动引入

```php
require_once 'kenaito-conf-php-client/autoload.php';
```

## 使用

```php
use KenaitoConf\Config;
use KenaitoConf\ConfigClient;

$config = new Config([
    'baseUrl' => 'http://localhost:8080',
    'appName' => 'kenaito-springboot-demo',
    'env' => 'daily',
    'version' => 'latest',
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

$client->stop();
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
| `ignoreGlobal` | 是否忽略全局配置 | `false` |
| `cacheDir` | 本地缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-php-client-demo
php app.php
```
