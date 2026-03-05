<?php

require_once __DIR__ . '/../kenaito-conf-php-client/autoload.php';

use KenaitoConf\Config;
use KenaitoConf\ConfigClient;

$configValues = [
    'database.host' => getenv('DATABASE_HOST') ?: '192.168.1.1',
    'database.port' => getenv('DATABASE_PORT') ?: '8888',
    'database.name' => getenv('DATABASE_NAME') ?: 'xxxx',
    'database.user' => getenv('DATABASE_USER') ?: 'xxxx',
    'database.password' => getenv('DATABASE_PASSWORD') ?: 'xxxxxxxx',
];

$config = new Config([
    'baseUrl' => getenv('KENAITO_BASE_URL') ?: 'http://localhost:8080',
    'appName' => getenv('KENAITO_APP_NAME') ?: 'kenaito-springboot-demo',
    'env' => getenv('KENAITO_PROFILE') ?: 'daily',
    'version' => getenv('KENAITO_VERSION') ?: 'latest',
    'token' => getenv('KENAITO_TOKEN') ?: '',
    'syncInterval' => (int)(getenv('KENAITO_SYNC_INTERVAL') ?: 30),
    'cacheDir' => getenv('KENAITO_CACHE_DIR') ?: '.',
]);

$client = new ConfigClient($config);

$client->onChange(function ($changed) use (&$configValues) {
    foreach ($changed as $k => $v) {
        if (isset($configValues[$k])) {
            $configValues[$k] = $v;
            error_log("config updated: {$k} = {$v}");
        }
    }
});

try {
    $client->start();
    $remote = $client->getAll();
    foreach ($configValues as $k => &$v) {
        if (isset($remote[$k])) {
            $v = $remote[$k];
        }
    }
    error_log("loaded " . count($remote) . " configs from remote");
} catch (Exception $e) {
    error_log("failed to init config client: " . $e->getMessage());
}

$port = (int)(getenv('SERVER_PORT') ?: 8084);

echo "Server started on port {$port}\n";
echo "Current configs:\n";
foreach ($configValues as $k => $v) {
    echo "  {$k} = {$v}\n";
}

while (true) {
    pcntl_signal_dispatch();
    usleep(100000);
}
