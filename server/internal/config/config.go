package config

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

var Cfg *Config

func Load(path string) error {
	baseProps, err := parseProps(path)
	if err != nil {
		return err
	}

	active := baseProps["active"]
	if active == "" {
		return fmt.Errorf("missing required config: active")
	}

	dir := filepath.Dir(path)
	profilePath := filepath.Join(dir, "application-"+active+".properties")
	profileProps, err := parseProps(profilePath)
	if err != nil {
		return fmt.Errorf("load profile %s: %w", active, err)
	}

	for k, v := range profileProps {
		baseProps[k] = v
	}

	Cfg = &Config{}

	Cfg.Server.Port = intProp(baseProps, "server.port", 8080)
	setServerMode(active)

	Cfg.Database.Host = baseProps["database.host"]
	Cfg.Database.Port = intProp(baseProps, "database.port", 3306)
	Cfg.Database.Name = baseProps["database.name"]
	Cfg.Database.User = baseProps["database.user"]
	Cfg.Database.Password = baseProps["database.password"]

	Cfg.JWT.Secret = baseProps["jwt.secret"]
	Cfg.JWT.ExpireHours = intProp(baseProps, "jwt.expireHours", 24)

	if Cfg.Database.Host == "" {
		return fmt.Errorf("missing required config: database.host")
	}

	return nil
}

func setServerMode(active string) {
	if "daily" == active {
		Cfg.Server.Mode = gin.DebugMode
	} else if "stage" == active || "online" == active {
		Cfg.Server.Mode = gin.ReleaseMode
	} else {
		Cfg.Server.Mode = gin.DebugMode
	}
}

func parseProps(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	props := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		props[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return props, nil
}

func intProp(props map[string]string, key string, defaultVal int) int {
	v, ok := props[key]
	if !ok || v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func (d *DatabaseConfig) DSN() string {
	return d.User + ":" + d.Password + "@tcp(" + d.Host + ":" + strconv.Itoa(d.Port) + ")/" + d.Name + "?charset=utf8&parseTime=true&loc=Local"
}
