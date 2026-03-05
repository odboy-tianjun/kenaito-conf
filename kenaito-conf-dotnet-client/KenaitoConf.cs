using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

namespace KenaitoConf
{
    public class Config
    {
        public string BaseUrl { get; set; } = "";
        public string AppName { get; set; } = "";
        public string Env { get; set; } = "dev";
        public string Version { get; set; } = "latest";
        public string Token { get; set; } = "";
        public int SyncInterval { get; set; } = 60;
        public bool IgnoreGlobal { get; set; } = false;
        public string CacheDir { get; set; } = ".";
    }

    public static class Crypto
    {
        private const string CryptPrefix = "{Cipher}";

        public static string DecryptConfig(string secret, string data)
        {
            if (!data.StartsWith(CryptPrefix))
                return data;
            return DecryptWithAES(secret, data.Substring(CryptPrefix.Length));
        }

        public static string DecryptWithAES(string key, string data)
        {
            using var md5 = MD5.Create();
            string md5Hash = BitConverter.ToString(md5.ComputeHash(Encoding.UTF8.GetBytes(key))).Replace("-", "").ToLower();
            string aesKeyStr = md5Hash.Substring(16);
            byte[] aesKey = Encoding.UTF8.GetBytes(aesKeyStr);
            byte[] cipherText = Convert.FromBase64String(data);
            byte[] iv = new byte[16];

            using var aes = Aes.Create();
            aes.Key = aesKey;
            aes.IV = iv;
            aes.Mode = CipherMode.CBC;
            aes.Padding = PaddingMode.None;

            using var decryptor = aes.CreateDecryptor();
            byte[] plainText = decryptor.TransformFinalBlock(cipherText, 0, cipherText.Length);
            int padding = plainText[plainText.Length - 1];
            return Encoding.UTF8.GetString(plainText, 0, plainText.Length - padding);
        }
    }

    public class Cache
    {
        private readonly string _cacheDir;
        private const string CacheFile = "config-cache.json";

        public Cache(string cacheDir = ".")
        {
            _cacheDir = cacheDir;
        }

        public void Write(Dictionary<string, object> data)
        {
            string path = Path.Combine(_cacheDir, CacheFile);
            File.WriteAllText(path, JsonSerializer.Serialize(data, new JsonSerializerOptions { WriteIndented = true }));
        }

        public Dictionary<string, object> Read()
        {
            string path = Path.Combine(_cacheDir, CacheFile);
            if (!File.Exists(path)) return null;
            string content = File.ReadAllText(path);
            return JsonSerializer.Deserialize<Dictionary<string, object>>(content);
        }
    }

    public class ConfigClient
    {
        private readonly Config _config;
        private readonly Cache _cache;
        private readonly HttpClient _http = new();
        private readonly Dictionary<string, string> _configs = new();
        private string _secret = "";
        private string _globalSecret = "";
        private long _lastTime = 0;
        private bool _started = false;
        private CancellationTokenSource _cts;
        private readonly List<Action<Dictionary<string, string>>> _handlers = new();

        public ConfigClient(Config config)
        {
            _config = config;
            _cache = new Cache(config.CacheDir);
            _http.Timeout = TimeSpan.FromSeconds(10);
        }

        public void OnChange(Action<Dictionary<string, string>> handler)
        {
            _handlers.Add(handler);
        }

        public string Get(string key)
        {
            return _configs.TryGetValue(key, out var value) ? value : null;
        }

        public Dictionary<string, string> GetAll()
        {
            return new Dictionary<string, string>(_configs);
        }

        public async Task StartAsync()
        {
            if (_started) return;
            _started = true;

            try
            {
                await FetchAllAsync();
            }
            catch (Exception e)
            {
                Console.Error.WriteLine($"[ConfigClient] fetch all configs error: {e.Message}, trying cache");
                var cached = _cache.Read();
                if (cached != null)
                {
                    ApplyConfigs(cached);
                    Console.Error.WriteLine($"[ConfigClient] loaded {_configs.Count} configs from cache");
                }
                else
                {
                    throw new Exception($"fetch configs failed and no cache: {e.Message}");
                }
            }

            _cts = new CancellationTokenSource();
            _ = SyncLoopAsync(_cts.Token);
        }

        public void Stop()
        {
            _cts?.Cancel();
        }

        private async Task SyncLoopAsync(CancellationToken ct)
        {
            while (!ct.IsCancellationRequested)
            {
                await Task.Delay(_config.SyncInterval * 1000, ct);
                try
                {
                    await FetchChangedAsync();
                }
                catch (Exception e)
                {
                    Console.Error.WriteLine($"[ConfigClient] sync error: {e.Message}");
                }
            }
        }

        private async Task FetchAllAsync()
        {
            string url = $"{_config.BaseUrl}/api/fetchAllConfigs";
            var payload = new
            {
                appName = _config.AppName,
                env = _config.Env,
                version = _config.Version,
                ignoreGlobal = _config.IgnoreGlobal,
                token = _config.Token
            };
            var content = new StringContent(JsonSerializer.Serialize(payload), Encoding.UTF8, "application/json");
            var resp = await _http.PostAsync(url, content);
            resp.EnsureSuccessStatusCode();
            string body = await resp.Content.ReadAsStringAsync();
            var data = JsonSerializer.Deserialize<Dictionary<string, object>>(body);

            if (data.ContainsKey("code"))
            {
                var code = Convert.ToInt32(data["code"]);
                if (code != 0)
                    throw new Exception($"server error: {(data.ContainsKey("msg") ? data["msg"] : "unknown")}");
            }

            if (data.TryGetValue("kenaito.config.encrypt-secret", out var s))
                _secret = s?.ToString() ?? "";
            if (data.TryGetValue("kenaito.config.global-encrypt-secret", out var gs))
                _globalSecret = gs?.ToString() ?? "";
            if (data.TryGetValue("latest.fetch.timestamp", out var ts))
                _lastTime = Convert.ToInt64(ts);

            _cache.Write(data);
            ApplyConfigs(data);
            Console.Error.WriteLine($"[ConfigClient] fetched {_configs.Count} configs");
        }

        private async Task FetchChangedAsync()
        {
            string url = $"{_config.BaseUrl}/api/fetchChangedConfigs";
            var payload = new
            {
                appName = _config.AppName,
                env = _config.Env,
                version = _config.Version,
                lastTime = _lastTime,
                ignoreGlobal = _config.IgnoreGlobal,
                token = _config.Token
            };
            var content = new StringContent(JsonSerializer.Serialize(payload), Encoding.UTF8, "application/json");
            var resp = await _http.PostAsync(url, content);
            resp.EnsureSuccessStatusCode();
            string body = await resp.Content.ReadAsStringAsync();
            var result = JsonSerializer.Deserialize<Dictionary<string, object>>(body);

            if (!result.ContainsKey("code") || Convert.ToInt32(result["code"]) != 0)
                throw new Exception("server error");

            var changedData = JsonSerializer.Deserialize<Dictionary<string, object>>(result["data"].ToString());
            if (changedData.TryGetValue("latest.fetch.timestamp", out var ts))
            {
                _lastTime = Convert.ToInt64(ts);
                changedData.Remove("latest.fetch.timestamp");
            }

            var changed = new Dictionary<string, string>();
            foreach (var kv in changedData)
            {
                changed[kv.Key] = TryDecrypt(kv.Value?.ToString() ?? "");
            }

            foreach (var kv in changed)
                _configs[kv.Key] = kv.Value;

            foreach (var handler in _handlers)
            {
                try { handler(changed); }
                catch (Exception e) { Console.Error.WriteLine($"[ConfigClient] handler error: {e.Message}"); }
            }
        }

        private void ApplyConfigs(Dictionary<string, object> data)
        {
            foreach (var kv in data)
            {
                if (kv.Key.StartsWith("kenaito.") || kv.Key == "latest.fetch.timestamp")
                    continue;
                if (kv.Value is JsonElement)
                    continue;
                _configs[kv.Key] = TryDecrypt(kv.Value?.ToString() ?? "");
            }
        }

        private string TryDecrypt(string value)
        {
            string decrypted = Crypto.DecryptConfig(_secret, value);
            if (decrypted == value && !string.IsNullOrEmpty(_globalSecret))
                decrypted = Crypto.DecryptConfig(_globalSecret, value);
            return decrypted;
        }
    }
}
