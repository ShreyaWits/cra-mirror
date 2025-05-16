package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateOTP generates a 6-digit numeric One-Time Password (OTP).
func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(1_000_000)     // Generates a number between 0 and 999999
	return fmt.Sprintf("%06d", otp) // Pads with leading zeros if needed
}
