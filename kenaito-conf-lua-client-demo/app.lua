local ConfigClient = require("kenaito-conf")

local config_values = {
    ["database.host"] = os.getenv("DATABASE_HOST") or "192.168.1.1",
    ["database.port"] = os.getenv("DATABASE_PORT") or "8888",
    ["database.name"] = os.getenv("DATABASE_NAME") or "xxxx",
    ["database.user"] = os.getenv("DATABASE_USER") or "xxxx",
    ["database.password"] = os.getenv("DATABASE_PASSWORD") or "xxxxxxxx",
}

local config = {
    baseUrl = os.getenv("KENAITO_BASE_URL") or "http://localhost:8080",
    appName = os.getenv("KENAITO_APP_NAME") or "kenaito-springboot-demo",
    env = os.getenv("KENAITO_PROFILE") or "daily",
    version = os.getenv("KENAITO_VERSION") or "latest",
    token = os.getenv("KENAITO_TOKEN") or "",
    syncInterval = tonumber(os.getenv("KENAITO_SYNC_INTERVAL") or "10"),
    cacheDir = os.getenv("KENAITO_CACHE_DIR") or ".",
}

local client = ConfigClient.new(config)

client:on_change(function(changed)
    for k, v in pairs(changed) do
        if config_values[k] then
            config_values[k] = v
            print("config updated: " .. k .. " = " .. v)
        end
    end
end)

client:start()

local remote = client:get_all()
for k, v in pairs(remote) do
    if config_values[k] then
        config_values[k] = v
    end
end

print("Current configs:")
for k, v in pairs(config_values) do
    print("  " .. k .. " = " .. v)
end

print("\nListening for config changes... (press Ctrl+C to stop)")
while true do
    os.execute("sleep 1")
end
