package = "kenaito-conf-client"
version = "1.0.0-1"
source = {
    url = "git+https://gitea.odboy.cn/odboy/kenaito-conf.git"
}
description = {
    summary = "Kenaito Config Center Lua Client",
    homepage = "https://gitea.odboy.cn/odboy/kenaito-conf",
    license = "MIT"
}
dependencies = {
    "lua >= 5.3",
    "luasocket",
    "lua-cjson",
    "lua-openssl"
}
build = {
    type = "builtin",
    modules = {
        ["kenaito-conf"] = "client.lua"
    }
}
