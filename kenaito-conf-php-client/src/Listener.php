<?php

namespace Kenaito\Conf;

class Listener
{
    private array $handlers = [];

    public function onChange(callable $handler): void
    {
        $this->handlers[] = $handler;
    }

    public function notify(array $changed): void
    {
        foreach ($this->handlers as $handler) {
            try {
                $handler($changed);
            } catch (\Throwable $e) {
                error_log("[ConfigClient] change handler error: " . $e->getMessage());
            }
        }
    }
}
