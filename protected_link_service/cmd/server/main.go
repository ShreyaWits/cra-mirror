package main

import (
	"log"

	"protected_link/pkg/initialization"
)

func main() {
	log.Println("🔧 Initializing protected_link service")

	app, err := initialization.InitializeApp()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application: %v", err)
	}

	if err := app.StartServer(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}

	app.WaitForShutdown()
}
