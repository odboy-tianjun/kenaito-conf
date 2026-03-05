<?php

namespace KenaitoConf;

class Cache
{
    private string $cacheDir;
    private const CACHE_FILE = 'config-cache.json';

    public function __construct(string $cacheDir = '.')
    {
        $this->cacheDir = $cacheDir;
    }

    public function write(array $data): void
    {
        $path = $this->cacheDir . DIRECTORY_SEPARATOR . self::CACHE_FILE;
        file_put_contents($path, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE));
    }

    public function read(): ?array
    {
        $path = $this->cacheDir . DIRECTORY_SEPARATOR . self::CACHE_FILE;
        if (!file_exists($path)) {
            return null;
        }
        $content = file_get_contents($path);
        return json_decode($content, true);
    }
}
