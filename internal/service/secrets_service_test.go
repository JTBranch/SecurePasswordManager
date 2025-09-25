package service_test

import (
	"encoding/json"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
	"go-password-manager/tests/testdata"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

type MockCryptoService struct {
	mock.Mock
	key []byte
}

func newMockCryptoService(key []byte) *MockCryptoService {
	m := &MockCryptoService{key: key}
	return m
}

// For versioning tests, override methods to return input data directly
func (m *MockCryptoService) EncryptSymmetric(data, key []byte, aad []byte) ([]byte, []byte, error) {
	return data, []byte("nonce"), nil
}

func (m *MockCryptoService) DecryptSymmetric(data, nonce, key []byte, aad []byte) ([]byte, error) {
	return data, nil
}

func (m *MockCryptoService) GetKey() []byte {
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
	require := require.New(t)
	require.NoError(err, "Failed to marshal empty secrets file")
	err = os.WriteFile(testFile, data, 0644)
	require.NoError(err, "Failed to write empty secrets file")

	cryptoService := newMockCryptoService([]byte(testdata.TestEncryptionKey))
	storageService := storage.NewFileStorage(testFile, "1.0.0", testdata.TestUsers.UnitTestUser.Name)

	svc := service.NewSecretsService(cryptoService, storageService)

	return svc
}

// Removed top-level TestSecretsService wrapper; each test is now a top-level function.
func TestCreateSecret(t *testing.T) {
	require := require.New(t)
	svc := setupTestService(t)

	err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
	require.NoError(err, errCreateSecret)

	fileData, err := svc.LoadAllSecrets()
	require.NoError(err, errLoadSecrets)

	require.Equal(1, len(fileData.Secrets), "Expected 1 secret")
	secret := fileData.Secrets[0]
	require.Equal(testdata.TestSecrets.Simple.Name, secret.SecretName, "Secret name should match")
	require.Equal(1, secret.CurrentVersion, "Current version should be 1")
	require.Equal(1, len(secret.Versions), "Expected 1 version")
}

func TestEditSecret(t *testing.T) {
	require := require.New(t)
	svc := setupTestService(t)

	err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, "initial-value")
	require.NoError(err, errCreateSecret)

	err = svc.UpdateSecret(testdata.TestSecrets.Simple.Name, "updated-value")
	require.NoError(err, "Expected no error editing secret")

	fileData, err := svc.LoadAllSecrets()
	require.NoError(err, errLoadSecrets)

	require.Equal(1, len(fileData.Secrets), "Should still have 1 secret")
	secret := fileData.Secrets[0]
	require.Equal(2, secret.CurrentVersion, "Version should be incremented")
	require.Equal(2, len(secret.Versions), "Should have 2 versions")
}

func TestDeleteSecret(t *testing.T) {
	require := require.New(t)
	svc := setupTestService(t)

	err := svc.SaveNewSecret(testdata.TestSecrets.Temporary.Name, testdata.TestSecrets.Temporary.Value)
	require.NoError(err, errCreateSecret)

	err = svc.DeleteSecret(testdata.TestSecrets.Temporary.Name)
	require.NoError(err, "Expected no error deleting secret")

	fileData, err := svc.LoadAllSecrets()
	require.NoError(err, errLoadSecrets)
	require.Equal(0, len(fileData.Secrets), "Expected 0 secrets after deletion")
}

func TestGetSecret(t *testing.T) {
	require := require.New(t)
	svc := setupTestService(t)

	err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
	require.NoError(err, errCreateSecret)

	secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
	require.NoError(err, "Expected no error getting secret")
	require.NotNil(secret, "Expected non-nil secret")
	require.Equal(testdata.TestSecrets.Simple.Name, secret.SecretName, "Secret name should match")
}

func TestGetSecretNonExistent(t *testing.T) {
	require := require.New(t)
	svc := setupTestService(t)

	_, err := svc.GetSecret(nonExistentName)
	require.Error(err, "Expected error getting non-existent secret")
}

func TestGetSecretValue(t *testing.T) {
	require := require.New(t)
	svc := setupTestService(t)

	err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
	require.NoError(err, errCreateSecret)
	secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
	require.NoError(err, errGettingSecretFailed)
	value, err := svc.GetSecretValue(secret)
	require.NoError(err, errGetSecretValue)
	require.Equal(testdata.TestSecrets.Simple.Value, value, secretValueShouldMatch)
}

func TestGetSecretValueByVersion(t *testing.T) {
	require := require.New(t)
	tempDir := t.TempDir()

	cryptoService := newMockCryptoService([]byte(testdata.TestEncryptionKey))
	storageService := storage.NewFileStorage(filepath.Join(tempDir, "secrets.json"), "1.0.0", testdata.TestUsers.UnitTestUser.Name)

	svc := service.NewSecretsService(cryptoService, storageService)

	err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, "value1")
	require.NoError(err, errCreateSecret)
	err = svc.UpdateSecret(testdata.TestSecrets.Simple.Name, "value2")
	require.NoError(err, "updating secret failed")
	secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
	require.NoError(err, errGettingSecretFailed)
	value, err := svc.GetSecretValueByVersion(secret, 1)
	require.NoError(err, errGetSecretValue)
	require.Equal("value1", value, secretValueShouldMatch)

	value, err = svc.GetSecretValueByVersion(secret, 2)
	require.NoError(err, errGetSecretValue)
	require.Equal("value2", value, secretValueShouldMatch)
}

func TestGetSecretValueInvalidVersion(t *testing.T) {
	require := require.New(t)
	svc := setupTestService(t)

	err := svc.SaveNewSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value)
	require.NoError(err, errCreateSecret)
	secret, err := svc.GetSecret(testdata.TestSecrets.Simple.Name)
	require.NoError(err, errGettingSecretFailed)

	_, err = svc.GetSecretValueByVersion(secret, 99)
	require.Error(err, "Expected error for invalid version")
}

func TestLoadAllSecretsFileNotFound(t *testing.T) {
	require := require.New(t)
	tempDir := t.TempDir()

	cryptoService := newMockCryptoService([]byte(testdata.TestEncryptionKey))
	storageService := storage.NewFileStorage(filepath.Join(tempDir, "non_existent_file.json"), "1.0.0", testdata.TestUsers.UnitTestUser.Name)

	svc := service.NewSecretsService(cryptoService, storageService)

	fileData, err := svc.LoadAllSecrets()
	require.NoError(err, "Expected no error when file does not exist")
	require.NotNil(fileData, "Expected non-nil file data")
	require.Equal(0, len(fileData.Secrets), "Expected no secrets in new file")
}
