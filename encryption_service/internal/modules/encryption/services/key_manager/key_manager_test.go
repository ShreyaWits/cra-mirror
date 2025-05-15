package keymanager_test

import (
	"testing"

	"encryption_microservice/internal/common/mocks"
	keymanager "encryption_microservice/internal/modules/encryption/services/key_manager"
	"encryption_microservice/pkg/errors"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestKeyManager_StoreKEK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mocks.NewMockKmsService(ctrl)
	manager := keymanager.NewKeyManager(mockKMS)

	tests := []struct {
		name    string
		keyID   string
		kek     []byte
		mockErr *errors.CustomError
		wantErr bool
	}{
		{
			name:    "Success",
			keyID:   "test-key-success",
			kek:     []byte("some-kek"),
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "Failure",
			keyID:   "test-key-fail",
			kek:     []byte("some-kek"),
			mockErr: errors.NewCustomError(errors.KMSerrStoreKEK, nil),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockKMS.EXPECT().StoreKEK(tc.keyID, tc.kek).Return(tc.mockErr)

			err := manager.StoreKEK(tc.keyID, tc.kek)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tc.mockErr, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestKeyManager_RetrieveKEK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mocks.NewMockKmsService(ctrl)
	manager := keymanager.NewKeyManager(mockKMS)

	tests := []struct {
		name      string
		keyID     string
		mockKEK   []byte
		mockErr   *errors.CustomError
		expectErr bool
	}{
		{
			name:      "Success",
			keyID:     "retrieve-key-success",
			mockKEK:   []byte("retrieved-kek"),
			mockErr:   nil,
			expectErr: false,
		},
		{
			name:      "Failure",
			keyID:     "retrieve-key-fail",
			mockKEK:   nil,
			mockErr:   errors.NewCustomError(errors.KMSerrRetrieveKEK, nil),
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockKMS.EXPECT().RetrieveKEK(tc.keyID).Return(tc.mockKEK, tc.mockErr)

			kek, err := manager.RetrieveKEK(tc.keyID)
			if tc.expectErr {
				assert.Nil(t, kek)
				assert.Equal(t, tc.mockErr, err)
			} else {
				assert.Equal(t, tc.mockKEK, kek)
				assert.Nil(t, err)
			}
		})
	}
}
