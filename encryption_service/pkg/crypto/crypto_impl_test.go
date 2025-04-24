package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustKey(n int) []byte {
	key := strings.Repeat("k", n)
	return []byte(key)
}

func TestEncryption_StringRoundTrip(t *testing.T) {
	e := NewEncryption()
	key := mustKey(32) // AES‑256

	plain := "hot-secret-payload"
	cipher, cerr := e.Encrypt(plain, key)

	require.Nil(t, cerr, "encrypt failed")

	out, derr := e.Decrypt(cipher, key)
	require.Nil(t, derr, "decrypt failed")
	require.Equal(t, plain, out)
}

func TestEncryption_StringRoundTripFail(t *testing.T) {
	e := NewEncryption()
	key := mustKey(89) // AES‑256

	plain := "hot-secret-payload"
	cipher, cerr := e.Encrypt(plain, key)

	require.NotNil(t, cerr, "encrypt failed")
	require.Empty(t, cipher, "encrypt failed")
}

func TestEncryption_StringRoundTripDecryptFail(t *testing.T) {
	e := NewEncryption()
	key := mustKey(32) // AES‑256

	plain := "hot-secret-payload"
	cipher, cerr := e.Encrypt(plain, key)

	require.Nil(t, cerr, "encrypt failed")

	key = mustKey(89) // AES‑256
	out, derr := e.Decrypt(cipher, key)
	require.NotNil(t, derr, "decrypt failed")
	require.Empty(t, out)
}

func TestEncryption_BytesRoundTrip(t *testing.T) {
	e := NewEncryption()
	key := mustKey(16) // AES‑128

	plain := []byte("byte slice payload")
	cipher, cerr := e.EncryptBytes(plain, key)
	require.Nil(t, cerr)

	out, derr := e.DecryptBytes(cipher, key)
	require.Nil(t, derr)
	require.Equal(t, plain, out)
}

func TestEncryption_KeyLengthErrors(t *testing.T) {
	e := NewEncryption()
	badKey := mustKey(10) // invalid size for AES

	_, cerr := e.Encrypt("data", badKey)
	require.NotNil(t, cerr)

	_, cerr = e.EncryptBytes([]byte("data"), badKey)
	require.NotNil(t, cerr)
}

func TestEncryption_ShortCipherErrors(t *testing.T) {
	e := NewEncryption()
	key := mustKey(32)

	// shorter than aes.BlockSize (16)
	shortBin := []byte("short")
	shortB64 := "c2hvcnQ=" // base64("short")

	_, derr := e.Decrypt(shortB64, key)
	require.NotNil(t, derr)

	_, derr = e.DecryptBytes(shortBin, key)
	require.NotNil(t, derr)
}

func TestEncryption_BadBase64Error(t *testing.T) {
	e := NewEncryption()
	key := mustKey(32)

	_, derr := e.Decrypt("###not‑b64###", key)
	require.NotNil(t, derr)
}
