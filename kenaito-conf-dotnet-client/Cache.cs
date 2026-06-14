using Newtonsoft.Json;

namespace Kenaito.Conf.Client;

internal class Cache
{
    private readonly string _cacheFile;

    public Cache(string cacheDir)
    {
        _cacheFile = Path.Combine(cacheDir, "config-cache.json");
    }

    public Dictionary<string, object>? Read()
    {
        try
        {
            if (!File.Exists(_cacheFile))
                return null;
            var json = File.ReadAllText(_cacheFile);
            return JsonConvert.DeserializeObject<Dictionary<string, object>>(json);
        }
        catch
        {
            return null;
        }
    }

    public void Write(Dictionary<string, object> data)
    {
        try
        {
            var json = JsonConvert.SerializeObject(data, Formatting.Indented);
            File.WriteAllText(_cacheFile, json);
        }
        catch
        {
            // ignore cache write errors
        }
    }
}
