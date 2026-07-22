package = "kenaito-conf"
version = "1.0.3-1"

source = {
    url = "git://github.com/odboy/kenaito-conf.git",
    tag = "v1.0.3",
    dir = "kenaito-conf-lua-client"
}

description = {
    summary = "Kenaito Config Center Lua Client",
    detailed = [[
        Lua client SDK for Kenaito Config Center.
        Supports config fetching, caching, and change notification.
    ]],
    homepage = "https://github.com/odboy/kenaito-conf",
    license = "Apache-2.0",
    maintainer = "odboy <tianjun@odboy.cn>"
}

dependencies = {
    "lua >= 5.1",
    "luasocket",
    "dkjson"
}

build = {
    type = "builtin",
    modules = {
        ["kenaito_conf"] = "kenaito_conf/init.lua",
        ["kenaito_conf.client"] = "kenaito_conf/client.lua",
        ["kenaito_conf.config"] = "kenaito_conf/config.lua",
        ["kenaito_conf.cache"] = "kenaito_conf/cache.lua",
        ["kenaito_conf.listener"] = "kenaito_conf/listener.lua"
    }
}
