package service

import (
	"encoding/json"
	"go-password-manager/internal/domain"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testSecretsFile   = "test_secrets.json"
	testSecretName    = "test-secret"
	testSecretValue   = "test-value"
	nonExistentName   = "non-existent"
	testUser          = "test-user"
	testEncryptionKey = "test-key-32-bytes-long-exactly!!"
	errCreateSecret   = "Expected no error creating secret, got: %v"
	errLoadSecrets    = "Expected no error loading secrets, got: %v"
)

func setupTestService(t *testing.T) (*SecretsService, string) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, testSecretsFile)

	service := &SecretsService{
		AppVersion: "1.0.0",
		AppUser:    testUser,
		filePath:   testFile,
	}

	// Initialize with test encryption key
	service.encryptionKey = []byte(testEncryptionKey)

	// Create an empty secrets file to avoid "file not found" errors
	emptySecretsFile := domain.SecretsFile{
		AppVersion:  "1.0.0",
		AppUser:     testUser,
		LastUpdated: time.Now().Format(time.RFC3339),
		Secrets:     []domain.Secret{},
	}

	// Write the empty file
	data, _ := json.MarshalIndent(emptySecretsFile, "", "  ")
	os.WriteFile(testFile, data, 0644)

	return service, testFile
}

func TestSecretsServiceCreateSecret(t *testing.T) {
	service, _ := setupTestService(t)

	// Test creating a new secret
	err := service.SaveSecret(testSecretName, testSecretValue, "key_value")
	if err != nil {
		t.Fatalf(errCreateSecret, err)
	}

	// Verify the secret was created
	fileData, err := service.LoadLatestSecrets()
	if err != nil {
		t.Fatalf(errLoadSecrets, err)
	}

	if len(fileData.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got: %d", len(fileData.Secrets))
	}

	secret := fileData.Secrets[0]
	if secret.SecretName != testSecretName {
		t.Errorf("Expected secret name '%s', got: %s", testSecretName, secret.SecretName)
	}

	if secret.CurrentVersion != 1 {
		t.Errorf("Expected current version 1, got: %d", secret.CurrentVersion)
	}

	if len(secret.Versions) != 1 {
		t.Errorf("Expected 1 version, got: %d", len(secret.Versions))
	}
}

func TestSecretsServiceEditSecret(t *testing.T) {
	service, _ := setupTestService(t)

	// Create initial secret
	err := service.SaveSecret(testSecretName, "initial-value", "key_value")
	if err != nil {
		t.Fatalf(errCreateSecret, err)
	}

	// Edit the secret
	err = service.EditSecret(testSecretName, "updated-value")
	if err != nil {
		t.Fatalf("Expected no error editing secret, got: %v", err)
	}

	// Verify the secret was updated
	fileData, err := service.LoadLatestSecrets()
	if err != nil {
		t.Fatalf(errLoadSecrets, err)
	}

	secret := fileData.Secrets[0]
	if secret.CurrentVersion != 2 {
		t.Errorf("Expected current version 2, got: %d", secret.CurrentVersion)
	}

	if len(secret.Versions) != 2 {
		t.Errorf("Expected 2 versions, got: %d", len(secret.Versions))
	}

	// Verify we can decrypt the latest version
	decrypted, err := service.DisplaySecret(secret)
	if err != nil {
		t.Fatalf("Expected no error decrypting secret, got: %v", err)
	}

	if decrypted != "updated-value" {
		t.Errorf("Expected decrypted value 'updated-value', got: %s", decrypted)
	}
}

func TestSecretsServiceDeleteSecret(t *testing.T) {
	service, _ := setupTestService(t)

	// Create a secret
	err := service.SaveSecret(testSecretName, testSecretValue, "key_value")
	if err != nil {
		t.Fatalf(errCreateSecret, err)
	}

	// Delete the secret
	err = service.DeleteSecret(testSecretName)
	if err != nil {
		t.Fatalf("Expected no error deleting secret, got: %v", err)
	}

	// Verify the secret was deleted
	fileData, err := service.LoadLatestSecrets()
	if err != nil {
		t.Fatalf(errLoadSecrets, err)
	}

	if len(fileData.Secrets) != 0 {
		t.Errorf("Expected 0 secrets after deletion, got: %d", len(fileData.Secrets))
	}
}

func TestSecretsServiceDecryptSecretVersion(t *testing.T) {
	service, _ := setupTestService(t)

	// Create and edit a secret to have multiple versions
	err := service.SaveSecret(testSecretName, "version-1", "key_value")
	if err != nil {
		t.Fatalf(errCreateSecret, err)
	}

	err = service.EditSecret(testSecretName, "version-2")
	if err != nil {
		t.Fatalf("Expected no error editing secret, got: %v", err)
	}

	// Load the secret
	fileData, err := service.LoadLatestSecrets()
	if err != nil {
		t.Fatalf(errLoadSecrets, err)
	}

	secret := fileData.Secrets[0]

	// Test decrypting version 1
	version1 := secret.Versions[0]
	decrypted, err := service.DecryptSecretVersion(version1)
	if err != nil {
		t.Fatalf("Expected no error decrypting version 1, got: %v", err)
	}

	if decrypted != "version-1" {
		t.Errorf("Expected decrypted value 'version-1', got: %s", decrypted)
	}

	// Test decrypting version 2
	version2 := secret.Versions[1]
	decrypted, err = service.DecryptSecretVersion(version2)
	if err != nil {
		t.Fatalf("Expected no error decrypting version 2, got: %v", err)
	}

	if decrypted != "version-2" {
		t.Errorf("Expected decrypted value 'version-2', got: %s", decrypted)
	}
}

func TestSecretsServiceDuplicateSecretName(t *testing.T) {
	service, _ := setupTestService(t)

	// Create a secret
	err := service.SaveSecret(testSecretName, testSecretValue, "key_value")
	if err != nil {
		t.Fatalf(errCreateSecret, err)
	}

	// Try to create another secret with the same name (should create new version)
	err = service.SaveSecret(testSecretName, "another-value", "key_value")
	if err != nil {
		t.Fatalf("Expected no error when adding version to existing secret, got: %v", err)
	}

	// Verify that we now have 2 versions
	fileData, err := service.LoadLatestSecrets()
	if err != nil {
		t.Fatalf(errLoadSecrets, err)
	}

	if len(fileData.Secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(fileData.Secrets))
	}

	secret := fileData.Secrets[0]
	if secret.CurrentVersion != 2 {
		t.Errorf("Expected version 2, got %d", secret.CurrentVersion)
	}

	if len(secret.Versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(secret.Versions))
	}
}

func TestSecretsServiceNonExistentSecret(t *testing.T) {
	service, _ := setupTestService(t)

	// Try to edit non-existent secret
	err := service.EditSecret(nonExistentName, "value")
	if err == nil {
		t.Fatal("Expected error when editing non-existent secret, got nil")
	}

	// Try to delete non-existent secret (should be idempotent)
	err = service.DeleteSecret(nonExistentName)
	if err != nil {
		t.Fatalf("Delete should be idempotent, got error: %v", err)
	}
}
