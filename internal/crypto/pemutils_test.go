package crypto

import (
	"go-password-manager/internal/config/devicekeys"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeX25519PublicKeyPEM(t *testing.T) {
	pemUtil := &PemUtils{}

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	pemBytes, err := pemUtil.EncodeKeyToPEM(key, devicekeys.KeyTypeX25519Public)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(pemBytes), "X25519 PUBLIC KEY"))

	decoded, err := pemUtil.DecodeKeyFromPEM(pemBytes, devicekeys.KeyTypeX25519Public)
	assert.NoError(t, err)
	assert.Equal(t, key, decoded)
	assert.NoError(t, err)
}

func TestEncodeDecodeX25519PrivateKeyPEM(t *testing.T) {
	pemUtil := &PemUtils{}

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(255 - i)
	}
	pemBytes, err := pemUtil.EncodeKeyToPEM(key, devicekeys.KeyTypeX25519Private)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(pemBytes), "X25519 PRIVATE KEY"))

	decoded, err := pemUtil.DecodeKeyFromPEM(pemBytes, devicekeys.KeyTypeX25519Private)
	assert.NoError(t, err)
	assert.Equal(t, key, decoded)
	assert.NoError(t, err)
}

func TestEncodeDecodeEd25519PublicKeyPEM(t *testing.T) {
	pemUtil := &PemUtils{}

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	pemBytes, err := pemUtil.EncodeKeyToPEM(key, devicekeys.KeyTypeEd25519Public)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(pemBytes), "ED25519 PUBLIC KEY"))

	decoded, err := pemUtil.DecodeKeyFromPEM(pemBytes, devicekeys.KeyTypeEd25519Public)
	assert.NoError(t, err)
	assert.Equal(t, key, decoded)
	assert.NoError(t, err)
}

func TestEncodeDecodeEd25519PrivateKeyPEM(t *testing.T) {
	pemUtil := &PemUtils{}

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(255 - i)
	}
	pemBytes, err := pemUtil.EncodeKeyToPEM(key, devicekeys.KeyTypeEd25519Private)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(pemBytes), "ED25519 PRIVATE KEY"))

	decoded, err := pemUtil.DecodeKeyFromPEM(pemBytes, devicekeys.KeyTypeEd25519Private)
	assert.NoError(t, err)
	assert.Equal(t, key, decoded)
}

func TestEncodeSymmetricKeyToPEM(t *testing.T) {
	pemUtil := &PemUtils{}
	key := []byte("mytestkey1234567890")
	pemBytes, err := pemUtil.EncodeKeyToPEM(key, devicekeys.KeyTypeSymmetric)
	assert.NoError(t, err, "Failed to encode symmetric key to PEM")
	assert.Contains(t, string(pemBytes), "SYMMETRIC KEY", "PEM output missing block type")

	// Decode and check key bytes
	decodedKey, err := pemUtil.DecodeKeyFromPEM(pemBytes, devicekeys.KeyTypeSymmetric)
	assert.NoError(t, err, "Failed to decode symmetric key from PEM")
	assert.Equal(t, key, decodedKey, "Decoded key does not match original")
}

func TestDecodeSymmetricKeyFromPEM(t *testing.T) {
	pemUtil := &PemUtils{}
	key := []byte("mytestkey1234567890")
	pemBytes, _ := pemUtil.EncodeKeyToPEM(key, devicekeys.KeyTypeSymmetric)
	decodedKey, err := pemUtil.DecodeKeyFromPEM(pemBytes, devicekeys.KeyTypeSymmetric)
	assert.NoError(t, err, "Failed to decode symmetric key from PEM")
	assert.Equal(t, key, decodedKey, "Decoded key does not match original")
}
