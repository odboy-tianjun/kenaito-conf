# Kenaito Config Center Lua Client

Lua 语言的配置中心客户端 SDK。

## 依赖

- Lua 5.3+
- LuaSocket
- lua-cjson
- lua-openssl

## 安装

```bash
luarocks install kenaito-conf-client-1.0.0-1.rockspec
```

或手动安装依赖：

```bash
luarocks install luasocket
luarocks install lua-cjson
luarocks install lua-openssl
```

## 使用

```lua
local ConfigClient = require("kenaito-conf")

local config = {
    baseUrl = "http://localhost:8080",
    appName = "kenaito-springboot-demo",
    env = "daily",
    version = "latest",
    syncInterval = 30,
}

local client = ConfigClient.new(config)

client:on_change(function(changed)
    for k, v in pairs(changed) do
        print("config changed: " .. k .. " = " .. v)
    end
end)

client:start()

local value = client:get("database.host")
local all = client:get_all()

client:stop()
```

## 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `baseUrl` | 配置中心地址 | 必填 |
| `appName` | 应用名称 | 必填 |
| `env` | 环境 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 认证令牌 | 空 |
| `syncInterval` | 同步间隔（秒） | `60` |
| `ignoreGlobal` | 是否忽略全局配置 | `false` |
| `cacheDir` | 本地缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-lua-client-demo
lua app.lua
```
