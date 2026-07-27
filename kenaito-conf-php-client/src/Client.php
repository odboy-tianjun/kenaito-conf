<?php

namespace Kenaito\Conf;

class Client
{
    private Config $config;
    private Cache $cache;
    private Listener $listener;
    private array $configs = [];
    private int $lastTime = 0;
    private bool $started = false;
    public bool $stop = false;

    public function __construct(Config $config)
    {
        $this->config = $config;
        $this->cache = new Cache($config->cacheDir);
        $this->listener = new Listener();
    }

    public function onChange(callable $handler): void
    {
        $this->listener->onChange($handler);
    }

    public function get(string $key): ?string
    {
        return $this->configs[$key] ?? null;
    }

    public function getAll(): array
    {
        return $this->configs;
    }

    public function start(): void
    {
        if ($this->started) {
            return;
        }
        $this->started = true;

        try {
            $this->fetchAll();
        } catch (\Throwable $e) {
            error_log("[ConfigClient] fetch all configs error: " . $e->getMessage() . ", trying cache");
            $cached = $this->cache->read();
            if ($cached !== null) {
                $this->applyConfigs($cached);
                error_log("[ConfigClient] loaded " . count($this->configs) . " configs from cache");
            } else {
                throw new \RuntimeException("fetch configs failed and no cache: " . $e->getMessage(), 0, $e);
            }
        }
    }

    public function stop(): void
    {
        $this->stop = true;
    }

    public function syncOnce(): void
    {
        try {
            $this->fetchChanged();
        } catch (\Throwable $e) {
            error_log("[ConfigClient] sync error: " . $e->getMessage());
        }
    }

    private function httpPost(string $url, array $payload): array
    {
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
        curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
        curl_setopt($ch, CURLOPT_TIMEOUT, 10);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);

        if ($response === false) {
            throw new \RuntimeException("HTTP request failed: " . $error);
        }

        if ($httpCode !== 200) {
            throw new \RuntimeException("HTTP " . $httpCode);
        }

        $data = json_decode($response, true);
        if (!is_array($data)) {
            throw new \RuntimeException("JSON decode error");
        }

        return $data;
    }

    private function fetchAll(): void
    {
        $url = $this->config->baseUrl . '/api/fetchAllConfigs';
        $payload = [
            'appName' => $this->config->appName,
            'env' => $this->config->env,
            'version' => $this->config->version,
            'ignoreGlobal' => $this->config->ignoreGlobal,
            'token' => $this->config->token,
        ];

        $data = $this->httpPost($url, $payload);

        if (isset($data['code']) && $data['code'] !== 0) {
            throw new \RuntimeException("server error: " . ($data['msg'] ?? 'unknown'));
        }

        if (isset($data['latest.fetch.timestamp'])) {
            $this->lastTime = (int)$data['latest.fetch.timestamp'];
        }

        $this->cache->write($data);
        $this->applyConfigs($data);

        error_log("[ConfigClient] fetched " . count($this->configs) . " configs");
    }

    private function fetchChanged(): void
    {
        $url = $this->config->baseUrl . '/api/fetchChangedConfigs';
        $payload = [
            'appName' => $this->config->appName,
            'env' => $this->config->env,
            'version' => $this->config->version,
            'lastTime' => $this->lastTime,
            'ignoreGlobal' => $this->config->ignoreGlobal,
            'token' => $this->config->token,
        ];

        $result = $this->httpPost($url, $payload);

        if (isset($result['code']) && $result['code'] !== 0) {
            throw new \RuntimeException("server error code: " . $result['code']);
        }

        $changedData = $result['data'] ?? [];

        if (isset($changedData['latest.fetch.timestamp'])) {
            $this->lastTime = (int)$changedData['latest.fetch.timestamp'];
            unset($changedData['latest.fetch.timestamp']);
        }

        // 处理删除的配置
        if (isset($changedData['deletedKeys']) && $changedData['deletedKeys'] !== '') {
            $deletedKeys = explode(',', $changedData['deletedKeys']);
            foreach ($deletedKeys as $key) {
                unset($this->configs[$key]);
            }
            error_log("[ConfigClient] " . count($deletedKeys) . " configs deleted: " . implode(', ', $deletedKeys));
            unset($changedData['deletedKeys']);
        }

        // 处理变更的配置
        if (!empty($changedData)) {
            $changed = [];
            foreach ($changedData as $key => $value) {
                if (is_string($value)) {
                    $this->configs[$key] = $value;
                    $changed[$key] = $value;
                }
            }

            if (!empty($changed)) {
                $this->listener->notify($changed);
                error_log("[ConfigClient] " . count($changed) . " configs changed: " . implode(', ', array_keys($changed)));
            }
        }
    }

    private function applyConfigs(array $data): void
    {
        foreach ($data as $key => $value) {
            if (strpos($key, 'kenaito.') === 0 || $key === 'latest.fetch.timestamp') {
                continue;
            }
            if (is_string($value)) {
                $this->configs[$key] = $value;
            }
        }
    }
}
