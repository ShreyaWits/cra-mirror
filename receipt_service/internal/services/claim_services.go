package services

import (
	"context"
	errorResponse "nps-reciept-service/common"
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
	// Validate PRAN
	if req.Pran == "" || len(req.Pran) != 12 {
		utils.LogWarning("Validation failed: invalid PRAN", map[string]interface{}{
			"pran": req.Pran,
		})
		return nil, errorResponse.SendError("CLM0001") // custom error code
	}

	// Missing date
	if req.TransactionType == "" {
		utils.LogWarning("Validation failed: missing date", nil)
		return nil, errorResponse.SendError("CLM0002")
	}

	// Validate transaction_type
	if req.TransactionType == "" {
		utils.LogWarning("Validation failed: missing transaction type", nil)
		return nil, errorResponse.SendError("CLM0003")
	}

	// Validate date_of_claim
	date, err := time.Parse("2006-01-02", req.DateOfClaim)
	if err != nil {
		utils.LogError("Invalid date format", err, map[string]interface{}{
			"date_of_claim": req.DateOfClaim,
		})
		return nil, errorResponse.SendError("CLM0004")
	}

	last4 := req.Pran[len(req.Pran)-4:]
	datePart := date.Format("060102")

	key := "claim:sequence:" + datePart + ":" + last4
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
		return nil, errorResponse.SendError("CLM0005")
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
