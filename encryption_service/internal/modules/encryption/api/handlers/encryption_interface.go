package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"

	"github.com/gofiber/fiber/v2"
)

type EncryptionHandler interface {
	Decrypt(ctx context.Context, req *pb.DecryptRequest) (*pb.DecryptResponse, error)
	Encrypt(ctx context.Context, req *pb.EncryptDataRequest) (*pb.EncryptDataResponse, error)
	GenerateEDEK(ctx context.Context, req *pb.GenerateEDEKRequest) (*pb.GenerateEDEKResponse, error)
	HandleHealth(c *fiber.Ctx) error
}
