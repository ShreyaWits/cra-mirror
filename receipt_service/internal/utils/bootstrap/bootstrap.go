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
	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDRESS is not set")
	}
	config.InitRedis(redisAddr)

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
