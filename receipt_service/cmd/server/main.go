package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nps-reciept-service/internal/config"
	cacheclient "nps-reciept-service/internal/config"
	"nps-reciept-service/internal/utils"
	"nps-reciept-service/internal/utils/bootstrap"

	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("error loading environment variables: %v\n", err)
		}
	}

	// Load environment variables
	token := os.Getenv("JWT_TOKEN")
	if token == "" {
		log.Fatalf("JWT_TOKEN is missing in .env")
	}

	rawData, err := config.LoadConfigFromAPI(token)
	if err != nil {
		log.Fatalf("Failed to fetch config from API: %v", err)
	}

	cfg, err := config.NewConfig(rawData)
	if err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}
	// Create cache client
	cacheClient, err := cacheclient.NewRedisClient(cfg.CacheUrl)

	if err != nil {
		fmt.Println("failed to create cache client: %w", err)
	}

	if cacheClient == nil {
		fmt.Println("cache client is nil despite no error")
	}
	// Set initial config
	config.SetCurrentConfig(cfg)

	// Set up config webhook endpoint
	http.HandleFunc("/config/webhook", config.HandleConfigWebhook)
	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("Webhook server error: %v", err)
		}
	}()

	// Bootstrap the services
	server := bootstrap.BootstrapServices()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		utils.LogInfo("Received shutdown signal", nil)
		server.Stop()
	}()

	// Start the server
	if err := server.Start(); err != nil {
		utils.LogError("Server failed", err, nil)
		log.Fatalf("failed to serve: %v", err)
	}
}
