# Kenaito Config Center .NET Client

.NET client SDK for the Kenaito Config Center.

## Installation

### Local Build

```bash
cd kenaito-conf-dotnet-client
dotnet build
```

### As Local Dependency

Reference in the demo project's csproj:

```xml
<ItemGroup>
  <ProjectReference Include="../kenaito-conf-dotnet-client/KenaitoConfClient.csproj" />
</ItemGroup>
```

## Quick Start

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

## Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| BaseUrl | Config center URL | - | Yes |
| AppName | Application name | - | Yes |
| Env | Environment | dev | No |
| Version | Config version | latest | No |
| Token | Auth token | null | No |
| SyncInterval | Sync interval (seconds) | 10 | No |
| CacheDir | Cache directory | . | No |
| IgnoreGlobal | Whether to ignore global configs | false | No |

## Features

- Fetch all configs on startup
- Periodic incremental config sync
- Local file cache, load from cache when service is unavailable
- Config change callback notification
