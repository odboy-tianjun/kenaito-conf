package main

import (
	"fmt"
	"kenaito-conf/go-client"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// create config
	cfg := &client.Config{
		BaseURL:      "http://localhost:8080",
		AppName:      "myapp",
		Env:          "dev",
		Version:      "latest",
		SyncInterval: 30, // sync every 30 seconds
		CacheDir:     ".",
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
	dbHost := c.Get("db.host")
	fmt.Printf("\ndb.host = %s\n", dbHost)

	// keep running until signal
	fmt.Println("\nListening for config changes... (press Ctrl+C to stop)")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
}
