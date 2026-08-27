# Kenaito Config Center 客户端 Docker 部署指南

本文档说明如何使用 Dockerfile 构建和测试各语言客户端 demo，以及如何使用 Docker Compose 一键部署管理端。

所有命令均在项目根目录执行。

## 服务端一键部署（Docker Compose）

```bash
# 可选：按需修改端口、数据库密码和 JWT 密钥
cp .env.example .env

docker compose up -d --build
```

Compose 会启动三个服务：

| 服务 | 说明 | 默认端口 |
| --- | --- | --- |
| mysql | MySQL 8.0，数据持久化到 `mysql-data` 卷 | 仅容器网络 |
| server | Go 后端，首次启动自动建表并初始化管理员 | 8080 |
| web | Nginx 托管的 Vue 前端，并反向代理 `/api`、`/admin`、`/user` | 3000 |

启动后访问 http://localhost:3000 ，默认管理员 `admin` / `admin123`。

常用命令：

```bash
docker compose ps            # 查看服务与健康状态
docker compose logs -f server
docker compose down          # 停止服务，保留 MySQL 数据
docker compose down -v       # 停止服务并删除 MySQL 数据卷
```

环境变量说明（写入 `.env` 即可生效）：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| WEB_PORT | 3000 | 前端对外端口 |
| SERVER_PORT | 8080 | 后端对外端口 |
| MYSQL_DATABASE | kenaito_conf | 数据库名 |
| MYSQL_USER / MYSQL_PASSWORD | kenaito / kenaito123 | 后端使用的数据库账号 |
| MYSQL_ROOT_PASSWORD | root123456 | MySQL root 密码 |
| JWT_SECRET | kenaito-conf-docker-secret-change-me | JWT 密钥，生产环境必须修改 |
| JWT_EXPIRE_HOURS | 24 | JWT 有效期（小时） |
| TZ | Asia/Shanghai | 容器时区 |

后端配置文件为 `server/application-docker.properties`，以下 `KENAITO_*` 环境变量优先级更高：`KENAITO_ACTIVE`、`KENAITO_SERVER_PORT`、`KENAITO_DB_HOST`、`KENAITO_DB_PORT`、`KENAITO_DB_NAME`、`KENAITO_DB_USER`、`KENAITO_DB_PASSWORD`、`KENAITO_JWT_SECRET`、`KENAITO_JWT_EXPIRE_HOURS`。

后端健康检查接口：`GET /health`，数据库不可用时返回 503。

## Go 客户端

```bash
docker build -f Dockerfile.go -t kenaito-conf-go-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-golang-demo
  
# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-golang-demo
```

## Node.js 客户端

```bash
docker build -f Dockerfile.nodejs -t kenaito-conf-nodejs-demo .

# linux or mac
docker run -d --rm --name nodejs-demo \
  -p 8083:8083 \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-nodejs-demo
  
# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-nodejs-demo

# 查看日志
docker logs -f nodejs-demo

# 测试接口
curl http://localhost:8083/
curl http://localhost:8083/health

# 停止
docker stop nodejs-demo
```

## Python 客户端

```bash
docker build -f Dockerfile.python -t kenaito-conf-python-demo .

# linux or mac
docker run -d --rm --name python-demo \
  -p 8082:8082 \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-python-demo
  
# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-python-demo

# 查看日志
docker logs -f python-demo

# 测试接口
curl http://localhost:8082/
curl http://localhost:8082/health

# 停止
docker stop python-demo
```

## .NET 客户端

```bash
docker build -f Dockerfile.dotnet -t kenaito-conf-dotnet-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-dotnet-demo

# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-dotnet-demo
```

## Perl 客户端

```bash
docker build -f Dockerfile.perl -t kenaito-conf-perl-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-perl-demo
  
# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-perl-demo
```

## Lua 客户端

```bash
docker build -f Dockerfile.lua -t kenaito-conf-lua-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-lua-demo
  
# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-lua-demo
```

## PHP 客户端

```bash
docker build -f Dockerfile.php -t kenaito-conf-php-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-php-demo

# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-php-demo
```

## Spring Boot 2 客户端

```bash
docker build -f Dockerfile.springboot2 -t kenaito-conf-springboot2-demo .

# linux or mac
docker run -d --rm --name springboot2-demo \
  -p 8081:8081 \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-springboot2-demo
  
# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-springboot2-demo

# 查看日志
docker logs -f springboot2-demo

# 测试接口
curl http://localhost:8081/
curl http://localhost:8081/health

# 停止
docker stop springboot2-demo
```

## Spring Boot 3 客户端

Spring Boot 3 客户端要求 JDK 17+，Demo 默认监听 `8084`。构建阶段使用 Maven 3 + JDK 21 镜像编译，运行阶段使用 `eclipse-temurin:17-jre`。

```bash
docker build -f Dockerfile.springboot3 -t kenaito-conf-springboot3-demo .

# linux or mac
docker run -d --rm --name springboot3-demo \
  -p 8084:8084 \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-springboot3-demo

# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-springboot3-demo

# 查看日志
docker logs -f springboot3-demo

# 测试接口
curl http://localhost:8084/config
curl http://localhost:8084/configprops
curl http://localhost:8084/changes
curl http://localhost:8084/health

# 停止
docker stop springboot3-demo
```

## C++ 客户端

```bash
docker build -f Dockerfile.cpp -t kenaito-conf-cpp-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-cpp-demo
  
# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-cpp-demo
```

## Rust 客户端

```bash
docker build -f Dockerfile.rust -t kenaito-conf-rust-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-rust-demo

# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-rust-demo
```

## Ruby 客户端

```bash
docker build -f Dockerfile.ruby -t kenaito-conf-ruby-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-ruby-demo

# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-ruby-demo
```

## Bash 客户端

```bash
docker build -f Dockerfile.bash -t kenaito-conf-bash-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-bash-demo

# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-bash-demo
```

## Dart 客户端

```bash
docker build -f Dockerfile.dart -t kenaito-conf-dart-demo .

# linux or mac
docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-dart-demo

# windows
docker run -it --rm -e KENAITO_BASE_URL=http://host.docker.internal:8080 -e KENAITO_APP_NAME=kenaito-springboot-demo -e KENAITO_ENV=daily kenaito-conf-dart-demo
```

## 环境变量说明

| 变量名                     | 说明       | 默认值                       |
|-------------------------|----------|---------------------------|
| `KENAITO_BASE_URL`      | 配置中心地址   | `http://localhost:8080`   |
| `KENAITO_APP_NAME`      | 应用名称     | `kenaito-springboot-demo` |
| `KENAITO_ENV`           | 环境标识     | `daily`                   |
| `KENAITO_VERSION`       | 配置版本     | `latest`                  |
| `KENAITO_TOKEN`         | 应用 Token | 空                         |
| `KENAITO_SYNC_INTERVAL` | 同步间隔（秒）  | `10`                      |
| `KENAITO_CACHE_DIR`     | 缓存目录     | `.`                       |

## 注意事项

1. **host.docker.internal**：容器内访问宿主机服务时使用 `host.docker.internal:8080`，Linux 系统可能需要添加
   `--add-host host.docker.internal:host-gateway`

2. **网络模式**：如果配置中心也运行在 Docker 中，建议使用自定义网络：
   ```bash
   docker network create kenaito-net
   docker run --network kenaito-net --name kenaito-server ...
   docker run --network kenaito-net -e KENAITO_BASE_URL=http://kenaito-server:8080 ...
   ```

3. **配置缓存**：各客户端会在运行目录生成 `config-cache.*` 缓存文件（Spring Boot 客户端为 `config-cache.log`），容器重启后可从缓存加载配置

4. **日志查看**：
    - Go/.NET：直接输出到 stdout
    - Node.js/Python/Spring Boot：使用各自日志框架

## 测试配置变更

1. 启动客户端容器
2. 在管理后台修改对应应用的配置
3. 等待同步间隔（默认 10 秒）
4. 观察客户端日志是否输出配置变更
5. 访问 HTTP 接口验证配置是否更新（Node.js/Python/Spring Boot）
