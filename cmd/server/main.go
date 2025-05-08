package main

import (
	"fmt"
	"log"
	"net"
	"template-services/internal/di"
	"template-services/internal/template/routes"
	pb "template-services/proto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/grpc"
)

func main() {
	// Initialize DI container
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize container: %v", err)
	}

	// Get handlers from container
	templateHandler, templateGRPCHandler, err := di.InitHandlers(container)
	if err != nil {
		log.Fatalf("❌ Failed to initialize handlers: %v", err)
	}

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
