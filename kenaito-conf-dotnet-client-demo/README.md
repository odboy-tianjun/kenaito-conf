# Kenaito Config Center .NET Demo

演示如何使用 kenaito-conf-dotnet-client 客户端库。

## 本地运行

```bash
cd kenaito-conf-dotnet-client-demo
dotnet run
```

## Docker 运行

```bash
# 在项目根目录执行
docker build -f Dockerfile.dotnet -t kenaito-conf-dotnet-demo .

docker run -it --rm \
  -e CONFIG_CENTER_BASE_URL=http://host.docker.internal:8080 \
  kenaito-conf-dotnet-demo
```

## 配置项

修改 `Program.cs` 中的配置：

```csharp
var config = new Config
{
    BaseUrl = "http://localhost:8080",  // 配置中心地址
    AppName = "kenaito-dotnet-demo",     // 应用名称
    Env = "daily",                       // 环境
    SyncInterval = 10                    // 同步间隔（秒）
};
```

## 功能演示

- 启动时拉取全量配置并打印
- 监听配置变更，实时输出变更内容
- 按 Ctrl+C 优雅退出
