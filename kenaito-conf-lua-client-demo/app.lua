#!/usr/bin/env lua

-- 添加客户端库路径
local script_dir = debug.getinfo(1, "S").source:match("@(.+)/[^/]+$") or "."
package.path = package.path .. ";" .. script_dir .. "/../kenaito-conf-lua-client/?.lua"

local kenaito = require("kenaito_conf")

-- 读取环境变量
local function get_env(key, default)
    local value = os.getenv(key)
    if value and value ~= "" then
        return value
    end
    return default
end

-- 配置值（可被远程配置覆盖）
local config_values = {
    ["database.host"] = get_env("DATABASE_HOST", "192.168.1.1"),
    ["database.port"] = get_env("DATABASE_PORT", "8888"),
    ["database.name"] = get_env("DATABASE_NAME", "xxxx"),
    ["database.user"] = get_env("DATABASE_USER", "xxxx"),
    ["database.password"] = get_env("DATABASE_PASSWORD", "xxxxxxxx"),
}

-- 创建配置
local config = kenaito.Config.new({
    base_url = get_env("KENAITO_BASE_URL", "http://localhost:8080"),
    app_name = get_env("KENAITO_APP_NAME", "kenaito-springboot-demo"),
    env = get_env("KENAITO_ENV", "daily"),
    version = get_env("KENAITO_VERSION", "latest"),
    token = get_env("KENAITO_TOKEN", ""),
    sync_interval = tonumber(get_env("KENAITO_SYNC_INTERVAL", "10")),
    cache_dir = get_env("KENAITO_CACHE_DIR", "."),
})

-- 创建客户端
local client = kenaito.Client.new({ config = config })

-- 监听配置变更
client:on_change(function(changed)
    print("Config changed:")
    for k, v in pairs(changed) do
        print("  " .. k .. " = " .. v)
        if config_values[k] then
            config_values[k] = v
            print("  [updated] " .. k .. " = " .. v)
        end
    end
end)

-- 启动客户端
local ok, err = pcall(function()
    client:start()
end)

if not ok then
    print("Failed to init config client, using local defaults: " .. tostring(err))
end

-- 应用远程配置
local remote = client:get_all()
local count = 0
for _ in pairs(remote) do
    count = count + 1
end

for k, v in pairs(remote) do
    if config_values[k] then
        config_values[k] = v
    end
end

print("loaded " .. count .. " configs from remote")
print()

-- 显示当前配置
print("Current configs:")
local keys = {}
for k in pairs(config_values) do
    table.insert(keys, k)
end
table.sort(keys)
for _, k in ipairs(keys) do
    print("  " .. k .. " = " .. config_values[k])
end

print()
print("Listening for config changes... (press Ctrl+C to stop)")

-- 同步循环
local running = true

-- 注册信号处理（如果支持）
if package.loaded["posix.signal"] then
    local signal = require("posix.signal")
    signal.signal(signal.SIGINT, function()
        print("\nShutting down...")
        running = false
        client:stop()
    end)
end

-- 主循环
while running do
    client:sync_once()
    
    -- 等待 sync_interval 秒，期间检查是否应该停止
    for i = 1, config.sync_interval * 10 do
        if not running then
            break
        end
        -- 休眠 0.1 秒
        local start = os.clock()
        while os.clock() - start < 0.1 do
            -- busy wait
        end
    end
end
