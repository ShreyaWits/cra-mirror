package bootstrap

import (
	"log"
	"nps-reciept-service/internal/config"
	"os"
	"strconv"

	grpcServer "nps-reciept-service/internal/grpc"

	"github.com/joho/godotenv"
)

// Bootstrap initializes env, Redis and returns the gRPC server
func BootstrapServices() *grpcServer.GrpcServer {
	// Load .env file
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: No .env file found. Proceeding without it. Error: %v", err)
		} else {
			log.Println("Loaded .env file")
		}
	}

	// Redis setup
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisUsername := os.Getenv("REDIS_USERNAME")
	redisTTL := os.Getenv("RECEIPT_SERVICE_REDIS_TTL")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" {
		log.Fatal("REDIS_HOST is not set")
	}
	if redisPassword == "" {
		log.Fatal("REDIS_PASSWORD is not set")
	}
	if redisUsername == "" {
		log.Fatal("REDIS_USERNAME is not set")
	}
	if redisTTL == "" {
		log.Fatal("RECEIPT_SERVICE_REDIS_TTL is not set")
	}
	if redisPort == "" {
		log.Fatal("REDIS_PORT is not set")
	}

	// Validate Redis port is a valid number
	if _, err := strconv.Atoi(redisPort); err != nil {
		log.Fatalf("Invalid REDIS_PORT: %v", err)
	}

	config.InitRedis(redisHost, redisUsername, redisPassword, redisTTL)

	// gRPC setup
	grpcAddr := os.Getenv("GRPC_PORT")
	if grpcAddr == "" {
		log.Fatal("GRPC_PORT is not set")
	}
	grpcPort, err := strconv.Atoi(grpcAddr)
	if err != nil {
		log.Fatalf("Invalid GRPC_PORT: %v", err)
	}

	return grpcServer.NewGrpcServer(grpcPort)
}
