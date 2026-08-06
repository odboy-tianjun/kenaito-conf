# Kenaito Config Center Lua Client

Lua client SDK for the Kenaito Config Center, suitable for Lua 5.1+ projects.

## Features

- Fetch application configs from the config center
- Periodic config sync
- Config change callback listener
- Local file cache (fallback when config center is unavailable)

## Installation

### Using LuaRocks (Recommended)

```bash
cd kenaito-conf-lua-client
luarocks make
```

### Manual Install

Copy the `kenaito_conf` directory to your Lua path, or add it to `package.path`:

```lua
package.path = package.path .. ";/path/to/kenaito-conf-lua-client/?.lua"
```

## Dependencies

- Lua 5.1+
- LuaSocket (`luasocket`)
- dkjson (`dkjson`)

## Usage

```lua
local kenaito = require("kenaito_conf")

local config = kenaito.Config.new({
    base_url = "http://localhost:8080",
    app_name = "my-app",
    env = "daily",
    sync_interval = 10
})

local client = kenaito.Client.new({ config = config })

-- Listen for config changes
client:on_change(function(changed)
    for k, v in pairs(changed) do
        print("config changed: " .. k .. " = " .. v)
    end
end)

-- Start client (fetch configs)
client:start()

-- Get config
local value = client:get("database.host")
local all = client:get_all()

-- Manual sync
client:sync_once()

-- Stop
client:stop()
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `base_url` | Config center URL | `http://localhost:8080` |
| `app_name` | Application name | (Required) |
| `env` | Environment | `dev` |
| `version` | Config version | `latest` |
| `token` | App token | Empty |
| `sync_interval` | Sync interval (seconds) | `10` |
| `ignore_global` | Whether to ignore global configs | `false` |
| `cache_dir` | Cache directory | `.` |

## Demo

```bash
cd kenaito-conf-lua-client-demo
lua app.lua
```

## Environment Variables

The demo supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `KENAITO_BASE_URL` | Config center URL | `http://localhost:8080` |
| `KENAITO_APP_NAME` | Application name | `kenaito-springboot-demo` |
| `KENAITO_ENV` | Environment | `daily` |
| `KENAITO_VERSION` | Config version | `latest` |
| `KENAITO_TOKEN` | App token | Empty |
| `KENAITO_SYNC_INTERVAL` | Sync interval (seconds) | `10` |
| `KENAITO_CACHE_DIR` | Cache directory | `.` |
