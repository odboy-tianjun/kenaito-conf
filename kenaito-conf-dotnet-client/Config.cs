namespace Kenaito.Conf.Client;

public class Config
{
    public string BaseUrl { get; set; } = "";
    public string AppName { get; set; } = "";
    public string Env { get; set; } = "dev";
    public string Version { get; set; } = "latest";
    public string? Token { get; set; }
    public int SyncInterval { get; set; } = 10;
    public string CacheDir { get; set; } = ".";
    public bool IgnoreGlobal { get; set; } = false;
}
