package keymanager

import (
	"context"
	"fmt"
	"time"

	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/kms"
	"encryption_microservice/pkg/observability"

	otelcodes "go.opentelemetry.io/otel/codes"
)

type KeyManagerImpl struct {
	kms kms.KmsService
	obs *observability.ObservabilityStack
}

func NewKeyManager(kms kms.KmsService, obs *observability.ObservabilityStack) KeyManager {
	return &KeyManagerImpl{
		kms: kms,
		obs: obs,
	}
}

func (k *KeyManagerImpl) StoreKEK(ctx context.Context, keyID string, kek []byte) *errors.CustomError {
	functionName := "StoreKEK"

	ctx, span := k.obs.TracerService.StartTracer(ctx, functionName)
	defer k.obs.TracerService.StopSpan(span)

	// Set tracing attributes
	k.obs.TracerService.SetAttributes(span, map[string]string{
		"key_id": keyID,
	})

	k.obs.LoggerService.Info(ctx, fmt.Sprintf("Storing KEK for key ID: %s", keyID))
	startTime := time.Now()

	if err := k.kms.StoreKEK(keyID, kek); err != nil {
		errMsg := fmt.Sprintf("Failed to store KEK: %v", err)
		k.obs.LoggerService.Error(ctx, errMsg)
		k.obs.TracerService.SetStatus(span, otelcodes.Error, errMsg)
		k.obs.MetricsService.IncrementCounter(ctx, "kek_store_failure", 1, map[string]string{
			"key_id": keyID,
		})
		return err
	}

	processingTime := time.Since(startTime).Milliseconds()
	k.obs.MetricsService.RecordHistogram(ctx, "kek_store_time_ms", float64(processingTime), map[string]string{
		"key_id": keyID,
	})
	k.obs.MetricsService.IncrementCounter(ctx, "kek_store_success", 1, nil)
	k.obs.LoggerService.Info(ctx, fmt.Sprintf("KEK stored successfully for key ID %s in %dms", keyID, processingTime))
	k.obs.TracerService.SetStatus(span, otelcodes.Ok, "KEK stored successfully")

	return nil
}

func (k *KeyManagerImpl) RetrieveKEK(ctx context.Context, keyID string) ([]byte, *errors.CustomError) {
	functionName := "RetrieveKEK"

	ctx, span := k.obs.TracerService.StartTracer(ctx, functionName)
	defer k.obs.TracerService.StopSpan(span)

	// Set tracing attributes
	k.obs.TracerService.SetAttributes(span, map[string]string{
		"key_id": keyID,
	})

	k.obs.LoggerService.Info(ctx, fmt.Sprintf("Retrieving KEK for key ID: %s", keyID))
	startTime := time.Now()

	kek, err := k.kms.RetrieveKEK(keyID)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve KEK: %v", err)
		k.obs.LoggerService.Error(ctx, errMsg)
		k.obs.TracerService.SetStatus(span, otelcodes.Error, errMsg)
		k.obs.MetricsService.IncrementCounter(ctx, "kek_retrieve_failure", 1, map[string]string{
			"key_id": keyID,
		})
		return nil, err
	}

	processingTime := time.Since(startTime).Milliseconds()
	k.obs.MetricsService.RecordHistogram(ctx, "kek_retrieve_time_ms", float64(processingTime), map[string]string{
		"key_id": keyID,
	})
	k.obs.MetricsService.IncrementCounter(ctx, "kek_retrieve_success", 1, nil)
	k.obs.LoggerService.Info(ctx, fmt.Sprintf("KEK retrieved successfully for key ID %s in %dms", keyID, processingTime))
	k.obs.TracerService.SetStatus(span, otelcodes.Ok, "KEK retrieved successfully")

	return kek, nil
}
