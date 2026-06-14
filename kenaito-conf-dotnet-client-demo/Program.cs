using Kenaito.Conf.Client;

var config = new Config
{
    BaseUrl = "http://localhost:8080",
    AppName = "kenaito-springboot-demo",
    Env = "daily",
    Version = "latest",
    SyncInterval = 10,
    CacheDir = "."
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
