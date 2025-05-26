package encryptionengine_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"encryption_microservice/internal/common/mocks"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	pkgErr "encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/observability"
)

func TestEncryptionEngine_GenerateEncryptionKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create real observability stack with nil env (disabled telemetry)
	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, nil, obs)

	key, err := engine.GenerateEncryptionKey(ctx)
	require.Nil(t, err)
	require.Len(t, key, 32)
}

func TestEncryptionEngine_Encrypt_Decrypt_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, encMock, obs)

	dek := make([]byte, 32)
	rand.Read(dek) // valid key

	plain := "hello"
	cipher := "cipher‑text"

	// expectations
	encMock.EXPECT().Encrypt(plain, dek).Return(cipher, nil)
	encMock.EXPECT().Decrypt(cipher, dek).Return(plain, nil)

	outCipher, err := engine.Encrypt(ctx, plain, dek)
	require.Nil(t, err)
	require.Equal(t, cipher, outCipher)

	outPlain, err := engine.Decrypt(ctx, cipher, dek)
	require.Nil(t, err)
	require.Equal(t, plain, outPlain)
}

func TestEncryptionEngine_Encrypt_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, encMock, obs)

	dek := bytes.Repeat([]byte{'k'}, 32)
	encMock.EXPECT().
		Encrypt("data", dek).
		Return("", pkgErr.NewCustomError(pkgErr.CRYPerrEncryptData, errors.New("fail")))

	_, err := engine.Encrypt(ctx, "data", dek)
	require.Error(t, err)
}

func TestEncryptionEngine_EncryptDEK_DecryptDEK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, encMock, obs)

	dek := bytes.Repeat([]byte{'d'}, 32)
	kek := bytes.Repeat([]byte{'k'}, 32)
	encryptedDEK := []byte("encrypted‑dek")

	encMock.EXPECT().EncryptBytes(dek, kek).Return(encryptedDEK, nil)
	encMock.EXPECT().DecryptBytes(encryptedDEK, kek).Return(dek, nil)

	edekStr, err := engine.EncryptDEK(ctx, dek, kek)
	require.Nil(t, err)
	require.Equal(t, base64.StdEncoding.EncodeToString(encryptedDEK), edekStr)

	out, err := engine.DecryptDEK(ctx, edekStr, kek)
	require.Nil(t, err)
	require.Equal(t, dek, out)
}

func TestEncryptionEngine_DecryptDEK_BadBase64(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, nil, obs)

	_, err := engine.DecryptDEK(ctx, "##not‑b64##", nil)
	require.Error(t, err)
}

func TestEncryptionEngine_Encrypt_Decrypt_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, encMock, obs)

	dek := make([]byte, 32)
	rand.Read(dek) // valid key

	plain := "hello"
	cipher := "cipher‑text"

	// expectations
	encMock.EXPECT().Encrypt(plain, dek).Return("", pkgErr.NewCustomError(pkgErr.CRYPerrEncryptData, errors.New("fail")))
	encMock.EXPECT().Decrypt(cipher, dek).Return(plain, pkgErr.NewCustomError(pkgErr.CRYPerrDecryptData, errors.New("fail")))

	outCipher, err := engine.Encrypt(ctx, plain, dek)
	require.NotNil(t, err)
	require.Empty(t, outCipher)

	outPlain, err := engine.Decrypt(ctx, cipher, dek)
	require.NotNil(t, err)
	require.Empty(t, outPlain)
}

func TestEncryptionEngine_EncryptDEK_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, encMock, obs)

	dek := bytes.Repeat([]byte{'d'}, 32)
	kek := bytes.Repeat([]byte{'k'}, 32)
	encMock.EXPECT().EncryptBytes(dek, kek).Return([]byte(""), pkgErr.NewCustomError(pkgErr.CRYPerrDecryptData, errors.New("fail")))

	edekStr, err := engine.EncryptDEK(ctx, dek, kek)
	require.NotNil(t, err)
	require.Empty(t, edekStr)
}

func TestEncryptionEngine_EncryptDEK_DecryptDEK_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	obs := observability.NewObservabilityStack(nil)
	ctx := context.Background()

	engine := encryptionengine.NewEncryptionEngine(nil, encMock, obs)

	dek := bytes.Repeat([]byte{'d'}, 32)
	kek := bytes.Repeat([]byte{'k'}, 32)
	encryptedDEK := []byte("encrypted‑dek")

	encMock.EXPECT().EncryptBytes(dek, kek).Return(encryptedDEK, nil)
	encMock.EXPECT().DecryptBytes(encryptedDEK, kek).Return([]byte(""), pkgErr.NewCustomError(pkgErr.CRYPerrDecryptData, errors.New("fail")))

	edekStr, err := engine.EncryptDEK(ctx, dek, kek)
	require.Nil(t, err)
	require.Equal(t, base64.StdEncoding.EncodeToString(encryptedDEK), edekStr)

	out, err := engine.DecryptDEK(ctx, edekStr, kek)
	require.NotNil(t, err)
	require.Empty(t, out)
}
