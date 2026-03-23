using System;
using System.Collections.Generic;
using KenaitoConf;

var configValues = new Dictionary<string, string>
{
    { "database.host", Environment.GetEnvironmentVariable("DATABASE_HOST") ?? "192.168.1.1" },
    { "database.port", Environment.GetEnvironmentVariable("DATABASE_PORT") ?? "8888" },
    { "database.name", Environment.GetEnvironmentVariable("DATABASE_NAME") ?? "xxxx" },
    { "database.user", Environment.GetEnvironmentVariable("DATABASE_USER") ?? "xxxx" },
    { "database.password", Environment.GetEnvironmentVariable("DATABASE_PASSWORD") ?? "xxxxxxxx" },
};

var config = new Config
{
    BaseUrl = Environment.GetEnvironmentVariable("KENAITO_BASE_URL") ?? "http://localhost:8080",
    AppName = Environment.GetEnvironmentVariable("KENAITO_APP_NAME") ?? "kenaito-springboot-demo",
    Env = Environment.GetEnvironmentVariable("KENAITO_PROFILE") ?? "daily",
    Version = Environment.GetEnvironmentVariable("KENAITO_VERSION") ?? "latest",
    Token = Environment.GetEnvironmentVariable("KENAITO_TOKEN") ?? "",
    SyncInterval = int.Parse(Environment.GetEnvironmentVariable("KENAITO_SYNC_INTERVAL") ?? "10"),
    CacheDir = Environment.GetEnvironmentVariable("KENAITO_CACHE_DIR") ?? ".",
};

var client = new ConfigClient(config);

client.OnChange(changed =>
{
    foreach (var kv in changed)
    {
        if (configValues.ContainsKey(kv.Key))
        {
            configValues[kv.Key] = kv.Value;
            Console.Error.WriteLine($"config updated: {kv.Key} = {kv.Value}");
        }
    }
});

try
{
    await client.StartAsync();
    var remote = client.GetAll();
    foreach (var key in new List<string>(configValues.Keys))
    {
        if (remote.ContainsKey(key))
            configValues[key] = remote[key];
    }
    Console.Error.WriteLine($"loaded {remote.Count} configs from remote");
}
catch (Exception e)
{
    Console.Error.WriteLine($"failed to init config client: {e.Message}");
}

Console.WriteLine("Current configs:");
foreach (var kv in configValues)
    Console.WriteLine($"  {kv.Key} = {kv.Value}");

Console.WriteLine("\nListening for config changes... (press Ctrl+C to stop)");
var cts = new System.Threading.CancellationTokenSource();
Console.CancelKeyPress += (_, e) => { e.Cancel = true; cts.Cancel(); };
try { await System.Threading.Tasks.Task.Delay(System.Threading.Timeout.Infinite, cts.Token); }
catch (System.Threading.Tasks.TaskCanceledException) { }
client.Stop();
Console.WriteLine("Shutting down...");
