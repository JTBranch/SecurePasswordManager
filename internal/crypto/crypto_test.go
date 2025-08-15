package crypto_test

import (
	"go-password-manager/internal/crypto"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/testdata"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecrypt", func(tc *helpers.UnitTestCase) {
		// Test data
		plaintext := testdata.TestSecrets.Simple.Value
		key := []byte(testdata.TestEncryptionKey)

		// Test encryption
		encrypted, err := crypto.Encrypt([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting")

		// Encrypted data should be different from plaintext
		tc.Assert.NotEqual(string(encrypted), plaintext, "Encrypted data should not equal plaintext")

		// Test decryption
		decrypted, err := crypto.Decrypt(encrypted, key)
		tc.Require.NoError(err, "Expected no error decrypting")

		// Decrypted data should match original plaintext
		tc.Assert.Equal(string(decrypted), plaintext, "Decrypted text should match original")
	})
}

func TestEncryptDecryptEmpty(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecryptEmpty", func(tc *helpers.UnitTestCase) {
		// Test with empty string
		plaintext := ""
		key := []byte(testdata.TestEncryptionKey)

		encrypted, err := crypto.Encrypt([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting empty string")

		decrypted, err := crypto.Decrypt(encrypted, key)
		tc.Require.NoError(err, "Expected no error decrypting empty string")

		tc.Assert.Equal(string(decrypted), plaintext, "Expected empty string")
	})
}

func TestEncryptDecryptLongText(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecryptLongText", func(tc *helpers.UnitTestCase) {
		// Test with longer text
		plaintext := testdata.TestSecrets.Long.Value
		key := []byte(testdata.TestEncryptionKey)

		encrypted, err := crypto.Encrypt([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting long text")

		decrypted, err := crypto.Decrypt(encrypted, key)
		tc.Require.NoError(err, "Expected no error decrypting long text")

		tc.Assert.Equal(string(decrypted), plaintext, "Expected decrypted text to match original")
	})
}

func TestDecryptWithWrongKey(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestDecryptWithWrongKey", func(tc *helpers.UnitTestCase) {
		plaintext := testdata.TestSecrets.Simple.Value
		key1 := []byte(testdata.TestEncryptionKey)
		key2 := []byte(testdata.DifferentEncryptionKey)

		// Encrypt with key1
		encrypted, err := crypto.Encrypt([]byte(plaintext), key1)
		tc.Require.NoError(err, "Expected no error encrypting")

		// Try to decrypt with key2 (should fail)
		_, err = crypto.Decrypt(encrypted, key2)
		tc.Assert.Error(err, "Expected error when decrypting with wrong key")
	})
}

func TestEncryptDecryptSpecialCharacters(t *testing.T) {
	helpers.WithUnitTestCase(t, "TestEncryptDecryptSpecialCharacters", func(tc *helpers.UnitTestCase) {
		// Test with special characters and unicode
		plaintext := testdata.TestSecrets.Special.Value
		key := []byte(testdata.TestEncryptionKey)

		encrypted, err := crypto.Encrypt([]byte(plaintext), key)
		tc.Require.NoError(err, "Expected no error encrypting special characters")

		decrypted, err := crypto.Decrypt(encrypted, key)
		tc.Require.NoError(err, "Expected no error decrypting special characters")

		tc.Assert.Equal(string(decrypted), plaintext, "Expected decrypted text to match original")
	})
}
