<?php

namespace Kenaito\Conf;

class Cache
{
    private string $cacheDir;
    private const CACHE_FILE = 'config-cache.json';

    public function __construct(string $cacheDir = '.')
    {
        $this->cacheDir = $cacheDir;
    }

    public function write(array $data): bool
    {
        $path = $this->cacheDir . DIRECTORY_SEPARATOR . self::CACHE_FILE;
        $json = json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
        return file_put_contents($path, $json) !== false;
    }

    public function read(): ?array
    {
        $path = $this->cacheDir . DIRECTORY_SEPARATOR . self::CACHE_FILE;
        if (!file_exists($path)) {
            return null;
        }
        $json = file_get_contents($path);
        $data = json_decode($json, true);
        return is_array($data) ? $data : null;
    }
}
