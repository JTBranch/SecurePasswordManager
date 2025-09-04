package crypto_test

import (
	"bytes"
	"go-password-manager/internal/crypto"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
)

const errNonceNotNil = "Nonce should not be nil"

func TestEncryptDecrypt(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecrypt", func(tc *helpers.UnitTestCase) {
		// Test data
		plaintext := testdata.TestSecrets.Simple.Value
		key := []byte(testdata.TestEncryptionKey)

		// Test encryption
		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting")

		// Encrypted data should be different from plaintext
		tc.Assert.NotEqual(string(encrypted), plaintext, "Encrypted data should not equal plaintext")

		// Test decryption
		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key)
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

		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting empty string")

		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key)
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

		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting long text")

		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key)
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
		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key1)
		tc.Require.NoError(err, "Expected no error encrypting")

		// Try to decrypt with key2 (should fail)
		_, err = crypto.DecryptSymmetric(encrypted, nonce, key2)
		tc.Assert.Error(err, "Expected error when decrypting with wrong key")
	})
}

func TestEncryptDecryptSpecialCharacters(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecryptSpecialCharacters", func(tc *helpers.UnitTestCase) {
		// Test with special characters and unicode
		plaintext := testdata.TestSecrets.Special.Value
		key := []byte(testdata.TestEncryptionKey)

		encrypted, nonce, err := crypto.EncryptSymmetric([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting special characters")

		decrypted, err := crypto.DecryptSymmetric(encrypted, nonce, key)
		tc.Require.NoError(err, "Expected no error decrypting special characters")

		tc.Assert.Equal(string(decrypted), plaintext, "Expected decrypted text to match original")
		tc.Assert.NotNil(nonce, errNonceNotNil)
	})
}

// --- X25519 Asymmetric Crypto Tests ---

const (
	errKeyGen  = "Key pair generation failed: %v"
	errEncrypt = "Encryption failed: %v"
	errDecrypt = "Decryption failed: %v"
)

func TestGenerateX25519KeyPair(t *testing.T) {
	pub, _, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf(errKeyGen, err)
	}
	if len(pub) != 32 {
		t.Errorf("Expected 32-byte public key, got %d", len(pub))
	}
}

func TestEncryptDecryptAsymmetricSuccess(t *testing.T) {
	pub, priv, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf(errKeyGen, err)
	}
	plaintext := []byte("hello asymmetric world")
	result, err := crypto.EncryptAsymmetric(plaintext, pub)
	if err != nil {
		t.Fatalf(errEncrypt, err)
	}
	// Check all fields
	if len(result.Ciphertext) == 0 {
		t.Errorf("Ciphertext should not be empty")
	}
	if len(result.Nonce) == 0 {
		t.Errorf("Nonce should not be empty")
	}
	if len(result.EphemeralPublicKey) != 32 {
		t.Errorf("Ephemeral public key should be 32 bytes, got %d", len(result.EphemeralPublicKey))
	}
	fullCipher := append(result.EphemeralPublicKey, result.Ciphertext...)
	// Decrypt using priv
	decrypted, err := crypto.DecryptAsymmetric(fullCipher, result.Nonce, priv)
	if err != nil {
		t.Fatalf(errDecrypt, err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text does not match original")
	}
}

func TestEncryptDecryptAsymmetricEmptyPlaintext(t *testing.T) {
	pub, priv, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf(errKeyGen, err)
	}
	plaintext := []byte("")
	result, err := crypto.EncryptAsymmetric(plaintext, pub)
	if err != nil {
		t.Fatalf(errEncrypt, err)
	}
	if len(result.Ciphertext) == 0 {
		t.Errorf("Ciphertext should not be empty for empty plaintext")
	}
	fullCipher := append(result.EphemeralPublicKey, result.Ciphertext...)
	decrypted, err := crypto.DecryptAsymmetric(fullCipher, result.Nonce, priv)
	if err != nil {
		t.Fatalf(errDecrypt, err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text does not match original (empty)")
	}
}

func TestDecryptAsymmetricWrongKey(t *testing.T) {
	pub, _, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf(errKeyGen, err)
	}
	_, otherPriv, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf(errKeyGen, err)
	}
	plaintext := []byte("should not decrypt")
	result, err := crypto.EncryptAsymmetric(plaintext, pub)
	if err != nil {
		t.Fatalf(errEncrypt, err)
	}
	// Try to decrypt with wrong private key
	encrypted := append(result.Nonce, result.Ciphertext...)
	fullCipher := append(result.EphemeralPublicKey, encrypted...)
	_, err = crypto.DecryptAsymmetric(fullCipher, result.Nonce, otherPriv)
	if err == nil {
		t.Errorf("Expected error when decrypting with wrong private key")
	}
}

func TestDecryptAsymmetricInvalidCiphertext(t *testing.T) {
	_, priv, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf(errKeyGen, err)
	}
	// Too short ciphertext
	_, err = crypto.DecryptAsymmetric([]byte("short"), []byte("shortnonce"), priv)
	if err == nil {
		t.Errorf("Expected error for short ciphertext")
	}
}
func TestDeriveSymmetricKeyHKDF(t *testing.T) {
	sharedSecret := []byte("shared-secret-value")
	salt := []byte("test-salt")
	info := []byte("test-info")
	key, err := crypto.DeriveSymmetricKeyHKDF(sharedSecret, salt, info)
	assert.NoError(t, err, "DeriveSymmetricKeyHKDF failed")
	assert.Equal(t, 32, len(key), "Expected key length 32")
	assert.False(t, bytes.Equal(key, make([]byte, 32)), "Derived key should not be all zeros")
}

func TestGenerateEd25519KeyPair(t *testing.T) {
	pub, priv, err := crypto.GenerateEd25519KeyPair()
	assert.NoError(t, err, "Key pair generation should not error")
	assert.NotNil(t, pub, "Public key should not be nil")
	assert.NotNil(t, priv, "Private key should not be nil")
	assert.Equal(t, 32, len(pub), "Public key should be 32 bytes")
	assert.Equal(t, 64, len(priv), "Private key should be 64 bytes")
}
