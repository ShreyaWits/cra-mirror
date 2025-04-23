package dtos

type BaseResponse struct {
	Message   string `json:"message" example:"Success"`
	Success   bool   `json:"success" example:"true"`
	Code      int    `json:"code,omitempty" example:"200"`
	ErrorCode string `json:"errorCode,omitempty"`
	Data      any    `json:"data,omitempty"` // Optional data, example omitted since it's dynamic
}

type ErrorResponse struct {
	Message   string `json:"message" example:"Invalid Request"`
	Success   bool   `json:"success" example:"false"`
	ErrorCode string `json:"errorCode,omitempty"`
	Code      int    `json:"code,omitempty" example:"400"` // Optional error code
}
