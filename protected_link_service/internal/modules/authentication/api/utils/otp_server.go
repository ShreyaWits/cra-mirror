package utils

import (
	"fmt"
	"math/rand"
	"time"
)

type OTPService struct{}

func NewOTPService() *OTPService {
	return &OTPService{}
}

func (s *OTPService) GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000)) // 6-digit OTP
}

func (s *OTPService) ValidateOTP(inputOTP, storedOTP string) bool {
	return inputOTP == storedOTP
}
