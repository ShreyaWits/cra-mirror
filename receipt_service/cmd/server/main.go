package main

import (
	"log"
	"nps-reciept-service/internal/utils"
	"nps-reciept-service/internal/utils/bootstrap"
	"os"
	"os/signal"
	"syscall"
)

func main() {

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
