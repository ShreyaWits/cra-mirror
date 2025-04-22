package services

import (
	"context"
	"nps-reciept-service/internal/config"
	"nps-reciept-service/internal/utils"
	pb "nps-reciept-service/proto"
	"time"
)

type ClaimServer struct {
	pb.UnimplementedClaimServiceServer
}

func NewClaimServer() *ClaimServer {
	return &ClaimServer{}
}

func (s *ClaimServer) ProcessClaim(ctx context.Context, req *pb.ClaimRequest) (*pb.ClaimResponse, error) {
	date, err := time.Parse("2006-01-02", req.DateOfClaim)
	if err != nil {
		utils.LogError("Invalid date format", err, nil)
		return nil, utils.SendError{
			Status:     "error",
			Message:    "Invalid date format",
			ErrContent: err,
		}
	}

	last4 := req.Pran[len(req.Pran)-4:]
	datePart := date.Format("060102")

	// Create a Redis key based only on the date to maintain daily counter
	key := "claim:sequence:" + datePart
	utils.LogInfo("Generating sequence key", map[string]interface{}{
		"key": key,
	})

	redisClient := config.GetRedisClient()

	// Increment the sequence number in Redis
	sequence, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		utils.LogError("Failed to increment Redis sequence", err, map[string]interface{}{
			"key": key,
		})
		return nil, err
	}

	// Set expiry for 24 hours only when the key is newly created
	if sequence == 1 {
		expiry := 24 * time.Hour
		err = redisClient.Expire(ctx, key, expiry).Err()
		if err != nil {
			utils.LogError("Failed to set expiry for Redis key", err, map[string]interface{}{
				"key": key,
			})
			// Optional: decide if you want to return error or continue
		}
	}

	claimID := "CLM" + datePart + last4 + formatSequence(sequence)
	utils.LogInfo("Claim ID generated", map[string]interface{}{
		"claim_id": claimID,
	})

	return &pb.ClaimResponse{
		Status:  "success",
		Message: "claim-id successfully generated",
		Data: &pb.ClaimData{
			ClaimId: claimID,
		},
	}, nil
}

func formatSequence(seq int64) string {
	return utils.PadLeft(int(seq), 4)
}
