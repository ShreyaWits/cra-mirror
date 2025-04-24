package encryptionengine_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"encryption_microservice/internal/common/mocks"
	encryptionengine "encryption_microservice/internal/modules/encryption/services/encryption_engine"
	pkgErr "encryption_microservice/pkg/errors"
)

func TestEncryptionEngine_GenerateEncryptionKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	engine := encryptionengine.NewEncryptionEngine(nil, nil)

	key, err := engine.GenerateEncryptionKey()
	require.Nil(t, err)
	require.Len(t, key, 32)
}

func TestEncryptionEngine_Encrypt_Decrypt_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	kekMock := mocks.NewMockKeyManager(ctrl)

	engine := encryptionengine.NewEncryptionEngine(kekMock, encMock)

	dek := make([]byte, 32)
	rand.Read(dek) // valid key

	plain := "hello"
	cipher := "cipher‑text"

	// expectations
	encMock.EXPECT().Encrypt(plain, dek).Return(cipher, nil)
	encMock.EXPECT().Decrypt(cipher, dek).Return(plain, nil)

	outCipher, err := engine.Encrypt(plain, dek)
	require.Nil(t, err)
	require.Equal(t, cipher, outCipher)

	outPlain, err := engine.Decrypt(cipher, dek)
	require.Nil(t, err)
	require.Equal(t, plain, outPlain)
}

func TestEncryptionEngine_Encrypt_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	engine := encryptionengine.NewEncryptionEngine(nil, encMock)

	dek := bytes.Repeat([]byte{'k'}, 32)
	encMock.EXPECT().
		Encrypt("data", dek).
		Return("", pkgErr.NewCustomError(pkgErr.CRYPerrEncryptData, errors.New("fail")))

	_, err := engine.Encrypt("data", dek)
	require.Error(t, err)
}

func TestEncryptionEngine_EncryptDEK_DecryptDEK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	encMock := mocks.NewMockEncrypter(ctrl)
	engine := encryptionengine.NewEncryptionEngine(nil, encMock)

	dek := bytes.Repeat([]byte{'d'}, 32)
	kek := bytes.Repeat([]byte{'k'}, 32)
	encryptedDEK := []byte("encrypted‑dek")

	encMock.EXPECT().EncryptBytes(dek, kek).Return(encryptedDEK, nil)
	encMock.EXPECT().DecryptBytes(encryptedDEK, kek).Return(dek, nil)

	edekStr, err := engine.EncryptDEK(dek, kek)
	require.Nil(t, err)
	require.Equal(t, base64.StdEncoding.EncodeToString(encryptedDEK), edekStr)

	out, err := engine.DecryptDEK(edekStr, kek)
	require.Nil(t, err)
	require.Equal(t, dek, out)
}

func TestEncryptionEngine_DecryptDEK_BadBase64(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	engine := encryptionengine.NewEncryptionEngine(nil, nil)

	_, err := engine.DecryptDEK("##not‑b64##", nil)
	require.Error(t, err)
}
