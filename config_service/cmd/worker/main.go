package main

import (
	"fmt"
	"log"
	"nps-config-service/internal/constants"
	workflows "nps-config-service/pkg/temporal"
	"os"

	"time" // Required for rand.Seed

	"math/rand" // Required for random number generation

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: No .env file found. Proceeding without it. Error: %v AND IS_DOCKER not true", err)
		} else {
			log.Println("Loaded .env file")
		}
	}
	// Seed the random number generator once in main
	rand.Seed(time.Now().UnixNano())
	endpoint := os.Getenv("TEMPORAL_SERVER_ENDPOINT")
	//log the endpoint
	fmt.Println("Temporal Server Endpoint:", endpoint)
	// Create a Temporal Client
	c, err := client.Dial(client.Options{
		HostPort: endpoint,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Create a new worker
	// "my-task-queue" is the name of the Task Queue this worker will poll
	w := worker.New(c, constants.SendWebhookTaskQueueName, worker.Options{})

	// Register our Workflow and Activity functions
	// w.RegisterWorkflow(app.SimpleWorkflow)
	w.RegisterWorkflow(workflows.WebhookWorkflow)
	w.RegisterActivity(workflows.SendWebhookActivity)

	// Start the worker and block until it stops
	log.Println("Starting Worker...")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
