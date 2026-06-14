# Kenaito Config Center .NET Client

kenaito-conf 配置中心的 .NET 客户端 SDK。

## 安装

### 本地编译

```bash
cd kenaito-conf-dotnet-client
dotnet build
```

### 作为本地依赖

在 demo 项目的 csproj 中引用：

```xml
<ItemGroup>
  <ProjectReference Include="../kenaito-conf-dotnet-client/KenaitoConfClient.csproj" />
</ItemGroup>
```

## 快速开始

```csharp
using Kenaito.Conf.Client;

var config = new Config
{
    BaseUrl = "http://localhost:8080",
    AppName = "kenaito-dotnet-demo",
    Env = "daily",
    SyncInterval = 10
};

var client = new Client(config);

client.OnChange(changed =>
{
    Console.WriteLine("Config changed:");
    foreach (var kv in changed)
    {
        Console.WriteLine($"  {kv.Key} = {kv.Value}");
    }
});

await client.StartAsync();

var dbHost = client.Get("database.host");
Console.WriteLine($"database.host = {dbHost}");

Console.ReadLine();
client.Stop();
```

## 配置项

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| BaseUrl | 配置中心地址 | - | 是 |
| AppName | 应用名称 | - | 是 |
| Env | 环境标识 | dev | 否 |
| Version | 配置版本 | latest | 否 |
| Token | 认证令牌 | null | 否 |
| SyncInterval | 同步间隔（秒） | 10 | 否 |
| CacheDir | 缓存目录 | . | 否 |
| IgnoreGlobal | 是否忽略全局配置 | false | 否 |

## 功能特性

- 启动时拉取全量配置
- 定时轮询增量配置变更
- 本地文件缓存，服务不可用时从缓存加载
- 配置变更回调通知
