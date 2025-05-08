package services

import (
	"context"
	"errors"
	pb "nps-reciept-service/proto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClaimRepository is a mock implementation of the ClaimRepository interface
type MockClaimRepository struct {
	mock.Mock
}

func (m *MockClaimRepository) IncrementClaimSequence(ctx context.Context, datePart string) (int64, error) {
	args := m.Called(ctx, datePart)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClaimRepository) SetClaimSequenceExpiry(ctx context.Context, datePart string, expiry time.Duration) error {
	args := m.Called(ctx, datePart, expiry)
	return args.Error(0)
}

func TestProcessClaim(t *testing.T) {
	tests := []struct {
		name              string
		req               *pb.ClaimRequest
		mockRepoSetup     func(*MockClaimRepository)
		expectedClaimID   string
		wantErr           bool
		expectedErrorCode string
	}{
		{
			name: "Successful Claim Processing",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
				TransactionType: "debit",
			},
			mockRepoSetup: func(m *MockClaimRepository) {
				m.On("IncrementClaimSequence", mock.Anything, "240424").Return(int64(1), nil).Once()
				m.On("SetClaimSequenceExpiry", mock.Anything, "240424", 24*time.Hour).Return(nil).Once()
			},
			expectedClaimID: "CLM24042490120001",
			wantErr:         false,
		},
		{
			name: "Successful Claim Processing - Sequence > 1",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
				TransactionType: "credit",
			},
			mockRepoSetup: func(m *MockClaimRepository) {
				m.On("IncrementClaimSequence", mock.Anything, "240424").Return(int64(15), nil).Once()
				// SetClaimSequenceExpiry should not be called for sequence > 1
			},
			expectedClaimID: "CLM24042490120015",
			wantErr:         false,
		},
		{
			name: "Increment Claim Sequence Error",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
				TransactionType: "debit",
			},
			mockRepoSetup: func(m *MockClaimRepository) {
				m.On("IncrementClaimSequence", mock.Anything, "240424").Return(int64(0), errors.New("redis error")).Once()
			},
			expectedClaimID:   "",
			wantErr:           true,
			expectedErrorCode: "CLM0005",
		},
		{
			name: "Set Claim Sequence Expiry Error",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
				TransactionType: "debit",
			},
			mockRepoSetup: func(m *MockClaimRepository) {
				m.On("IncrementClaimSequence", mock.Anything, "240424").Return(int64(1), nil).Once()
				m.On("SetClaimSequenceExpiry", mock.Anything, "240424", 24*time.Hour).Return(errors.New("redis expiry error")).Once()
			},
			expectedClaimID:   "CLM24042490120001", // Claim ID is still generated
			wantErr:           false, // Based on the comment in the code, this might not return an error
			expectedErrorCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockClaimRepository)
			if tt.mockRepoSetup != nil {
				tt.mockRepoSetup(mockRepo)
			}

			server := NewClaimServer(mockRepo)

			resp, err := server.ProcessClaim(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				// You might want to add more specific error checking here
				// based on the actual error structure returned by SendError
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, "claim-id successfully generated", resp.Message)
				assert.NotNil(t, resp.Data)
				assert.Equal(t, tt.expectedClaimID, resp.Data.ClaimId)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}