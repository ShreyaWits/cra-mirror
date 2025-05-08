package usecases_test

import (
	"context"
	"testing"

	"encryption_microservice/internal/common/mocks"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	enums "encryption_microservice/internal/modules/encryption/api/enum"
	usecases "encryption_microservice/internal/modules/encryption/usecases/encryption"
	"encryption_microservice/pkg/errors"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEncryptionUsecase_GenerateEDEK(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		mockSetup    func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine)
		expectError  bool
		expectedPriv string
		expectedPub  string
	}{
		{
			name:   "successfully generates EDEK",
			userID: "user1",
			mockSetup: func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine) {
				dek1 := []byte("dek1")
				kek1 := []byte("kek1")
				dek2 := []byte("dek2")
				kek2 := []byte("kek2")

				gomock.InOrder(
					eng.EXPECT().GenerateEncryptionKey().Return(dek1, nil),
					eng.EXPECT().GenerateEncryptionKey().Return(kek1, nil),
					km.EXPECT().StoreKEK("private_user1", kek1).Return(nil),
					eng.EXPECT().EncryptDEK(dek1, kek1).Return("edekPriv", nil),

					eng.EXPECT().GenerateEncryptionKey().Return(dek2, nil),
					eng.EXPECT().GenerateEncryptionKey().Return(kek2, nil),
					km.EXPECT().StoreKEK("public_user1", kek2).Return(nil),
					eng.EXPECT().EncryptDEK(dek2, kek2).Return("edekPub", nil),
				)
			},
			expectError:  false,
			expectedPriv: "edekPriv",
			expectedPub:  "edekPub",
		},
		{
			name:   "DEK generation fails",
			userID: "user1",
			mockSetup: func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine) {
				eng.EXPECT().GenerateEncryptionKey().Return(nil, &errors.CustomError{ErrorCode: errors.ENGErrGenerateDEK, Err: assert.AnError})
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockKM := mocks.NewMockKeyManager(ctrl)
			mockEng := mocks.NewMockEncryptionEngine(ctrl)
			tt.mockSetup(mockKM, mockEng)

			uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
			resp, err := uc.GenerateEDEK(context.Background(), tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				if err != nil {
					t.Logf("GenerateEDEK returned error: %v", err)
				}
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedPriv, resp.EDEKPrivate)
				assert.Equal(t, tt.expectedPub, resp.EDEKPublic)
			}
		})
	}
}

func TestEncryptionUsecase_Encrypt(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		edekPrivate   string
		edekPublic    string
		req           dtos.EncryptRequest
		mockSetup     func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine)
		expectError   bool
		expectedField string
	}{
		{
			name:        "success private encryption",
			userID:      "user1",
			edekPrivate: "edekPriv",
			edekPublic:  "edekPub",
			req: dtos.EncryptRequest{
				Data: []map[string]string{{"e_type": string(enums.KeyPrivate), "field": "value"}},
			},
			mockSetup: func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine) {
				kek := []byte("kek")
				dek := []byte("dek")
				km.EXPECT().RetrieveKEK("private_user1").Return(kek, nil)
				eng.EXPECT().DecryptDEK("edekPriv", kek).Return(dek, nil)
				eng.EXPECT().Encrypt("value", dek).Return("encryptedValue", nil)
			},
			expectError:   false,
			expectedField: "private_encryptedValue",
		},
		{
			name:        "RetrieveKEK fails",
			userID:      "user1",
			edekPrivate: "edekPriv",
			edekPublic:  "edekPub",
			req: dtos.EncryptRequest{
				Data: []map[string]string{{"e_type": string(enums.KeyPrivate), "field": "value"}},
			},
			mockSetup: func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine) {
				km.EXPECT().RetrieveKEK("private_user1").Return(nil, &errors.CustomError{ErrorCode: errors.ESErrRetrieveKEK, Err: assert.AnError})
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockKM := mocks.NewMockKeyManager(ctrl)
			mockEng := mocks.NewMockEncryptionEngine(ctrl)
			tt.mockSetup(mockKM, mockEng)

			uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
			resp, err := uc.Encrypt(context.Background(), tt.userID, tt.edekPrivate, tt.edekPublic, &tt.req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				if err != nil {
					t.Logf("Encrypt returned error: %v", err)
				}
				assert.Nil(t, err)
				assert.Len(t, resp.Data, 1)
				assert.Equal(t, tt.expectedField, resp.Data[0]["field"])
			}
		})
	}
}

func TestEncryptionUsecase_Decrypt(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		edekPrivate   string
		edekPublic    string
		req           dtos.DecryptRequest
		mockSetup     func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine)
		expectError   bool
		expectedValue string
	}{
		{
			name:        "success private decryption",
			userID:      "user1",
			edekPrivate: "edekPriv",
			edekPublic:  "edekPub",
			req: dtos.DecryptRequest{
				Data: []map[string]string{{"field": string(enums.KeyPrivate) + "_cipher"}},
			},
			mockSetup: func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine) {
				kek := []byte("kek")
				dek := []byte("dek")
				km.EXPECT().RetrieveKEK("private_user1").Return(kek, nil)
				eng.EXPECT().DecryptDEK("edekPriv", kek).Return(dek, nil)
				eng.EXPECT().Decrypt("cipher", dek).Return("value", nil)
			},
			expectError:   false,
			expectedValue: "value",
		},
		{
			name:        "DecryptEDEK fails",
			userID:      "user1",
			edekPrivate: "edekPriv",
			edekPublic:  "edekPub",
			req: dtos.DecryptRequest{
				Data: []map[string]string{{"field": string(enums.KeyPrivate) + "_cipher"}},
			},
			mockSetup: func(km *mocks.MockKeyManager, eng *mocks.MockEncryptionEngine) {
				kek := []byte("kek")
				km.EXPECT().RetrieveKEK("private_user1").Return(kek, nil)
				eng.EXPECT().DecryptDEK("edekPriv", kek).Return(nil, &errors.CustomError{ErrorCode: errors.ENGErrDecryptDEK, Err: assert.AnError})
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockKM := mocks.NewMockKeyManager(ctrl)
			mockEng := mocks.NewMockEncryptionEngine(ctrl)
			tt.mockSetup(mockKM, mockEng)

			uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
			resp, err := uc.Decrypt(context.Background(), tt.userID, tt.edekPrivate, tt.edekPublic, &tt.req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				if err != nil {
					t.Logf("Decrypt returned error: %v", err)
				}
				assert.Nil(t, err)
				assert.Len(t, resp.Data, 1)
				assert.Equal(t, tt.expectedValue, resp.Data[0]["field"])
			}
		})
	}
}

// Test public encryption branch
func TestEncryptionUsecase_Encrypt_PublicBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockKM := mocks.NewMockKeyManager(ctrl)
	mockEng := mocks.NewMockEncryptionEngine(ctrl)
	// Setup: public path
	kek := []byte("kekP")
	dek := []byte("dekP")
	mockKM.EXPECT().RetrieveKEK("public_user1").Return(kek, nil)
	mockEng.EXPECT().DecryptDEK("edekPub", kek).Return(dek, nil)
	mockEng.EXPECT().Encrypt("valuePub", dek).Return("encryptedPub", nil)

	uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
	req := dtos.EncryptRequest{Data: []map[string]string{{"e_type": string(enums.KeyPublic), "field": "valuePub"}}}
	resp, err := uc.Encrypt(context.Background(), "user1", "edekPriv", "edekPub", &req)
	assert.Nil(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "public_encryptedPub", resp.Data[0]["field"])
}

// Test invalid e_type error
func TestEncryptionUsecase_Encrypt_InvalidEType(t *testing.T) {
	uc := usecases.NewEncryptionUseCase(nil, nil)
	req := dtos.EncryptRequest{Data: []map[string]string{{"field": "valueNoType"}}}
	_, err := uc.Encrypt(context.Background(), "user1", "", "", &req)
	assert.Error(t, err)
	assert.Equal(t, errors.ESErrETYPEKeyMissing, err.ErrorCode)
}

// Test encryption engine error
func TestEncryptionUsecase_Encrypt_EncryptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockKM := mocks.NewMockKeyManager(ctrl)
	mockEng := mocks.NewMockEncryptionEngine(ctrl)
	kek := []byte("kekE")
	dek := []byte("dekE")
	mockKM.EXPECT().RetrieveKEK("private_user1").Return(kek, nil)
	mockEng.EXPECT().DecryptDEK("edekPriv", kek).Return(dek, nil)
	mockEng.EXPECT().Encrypt("value", dek).Return("", &errors.CustomError{ErrorCode: errors.ENGErrEncryptData, Err: assert.AnError})

	uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
	req := dtos.EncryptRequest{Data: []map[string]string{{"e_type": string(enums.KeyPrivate), "field": "value"}}}
	_, err := uc.Encrypt(context.Background(), "user1", "edekPriv", "edekPub", &req)
	assert.Error(t, err)
}

// Test public decryption branch
func TestEncryptionUsecase_Decrypt_PublicBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockKM := mocks.NewMockKeyManager(ctrl)
	mockEng := mocks.NewMockEncryptionEngine(ctrl)
	kek := []byte("kekD")
	dek := []byte("dekD")
	mockKM.EXPECT().RetrieveKEK("public_user1").Return(kek, nil)
	mockEng.EXPECT().DecryptDEK("edekPub", kek).Return(dek, nil)
	mockEng.EXPECT().Decrypt("cipherPub", dek).Return("valuePub", nil)

	uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
	req := dtos.DecryptRequest{Data: []map[string]string{{"field": string(enums.KeyPublic) + "_cipherPub"}}}
	resp, err := uc.Decrypt(context.Background(), "user1", "edekPriv", "edekPub", &req)
	assert.Nil(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "valuePub", resp.Data[0]["field"])
}

// Test skip plain fields in decryption
func TestEncryptionUsecase_Decrypt_SkipPlainFields(t *testing.T) {
	uc := usecases.NewEncryptionUseCase(nil, nil)
	req := dtos.DecryptRequest{Data: []map[string]string{{"field": "plainValue"}}}
	resp, err := uc.Decrypt(context.Background(), "user1", "edekPriv", "edekPub", &req)
	assert.Nil(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "plainValue", resp.Data[0]["field"])
}

// Test KEK store failure in GenerateEDEK
func TestEncryptionUsecase_GenerateEDEK_StoreKEKFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockKM := mocks.NewMockKeyManager(ctrl)
	mockEng := mocks.NewMockEncryptionEngine(ctrl)
	dek1 := []byte("d1")
	kek1 := []byte("k1")
	mockEng.EXPECT().GenerateEncryptionKey().Return(dek1, nil)
	mockEng.EXPECT().GenerateEncryptionKey().Return(kek1, nil)
	mockKM.EXPECT().StoreKEK("private_user1", kek1).Return(&errors.CustomError{ErrorCode: errors.KMGErrStoreKEK, Err: assert.AnError})

	uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
	_, err := uc.GenerateEDEK(context.Background(), "user1")
	assert.Error(t, err)
	assert.Equal(t, errors.KMGErrStoreKEK, err.ErrorCode)
}

// Test EncryptDEK failure in GenerateEDEK
func TestEncryptionUsecase_GenerateEDEK_EncryptDEKFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockKM := mocks.NewMockKeyManager(ctrl)
	mockEng := mocks.NewMockEncryptionEngine(ctrl)
	dek1 := []byte("d2")
	kek1 := []byte("k2")
	mockEng.EXPECT().GenerateEncryptionKey().Return(dek1, nil)
	mockEng.EXPECT().GenerateEncryptionKey().Return(kek1, nil)
	mockKM.EXPECT().StoreKEK("private_user1", kek1).Return(nil)
	mockEng.EXPECT().EncryptDEK(dek1, kek1).Return("", &errors.CustomError{ErrorCode: errors.ENGErrEncryptDEK, Err: assert.AnError})

	uc := usecases.NewEncryptionUseCase(mockKM, mockEng)
	_, err := uc.GenerateEDEK(context.Background(), "user1")
	assert.Error(t, err)
	assert.Equal(t, errors.ENGErrEncryptDEK, err.ErrorCode)
}
