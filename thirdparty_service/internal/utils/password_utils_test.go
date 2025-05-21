package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword_Success(t *testing.T) {
	password := "My$ecureP@ss123"
	hashed, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.NotEqual(t, password, hashed)
}

func TestHashPassword_Error(t *testing.T) {
	// Backup the real generator
	originalGen := hashPasswordGenerator

	// Inject mock failure
	hashPasswordGenerator = func([]byte, int) ([]byte, error) {
		return nil, errors.New("bcrypt fail")
	}

	defer func() {
		hashPasswordGenerator = originalGen
	}()

	hash, err := HashPassword("any")
	assert.Error(t, err)
	assert.Equal(t, ErrHashingPassword, err)
	assert.Empty(t, hash)
}

func TestValidatePassword_Correct(t *testing.T) {
	password := "Secret123!"
	hashed, _ := HashPassword(password)
	assert.True(t, ValidatePassword(hashed, password))
}

func TestValidatePassword_Incorrect(t *testing.T) {
	password := "Secret123!"
	hashed, _ := HashPassword(password)
	assert.False(t, ValidatePassword(hashed, "WrongOne"))
}
