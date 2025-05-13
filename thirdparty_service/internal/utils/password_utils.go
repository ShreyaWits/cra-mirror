package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordCost = 12
)

var (
	ErrHashingPassword = fmt.Errorf("failed to hash the password")
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)
	if err != nil {
		return "", ErrHashingPassword
	}
	return string(hash), nil
}

func ValidatePassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
