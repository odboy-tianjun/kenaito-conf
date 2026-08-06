# Kenaito Config Center PHP Client

PHP client SDK for the Kenaito Config Center, suitable for PHP 7.4+ projects.

## Features

- Fetch application configs from the config center
- Periodic config sync
- Config change callback listener
- Local file cache (fallback when config center is unavailable)

## Installation

### Using Composer (Recommended)

```bash
cd kenaito-conf-php-client
composer install
```

In your project:

```bash
composer require kenaito/conf-client
```

### Manual Include

```php
require_once '/path/to/kenaito-conf-php-client/autoload.php';
```

## Usage

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

// Get config
$value = $client->get('database.host');
$all = $client->getAll();

// Manual sync
$client->syncOnce();

// Stop
$client->stop();
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `baseUrl` | Config center URL | `http://localhost:8080` |
| `appName` | Application name | (Required) |
| `env` | Environment | `dev` |
| `version` | Config version | `latest` |
| `token` | App token | Empty |
| `syncInterval` | Sync interval (seconds) | `10` |
| `ignoreGlobal` | Whether to ignore global configs | `false` |
| `cacheDir` | Cache directory | `.` |

## Dependencies

- PHP 7.4+
- ext-curl
- ext-json

## Demo

```bash
cd kenaito-conf-php-client-demo
php app.php
```

## Environment Variables

The demo supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `KENAITO_BASE_URL` | Config center URL | `http://localhost:8080` |
| `KENAITO_APP_NAME` | Application name | `kenaito-springboot-demo` |
| `KENAITO_ENV` | Environment | `daily` |
| `KENAITO_VERSION` | Config version | `latest` |
| `KENAITO_TOKEN` | App token | Empty |
| `KENAITO_SYNC_INTERVAL` | Sync interval (seconds) | `10` |
| `KENAITO_CACHE_DIR` | Cache directory | `.` |
