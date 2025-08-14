package crypto

import (
	"testing"
)

const testKey = "test-key-32-bytes-long-exactly!!"

func TestEncryptDecrypt(t *testing.T) {
	// Test data
	plaintext := "test-secret-value"
	key := []byte(testKey)

	// Test encryption
	encrypted, err := Encrypt([]byte(plaintext), key)
	if err != nil {
		t.Fatalf("Expected no error encrypting, got: %v", err)
	}

	// Encrypted data should be different from plaintext
	if encrypted == plaintext {
		t.Error("Encrypted data should not equal plaintext")
	}

	// Test decryption
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Expected no error decrypting, got: %v", err)
	}

	// Decrypted data should match original plaintext
	if string(decrypted) != plaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", plaintext, string(decrypted))
	}
}

func TestEncryptDecryptEmpty(t *testing.T) {
	// Test with empty string
	plaintext := ""
	key := []byte(testKey)

	encrypted, err := Encrypt([]byte(plaintext), key)
	if err != nil {
		t.Fatalf("Expected no error encrypting empty string, got: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Expected no error decrypting empty string, got: %v", err)
	}

	if string(decrypted) != plaintext {
		t.Errorf("Expected empty string, got '%s'", string(decrypted))
	}
}

func TestEncryptDecryptLongText(t *testing.T) {
	// Test with longer text
	plaintext := "This is a much longer secret value that contains multiple words and special characters! @#$%^&*()"
	key := []byte(testKey)

	encrypted, err := Encrypt([]byte(plaintext), key)
	if err != nil {
		t.Fatalf("Expected no error encrypting long text, got: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Expected no error decrypting long text, got: %v", err)
	}

	if string(decrypted) != plaintext {
		t.Errorf("Expected decrypted text to match original")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	plaintext := "test-secret-value"
	key1 := []byte(testKey)
	key2 := []byte("different-key-32-bytes-exactly!")

	// Encrypt with key1
	encrypted, err := Encrypt([]byte(plaintext), key1)
	if err != nil {
		t.Fatalf("Expected no error encrypting, got: %v", err)
	}

	// Try to decrypt with key2 (should fail)
	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Fatal("Expected error when decrypting with wrong key, got nil")
	}
}

func TestEncryptDecryptSpecialCharacters(t *testing.T) {
	// Test with special characters and unicode
	plaintext := "密码test🔐password!@#$%^&*()_+-=[]{}|;':\",./<>?"
	key := []byte(testKey)

	encrypted, err := Encrypt([]byte(plaintext), key)
	if err != nil {
		t.Fatalf("Expected no error encrypting special chars, got: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Expected no error decrypting special chars, got: %v", err)
	}

	if string(decrypted) != plaintext {
		t.Errorf("Expected decrypted text to match original with special characters")
	}
}

func TestLoadOrCreateKey(t *testing.T) {
	// This test mainly checks that the function doesn't panic
	// and returns a key of appropriate length
	key, err := LoadOrCreateKey()
	if err != nil {
		t.Fatalf("Expected no error loading/creating key, got: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got: %d", len(key))
	}
}
