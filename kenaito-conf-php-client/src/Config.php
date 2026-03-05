<?php

namespace KenaitoConf;

class Config
{
    public string $baseUrl = '';
    public string $appName = '';
    public string $env = 'dev';
    public string $version = 'latest';
    public string $token = '';
    public int $syncInterval = 60;
    public bool $ignoreGlobal = false;
    public string $cacheDir = '.';

    public function __construct(array $options = [])
    {
        foreach ($options as $key => $value) {
            if (property_exists($this, $key)) {
                $this->$key = $value;
            }
        }
    }
}
