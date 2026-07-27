<?php

require_once __DIR__ . '/../kenaito-conf-php-client/autoload.php';

use Kenaito\Conf\Config;
use Kenaito\Conf\Client;

function getEnvVar(string $key, string $default): string
{
    $value = getenv($key);
    return ($value !== false && $value !== '') ? $value : $default;
}

// 配置值（可被远程配置覆盖）
$configValues = [
    'database.host' => getEnvVar('DATABASE_HOST', '192.168.1.1'),
    'database.port' => getEnvVar('DATABASE_PORT', '8888'),
    'database.name' => getEnvVar('DATABASE_NAME', 'xxxx'),
    'database.user' => getEnvVar('DATABASE_USER', 'xxxx'),
    'database.password' => getEnvVar('DATABASE_PASSWORD', 'xxxxxxxx'),
];

$config = new Config([
    'baseUrl' => getEnvVar('KENAITO_BASE_URL', 'http://localhost:8080'),
    'appName' => getEnvVar('KENAITO_APP_NAME', 'kenaito-springboot-demo'),
    'env' => getEnvVar('KENAITO_ENV', 'daily'),
    'version' => getEnvVar('KENAITO_VERSION', 'latest'),
    'token' => getEnvVar('KENAITO_TOKEN', ''),
    'syncInterval' => (int)getEnvVar('KENAITO_SYNC_INTERVAL', '10'),
    'cacheDir' => getEnvVar('KENAITO_CACHE_DIR', '.'),
]);

$client = new Client($config);

// 监听配置变更
$client->onChange(function(array $changed) use (&$configValues) {
    echo "Config changed:\n";
    foreach ($changed as $key => $value) {
        echo "  $key = $value\n";
        if (isset($configValues[$key])) {
            $configValues[$key] = $value;
            echo "  [updated] $key = $value\n";
        }
    }
});

// 启动客户端
try {
    $client->start();
    $remote = $client->getAll();
    foreach ($configValues as $key => $value) {
        if (isset($remote[$key])) {
            $configValues[$key] = $remote[$key];
        }
    }
    echo "loaded " . count($remote) . " configs from remote\n\n";
} catch (\Throwable $e) {
    echo "Failed to init config client, using local defaults: " . $e->getMessage() . "\n\n";
}

// 显示当前配置
echo "Current configs:\n";
ksort($configValues);
foreach ($configValues as $key => $value) {
    echo "  $key = $value\n";
}

echo "\nListening for config changes... (press Ctrl+C to stop)\n";

// 信号处理
pcntl_signal(SIGINT, function() use ($client) {
    echo "\nShutting down...\n";
    $client->stop();
    exit(0);
});

// 同步循环
while (!$client->stop) {
    pcntl_signal_dispatch();
    $client->syncOnce();
    sleep($config->syncInterval);
}
