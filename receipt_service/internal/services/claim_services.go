package services

import (
	"context"
	"time"

	claimErrorResponse "nps-reciept-service/common"
	"nps-reciept-service/internal/repositories"
	"nps-reciept-service/internal/utils"
	"nps-reciept-service/pkg/observability"
	pb "nps-reciept-service/proto"

	"go.opentelemetry.io/otel/codes"
)

type ClaimServer struct {
	pb.UnimplementedClaimServiceServer
	repo          repositories.ClaimRepository
	observability *observability.ObservabilityStack
}

func NewClaimServer(repo repositories.ClaimRepository, obs *observability.ObservabilityStack) *ClaimServer {
	return &ClaimServer{
		repo:          repo,
		observability: obs,
	}
}

func (s *ClaimServer) ProcessClaim(ctx context.Context, req *pb.ClaimRequest) (*pb.ClaimResponse, error) {
	ctx, span := s.observability.TracerService.StartTracer(ctx, "ProcessClaim")
	defer s.observability.TracerService.StopSpan(span)

	last4 := req.Pran[len(req.Pran)-4:]
	date, _ := time.Parse("2006-01-02", req.DateOfClaim)
	datePart := date.Format("060102")

	sequence, err := s.repo.IncrementClaimSequence(ctx, datePart)
	if err != nil {
		s.observability.TracerService.RecordError(span, err)
		s.observability.TracerService.SetStatus(span, codes.Error, "Failed to increment claim sequence")
		s.observability.MetricsService.IncrementCounter(ctx, "claim_sequence_failed", 1, map[string]string{"status": "error"})
		return nil, claimErrorResponse.SendError("CLM0005")
	}

	if sequence == 1 {
		expiry := 24 * time.Hour
		err = s.repo.SetClaimSequenceExpiry(ctx, datePart, expiry)
		if err != nil {
			s.observability.TracerService.RecordError(span, err)
			s.observability.TracerService.SetAttributes(span, map[string]string{"warn": "expiry_set_failed"})
		}
	}

	claimID := "CLM" + datePart + last4 + formatSequence(sequence)
	s.observability.TracerService.SetAttributes(span, map[string]string{"claim_id": claimID})
	s.observability.MetricsService.IncrementCounter(ctx, "claim_id_generated", 1, map[string]string{"status": "success"})

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
