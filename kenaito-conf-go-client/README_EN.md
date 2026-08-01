# Kenaito Config Center Go Client

Go client SDK for the Kenaito Config Center.

## Features

- Fetch application configs from the config center
- Periodic config sync
- Config change callback listener
- Local file cache (fallback when config center is unavailable)

## Installation

### Option 1: Using go mod (Recommended)

Run in your project root:

```bash
go get kenaito-conf/go-client
```

### Option 2: Local Development

If you are developing or modifying the client code locally, use the `replace` directive:

```bash
# Add to your go.mod
require kenaito-conf/go-client v0.0.0

replace kenaito-conf/go-client => ../kenaito-conf-go-client
```

Then run:

```bash
go mod tidy
```

## Quick Start

```go
package main

import (
    "fmt"
    "kenaito-conf/go-client"
    "log"
)

func main() {
    // Create config
    cfg := &client.Config{
        BaseURL:      "http://localhost:8080",
        AppName:      "my-app",
        Env:          "daily",
        Version:      "latest",
        Token:        "", // Optional, if app token verification is enabled
        SyncInterval: 30, // Sync interval in seconds
        CacheDir:     ".", // Cache directory
    }

    // Create client
    c := client.New(cfg)

    // Register config change callback
    c.OnChange(func(changed map[string]string) {
        fmt.Println("Config changed:")
        for k, v := range changed {
            fmt.Printf("  %s = %s\n", k, v)
        }
    })

    // Start client (fetch configs + enable periodic sync)
    if err := c.Start(); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    defer c.Stop()

    // Read config
    value := c.Get("database.host")
    fmt.Printf("database.host = %s\n", value)

    // Get all configs
    all := c.GetAll()
    for k, v := range all {
        fmt.Printf("%s = %s\n", k, v)
    }
}
```

## Configuration

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `BaseURL` | string | Config center URL | Required |
| `AppName` | string | Application name | Required |
| `Env` | string | Environment (dev/daily/prod) | `dev` |
| `Version` | string | Config version | `latest` |
| `Token` | string | App token (if token verification is enabled) | Empty |
| `SyncInterval` | int | Sync interval (seconds) | `60` |
| `IgnoreGlobal` | bool | Whether to ignore global configs | `false` |
| `CacheDir` | string | Local cache directory | `.` |

## API Reference

### `client.New(cfg *Config) *Client`

Create a client instance.

### `(*Client).Start() error`

Start the client, fetch configs and enable periodic sync.

### `(*Client).Stop()`

Stop the client and disable periodic sync.

### `(*Client).Get(key string) string`

Get the value of a specific config entry.

### `(*Client).GetAll() map[string]string`

Get all config entries.

### `(*Client).OnChange(handler func(map[string]string))`

Register a config change callback. The function is called when configs change, with the changed key-value pairs as argument.

## Demo

See the full example at: [kenaito-conf-go-client-demo](../kenaito-conf-go-client-demo)

Run the demo:

```bash
cd kenaito-conf-go-client-demo
go run main.go
```

## Notes

1. **Token Verification**: If the app has a token configured, pass the correct token value in `Config.Token`
2. **Cache File**: The client creates a `config-cache.json` file in the `CacheDir` directory as a fallback when the config center is unavailable
