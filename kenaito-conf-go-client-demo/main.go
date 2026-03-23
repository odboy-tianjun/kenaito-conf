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
		BaseURL: "http://localhost:8080",
		AppName: "kenaito-springboot-demo",
		Env:     "daily",
		Version: "latest",
		//Token:        "6f6966fe196572954b176c0e2a97bc399b050c6be31454b8d6b0f7f622b5d6f1",
		SyncInterval: 10, // sync every 10 seconds
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
	dbHost := c.Get("database.host")
	fmt.Printf("\ndatabase.host = %s\n", dbHost)

	// keep running until signal
	fmt.Println("\nListening for config changes... (press Ctrl+C to stop)")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
}
