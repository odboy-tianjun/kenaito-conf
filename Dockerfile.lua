# Build stage
FROM alpine:3.20 AS build

WORKDIR /src

# Install Lua and dependencies
RUN apk add --no-cache lua5.4 lua5.4-dev luarocks5.4 git build-base

# Create symlinks
RUN ln -sf /usr/bin/lua5.4 /usr/bin/lua && \
    ln -sf /usr/bin/luarocks-5.4 /usr/bin/luarocks

# Copy client library
COPY kenaito-conf-lua-client/ ./kenaito-conf-lua-client/

# Install Lua dependencies
WORKDIR /src/kenaito-conf-lua-client
RUN luarocks install luasocket && \
    luarocks install dkjson

# Copy demo project
WORKDIR /src
COPY kenaito-conf-lua-client-demo/ ./kenaito-conf-lua-client-demo/

# Runtime stage
FROM alpine:3.20

WORKDIR /app

# Install Lua runtime
RUN apk add --no-cache lua5.4

# Create symlink
RUN ln -sf /usr/bin/lua5.4 /usr/bin/lua

# Copy Lua modules from build stage
COPY --from=build /usr/local/share/lua/5.4 /usr/local/share/lua/5.4
COPY --from=build /usr/local/lib/lua/5.4 /usr/local/lib/lua/5.4

# Copy client library
COPY --from=build /src/kenaito-conf-lua-client/kenaito_conf /app/kenaito_conf

# Copy demo
COPY --from=build /src/kenaito-conf-lua-client-demo/app.lua /app/app.lua

ENV LUA_PATH="/app/?.lua;/usr/local/share/lua/5.4/?.lua;;"
ENV LUA_CPATH="/usr/local/lib/lua/5.4/?.so;;"

ENTRYPOINT ["lua", "app.lua"]
