package service_test

import (
	"encoding/json"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/testdata"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testSecretsFile = "test_secrets.json"
	nonExistentName = "non-existent"
	errCreateSecret = "Expected no error creating secret"
	errLoadSecrets  = "Expected no error loading secrets"
)

// setupTestService creates a new SecretsService for testing, with a temporary file.
func setupTestService(t *testing.T) *service.SecretsService {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, testSecretsFile)

	svc := service.New("1.0.0", testdata.TestUsers.UnitTestUser.Name, testFile)
	svc.SetEncryptionKey([]byte(testdata.TestEncryptionKey))

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

	return svc
}

func TestSecretsService(t *testing.T) {
	helpers.WithUnitTestCase(t, "CreateSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value, "key_value")
		tc.Require.NoError(err, errCreateSecret)

		fileData, err := svc.LoadLatestSecrets()
		tc.Require.NoError(err, errLoadSecrets)

		tc.Assert.Len(fileData.Secrets, 1, "Expected 1 secret")
		secret := fileData.Secrets[0]
		tc.Assert.Equal(testdata.TestSecrets.Simple.Name, secret.SecretName, "Secret name should match")
		tc.Assert.Equal(1, secret.CurrentVersion, "Current version should be 1")
		tc.Assert.Len(secret.Versions, 1, "Expected 1 version")
	})

	helpers.WithUnitTestCase(t, "EditSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveSecret(testdata.TestSecrets.Simple.Name, "initial-value", "key_value")
		tc.Require.NoError(err, errCreateSecret)

		err = svc.EditSecret(testdata.TestSecrets.Simple.Name, "updated-value")
		tc.Require.NoError(err, "Expected no error editing secret")

		fileData, err := svc.LoadLatestSecrets()
		tc.Require.NoError(err, errLoadSecrets)

		tc.Assert.Len(fileData.Secrets, 1, "Should still have 1 secret")
		secret := fileData.Secrets[0]
		tc.Assert.Equal(2, secret.CurrentVersion, "Version should be incremented")
		tc.Assert.Len(secret.Versions, 2, "Should have 2 versions")
	})

	helpers.WithUnitTestCase(t, "DeleteSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveSecret(testdata.TestSecrets.Temporary.Name, testdata.TestSecrets.Temporary.Value, "key_value")
		tc.Require.NoError(err, errCreateSecret)

		err = svc.DeleteSecret(testdata.TestSecrets.Temporary.Name)
		tc.Require.NoError(err, "Expected no error deleting secret")

		fileData, err := svc.LoadLatestSecrets()
		tc.Require.NoError(err, errLoadSecrets)
		tc.Assert.Len(fileData.Secrets, 0, "Expected 0 secrets after deletion")
	})

	helpers.WithUnitTestCase(t, "GetSecret", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value, "key_value")
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

		err := svc.SaveSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value, "key_value")
		tc.Require.NoError(err, errCreateSecret)

		value, err := svc.GetSecretValue(testdata.TestSecrets.Simple.Name, 1)
		tc.Require.NoError(err, "Expected no error getting secret value")
		tc.Assert.Equal(testdata.TestSecrets.Simple.Value, value, "Secret value should match")
	})

	helpers.WithUnitTestCase(t, "GetSecretValueInvalidVersion", func(tc *helpers.UnitTestCase) {
		svc := setupTestService(t)

		err := svc.SaveSecret(testdata.TestSecrets.Simple.Name, testdata.TestSecrets.Simple.Value, "key_value")
		tc.Require.NoError(err, errCreateSecret)

		_, err = svc.GetSecretValue(testdata.TestSecrets.Simple.Name, 99)
		tc.Assert.Error(err, "Expected error for invalid version")
	})

	helpers.WithUnitTestCase(t, "LoadLatestSecretsFileNotFound", func(tc *helpers.UnitTestCase) {
		tempDir := t.TempDir()
		svc := service.New("1.0.0", testdata.TestUsers.UnitTestUser.Name, filepath.Join(tempDir, "non_existent_file.json"))
		svc.SetEncryptionKey([]byte(testdata.TestEncryptionKey))

		fileData, err := svc.LoadLatestSecrets()
		tc.Require.NoError(err, "Expected no error when file does not exist")
		tc.Assert.NotNil(fileData, "Expected non-nil file data")
		tc.Assert.Empty(fileData.Secrets, "Expected no secrets in new file")
	})
}
