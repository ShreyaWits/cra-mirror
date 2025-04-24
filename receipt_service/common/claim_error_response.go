package errorResponse

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorResponse struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.ErrorMsg)
}

var errorMap = map[string]string{
	"CLM0001": "Validation failed: invalid PRAN",
	"CLM0002": "Validation failed: missing date",
	"CLM0003": "Validation failed: missing transaction type",
	"CLM0004": "Invalid date format",
	"CLM0005": "Failed to increment Redis sequence",
	"CLM0006": "Invalid request type",
	"CLM0007": "Validation failed: missing PRAN",
}

func SendError(code string, errorCode ...codes.Code) error {
	msg, exists := errorMap[code]
	if !exists {
		code = "UNKNOWN"
		msg = "Unknown error code"
	}

	errObj := ErrorResponse{
		ErrorCode: code,
		ErrorMsg:  msg,
	}

	jsonBytes, _ := json.Marshal(errObj)

	// Use provided gRPC code if available, otherwise default to Internal
	grpcCode := codes.Internal
	if len(errorCode) > 0 {
		grpcCode = errorCode[0]
	}

	return status.Error(grpcCode, string(jsonBytes))
}
