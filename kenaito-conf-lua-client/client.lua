local http = require("socket.http")
local ltn12 = require("ltn12")
local cjson = require("cjson")

local ConfigClient = {}
ConfigClient.__index = ConfigClient

function ConfigClient.new(config)
    local self = setmetatable({}, ConfigClient)
    self.config = config or {}
    self.config.baseUrl = self.config.baseUrl or ""
    self.config.appName = self.config.appName or ""
    self.config.env = self.config.env or "dev"
    self.config.version = self.config.version or "latest"
    self.config.token = self.config.token or ""
    self.config.syncInterval = self.config.syncInterval or 60
    self.config.ignoreGlobal = self.config.ignoreGlobal or false
    self.config.cacheDir = self.config.cacheDir or "."

    self.configs = {}
    self.lastTime = 0
    self.handlers = {}
    return self
end

function ConfigClient:on_change(handler)
    table.insert(self.handlers, handler)
end

function ConfigClient:get(key)
    return self.configs[key]
end

function ConfigClient:get_all()
    return self.configs
end

function ConfigClient:start()
    -- Fetch all configs
    local url = self.config.baseUrl .. "/api/fetchAllConfigs"
    local payload = cjson.encode({
        appName = self.config.appName,
        env = self.config.env,
        version = self.config.version,
        ignoreGlobal = self.config.ignoreGlobal,
        token = self.config.token
    })
    
    local response_body = {}
    local res, code, response_headers = http.request{
        url = url,
        method = "POST",
        headers = {
            ["Content-Type"] = "application/json",
            ["Content-Length"] = #payload
        },
        source = ltn12.source.string(payload),
        sink = ltn12.sink.table(response_body)
    }
    
    if code == 200 then
        local data = cjson.decode(table.concat(response_body))

        if data["latest.fetch.timestamp"] then
            self.lastTime = tonumber(data["latest.fetch.timestamp"])
        end

        for k, v in pairs(data) do
            if not string.match(k, "^kenaito%.") and k ~= "latest.fetch.timestamp" then
                if type(v) == "string" then
                    self.configs[k] = v
                end
            end
        end

        print("[ConfigClient] fetched " .. #self.configs .. " configs")
    end
end

function ConfigClient:stop()
    -- Cleanup
end

return ConfigClient
