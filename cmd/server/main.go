package main

import (
	"fmt"
	"log"
	"net"
	configEnv "template-services/internal/configs"
	"template-services/internal/pkg/cache"
	"template-services/internal/pkg/db"
	"template-services/internal/template/handler"
	"template-services/internal/template/repository"
	"template-services/internal/template/routes"
	"template-services/internal/template/service"
	pb "template-services/proto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/grpc"
)

func main() {
	configEnv, err := configEnv.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Initialize dependencies
	yugabyteDB := db.NewYugabyteDB(configEnv.YugabyteDBHost, configEnv.YugabyteDBPort, configEnv.YugabyteDBUser, configEnv.YugabyteDBPassword, configEnv.YugabyteDBName)
	if yugabyteDB == nil {
		log.Fatalf("❌ Failed to connect to YugabyteDB")
	}
	redisCache := cache.NewRedisCache(configEnv.RedisHost, configEnv.RedisPort, configEnv.RedisPassword)

	templateRepository := repository.NewTemplateRepository(yugabyteDB)
	templateService := service.NewTemplateService(templateRepository, redisCache)
	templateHandler := handler.NewTemplateHandler(templateService)
	templateGRPCHandler := handler.NewTemplateGRPCHandler(templateService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Setup routes
	routes.SetupTemplateRoutes(app, templateHandler)

	// Start gRPC Server
	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterTemplateServiceServer(grpcServer, templateGRPCHandler)
		fmt.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()
	// Start HTTP Server
	fmt.Println("HTTP server listening on :8080")
	log.Fatal(app.Listen(":8080"))
}
