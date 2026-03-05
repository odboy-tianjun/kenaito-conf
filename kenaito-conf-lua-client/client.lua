local http = require("socket.http")
local ltn12 = require("ltn12")
local cjson = require("cjson")
local openssl = require("openssl")

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
    self.secret = ""
    self.globalSecret = ""
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

local function decrypt_config(secret, data)
    local prefix = "{Cipher}"
    if string.sub(data, 1, #prefix) ~= prefix then
        return data
    end
    local encrypted = string.sub(data, #prefix + 1)
    
    -- MD5 hash
    local md5_hash = openssl.digest("md5", secret)
    local aes_key = string.sub(md5_hash, 17)
    
    -- Base64 decode
    local cipher_text = openssl.base64(encrypted, false)
    
    -- AES-128-CBC decrypt
    local iv = string.rep("\0", 16)
    local cipher = openssl.cipher.new("aes-128-cbc")
    local plain_text = cipher:decrypt(cipher_text, aes_key, iv)
    
    -- Remove PKCS7 padding
    local padding = string.byte(plain_text, -1)
    return string.sub(plain_text, 1, -padding - 1)
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
        
        if data["kenaito.config.encrypt-secret"] then
            self.secret = data["kenaito.config.encrypt-secret"]
        end
        if data["kenaito.config.global-encrypt-secret"] then
            self.globalSecret = data["kenaito.config.global-encrypt-secret"]
        end
        if data["latest.fetch.timestamp"] then
            self.lastTime = tonumber(data["latest.fetch.timestamp"])
        end
        
        for k, v in pairs(data) do
            if not string.match(k, "^kenaito%.") and k ~= "latest.fetch.timestamp" then
                if type(v) == "string" then
                    self.configs[k] = decrypt_config(self.secret, v)
                    if self.configs[k] == v and self.globalSecret ~= "" then
                        self.configs[k] = decrypt_config(self.globalSecret, v)
                    end
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
