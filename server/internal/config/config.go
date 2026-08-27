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

	active := os.Getenv("KENAITO_ACTIVE")
	if active == "" {
		active = baseProps["active"]
	}
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

	Cfg.Server.Port = envInt("KENAITO_SERVER_PORT", intProp(baseProps, "server.port", 8080))
	setServerMode(active)

	// 环境变量优先，便于容器化部署时覆盖配置文件
	Cfg.Database.Host = envStr("KENAITO_DB_HOST", baseProps["database.host"])
	Cfg.Database.Port = envInt("KENAITO_DB_PORT", intProp(baseProps, "database.port", 3306))
	Cfg.Database.Name = envStr("KENAITO_DB_NAME", baseProps["database.name"])
	Cfg.Database.User = envStr("KENAITO_DB_USER", baseProps["database.user"])
	Cfg.Database.Password = envStr("KENAITO_DB_PASSWORD", baseProps["database.password"])

	Cfg.JWT.Secret = envStr("KENAITO_JWT_SECRET", baseProps["jwt.secret"])
	Cfg.JWT.ExpireHours = envInt("KENAITO_JWT_EXPIRE_HOURS", intProp(baseProps, "jwt.expireHours", 24))

	if Cfg.Database.Host == "" {
		return fmt.Errorf("missing required config: database.host")
	}

	return nil
}

func setServerMode(active string) {
	if "daily" == active {
		Cfg.Server.Mode = gin.DebugMode
	} else if "stage" == active || "online" == active || "docker" == active {
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

// envStr 返回环境变量值，未设置时使用 fallback
func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envInt 返回整数环境变量值，未设置或非法时使用 fallback
func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func (d *DatabaseConfig) DSN() string {
	return d.User + ":" + d.Password + "@tcp(" + d.Host + ":" + strconv.Itoa(d.Port) + ")/" + d.Name + "?charset=utf8&parseTime=true&loc=Local"
}
