package main

import (
	"fmt"
	"kenaito-conf/go-client"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func main() {
	_ = godotenv.Load()

	// create config from environment variables
	cfg := &client.Config{
		BaseURL:      getEnv("KENAITO_BASE_URL", "http://localhost:8080"),
		AppName:      getEnv("KENAITO_APP_NAME", "kenaito-springboot-demo"),
		Env:          getEnv("KENAITO_ENV", "daily"),
		Version:      getEnv("KENAITO_VERSION", "latest"),
		Token:        getEnv("KENAITO_TOKEN", ""),
		SyncInterval: getEnvInt("KENAITO_SYNC_INTERVAL", 10),
		CacheDir:     getEnv("KENAITO_CACHE_DIR", "."),
	}

	// create client
	c := client.New(cfg)

	// register change handler
	c.OnChange(func(changed map[string]string) {
		fmt.Println("Config changed:")
		for k, v := range changed {
			fmt.Printf("  %s = %s\n", k, v)
		}
	})

	// start client
	if err := c.Start(); err != nil {
		log.Fatalf("Failed to start config client: %v", err)
	}
	defer c.Stop()

	// read configs
	fmt.Println("Current configs:")
	for k, v := range c.GetAll() {
		fmt.Printf("  %s = %s\n", k, v)
	}

	// get specific config
	dbHost := c.Get("database.host")
	fmt.Printf("\ndatabase.host = %s\n", dbHost)

	// keep running until signal
	fmt.Println("\nListening for config changes... (press Ctrl+C to stop)")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
}
