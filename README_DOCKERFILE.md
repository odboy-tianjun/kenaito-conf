# Kenaito Config Center 客户端 Docker 部署指南

本文档说明如何使用 Dockerfile 构建和测试各语言客户端 demo。

所有命令均在项目根目录执行。

## Go 客户端（测试通过）

```bash
docker build -f Dockerfile.go -t kenaito-conf-go-demo .

docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-go-demo
```

## Node.js 客户端（测试通过）

```bash
docker build -f Dockerfile.nodejs -t kenaito-conf-nodejs-demo .

docker run -d --rm --name nodejs-demo \
  -p 8083:8083 \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-nodejs-demo

# 查看日志
docker logs -f nodejs-demo

# 测试接口
curl http://localhost:8083/
curl http://localhost:8083/health

# 停止
docker stop nodejs-demo
```

## Python 客户端（测试通过）

```bash
docker build -f Dockerfile.python -t kenaito-conf-python-demo .

docker run -d --rm --name python-demo \
  -p 8082:8082 \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-python-demo

# 查看日志
docker logs -f python-demo

# 测试接口
curl http://localhost:8082/
curl http://localhost:8082/health

# 停止
docker stop python-demo
```

## .NET 客户端（测试通过）

```bash
docker build -f Dockerfile.dotnet -t kenaito-conf-dotnet-demo .

docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-dotnet-demo
```

## Perl 客户端（测试通过）

```bash
docker build -f Dockerfile.perl -t kenaito-conf-perl-demo .

docker run -it --rm \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-perl-demo
```

## Spring Boot 2 客户端（测试通过）

```bash
docker build -f Dockerfile.springboot2 -t kenaito-conf-springboot2-demo .

docker run -d --rm --name springboot2-demo \
  -p 8081:8081 \
  -e KENAITO_BASE_URL=http://host.docker.internal:8080 \
  -e KENAITO_APP_NAME=kenaito-springboot-demo \
  -e KENAITO_ENV=daily \
  kenaito-conf-springboot2-demo

# 查看日志
docker logs -f springboot2-demo

# 测试接口
curl http://localhost:8081/
curl http://localhost:8081/health

# 停止
docker stop springboot2-demo
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

3. **配置缓存**：各客户端会在运行目录生成 `config-cache.json` 文件，容器重启后可从缓存加载配置

4. **日志查看**：
    - Go/.NET：直接输出到 stdout
    - Node.js/Python/Spring Boot：使用各自日志框架

## 测试配置变更

1. 启动客户端容器
2. 在管理后台修改对应应用的配置
3. 等待同步间隔（默认 10 秒）
4. 观察客户端日志是否输出配置变更
5. 访问 HTTP 接口验证配置是否更新（Node.js/Python/Spring Boot）
