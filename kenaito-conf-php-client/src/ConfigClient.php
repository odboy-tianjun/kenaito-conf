<?php

namespace KenaitoConf;

class ConfigClient
{
    private Config $config;
    private Cache $cache;
    private array $configs = [];
    private int $lastTime = 0;
    private bool $started = false;
    private array $handlers = [];
    private $timer = null;

    public function __construct(Config $config)
    {
        $this->config = $config;
        $this->cache = new Cache($config->cacheDir);
    }

    public function onChange(callable $handler): void
    {
        $this->handlers[] = $handler;
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
        } catch (\Exception $e) {
            error_log("[ConfigClient] fetch all configs error: {$e->getMessage()}, trying cache");
            $cached = $this->cache->read();
            if ($cached !== null) {
                $this->applyConfigs($cached);
                error_log("[ConfigClient] loaded " . count($this->configs) . " configs from cache");
            } else {
                throw new \RuntimeException("fetch configs failed and no cache: {$e->getMessage()}");
            }
        }

        $this->timer = pcntl_alarm(0);
        pcntl_signal(SIGALRM, [$this, 'syncTick']);
        pcntl_alarm($this->config->syncInterval);
    }

    public function syncTick(): void
    {
        try {
            $this->fetchChanged();
        } catch (\Exception $e) {
            error_log("[ConfigClient] sync error: {$e->getMessage()}");
        }
        pcntl_alarm($this->config->syncInterval);
    }

    public function stop(): void
    {
        if ($this->timer !== null) {
            pcntl_alarm(0);
        }
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

        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
        curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, 10);
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($httpCode !== 200) {
            throw new \RuntimeException("HTTP {$httpCode}");
        }

        $data = json_decode($response, true);
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

        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
        curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, 10);
        $response = curl_exec($ch);
        curl_close($ch);

        $result = json_decode($response, true);
        if (!isset($result['code']) || $result['code'] !== 0) {
            throw new \RuntimeException("server error code: " . ($result['code'] ?? 'unknown'));
        }

        $changedData = $result['data'] ?? [];
        if (isset($changedData['latest.fetch.timestamp'])) {
            $this->lastTime = (int)$changedData['latest.fetch.timestamp'];
            unset($changedData['latest.fetch.timestamp']);
        }

        $changed = [];
        foreach ($changedData as $k => $v) {
            $changed[$k] = $v;
        }

        $this->configs = array_merge($this->configs, $changed);

        foreach ($this->handlers as $handler) {
            try {
                $handler($changed);
            } catch (\Exception $e) {
                error_log("[ConfigClient] change handler error: {$e->getMessage()}");
            }
        }

        error_log("[ConfigClient] " . count($changed) . " configs changed");
    }

    private function applyConfigs(array $data): void
    {
        foreach ($data as $k => $v) {
            if (str_starts_with($k, 'kenaito.') || $k === 'latest.fetch.timestamp') {
                continue;
            }
            if (!is_string($v)) {
                continue;
            }
            $this->configs[$k] = $v;
        }
    }
}
