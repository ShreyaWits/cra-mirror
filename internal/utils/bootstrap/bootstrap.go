package bootstrap

import (
	"log"
	"nps-reciept-service/config"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	grpcServer "nps-reciept-service/internal/grpc"
)

// Bootstrap initializes env, Redis and returns the gRPC server
func BootstrapServices() *grpcServer.GrpcServer {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
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