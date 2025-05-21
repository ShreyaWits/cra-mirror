//go:build !test
// build +test
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"Document-Processing/internal/config"
	"Document-Processing/internal/di"

	"github.com/joho/godotenv"
)

func startServer(cfg *config.Config) (*di.Container, error) {
	container, err := di.NewContainer(cfg)
	if err != nil {
		return nil, err
	}

	// Start server in a goroutine
	go func() {
		if err := container.Server.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	return container, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Failed to get .env")
	}

	// Load environment variables
	token := os.Getenv("JWT_TOKEN")

	rawData, err := config.LoadConfigFromAPI(token)
	if err != nil {
		log.Fatalf("Failed to fetch config from API: %v", err)
	}

	cfg, err := config.NewConfig(rawData)
	if err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	// Set initial config
	config.SetCurrentConfig(cfg)

	// Start initial server
	container, err := startServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Set up config webhook endpoint
	http.HandleFunc("/config/webhook", config.HandleConfigWebhook)
	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("Webhook server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Listen for config changes
	configChangeChan := config.GetConfigChangeChan()

	for {
		select {
		case <-sigChan:
			log.Println("Received shutdown signal")
			if err := container.Close(); err != nil {
				log.Printf("Error during cleanup: %v", err)
			}
			log.Println("Server shutdown complete")
			return

		case <-configChangeChan:
			log.Println("Received config change notification")

			// Get new config
			newCfg := config.GetCurrentConfig()
			if newCfg == nil {
				log.Println("No new config available")
				continue
			}

			// Close existing container
			if err := container.Close(); err != nil {
				log.Printf("Error closing existing container: %v", err)
			}

			// Start new server with new config
			newContainer, err := startServer(newCfg)
			if err != nil {
				log.Printf("Failed to start new server: %v", err)
				continue
			}

			container = newContainer
			log.Println("Server restarted with new configuration")
		}
	}
}
