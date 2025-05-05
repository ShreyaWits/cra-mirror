package services

import (
	"context"
	claimErrorResponse "nps-reciept-service/common"
	"nps-reciept-service/internal/utils"
	"regexp"

	pb "nps-reciept-service/proto"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}
type ClaimServer struct {
	pb.UnimplementedClaimServiceServer
	redisClient RedisClient
}

func NewClaimServer(client RedisClient) *ClaimServer {
	return &ClaimServer{
		redisClient: client,
	}
}

func (s *ClaimServer) ProcessClaim(ctx context.Context, req *pb.ClaimRequest) (*pb.ClaimResponse, error) {
	// Validate PRAN
	if len(req.Pran) != 12 {
		utils.LogWarning("Validation failed: invalid PRAN", map[string]interface{}{
			"pran": req.Pran,
		})
		return nil, claimErrorResponse.SendError("CLM0001") // custom error code
	}

	// Validate date_of_claim
	date, err := time.Parse("2006-01-02", req.DateOfClaim)
	if err != nil {
		utils.LogError("Invalid date format", err, map[string]interface{}{
			"date_of_claim": req.DateOfClaim,
		})
		return nil, claimErrorResponse.SendError("CLM0004")
	}

	// 💥 New validation: TransactionType must be non-numeric
	isNumeric := regexp.MustCompile(`^\d+$`).MatchString
	if req.TransactionType != "" && isNumeric(req.TransactionType) {
		return nil, claimErrorResponse.SendError("CLM0008")
	}

	last4 := req.Pran[len(req.Pran)-4:]
	datePart := date.Format("060102")

	key := "claim:sequence:" + datePart 
	utils.LogInfo("Generating sequence key", map[string]interface{}{
		"key": key,
	})

	redisClient := s.redisClient

	// Increment the sequence number in Redis
	sequence, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		utils.LogError("Failed to increment Redis sequence", err, map[string]interface{}{
			"key": key,
		})
		return nil, claimErrorResponse.SendError("CLM0005")
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
