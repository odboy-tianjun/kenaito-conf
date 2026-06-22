package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the config center client
type Client struct {
	config     *Config
	cache      *Cache
	listener   *Listener
	configs    map[string]string
	lastTime   int64
	mu         sync.RWMutex
	httpClient *http.Client
	stopCh     chan struct{}
	stopOnce   sync.Once
	started    bool
}

// New creates a new config client
func New(config *Config) *Client {
	if config.Version == "" {
		config.Version = "latest"
	}
	if config.SyncInterval <= 0 {
		config.SyncInterval = 60
	}
	if config.CacheDir == "" {
		config.CacheDir = "."
	}

	return &Client{
		config:     config,
		cache:      NewCache(config.CacheDir),
		listener:   NewListener(),
		configs:    make(map[string]string),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		stopCh:     make(chan struct{}),
	}
}

// OnChange registers a config change handler
func (c *Client) OnChange(handler ChangeHandler) {
	c.listener.OnChange(handler)
}

// Get returns a config value by key
func (c *Client) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.configs[key]
}

// GetAll returns all configs
func (c *Client) GetAll() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]string, len(c.configs))
	for k, v := range c.configs {
		result[k] = v
	}
	return result
}

// Start initializes the client and starts sync loop
func (c *Client) Start() error {
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return nil
	}
	c.started = true
	c.mu.Unlock()

	// fetch all configs
	if err := c.fetchAll(); err != nil {
		log.Printf("[ConfigClient] fetch all configs error: %v, trying cache", err)
		if cached, cacheErr := c.cache.Read(); cacheErr == nil {
			c.applyConfigs(cached)
			log.Printf("[ConfigClient] loaded %d configs from cache", len(c.configs))
		} else {
			return fmt.Errorf("fetch configs failed and no cache: %w", err)
		}
	}

	// start sync loop
	go c.syncLoop()

	return nil
}

// Stop stops the sync loop
func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

func (c *Client) fetchAll() error {
	url := fmt.Sprintf("%s/api/fetchAllConfigs", c.config.BaseURL)

	payload := map[string]interface{}{
		"appName":      c.config.AppName,
		"env":          c.config.Env,
		"version":      c.config.Version,
		"ignoreGlobal": c.config.IgnoreGlobal,
		"token":        c.config.Token,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	// check error response
	if code, ok := result["code"]; ok {
		codeInt, _ := code.(float64)
		if codeInt != 0 {
			msg, _ := result["msg"].(string)
			return fmt.Errorf("server error: %s", msg)
		}
	}

	// extract timestamp
	if ts, ok := result["latest.fetch.timestamp"]; ok {
		if tsStr, ok := ts.(string); ok {
			fmt.Sscanf(tsStr, "%d", &c.lastTime)
		}
	}

	// save to cache
	c.cache.Write(result)

	// apply configs
	c.applyConfigs(result)

	log.Printf("[ConfigClient] fetched %d configs", len(c.configs))
	return nil
}

func (c *Client) fetchChanged() error {
	url := fmt.Sprintf("%s/api/fetchChangedConfigs", c.config.BaseURL)

	payload := map[string]interface{}{
		"appName":      c.config.AppName,
		"env":          c.config.Env,
		"version":      c.config.Version,
		"lastTime":     c.lastTime,
		"ignoreGlobal": c.config.IgnoreGlobal,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var wrapper struct {
		Code int                 `json:"code"`
		Data map[string]string   `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return err
	}

	if wrapper.Code != 0 {
		return fmt.Errorf("server error code: %d", wrapper.Code)
	}

	// update timestamp
	if ts, ok := wrapper.Data["latest.fetch.timestamp"]; ok {
		fmt.Sscanf(ts, "%d", &c.lastTime)
		delete(wrapper.Data, "latest.fetch.timestamp")
	}

	// 处理删除的配置
	if deletedKeysStr, ok := wrapper.Data["deletedKeys"]; ok && deletedKeysStr != "" {
		deletedKeys := strings.Split(deletedKeysStr, ",")
		c.mu.Lock()
		for _, key := range deletedKeys {
			delete(c.configs, key)
		}
		c.mu.Unlock()
		log.Printf("[ConfigClient] %d configs deleted: %s", len(deletedKeys), strings.Join(deletedKeys, ", "))
		delete(wrapper.Data, "deletedKeys")
	}

	if len(wrapper.Data) > 0 {
		changed := make(map[string]string)
		for k, v := range wrapper.Data {
			changed[k] = v
		}

		// update local configs
		c.mu.Lock()
		for k, v := range changed {
			c.configs[k] = v
		}
		c.mu.Unlock()

		// notify listeners
		c.listener.Notify(changed)
		changedKeys := make([]string, 0, len(changed))
		for k := range changed {
			changedKeys = append(changedKeys, k)
		}
		log.Printf("[ConfigClient] %d configs changed: %s", len(changed), strings.Join(changedKeys, ", "))
	}

	return nil
}

func (c *Client) applyConfigs(data map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range data {
		// skip meta keys
		if strings.HasPrefix(k, "kenaito.") || k == "latest.fetch.timestamp" {
			continue
		}
		strVal, ok := v.(string)
		if !ok {
			continue
		}
		c.configs[k] = strVal
	}
}

func (c *Client) syncLoop() {
	ticker := time.NewTicker(time.Duration(c.config.SyncInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.fetchChanged(); err != nil {
				log.Printf("[ConfigClient] sync error: %v", err)
			}
		case <-c.stopCh:
			return
		}
	}
}
