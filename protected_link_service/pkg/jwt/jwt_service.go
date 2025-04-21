package jwt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	configEnv "protected_link/internal/configs"
	apiDtos "protected_link/internal/module/apis/dtos"
	"strings"
	"time"
)

type JwtCreation struct {
	JWTSecret []byte
}

func NewJwtCreation() (*JwtCreation, error) {
	cfg, err := configEnv.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	secret := strings.TrimSpace(cfg.JWTSecret) // in case there are spaces/newlines

	if len(secret) != 32 {
		return nil, fmt.Errorf("JWT secret must be exactly 32 bytes long for AES-256, got %d bytes", len(secret))
	}

	return &JwtCreation{
		JWTSecret: []byte(secret),
	}, nil
}

// Encrypt encodes any struct into an encrypted URL-safe string
func parseExpiration(expireIn string) (int64, error) {
	duration, err := time.ParseDuration(expireIn)
	if err != nil {
		return 0, fmt.Errorf("invalid expire_in format: %w", err)
	}
	return time.Now().Add(duration).Unix(), nil
}

func (j *JwtCreation) Encrypt(data *apiDtos.GenerateUrlRequest) (string, error) {

	expireStr := data.ExpireIn

	expiresAt, err := parseExpiration(expireStr)
	if err != nil {
		return "", err
	}

	// Wrap the data with expiration
	payload := apiDtos.SecurePayload{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	block, err := aes.NewCipher(j.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, jsonData, nil)
	encoded := base64.RawURLEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

// Decrypt decrypts the ciphertext and returns raw JSON bytes
func (j *JwtCreation) Decrypt(token string) ([]byte, error) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 token: %w", err)
	}

	block, err := aes.NewCipher(j.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// ✅ Decrypt the actual payload
	plainText, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	var payload apiDtos.SecurePayload
	// ✅ Unmarshal into temporary struct to validate time
	if err := json.Unmarshal(plainText, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted payload: %w", err)
	}

	// ✅ Convert ExpiresAt to time.Time and check expiration
	expirationTime := time.Unix(payload.ExpiresAt, 0)
	if time.Now().After(expirationTime) {
		return nil, fmt.Errorf("⏰ link has expired")
	}

	// Return decrypted plain JSON bytes
	return plainText, nil
}
