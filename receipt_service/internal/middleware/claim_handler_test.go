package middleware

import (
	"context"
	pb "nps-reciept-service/proto"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateClaimRequestInterceptor(t *testing.T) {
	interceptor := ValidateClaimRequestInterceptor()

	// Define test cases
	testCases := []struct {
		name        string
		req         interface{}
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Valid Claim Request",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
				TransactionType: "debit",
			},
			wantErr: false,
		},
		{
			name: "Invalid Claim Request Type",
			req:         "invalid request",
			wantErr:     true,
			expectedErr: `rpc error: code = InvalidArgument desc = {"errorCode":"CLM0006","errorMsg":"Invalid request type"}`,
		},
		{
			name: "Missing PRAN",
			req: &pb.ClaimRequest{
				DateOfClaim:     "2024-04-24",
				TransactionType: "debit",
			},
			wantErr:     true,
			expectedErr: `rpc error: code = InvalidArgument desc = {"errorCode":"CLM0007","errorMsg":"Validation failed: missing PRAN"}`,
		},
		{
			name: "Missing DateOfClaim",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				TransactionType: "debit",
			},
			wantErr:     true,
			expectedErr: `rpc error: code = InvalidArgument desc = {"errorCode":"CLM0002","errorMsg":"Validation failed: missing date"}`,
		},
		{
			name: "Missing TransactionType",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
			},
			wantErr:     true,
			expectedErr: `rpc error: code = InvalidArgument desc = {"errorCode":"CLM0003","errorMsg":"Validation failed: missing transaction type"}`,
		},
		{
			name: "Invalid PRAN Format",
			req: &pb.ClaimRequest{
				Pran:            "123", // Invalid PRAN
				DateOfClaim:     "2024-04-24",
				TransactionType: "debit",
			},
			wantErr:     true,
			expectedErr: `rpc error: code = InvalidArgument desc = {"errorCode":"CLM0001","errorMsg":"Validation failed: invalid PRAN"}`,
		},
		{
			name: "Invalid DateOfClaim Format",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "24-04-2024", // Invalid format
				TransactionType: "debit",
			},
			wantErr:     true,
			expectedErr: `rpc error: code = InvalidArgument desc = {"errorCode":"CLM0004","errorMsg":"Invalid date format"}`,
		},
		{
			name: "Invalid TransactionType",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
				TransactionType: "invalid_type", // Invalid transaction type
			},
			wantErr:     false,
			expectedErr: ``,
		},
		{
			name: "Numeric TransactionType",
			req: &pb.ClaimRequest{
				Pran:            "123456789012",
				DateOfClaim:     "2024-04-24",
				TransactionType: "123", // Numeric transaction type
			},
			wantErr:     true,
			expectedErr: `rpc error: code = InvalidArgument desc = {"errorCode":"CLM0008","errorMsg":"Validation failed: invalid Transaction type"}`,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock handler
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			}

			// Call the interceptor
			_, err := interceptor(context.Background(), tc.req, &grpc.UnaryServerInfo{}, handler)

			// Check for errors
			if (err != nil) != tc.wantErr {
				t.Fatalf("interceptor() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr && err != nil {
				// Check for custom error code
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("status.FromError(%v) = false; want true", err)
				}
				if st.Code() != codes.InvalidArgument {
					t.Errorf("interceptor() code = %v, expectedCode %v", st.Code(), codes.InvalidArgument)
				}
				if !reflect.DeepEqual(err.Error(), tc.expectedErr) {
					t.Errorf("interceptor() error = %v, expectedErr %v", err.Error(), tc.expectedErr)
				}
			}
		})
	}
}