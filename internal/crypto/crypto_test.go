package crypto_test

import (
	"go-password-manager/internal/crypto"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const errNonceNotNil = "Nonce should not be nil"

func TestEncryptDecrypt(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecrypt", func(tc *helpers.UnitTestCase) {
		// Test data
		plaintext := testdata.TestSecrets.Simple.Value
		key := []byte(testdata.TestEncryptionKey)

		// Test encryption
		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key, nil)
		tc.Require.NoError(err, "Expected no error encrypting")

		// Encrypted data should be different from plaintext
		tc.Assert.NotEqual(string(encrypted), plaintext, "Encrypted data should not equal plaintext")

		// Test decryption
		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key, nil)
		tc.Require.NoError(err, "Expected no error decrypting")

		// Decrypted data should match original plaintext
		tc.Assert.Equal(string(decrypted), plaintext, "Decrypted text should match original")
		tc.Assert.NotNil(nonce, errNonceNotNil)
	})
}

func TestEncryptDecryptEmpty(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecryptEmpty", func(tc *helpers.UnitTestCase) {
		// Test with empty string
		plaintext := ""
		key := []byte(testdata.TestEncryptionKey)

		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key, nil)
		tc.Require.NoError(err, "Expected no error encrypting empty string")

		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key, nil)
		tc.Require.NoError(err, "Expected no error decrypting empty string")

		tc.Assert.Equal(string(decrypted), plaintext, "Expected empty string")
		tc.Assert.NotNil(nonce, errNonceNotNil)
	})
}

func TestEncryptDecryptLongText(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecryptLongText", func(tc *helpers.UnitTestCase) {
		// Test with longer text
		plaintext := testdata.TestSecrets.Long.Value
		key := []byte(testdata.TestEncryptionKey)

		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key, nil)
		tc.Require.NoError(err, "Expected no error encrypting long text")

		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key, nil)
		tc.Require.NoError(err, "Expected no error decrypting long text")

		tc.Assert.Equal(string(decrypted), plaintext, "Expected decrypted text to match original")
		tc.Assert.NotNil(nonce, errNonceNotNil)
	})
}

func TestDecryptWithWrongKey(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestDecryptWithWrongKey", func(tc *helpers.UnitTestCase) {
		plaintext := testdata.TestSecrets.Simple.Value
		key1 := []byte(testdata.TestEncryptionKey)
		key2 := []byte(testdata.DifferentEncryptionKey)

		// Encrypt with key1
		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key1, nil)
		tc.Require.NoError(err, "Expected no error encrypting")

		// Try to decrypt with key2 (should fail)
		_, err = crypto.DecryptSymmetric(encrypted, nonce, key2, nil)
		tc.Assert.Error(err, "Expected error when decrypting with wrong key")
	})
}

func TestEncryptDecryptSpecialCharacters(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecryptSpecialCharacters", func(tc *helpers.UnitTestCase) {
		// Test with special characters and unicode
		plaintext := testdata.TestSecrets.Special.Value
		key := []byte(testdata.TestEncryptionKey)

		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key, nil)
		tc.Require.NoError(err, "Expected no error encrypting special characters")

		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key, nil)
		tc.Require.NoError(err, "Expected no error decrypting special characters")

		tc.Assert.Equal(string(decrypted), plaintext, "Expected decrypted text to match original")
		tc.Assert.NotNil(nonce, errNonceNotNil)
	})
}

// Legacy asymmetric & HKDF tests removed (protocol now uses key box only).

func TestGenerateX25519KeyPair(t *testing.T) {
	pub, _, err := crypto.GenerateX25519KeyPair()
	require.NoError(t, err)
	if len(pub) != 32 {
		t.Errorf("Expected 32-byte public key, got %d", len(pub))
	}
}

func TestGenerateEd25519KeyPair(t *testing.T) {
	pub, priv, err := crypto.GenerateEd25519KeyPair()
	assert.NoError(t, err, "Key pair generation should not error")
	assert.NotNil(t, pub, "Public key should not be nil")
	assert.NotNil(t, priv, "Private key should not be nil")
	assert.Equal(t, 32, len(pub), "Public key should be 32 bytes")
	assert.Equal(t, 64, len(priv), "Private key should be 64 bytes")
}
