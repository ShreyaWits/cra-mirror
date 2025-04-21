package services

import (
	"context"
	"nps-reciept-service/config"
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
			Status:  "error",
			Message: "Invalid date format",
			ErrContent:   err,
		}
	}

	last4 := req.Pran[len(req.Pran)-4:]
	datePart := date.Format("060102")

	// Create a unique key for each day and PRAN combination
	key := "claim:sequence:" + datePart + ":" + last4
	utils.LogInfo("Generating sequence key", map[string]interface{}{
		"key": key,
	})

	// Increment the sequence number in Redis
	sequence, err := config.GetRedisClient().Incr(ctx, key).Result()
	if err != nil {
		utils.LogError("Failed to increment Redis sequence", err, map[string]interface{}{
			"key": key,
		})
		return nil, err
	}

	claimID := "CLM" + datePart + last4 + formatSequence(sequence)
	utils.LogInfo("Claim ID generated", map[string]interface{}{
		"claim_id": claimID,
	})

	return &pb.ClaimResponse{
		Status: "success",
		Message: "claim-id successfully generated",
		Data: &pb.ClaimData{
			ClaimId: claimID,
		},
	}, nil
}

func formatSequence(seq int64) string {
	return utils.PadLeft(int(seq), 4)
}
