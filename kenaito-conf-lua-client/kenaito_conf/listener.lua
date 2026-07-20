local Listener = {}
Listener.__index = Listener

function Listener.new()
    local self = setmetatable({}, Listener)
    self.handlers = {}
    return self
end

function Listener:on_change(handler)
    table.insert(self.handlers, handler)
end

function Listener:notify(changed)
    for _, handler in ipairs(self.handlers) do
        local ok, err = pcall(handler, changed)
        if not ok then
            print("[ConfigClient] change handler error: " .. tostring(err))
        end
    end
end

return Listener
