package services

import (
	"context"
	claimErrorResponse "nps-reciept-service/common"
	"nps-reciept-service/internal/repositories"
	"nps-reciept-service/internal/utils"
	pb "nps-reciept-service/proto"
	"time"
)
type ClaimServer struct {
	pb.UnimplementedClaimServiceServer
	repo repositories.ClaimRepository
}

func NewClaimServer(repo repositories.ClaimRepository) *ClaimServer {
	return &ClaimServer{
		repo: repo,
	}
}

func (s *ClaimServer) ProcessClaim(ctx context.Context, req *pb.ClaimRequest) (*pb.ClaimResponse, error) {
	last4 := req.Pran[len(req.Pran)-4:]
	date, _ := time.Parse("2006-01-02", req.DateOfClaim)
	datePart := date.Format("060102")

	// Increment the sequence number using the repository
	sequence, err := s.repo.IncrementClaimSequence(ctx, datePart)
	if err != nil {
		utils.LogError("Failed to increment claim sequence", err, map[string]interface{}{
			"date_part": datePart,
		})
		return nil, claimErrorResponse.SendError("CLM0005")
	}

	// Set expiry for 24 hours only when the key is newly created
	if sequence == 1 {
		expiry := 24 * time.Hour
		err = s.repo.SetClaimSequenceExpiry(ctx, datePart, expiry)
		if err != nil {
			utils.LogError("Failed to set expiry for claim sequence key", err, map[string]interface{}{
				"date_part": datePart,
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


