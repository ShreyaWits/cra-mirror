package main

import (
	"context"
	"log"
	"nps-reciept-service/internal/utils/bootstrap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := bootstrap.BootstrapServices()

	// Use observability logger if available
	server.Observability.LoggerService.Info("Starting gRPC server", nil)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		sig := <-sigChan

		server.Observability.LoggerService.Info("Received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})

		// Stop server
		server.Stop()

		// Shutdown observability providers (TracerProvider, MeterProvider)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if server.Observability.TracerService != nil {
			if err := server.Observability.TracerService.Shutdown(ctx); err != nil {
				log.Printf("Tracer shutdown error: %v", err)
			}
		}

		if server.Observability.MetricsService != nil {
			if err := server.Observability.MetricsService.Shutdown(ctx); err != nil {
				log.Printf("Metrics shutdown error: %v", err)
			}
		}
	}()

	// Start the server
	if err := server.Start(); err != nil {
		server.Observability.LoggerService.Error("Server failed to start", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("failed to serve: %v", err)
	}
}
