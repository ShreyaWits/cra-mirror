package encryptionengine

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/pkg/crypto"
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/observability"

	otelcodes "go.opentelemetry.io/otel/codes"
)

type EncryptionEngineImpl struct {
	keyManager keymanager.KeyManager
	crypto     crypto.Encrypter
	obs        *observability.ObservabilityStack
}

func NewEncryptionEngine(
	keyManager keymanager.KeyManager,
	crypto crypto.Encrypter,
	obs *observability.ObservabilityStack,
) EncryptionEngine {
	return &EncryptionEngineImpl{
		keyManager: keyManager,
		crypto:     crypto,
		obs:        obs,
	}
}

func (e *EncryptionEngineImpl) GenerateEncryptionKey(ctx context.Context) ([]byte, *errors.CustomError) {
	functionName := "GenerateEncryptionKey"

	ctx, span := e.obs.TracerService.StartTracer(ctx, functionName)
	defer e.obs.TracerService.StopSpan(span)

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Generating key", requestID))
	startTime := time.Now()

	dek := make([]byte, 32) // 256 bits for AES-*errors.CustomError
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		e.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Key generation failed", requestID))
		e.obs.TracerService.SetStatus(span, otelcodes.Error, "Key generation failed")
		e.obs.MetricsService.IncrementCounter(ctx, "operation_failed", 1, nil)
		return nil, errors.NewCustomError(errors.ENGErrGenerateDEK, err)
	}

	processingTime := time.Since(startTime).Milliseconds()
	e.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)
	e.obs.MetricsService.IncrementCounter(ctx, "operation_success", 1, nil)
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	e.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return dek, nil
}

func (e *EncryptionEngineImpl) Encrypt(ctx context.Context, data string, dek []byte) (string, *errors.CustomError) {
	functionName := "Encrypt"

	ctx, span := e.obs.TracerService.StartTracer(ctx, functionName)
	defer e.obs.TracerService.StopSpan(span)

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Encrypting data", requestID))
	startTime := time.Now()

	// Encrypt the data with DEK
	encryptedData, customErr := e.crypto.Encrypt(data, dek)
	if customErr != nil {
		e.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Encryption failed", requestID))
		e.obs.TracerService.SetStatus(span, otelcodes.Error, "Encryption failed")
		e.obs.MetricsService.IncrementCounter(ctx, "operation_failed", 1, nil)
		return "", customErr
	}

	processingTime := time.Since(startTime).Milliseconds()
	e.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)
	e.obs.MetricsService.IncrementCounter(ctx, "operation_success", 1, nil)
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	e.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return encryptedData, nil
}

func (e *EncryptionEngineImpl) Decrypt(ctx context.Context, encryptedData string, dek []byte) (string, *errors.CustomError) {
	functionName := "Decrypt"

	ctx, span := e.obs.TracerService.StartTracer(ctx, functionName)
	defer e.obs.TracerService.StopSpan(span)

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Decrypting data", requestID))
	startTime := time.Now()

	// Decrypt the data with DEK
	decryptedData, customErr := e.crypto.Decrypt(encryptedData, dek)
	if customErr != nil {
		e.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Decryption failed", requestID))
		e.obs.TracerService.SetStatus(span, otelcodes.Error, "Decryption failed")
		e.obs.MetricsService.IncrementCounter(ctx, "operation_failed", 1, nil)
		return "", customErr
	}

	processingTime := time.Since(startTime).Milliseconds()
	e.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)
	e.obs.MetricsService.IncrementCounter(ctx, "operation_success", 1, nil)
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	e.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return decryptedData, nil
}

func (e *EncryptionEngineImpl) EncryptDEK(ctx context.Context, dek []byte, kek []byte) (string, *errors.CustomError) {
	functionName := "EncryptDEK"

	ctx, span := e.obs.TracerService.StartTracer(ctx, functionName)
	defer e.obs.TracerService.StopSpan(span)

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Encrypting DEK", requestID))
	startTime := time.Now()

	// Create AES cipher with KEK
	edek, customErr := e.crypto.EncryptBytes(dek, kek)
	if customErr != nil {
		e.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: DEK encryption failed", requestID))
		e.obs.TracerService.SetStatus(span, otelcodes.Error, "DEK encryption failed")
		e.obs.MetricsService.IncrementCounter(ctx, "operation_failed", 1, nil)
		return "", customErr
	}

	// Encode the encrypted DEK
	edekString := base64.StdEncoding.EncodeToString(edek)

	processingTime := time.Since(startTime).Milliseconds()
	e.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)
	e.obs.MetricsService.IncrementCounter(ctx, "operation_success", 1, nil)
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	e.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return edekString, nil
}

func (e *EncryptionEngineImpl) DecryptDEK(ctx context.Context, edek string, kek []byte) ([]byte, *errors.CustomError) {
	functionName := "DecryptDEK"

	ctx, span := e.obs.TracerService.StartTracer(ctx, functionName)
	defer e.obs.TracerService.StopSpan(span)

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STEP: Decrypting DEK", requestID))
	startTime := time.Now()

	// Decode the encrypted DEK
	decryptedEDEK, err := base64.StdEncoding.DecodeString(edek)
	if err != nil {
		e.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: Base64 decoding failed", requestID))
		e.obs.TracerService.SetStatus(span, otelcodes.Error, "Base64 decoding failed")
		e.obs.MetricsService.IncrementCounter(ctx, "operation_failed", 1, nil)
		return nil, errors.NewCustomError(errors.ENGErrDecryptDEK, err)
	}

	// Decrypt the DEK with KEK
	decryptedData, customErr := e.crypto.DecryptBytes(decryptedEDEK, kek)
	if customErr != nil {
		e.obs.LoggerService.Error(ctx, fmt.Sprintf("[%s] STATUS: DEK decryption failed", requestID))
		e.obs.TracerService.SetStatus(span, otelcodes.Error, "DEK decryption failed")
		e.obs.MetricsService.IncrementCounter(ctx, "operation_failed", 1, nil)
		return nil, customErr
	}

	processingTime := time.Since(startTime).Milliseconds()
	e.obs.MetricsService.RecordHistogram(ctx, "processing_time_ms", float64(processingTime), nil)
	e.obs.MetricsService.IncrementCounter(ctx, "operation_success", 1, nil)
	e.obs.LoggerService.Info(ctx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	e.obs.TracerService.SetStatus(span, otelcodes.Ok, "Success")

	return decryptedData, nil
}
