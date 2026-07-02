using Kenaito.Conf.Client;

var config = new Config
{
    BaseUrl = Environment.GetEnvironmentVariable("KENAITO_BASE_URL") ?? "http://localhost:8080",
    AppName = Environment.GetEnvironmentVariable("KENAITO_APP_NAME") ?? "kenaito-springboot-demo",
    Env = Environment.GetEnvironmentVariable("KENAITO_ENV") ?? "daily",
    Version = Environment.GetEnvironmentVariable("KENAITO_VERSION") ?? "latest",
    SyncInterval = int.TryParse(Environment.GetEnvironmentVariable("KENAITO_SYNC_INTERVAL"), out var si) ? si : 10,
    CacheDir = Environment.GetEnvironmentVariable("KENAITO_CACHE_DIR") ?? ".",
    Token = Environment.GetEnvironmentVariable("KENAITO_TOKEN")
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

Console.WriteLine("Current configs:");
foreach (var kv in client.GetAll())
{
    Console.WriteLine($"  {kv.Key} = {kv.Value}");
}

var dbHost = client.Get("database.host");
Console.WriteLine($"\ndatabase.host = {dbHost}");

Console.WriteLine("\nListening for config changes... (press Ctrl+C to stop)");
var cts = new CancellationTokenSource();
Console.CancelKeyPress += (s, e) =>
{
    e.Cancel = true;
    cts.Cancel();
};

try
{
    await Task.Delay(Timeout.Infinite, cts.Token);
}
catch (TaskCanceledException)
{
}

Console.WriteLine("\nShutting down...");
client.Stop();
