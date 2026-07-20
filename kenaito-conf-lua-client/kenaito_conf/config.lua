local Config = {}
Config.__index = Config

function Config.new(opts)
    opts = opts or {}
    local self = setmetatable({}, Config)
    self.base_url = opts.base_url or "http://localhost:8080"
    self.app_name = opts.app_name or ""
    self.env = opts.env or "dev"
    self.version = opts.version or "latest"
    self.token = opts.token or ""
    self.sync_interval = opts.sync_interval or 10
    self.ignore_global = opts.ignore_global or false
    self.cache_dir = opts.cache_dir or "."
    return self
end

return Config
