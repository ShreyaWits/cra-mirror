package usecases

import (
	"context"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	enums "encryption_microservice/internal/modules/encryption/api/enum"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/observability"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// EncryptionUseCaseImpl implements the EncryptionUseCase interface
type EncryptionUseCaseImpl struct {
	keyManager       keymanager.KeyManager
	encryptionEngine encryptionengine.EncryptionEngine
	obs              *observability.ObservabilityStack
}

// NewEncryptionUseCase creates a new instance of EncryptionUseCaseImpl
func NewEncryptionUseCase(
	keyManager keymanager.KeyManager,
	encryptionEngine encryptionengine.EncryptionEngine) EncryptionUseCase {
	return &EncryptionUseCaseImpl{
		keyManager:       keyManager,
		encryptionEngine: encryptionEngine,
	}
}

// NewEncryptionUseCaseWithObs creates a new instance with observability
func NewEncryptionUseCaseWithObs(
	keyManager keymanager.KeyManager,
	encryptionEngine encryptionengine.EncryptionEngine,
	obs *observability.ObservabilityStack) EncryptionUseCase {
	return &EncryptionUseCaseImpl{
		keyManager:       keyManager,
		encryptionEngine: encryptionEngine,
		obs:              obs,
	}
}

func (u *EncryptionUseCaseImpl) getDek(ctx context.Context, kekID string, edek string) ([]byte, *errors.CustomError) {
	// Create span if observability is available
	if u.obs != nil {
		var span trace.Span
		ctx, span = u.obs.TracerService.StartTracer(ctx, "EncryptionUseCase.getDek")
		defer u.obs.TracerService.StopSpan(span)

		// Add attributes to span
		u.obs.TracerService.SetAttributes(span, map[string]string{
			"kek_id": kekID,
		})
	}

	// Track operation time if observability is available
	var startTime time.Time
	if u.obs != nil {
		startTime = time.Now()
	}

	kek, err := u.keyManager.RetrieveKEK(ctx, kekID)
	if err != nil {
		if u.obs != nil {
			u.obs.LoggerService.Error(ctx, "Failed to retrieve KEK", "kek_id", kekID, "error", err)
		}
		return nil, err
	}

	dek, err := u.encryptionEngine.DecryptDEK(ctx, edek, kek)
	if err != nil {
		if u.obs != nil {
			u.obs.LoggerService.Error(ctx, "Failed to decrypt DEK", "kek_id", kekID, "error", err)
		}
		return nil, err
	}

	if u.obs != nil {
		// Record processing time
		processingTime := time.Since(startTime).Milliseconds()
		u.obs.MetricsService.RecordHistogram(ctx, "encryption_getDek_time_ms", float64(processingTime), nil)
	}

	return dek, nil
}

// Encrypt implements the encryption use case
func (u *EncryptionUseCaseImpl) Encrypt(ctx context.Context, userID string, edekPrivate, edekPublic string, req *dtos.EncryptRequest) (*dtos.EncryptResponse, *errors.CustomError) {
	// Create span if observability is available
	if u.obs != nil {
		var span trace.Span
		ctx, span = u.obs.TracerService.StartTracer(ctx, "EncryptionUseCase.Encrypt")
		defer u.obs.TracerService.StopSpan(span)

		// Add attributes to span
		u.obs.TracerService.SetAttributes(span, map[string]string{
			"user_id":     userID,
			"items_count": fmt.Sprintf("%d", len(req.Data)),
		})

		// Record metrics
		u.obs.MetricsService.IncrementCounter(ctx, "encryption_encrypt_calls", 1, nil)

		// Log the operation
		u.obs.LoggerService.Debug(ctx, "Starting encryption operation",
			"user_id", userID,
			"items_count", len(req.Data))
	}

	// Track operation time if observability is available
	var startTime time.Time
	if u.obs != nil {
		startTime = time.Now()
	}

	var (
		DEKMap         = make(map[enums.KeyType][]byte)
		retrievedFlags = make(map[enums.KeyType]bool)
		encryptedItems []map[string]string
	)

	for _, item := range req.Data {
		eType, ok := item["e_type"]
		isValid, keyType := enums.IsValidKeyType(eType)
		if !ok || !isValid {
			err := fmt.Errorf("e_type is missing or invalid")
			if u.obs != nil {
				u.obs.LoggerService.Error(ctx, "Invalid e_type", "error", err)
			}
			return nil, errors.NewCustomError(errors.ESErrETYPEKeyMissing, err)
		}

		// Fetch DEK if not already fetched
		if !retrievedFlags[keyType] {
			kekID := keyType.AddKeyTypeIdentifier(userID)
			var err *errors.CustomError

			// Create child span for DEK retrieval
			var dekSpan trace.Span
			var dekCtx context.Context
			if u.obs != nil {
				dekCtx, dekSpan = u.obs.TracerService.StartTracer(ctx, "RetrieveDEK")
				u.obs.TracerService.SetAttributes(dekSpan, map[string]string{
					"key_type": string(keyType),
					"kek_id":   kekID,
				})
			} else {
				dekCtx = ctx
			}

			switch keyType {
			case enums.KeyPrivate:
				DEKMap[keyType], err = u.getDek(dekCtx, kekID, edekPrivate)
			case enums.KeyPublic:
				DEKMap[keyType], err = u.getDek(dekCtx, kekID, edekPublic)
			default:
				err = errors.NewCustomError(errors.ESErrRetrieveKEK, fmt.Errorf("unsupported KMS type: %s", eType))
			}

			if u.obs != nil && dekSpan != nil {
				if err != nil {
					u.obs.TracerService.RecordError(dekSpan, err)
				}
				u.obs.TracerService.StopSpan(dekSpan)
			}

			if err != nil {
				if u.obs != nil {
					u.obs.LoggerService.Error(ctx, "Failed to retrieve DEK",
						"key_type", keyType,
						"kek_id", kekID,
						"error", err)
					u.obs.MetricsService.IncrementCounter(ctx, "encryption_encrypt_errors", 1, map[string]string{
						"error_type": "dek_retrieval",
					})
				}
				return nil, err
			}
			retrievedFlags[keyType] = true
		}

		// Create child span for item encryption
		var itemSpan trace.Span
		var itemCtx context.Context
		if u.obs != nil {
			itemCtx, itemSpan = u.obs.TracerService.StartTracer(ctx, "EncryptItem")
			u.obs.TracerService.SetAttributes(itemSpan, map[string]string{
				"key_type": string(keyType),
			})
		} else {
			itemCtx = ctx
		}

		// Encrypt the item using the proper DEK
		encryptItem, err := u.encryptItemFields(itemCtx, item, DEKMap[keyType], keyType)

		if u.obs != nil && itemSpan != nil {
			if err != nil {
				u.obs.TracerService.RecordError(itemSpan, err)
			}
			u.obs.TracerService.StopSpan(itemSpan)
		}

		if err != nil {
			if u.obs != nil {
				u.obs.LoggerService.Error(ctx, "Failed to encrypt item fields",
					"key_type", keyType,
					"error", err)
				u.obs.MetricsService.IncrementCounter(ctx, "encryption_encrypt_errors", 1, map[string]string{
					"error_type": "item_encryption",
				})
			}
			return nil, err
		}
		encryptedItems = append(encryptedItems, encryptItem)
	}

	if u.obs != nil {
		// Record processing time
		processingTime := time.Since(startTime).Milliseconds()
		u.obs.MetricsService.RecordHistogram(ctx, "encryption_encrypt_time_ms", float64(processingTime), nil)

		// Log completion
		u.obs.LoggerService.Debug(ctx, "Completed encryption operation",
			"user_id", userID,
			"processed_items", len(encryptedItems),
			"duration_ms", processingTime)
	}

	return &dtos.EncryptResponse{Data: encryptedItems}, nil
}

func (u *EncryptionUseCaseImpl) encryptItemFields(ctx context.Context, item map[string]string, dek []byte, prefix enums.KeyType) (map[string]string, *errors.CustomError) {
	// Create span if observability is available
	if u.obs != nil {
		var span trace.Span
		ctx, span = u.obs.TracerService.StartTracer(ctx, "EncryptionUseCase.encryptItemFields")
		defer u.obs.TracerService.StopSpan(span)

		// Add attributes to span
		u.obs.TracerService.SetAttributes(span, map[string]string{
			"key_type":    string(prefix),
			"field_count": fmt.Sprintf("%d", len(item)),
		})
	}

	encryptedItem := make(map[string]string)

	for key, value := range item {
		if key == "e_type" {
			continue // Skip e_type
		}

		// Create child span for field encryption
		var fieldSpan trace.Span
		var fieldCtx context.Context
		if u.obs != nil {
			fieldCtx, fieldSpan = u.obs.TracerService.StartTracer(ctx, "EncryptField")
			u.obs.TracerService.SetAttributes(fieldSpan, map[string]string{
				"field_key": key,
			})
		} else {
			fieldCtx = ctx
		}

		strVal := fmt.Sprintf("%v", value)
		encryptedVal, err := u.encryptionEngine.Encrypt(fieldCtx, strVal, dek)

		if u.obs != nil && fieldSpan != nil {
			if err != nil {
				u.obs.TracerService.RecordError(fieldSpan, err)
			}
			u.obs.TracerService.StopSpan(fieldSpan)
		}

		if err != nil {
			if u.obs != nil {
				u.obs.LoggerService.Error(ctx, "Failed to encrypt field",
					"field_key", key,
					"error", err)
			}
			return nil, err
		}
		encryptedItem[key] = prefix.AddKeyTypeIdentifier(encryptedVal)
	}

	return encryptedItem, nil
}

func (u *EncryptionUseCaseImpl) Decrypt(ctx context.Context, userID string, edekPrivate string, edekPublic string, req *dtos.DecryptRequest) (*dtos.DecryptResponse, *errors.CustomError) {
	// Create span if observability is available
	if u.obs != nil {
		var span trace.Span
		ctx, span = u.obs.TracerService.StartTracer(ctx, "EncryptionUseCase.Decrypt")
		defer u.obs.TracerService.StopSpan(span)

		// Add attributes to span
		u.obs.TracerService.SetAttributes(span, map[string]string{
			"user_id":     userID,
			"items_count": fmt.Sprintf("%d", len(req.Data)),
		})

		// Record metrics
		u.obs.MetricsService.IncrementCounter(ctx, "encryption_decrypt_calls", 1, nil)

		// Log the operation
		u.obs.LoggerService.Debug(ctx, "Starting decryption operation",
			"user_id", userID,
			"items_count", len(req.Data))
	}

	// Track operation time if observability is available
	var startTime time.Time
	if u.obs != nil {
		startTime = time.Now()
	}

	var (
		DEKMap         = make(map[enums.KeyType][]byte)
		retrievedFlags = make(map[enums.KeyType]bool)
		decryptedItems = make([]map[string]string, 0, len(req.Data))
	)

	for _, item := range req.Data {
		decryptedItem := make(map[string]string)

		for key, value := range item {
			strVal := fmt.Sprintf("%v", value)

			// Skip if no KMS identifier
			if !enums.HasKeyTypeIdentifier(strVal) {
				decryptedItem[key] = strVal
				continue
			}

			encryptedString, keyTypePtr, _ := enums.RemoveKeyTypeIdentifier(strVal)
			keyType := *keyTypePtr

			var dek []byte
			var err *errors.CustomError

			// Retrieve DEK if not already cached
			if !retrievedFlags[keyType] {
				kekID := keyType.AddKeyTypeIdentifier(userID)
				edek := edekPrivate
				if keyType == enums.KeyPublic {
					edek = edekPublic
				}

				// Create child span for DEK retrieval
				var dekSpan trace.Span
				var dekCtx context.Context
				if u.obs != nil {
					dekCtx, dekSpan = u.obs.TracerService.StartTracer(ctx, "RetrieveDEK")
					u.obs.TracerService.SetAttributes(dekSpan, map[string]string{
						"key_type": string(keyType),
						"kek_id":   kekID,
					})
				} else {
					dekCtx = ctx
				}

				dek, err = u.getDek(dekCtx, kekID, edek)

				if u.obs != nil && dekSpan != nil {
					if err != nil {
						u.obs.TracerService.RecordError(dekSpan, err)
					}
					u.obs.TracerService.StopSpan(dekSpan)
				}

				if err != nil {
					if u.obs != nil {
						u.obs.LoggerService.Error(ctx, "Failed to retrieve DEK for decryption",
							"key_type", keyType,
							"kek_id", kekID,
							"error", err)
						u.obs.MetricsService.IncrementCounter(ctx, "encryption_decrypt_errors", 1, map[string]string{
							"error_type": "dek_retrieval",
						})
					}
					return nil, err
				}
				DEKMap[keyType] = dek
				retrievedFlags[keyType] = true
			} else {
				dek = DEKMap[keyType]
			}

			// Create child span for field decryption
			var fieldSpan trace.Span
			var fieldCtx context.Context
			if u.obs != nil {
				fieldCtx, fieldSpan = u.obs.TracerService.StartTracer(ctx, "DecryptField")
				u.obs.TracerService.SetAttributes(fieldSpan, map[string]string{
					"field_key": key,
					"key_type":  string(keyType),
				})
			} else {
				fieldCtx = ctx
			}

			// Decrypt field
			plainText, err := u.encryptionEngine.Decrypt(fieldCtx, encryptedString, dek)

			if u.obs != nil && fieldSpan != nil {
				if err != nil {
					u.obs.TracerService.RecordError(fieldSpan, err)
				}
				u.obs.TracerService.StopSpan(fieldSpan)
			}

			if err != nil {
				if u.obs != nil {
					u.obs.LoggerService.Error(ctx, "Failed to decrypt field",
						"field_key", key,
						"key_type", string(keyType),
						"error", err)
					u.obs.MetricsService.IncrementCounter(ctx, "encryption_decrypt_errors", 1, map[string]string{
						"error_type": "field_decryption",
					})
				}
				return nil, err
			}
			decryptedItem[key] = plainText
		}

		decryptedItems = append(decryptedItems, decryptedItem)
	}

	if u.obs != nil {
		// Record processing time
		processingTime := time.Since(startTime).Milliseconds()
		u.obs.MetricsService.RecordHistogram(ctx, "encryption_decrypt_time_ms", float64(processingTime), nil)

		// Log completion
		u.obs.LoggerService.Debug(ctx, "Completed decryption operation",
			"user_id", userID,
			"processed_items", len(decryptedItems),
			"duration_ms", processingTime)
	}

	return &dtos.DecryptResponse{
		Data: decryptedItems,
	}, nil
}

// GenerateEDEK implements the EDEK generation use case
func (u *EncryptionUseCaseImpl) GenerateEDEK(ctx context.Context, userID string) (*dtos.GenerateEDEKResponse, *errors.CustomError) {
	// Create span if observability is available
	if u.obs != nil {
		var span trace.Span
		ctx, span = u.obs.TracerService.StartTracer(ctx, "EncryptionUseCase.GenerateEDEK")
		defer u.obs.TracerService.StopSpan(span)

		// Add attributes to span
		u.obs.TracerService.SetAttributes(span, map[string]string{
			"user_id": userID,
		})

		// Record metrics
		u.obs.MetricsService.IncrementCounter(ctx, "encryption_generateEDEK_calls", 1, nil)

		// Log the operation
		u.obs.LoggerService.Debug(ctx, "Starting EDEK generation", "user_id", userID)
	}

	// Track operation time if observability is available
	var startTime time.Time
	if u.obs != nil {
		startTime = time.Now()
	}

	// Helper function to generate, store KEK and return EDEK
	generateEDEK := func(keyType enums.KeyType) (string, *errors.CustomError) {
		// Create child span for EDEK generation for specific key type
		var edekSpan trace.Span
		var edekCtx context.Context
		if u.obs != nil {
			edekCtx, edekSpan = u.obs.TracerService.StartTracer(ctx, "GenerateEDEKForType")
			u.obs.TracerService.SetAttributes(edekSpan, map[string]string{
				"key_type": string(keyType),
				"user_id":  userID,
			})
			defer u.obs.TracerService.StopSpan(edekSpan)
		} else {
			edekCtx = ctx
		}

		// Generate DEK
		dekGenCtx, dekGenSpan := u.createChildSpan(edekCtx, "GenerateDEK")
		dek, err := u.encryptionEngine.GenerateEncryptionKey(dekGenCtx)
		u.finalizeSpan(dekGenSpan, err)

		if err != nil {
			if u.obs != nil {
				u.obs.LoggerService.Error(edekCtx, "Failed to generate DEK",
					"key_type", keyType,
					"error", err)
				u.obs.MetricsService.IncrementCounter(edekCtx, "encryption_generateEDEK_errors", 1, map[string]string{
					"error_type": "dek_generation",
					"key_type":   string(keyType),
				})
			}
			return "", err
		}

		// Generate KEK
		kekGenCtx, kekGenSpan := u.createChildSpan(edekCtx, "GenerateKEK")
		kek, err := u.encryptionEngine.GenerateEncryptionKey(kekGenCtx)
		u.finalizeSpan(kekGenSpan, err)

		if err != nil {
			if u.obs != nil {
				u.obs.LoggerService.Error(edekCtx, "Failed to generate KEK",
					"key_type", keyType,
					"error", err)
				u.obs.MetricsService.IncrementCounter(edekCtx, "encryption_generateEDEK_errors", 1, map[string]string{
					"error_type": "kek_generation",
					"key_type":   string(keyType),
				})
			}
			return "", err
		}

		// Store KEK in KMS
		kmsKeyName := keyType.AddKeyTypeIdentifier(userID)
		kmsStoreCtx, kmsStoreSpan := u.createChildSpan(edekCtx, "StoreKEK")
		u.obs.TracerService.SetAttributes(kmsStoreSpan, map[string]string{
			"kms_key_name": kmsKeyName,
		})

		err = u.keyManager.StoreKEK(kmsStoreCtx, kmsKeyName, kek)
		u.finalizeSpan(kmsStoreSpan, err)

		if err != nil {
			if u.obs != nil {
				u.obs.LoggerService.Error(edekCtx, "Failed to store KEK",
					"key_type", keyType,
					"kms_key_name", kmsKeyName,
					"error", err)
				u.obs.MetricsService.IncrementCounter(edekCtx, "encryption_generateEDEK_errors", 1, map[string]string{
					"error_type": "kek_storage",
					"key_type":   string(keyType),
				})
			}
			return "", err
		}

		// Encrypt DEK to generate EDEK
		encDekCtx, encDekSpan := u.createChildSpan(edekCtx, "EncryptDEK")
		edek, err := u.encryptionEngine.EncryptDEK(encDekCtx, dek, kek)
		u.finalizeSpan(encDekSpan, err)

		if err != nil {
			if u.obs != nil {
				u.obs.LoggerService.Error(edekCtx, "Failed to encrypt DEK",
					"key_type", keyType,
					"error", err)
				u.obs.MetricsService.IncrementCounter(edekCtx, "encryption_generateEDEK_errors", 1, map[string]string{
					"error_type": "edek_generation",
					"key_type":   string(keyType),
				})
			}
			return "", err
		}

		return edek, nil
	}

	// Generate EDEKs for both private and public
	_, privateSpan := u.createChildSpan(ctx, "GeneratePrivateEDEK")
	edekPrivate, err := generateEDEK(enums.KeyPrivate)
	u.finalizeSpan(privateSpan, err)

	if err != nil {
		if u.obs != nil {
			u.obs.LoggerService.Error(ctx, "Failed to generate private EDEK", "error", err)
		}
		return nil, err
	}

	_, publicSpan := u.createChildSpan(ctx, "GeneratePublicEDEK")
	edekPublic, err := generateEDEK(enums.KeyPublic)
	u.finalizeSpan(publicSpan, err)

	if err != nil {
		if u.obs != nil {
			u.obs.LoggerService.Error(ctx, "Failed to generate public EDEK", "error", err)
		}
		return nil, err
	}

	if u.obs != nil {
		// Record processing time
		processingTime := time.Since(startTime).Milliseconds()
		u.obs.MetricsService.RecordHistogram(ctx, "encryption_generateEDEK_time_ms", float64(processingTime), nil)

		// Log completion
		u.obs.LoggerService.Debug(ctx, "Completed EDEK generation",
			"user_id", userID,
			"duration_ms", processingTime)
	}

	return &dtos.GenerateEDEKResponse{
		EDEKPrivate: edekPrivate,
		EDEKPublic:  edekPublic,
	}, nil
}

// Helper methods for creating and finalizing spans
func (u *EncryptionUseCaseImpl) createChildSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if u.obs != nil {
		newCtx, span := u.obs.TracerService.StartTracer(ctx, name)
		return newCtx, span
	}
	return ctx, nil
}

func (u *EncryptionUseCaseImpl) finalizeSpan(span trace.Span, err error) {
	if u.obs != nil && span != nil {
		if err != nil {
			u.obs.TracerService.RecordError(span, err)
		}
		u.obs.TracerService.StopSpan(span)
	}
}
