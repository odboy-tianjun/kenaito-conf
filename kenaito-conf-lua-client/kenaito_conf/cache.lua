local json = require("dkjson")

local Cache = {}
Cache.__index = Cache

local CACHE_FILE = "config-cache.json"

function Cache.new(cache_dir)
    local self = setmetatable({}, Cache)
    self.cache_dir = cache_dir or "."
    return self
end

function Cache:write(data)
    local path = self.cache_dir .. "/" .. CACHE_FILE
    local file = io.open(path, "w")
    if not file then
        return false, "cannot open cache file: " .. path
    end
    local str = json.encode(data, { indent = true })
    file:write(str)
    file:close()
    return true
end

function Cache:read()
    local path = self.cache_dir .. "/" .. CACHE_FILE
    local file = io.open(path, "r")
    if not file then
        return nil, "cache file not found"
    end
    local str = file:read("*a")
    file:close()
    local data, _, err = json.decode(str)
    if err then
        return nil, "json decode error: " .. err
    end
    return data
end

return Cache
