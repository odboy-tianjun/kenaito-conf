package client

// Config holds the client configuration
type Config struct {
	BaseURL      string `json:"baseUrl"`
	AppName      string `json:"appName"`
	Env          string `json:"env"`
	Version      string `json:"version"`
	Token        string `json:"token"`
	SyncInterval int    `json:"syncInterval"` // seconds
	IgnoreGlobal bool   `json:"ignoreGlobal"`
	CacheDir     string `json:"cacheDir"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		Version:      "latest",
		SyncInterval: 60,
		CacheDir:     ".",
	}
}
