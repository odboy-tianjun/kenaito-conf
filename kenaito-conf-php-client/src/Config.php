<?php

namespace Kenaito\Conf;

class Config
{
    public string $baseUrl;
    public string $appName;
    public string $env;
    public string $version;
    public string $token;
    public int $syncInterval;
    public bool $ignoreGlobal;
    public string $cacheDir;

    public function __construct(array $options = [])
    {
        $this->baseUrl = $options['baseUrl'] ?? 'http://localhost:8080';
        $this->appName = $options['appName'] ?? '';
        $this->env = $options['env'] ?? 'dev';
        $this->version = $options['version'] ?? 'latest';
        $this->token = $options['token'] ?? '';
        $this->syncInterval = $options['syncInterval'] ?? 10;
        $this->ignoreGlobal = $options['ignoreGlobal'] ?? false;
        $this->cacheDir = $options['cacheDir'] ?? '.';
    }
}
