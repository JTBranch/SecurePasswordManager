package service_test

import (
	"encoding/json"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/testdata"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testSecretsFile        = "test_secrets.json"
	nonExistentName        = "non-existent"
	errCreateSecret        = "Expected no error creating secret"
	errLoadSecrets         = "Expected no error loading secrets"
	errGetSecretValue      = "Expected no error getting secret value"
	errGettingSecretFailed = "getting secret failed"
	secretValueShouldMatch = "Secret value should match"
)

// Mock CryptoService for testing
type mockCryptoService struct {
	key []byte
}

func newMockCryptoService(key []byte) service.CryptoService {
	return &mockCryptoService{key: key}
}

func (m *mockCryptoService) Encrypt(data, key []byte) ([]byte, error) {
	s, err := crypto.Encrypt(data, key)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (m *mockCryptoService) Decrypt(data, key []byte) ([]byte, error) {
	return crypto.Decrypt(string(data), key)
}

func (m *mockCryptoService) GetKey() []byte {
	return m.key
}

// setupTestService creates a new SecretsService for testing, with a temporary file.
func setupTestService(t *testing.T) *service.SecretsService {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, testSecretsFile)

	// Create an empty secrets file to ensure tests start with a clean slate.
	emptySecretsFile := domain.SecretsFile{
		AppVersion:  "1.0.0",
		AppUser:     testdata.TestUsers.UnitTestUser.Name,
		LastUpdated: time.Now().Format(time.RFC3339),
		Secrets:     []domain.Secret{},
	}

	data, err := json.MarshalIndent(emptySecretsFile, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal empty secrets file: %v", err)
	}
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to write empty secrets file: %v", err)
	}

	cryptoService := newMockCryptoService([]byte(testdata.TestEncryptionKey))
	storageService := storage.NewFileStorage(testFile, "1.0.0", testdata.TestUsers.UnitTestUser.Name)

	svc := service.NewSecretsService(cryptoService, storageService)

	return svc
}

func TestSecretsService(t *testing.T) {
	helpers.WithUnitTestCase(t, "CreateSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
		tc.Require.NoError(err, errCreateSecret)

		fileData, err := svc.LoadAllSecrets()
		tc.Require.NoError(err, errLoadSecrets)

		tc.Assert.Len(fileData.Secrets, 1, "Expected 1 secret")
		secret := fileData.Secrets[0]
		tc.Assert.Equal(testdata.TestSecrets.Simple.Name, secret.SecretName, "Secret name should match")
		tc.Assert.Equal(1, secret.CurrentVersion, "Current version should be 1")
		tc.Assert.Len(secret.Versions, 1, "Expected 1 version")
	})

	helpers.WithUnitTestCase(t, "EditSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, "initial-value")
		tc.Require.NoError(err, errCreateSecret)

		err = svc.UpdateSecret(testdata.TestSecrets.Simple.Name, "updated-value")
		tc.Require.NoError(err, "Expected no error editing secret")

		fileData, err := svc.LoadAllSecrets()
		tc.Require.NoError(err, errLoadSecrets)

		tc.Assert.Len(fileData.Secrets, 1, "Should still have 1 secret")
		secret := fileData.Secrets[0]
		tc.Assert.Equal(2, secret.CurrentVersion, "Version should be incremented")
		tc.Assert.Len(secret.Versions, 2, "Should have 2 versions")
	})

	helpers.WithUnitTestCase(t, "DeleteSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveNewSecret(testdata.TestSecrets.Temporary.Name, testdata.TestSecrets.Temporary.Value)
		tc.Require.NoError(err, errCreateSecret)

		err = svc.DeleteSecret(testdata.TestSecrets.Temporary.Name)
		tc.Require.NoError(err, "Expected no error deleting secret")

		fileData, err := svc.LoadAllSecrets()
		tc.Require.NoError(err, errLoadSecrets)
		tc.Assert.Len(fileData.Secrets, 0, "Expected 0 secrets after deletion")
	})

	helpers.WithUnitTestCase(t, "GetSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
		tc.Require.NoError(err, errCreateSecret)

		secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
		tc.Require.NoError(err, "Expected no error getting secret")
		tc.Assert.NotNil(secret, "Expected non-nil secret")
		tc.Assert.Equal(testdata.TestSecrets.Simple.Name, secret.SecretName, "Secret name should match")
	})

	helpers.WithUnitTestCase(t, "GetSecretNonExistent", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		_, err := svc.GetSecret(nonExistentName)
		tc.Assert.Error(err, "Expected error getting non-existent secret")
	})

	helpers.WithUnitTestCase(t, "GetSecretValue", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
		tc.Require.NoError(err, errCreateSecret)
		secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
		tc.Require.NoError(err, errGettingSecretFailed)
		value, err := svc.GetSecretValue(secret)
		tc.Require.NoError(err, errGetSecretValue)
		tc.Assert.Equal(testdata.TestSecrets.Simple.Value, value, secretValueShouldMatch)
		tc.Assert.Equal(testdata.TestSecrets.Simple.Value, value, secretValueShouldMatch)
		tc.Assert.Equal(testdata.TestSecrets.Simple.Value, value, secretValueShouldMatch)
	})

	helpers.WithUnitTestCase(t, "GetSecretValueByVersion", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, "value1")
		tc.Require.NoError(err, errCreateSecret)
		err = svc.UpdateSecret(testdata.TestSecrets.Simple.Name, "value2")
		tc.Require.NoError(err, "updating secret failed")
		secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
		tc.Require.NoError(err, errGettingSecretFailed)
		value, err := svc.GetSecretValueByVersion(secret, 1)
		tc.Require.NoError(err, errGetSecretValue)
		tc.Assert.Equal("value1", value, secretValueShouldMatch)

		value, err = svc.GetSecretValueByVersion(secret, 2)
		tc.Require.NoError(err, errGetSecretValue)
		tc.Assert.Equal("value2", value, secretValueShouldMatch)
		tc.Assert.Equal("value2", value, secretValueShouldMatch)
		tc.Assert.Equal("value2", value, secretValueShouldMatch)
	})

	helpers.WithUnitTestCase(t, "GetSecretValueInvalidVersion", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
		tc.Require.NoError(err, errCreateSecret)
		secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
		tc.Require.NoError(err, errGettingSecretFailed)

		_, err = svc.GetSecretValueByVersion(secret, 99)
		tc.Assert.Error(err, "Expected error for invalid version")
		tc.Assert.Error(err, "Expected error for invalid version")
	})

	helpers.WithUnitTestCase(t, "LoadAllSecretsFileNotFound", func(tc *helpers.UnitTestCase) {
		tempDir := t.TempDir()

		cryptoService := newMockCryptoService([]byte(testdata.TestEncryptionKey))
		storageService := storage.NewFileStorage(filepath.Join(tempDir, "non_existent_file.json"), "1.0.0", testdata.TestUsers.UnitTestUser.Name)

		svc := service.NewSecretsService(cryptoService, storageService)

		fileData, err := svc.LoadAllSecrets()
		tc.Require.NoError(err, "Expected no error when file does not exist")
		tc.Assert.NotNil(fileData, "Expected non-nil file data")
		tc.Assert.Empty(fileData.Secrets, "Expected no secrets in new file")
	})
}
