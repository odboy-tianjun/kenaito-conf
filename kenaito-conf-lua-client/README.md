# Kenaito Config Center Lua Client

Lua 客户端 SDK，适用于 Lua 5.1+ 项目。

## 安装

### 使用 LuaRocks（推荐）

```bash
cd kenaito-conf-lua-client
luarocks make
```

### 手动安装

将 `kenaito_conf` 目录复制到你的 Lua 路径中，或添加到 `package.path`：

```lua
package.path = package.path .. ";/path/to/kenaito-conf-lua-client/?.lua"
```

## 依赖

- Lua 5.1+
- LuaSocket (`luasocket`)
- dkjson (`dkjson`)

## 使用

```lua
local kenaito = require("kenaito_conf")

local config = kenaito.Config.new({
    base_url = "http://localhost:8080",
    app_name = "my-app",
    env = "daily",
    sync_interval = 10
})

local client = kenaito.Client.new({ config = config })

-- 监听配置变更
client:on_change(function(changed)
    for k, v in pairs(changed) do
        print("config changed: " .. k .. " = " .. v)
    end
end)

-- 启动客户端（拉取配置）
client:start()

-- 获取配置
local value = client:get("database.host")
local all = client:get_all()

-- 手动同步一次
client:sync_once()

-- 停止
client:stop()
```

## 配置项

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `base_url` | 配置中心地址 | `http://localhost:8080` |
| `app_name` | 应用名称 | （必填） |
| `env` | 环境标识 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 应用 Token | 空 |
| `sync_interval` | 同步间隔（秒） | `10` |
| `ignore_global` | 忽略全局配置 | `false` |
| `cache_dir` | 缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-lua-client-demo
lua app.lua
```

## 环境变量

Demo 支持以下环境变量：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `KENAITO_BASE_URL` | 配置中心地址 | `http://localhost:8080` |
| `KENAITO_APP_NAME` | 应用名称 | `kenaito-springboot-demo` |
| `KENAITO_ENV` | 环境标识 | `daily` |
| `KENAITO_VERSION` | 配置版本 | `latest` |
| `KENAITO_TOKEN` | 应用 Token | 空 |
| `KENAITO_SYNC_INTERVAL` | 同步间隔（秒） | `10` |
| `KENAITO_CACHE_DIR` | 缓存目录 | `.` |
