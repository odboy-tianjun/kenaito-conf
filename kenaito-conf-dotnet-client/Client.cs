using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace Kenaito.Conf.Client;

public class Client
{
    private readonly Config _config;
    private readonly Cache _cache;
    private readonly Listener _listener;
    private readonly HttpClient _httpClient;
    private readonly Dictionary<string, string> _configs;
    private readonly object _lock = new();
    private long _lastTime;
    private bool _started;
    private CancellationTokenSource? _cts;

    public Client(Config config)
    {
        _config = config;
        if (string.IsNullOrEmpty(_config.Version))
            _config.Version = "latest";
        if (_config.SyncInterval <= 0)
            _config.SyncInterval = 10;
        if (string.IsNullOrEmpty(_config.CacheDir))
            _config.CacheDir = ".";

        _cache = new Cache(_config.CacheDir);
        _listener = new Listener();
        _configs = new Dictionary<string, string>();
        _httpClient = new HttpClient { Timeout = TimeSpan.FromSeconds(10) };
    }

    public void OnChange(ChangeHandler handler)
    {
        _listener.OnChange(handler);
    }

    public string? Get(string key)
    {
        lock (_lock)
        {
            return _configs.TryGetValue(key, out var value) ? value : null;
        }
    }

    public Dictionary<string, string> GetAll()
    {
        lock (_lock)
        {
            return new Dictionary<string, string>(_configs);
        }
    }

    public async Task StartAsync()
    {
        if (_started)
            return;
        _started = true;

        try
        {
            await FetchAllAsync();
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[ConfigClient] fetch all configs error: {ex.Message}, trying cache");
            var cached = _cache.Read();
            if (cached != null)
            {
                ApplyConfigs(cached);
                Console.WriteLine($"[ConfigClient] loaded {CountConfigs(cached)} configs from cache");
            }
            else
            {
                throw new Exception("fetch configs failed and no cache", ex);
            }
        }

        _cts = new CancellationTokenSource();
        _ = SyncLoopAsync(_cts.Token);
    }

    public void Stop()
    {
        _cts?.Cancel();
    }

    private async Task FetchAllAsync()
    {
        var url = $"{_config.BaseUrl.TrimEnd('/')}/api/fetchAllConfigs";
        var payload = new
        {
            appName = _config.AppName,
            env = _config.Env,
            version = _config.Version,
            ignoreGlobal = _config.IgnoreGlobal,
            token = _config.Token
        };
        var content = new StringContent(JsonConvert.SerializeObject(payload), System.Text.Encoding.UTF8, "application/json");
        var response = await _httpClient.PostAsync(url, content);
        response.EnsureSuccessStatusCode();

        var body = await response.Content.ReadAsStringAsync();
        var result = JsonConvert.DeserializeObject<Dictionary<string, object>>(body);
        if (result == null)
            throw new Exception("empty response");

        if (result.TryGetValue("code", out var codeObj))
        {
            var code = Convert.ToInt64(codeObj);
            if (code != 0)
            {
                var msg = result.TryGetValue("msg", out var msgObj) ? msgObj?.ToString() : "unknown error";
                throw new Exception($"server error: {msg}");
            }
        }

        if (result.TryGetValue("latest.fetch.timestamp", out var tsObj))
        {
            long.TryParse(tsObj?.ToString(), out _lastTime);
        }

        _cache.Write(result);
        ApplyConfigs(result);

        Console.WriteLine($"[ConfigClient] fetched {CountConfigs(result)} configs");
    }

    private async Task FetchChangedAsync()
    {
        var url = $"{_config.BaseUrl.TrimEnd('/')}/api/fetchChangedConfigs";
        var payload = new
        {
            appName = _config.AppName,
            env = _config.Env,
            version = _config.Version,
            lastTime = _lastTime,
            ignoreGlobal = _config.IgnoreGlobal,
            token = _config.Token
        };
        var content = new StringContent(JsonConvert.SerializeObject(payload), System.Text.Encoding.UTF8, "application/json");
        var response = await _httpClient.PostAsync(url, content);
        response.EnsureSuccessStatusCode();

        var body = await response.Content.ReadAsStringAsync();
        var wrapper = JsonConvert.DeserializeObject<FetchChangedResponse>(body);
        if (wrapper == null)
            return;

        if (wrapper.Code != 0)
            throw new Exception($"server error code: {wrapper.Code}");

        if (wrapper.Data == null)
            return;

        if (wrapper.Data.TryGetValue("latest.fetch.timestamp", out var ts))
        {
            long.TryParse(ts, out _lastTime);
            wrapper.Data.Remove("latest.fetch.timestamp");
        }

        if (wrapper.Data.Count > 0)
        {
            lock (_lock)
            {
                foreach (var kv in wrapper.Data)
                {
                    _configs[kv.Key] = kv.Value;
                }
            }
            _listener.Notify(wrapper.Data);
            Console.WriteLine($"[ConfigClient] {wrapper.Data.Count} configs changed");
        }
    }

    private void ApplyConfigs(Dictionary<string, object> data)
    {
        lock (_lock)
        {
            foreach (var kv in data)
            {
                if (kv.Key.StartsWith("kenaito.") || kv.Key == "latest.fetch.timestamp")
                    continue;
                if (kv.Value is string strVal)
                {
                    _configs[kv.Key] = strVal;
                }
            }
        }
    }

    private int CountConfigs(Dictionary<string, object> data)
    {
        return data.Count(kv => !kv.Key.StartsWith("kenaito.") && kv.Key != "latest.fetch.timestamp" && kv.Value is string);
    }

    private async Task SyncLoopAsync(CancellationToken ct)
    {
        while (!ct.IsCancellationRequested)
        {
            try
            {
                await Task.Delay(_config.SyncInterval * 1000, ct);
                await FetchChangedAsync();
            }
            catch (OperationCanceledException)
            {
                break;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[ConfigClient] sync error: {ex.Message}");
            }
        }
    }

    private class FetchChangedResponse
    {
        public int Code { get; set; }
        public Dictionary<string, string>? Data { get; set; }
    }
}
