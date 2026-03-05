# Kenaito Config Center .NET Client

.NET 语言的配置中心客户端 SDK，支持 .NET 6.0+。

## 依赖

- .NET 6.0+
- System.Text.Json（内置）

## 安装

### 项目引用

在 `.csproj` 中添加：

```xml
<ItemGroup>
  <ProjectReference Include="..\kenaito-conf-dotnet-client\KenaitoConf.csproj" />
</ItemGroup>
```

### NuGet 包（待发布）

```bash
dotnet add package KenaitoConf
```

## 使用

```csharp
using KenaitoConf;

var config = new Config
{
    BaseUrl = "http://localhost:8080",
    AppName = "kenaito-springboot-demo",
    Env = "daily",
    Version = "latest",
    SyncInterval = 30,
};

var client = new ConfigClient(config);

client.OnChange(changed =>
{
    foreach (var kv in changed)
        Console.WriteLine($"config changed: {kv.Key} = {kv.Value}");
});

await client.StartAsync();

var value = client.Get("database.host");
var all = client.GetAll();

client.Stop();
```

## 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `BaseUrl` | 配置中心地址 | 必填 |
| `AppName` | 应用名称 | 必填 |
| `Env` | 环境 | `dev` |
| `Version` | 配置版本 | `latest` |
| `Token` | 认证令牌 | 空 |
| `SyncInterval` | 同步间隔（秒） | `60` |
| `IgnoreGlobal` | 是否忽略全局配置 | `false` |
| `CacheDir` | 本地缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-dotnet-client-demo
dotnet run
```
