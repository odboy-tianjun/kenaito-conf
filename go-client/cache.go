package client

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const cacheFile = "config-cache.json"

// Cache provides local file caching for configs
type Cache struct {
	dir string
}

func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

func (c *Cache) Write(data map[string]interface{}) error {
	path := filepath.Join(c.dir, cacheFile)
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

func (c *Cache) Read() (map[string]interface{}, error) {
	path := filepath.Join(c.dir, cacheFile)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	return data, err
}
