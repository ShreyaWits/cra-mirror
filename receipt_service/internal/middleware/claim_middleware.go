package middleware

import (
	"context"
	claimErrorResponse "nps-reciept-service/common"
	"nps-reciept-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func ValidateClaimRequestInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		claimReq, ok := req.(*proto.ClaimRequest)
		if !ok {
			return nil, claimErrorResponse.SendError("CLM0006", codes.InvalidArgument)
		}

		if claimReq.Pran == "" {
			return nil, claimErrorResponse.SendError("CLM0007", codes.InvalidArgument)
		}
		if claimReq.DateOfClaim == "" {
			return nil, claimErrorResponse.SendError("CLM0002", codes.InvalidArgument)
		}
		if claimReq.TransactionType == "" {
			return nil, claimErrorResponse.SendError("CLM0003", codes.InvalidArgument)
		}

		// All good → call the actual handler
		return handler(ctx, req)
	}
}
