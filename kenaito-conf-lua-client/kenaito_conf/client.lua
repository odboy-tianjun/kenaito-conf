local http = require("socket.http")
local ltn12 = require("ltn12")
local json = require("dkjson")

local Config = require("kenaito_conf.config")
local Cache = require("kenaito_conf.cache")
local Listener = require("kenaito_conf.listener")

local Client = {}
Client.__index = Client

function Client.new(opts)
    local config = opts.config
    if not config then
        error("config is required")
    end
    
    local self = setmetatable({}, Client)
    self.config = config
    self.cache = Cache.new(config.cache_dir)
    self.listener = Listener.new()
    self.configs = {}
    self.last_time = 0
    self.started = false
    self.stop = false
    
    return self
end

function Client:on_change(handler)
    self.listener:on_change(handler)
end

function Client:get(key)
    return self.configs[key]
end

function Client:get_all()
    local result = {}
    for k, v in pairs(self.configs) do
        result[k] = v
    end
    return result
end

function Client:start()
    if self.started then
        return
    end
    self.started = true
    
    local ok, err = self:_fetch_all()
    if not ok then
        print("[ConfigClient] fetch all configs error: " .. tostring(err) .. ", trying cache")
        local cached, cache_err = self.cache:read()
        if cached then
            self:_apply_configs(cached)
            local count = 0
            for _ in pairs(self.configs) do
                count = count + 1
            end
            print("[ConfigClient] loaded " .. count .. " configs from cache")
        else
            error("fetch configs failed and no cache: " .. tostring(err))
        end
    end
end

function Client:stop()
    self.stop = true
end

function Client:sync_once()
    local ok, err = self:_fetch_changed()
    if not ok then
        print("[ConfigClient] sync error: " .. tostring(err))
    end
end

function Client:_http_post(url, payload)
    local request_body = json.encode(payload)
    local response_body = {}
    
    local res, code, response_headers = http.request{
        url = url,
        method = "POST",
        headers = {
            ["Content-Type"] = "application/json",
            ["Content-Length"] = #request_body
        },
        source = ltn12.source.string(request_body),
        sink = ltn12.sink.table(response_body)
    }
    
    if res == nil then
        return nil, "HTTP request failed: " .. tostring(code)
    end
    
    if code ~= 200 then
        return nil, "HTTP " .. code
    end
    
    local response_str = table.concat(response_body)
    local data, _, err = json.decode(response_str)
    if err then
        return nil, "JSON decode error: " .. err
    end
    
    return data
end

function Client:_fetch_all()
    local url = self.config.base_url .. "/api/fetchAllConfigs"
    local payload = {
        appName = self.config.app_name,
        env = self.config.env,
        version = self.config.version,
        ignoreGlobal = self.config.ignore_global,
        token = self.config.token
    }
    
    local data, err = self:_http_post(url, payload)
    if not data then
        return false, err
    end
    
    if data.code and data.code ~= 0 then
        return false, "server error: " .. (data.msg or "unknown")
    end
    
    if data["latest.fetch.timestamp"] then
        self.last_time = tonumber(data["latest.fetch.timestamp"])
    end
    
    self.cache:write(data)
    self:_apply_configs(data)
    
    local count = 0
    for _ in pairs(self.configs) do
        count = count + 1
    end
    print("[ConfigClient] fetched " .. count .. " configs")
    
    return true
end

function Client:_fetch_changed()
    local url = self.config.base_url .. "/api/fetchChangedConfigs"
    local payload = {
        appName = self.config.app_name,
        env = self.config.env,
        version = self.config.version,
        lastTime = self.last_time,
        ignoreGlobal = self.config.ignore_global,
        token = self.config.token
    }
    
    local result, err = self:_http_post(url, payload)
    if not result then
        return false, err
    end
    
    if result.code and result.code ~= 0 then
        return false, "server error code: " .. result.code
    end
    
    local changed_data = result.data or {}
    
    if changed_data["latest.fetch.timestamp"] then
        self.last_time = tonumber(changed_data["latest.fetch.timestamp"])
        changed_data["latest.fetch.timestamp"] = nil
    end
    
    -- 处理删除的配置
    if changed_data.deletedKeys and changed_data.deletedKeys ~= "" then
        local deleted_keys = {}
        for key in string.gmatch(changed_data.deletedKeys, "([^,]+)") do
            table.insert(deleted_keys, key)
            self.configs[key] = nil
        end
        print("[ConfigClient] " .. #deleted_keys .. " configs deleted: " .. table.concat(deleted_keys, ", "))
        changed_data.deletedKeys = nil
    end
    
    -- 处理变更的配置
    local has_changes = false
    local changed_keys = {}
    for k, v in pairs(changed_data) do
        if type(v) == "string" then
            self.configs[k] = v
            table.insert(changed_keys, k)
            has_changes = true
        end
    end
    
    if has_changes then
        local changed = {}
        for _, k in ipairs(changed_keys) do
            changed[k] = self.configs[k]
        end
        self.listener:notify(changed)
        print("[ConfigClient] " .. #changed_keys .. " configs changed: " .. table.concat(changed_keys, ", "))
    end
    
    return true
end

function Client:_apply_configs(data)
    for k, v in pairs(data) do
        -- 跳过元数据
        if k:sub(1, 8) == "kenaito." or k == "latest.fetch.timestamp" then
            -- skip
        elseif type(v) == "string" then
            self.configs[k] = v
        end
    end
end

return Client
