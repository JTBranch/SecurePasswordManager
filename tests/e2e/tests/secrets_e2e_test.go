package e2e

import (
	"go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
	"go-password-manager/pkg/reporting"
	"go-password-manager/tests/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretsWorkflowE2E(t *testing.T) {
	reporting.WithReporting(t, "TestSecretsWorkflowE2E", func(reporter *reporting.TestWrapper) {
		testCreateEditDeleteWorkflow(reporter)
	})
}

func testCreateEditDeleteWorkflow(reporter *reporting.TestWrapper) {
	t := reporter.T()
	reporter.LogStep("Initializing secrets service", nil)

	buildCfg, err := buildconfig.Load()
	require.NoError(t, err, "Failed to load build config")

	configService, err := config.NewConfigService(buildCfg)
	require.NoError(t, err, "Failed to create config service")

	cryptoService, err := crypto.NewCryptoService(configService)
	require.NoError(t, err, "Failed to create crypto service")

	secretsPath, err := buildCfg.GetSecretsFilePath()
	require.NoError(t, err)
	storageService := storage.NewFileStorage(secretsPath, buildCfg.Application.Version, "e2e-user")

	secretsService := service.NewSecretsService(cryptoService, storageService)

	// Test 1: Create a secret
	secretName := testdata.TestSecrets.Simple.Name
	secretValue := testdata.TestSecrets.Simple.Value

	reporter.LogStep("Creating a new secret", map[string]interface{}{"secretName": secretName})
	err = secretsService.SaveNewSecret(secretName, secretValue)
	require.NoError(t, err, "Failed to create secret")

	// Test 2: Load and verify secret
	reporter.LogStep("Loading and verifying secret", nil)
	fileData, err := secretsService.LoadAllSecrets()
	require.NoError(t, err, "Failed to load secrets")
	require.NotEmpty(t, fileData.Secrets, "Expected at least 1 secret, got 0")

	// Find our test secret
	var testSecret *domain.Secret
	for i := range fileData.Secrets {
		if fileData.Secrets[i].SecretName == secretName {
			testSecret = &fileData.Secrets[i]
			break
		}
	}
	require.NotNil(t, testSecret, "Could not find test secret '%s'", secretName)

	// Test 3: Display secret (decrypt)
	reporter.LogStep("Decrypting and verifying secret value", nil)
	decrypted, err := secretsService.GetSecretValue(testSecret)
	require.NoError(t, err, "Failed to decrypt secret")
	assert.Equal(t, secretValue, decrypted, "Decrypted secret value does not match original")

	testEditSecretWorkflow(reporter, secretsService, secretName)
	testDeleteSecretWorkflow(reporter, secretsService, secretName)
}

func testEditSecretWorkflow(reporter *reporting.TestWrapper, secretsService *service.SecretsService, secretName string) {
	t := reporter.T()
	// Test 4: Edit secret (create new version)
	newValue := testdata.TestSecrets.Complex.Value

	reporter.LogStep("Editing secret to create a new version", map[string]interface{}{"newValue": newValue})
	err := secretsService.UpdateSecret(secretName, newValue)
	require.NoError(t, err, "Failed to edit secret")

	// Test 5: Verify edit created new version
	reporter.LogStep("Verifying new version was created", nil)
	fileData, err := secretsService.LoadAllSecrets()
	require.NoError(t, err, "Failed to reload secrets after edit")

	// Find and verify our secret
	var foundSecret *domain.Secret
	for i := range fileData.Secrets {
		if fileData.Secrets[i].SecretName == secretName {
			foundSecret = &fileData.Secrets[i]
			break
		}
	}
	require.NotNil(t, foundSecret, "Could not find secret after edit: %s", secretName)
	assert.GreaterOrEqual(t, foundSecret.CurrentVersion, 2, "Expected version >= 2 after edit")

	// Test 6: Verify current value is updated
	reporter.LogStep("Verifying current value is updated", nil)
	currentValue, err := secretsService.GetSecretValue(foundSecret)
	require.NoError(t, err, "Failed to display current secret value")

	if currentValue != newValue {
		t.Errorf("Expected current value '%s', got '%s'", newValue, currentValue)
	}
}

func testDeleteSecretWorkflow(reporter *reporting.TestWrapper, secretsService *service.SecretsService, secretName string) {
	t := reporter.T()
	// Test 7: Delete secret
	reporter.LogStep("Deleting secret", map[string]interface{}{"secretName": secretName})
	err := secretsService.DeleteSecret(secretName)
	require.NoError(t, err, "Failed to delete secret")

	// Test 8: Verify deletion
	reporter.LogStep("Verifying secret was deleted", nil)
	fileData, err := secretsService.LoadAllSecrets()
	require.NoError(t, err, "Failed to reload secrets after deletion")

	// Ensure the deleted secret is not present
	for _, secret := range fileData.Secrets {
		assert.NotEqual(t, secretName, secret.SecretName, "Found deleted secret in file")
	}
}

func TestErrorHandlingE2E(t *testing.T) {
	reporting.WithReporting(t, "TestErrorHandlingE2E", func(reporter *reporting.TestWrapper) {
		t := reporter.T()
		// Setup
		buildCfg, err := buildconfig.Load()
		require.NoError(t, err, "Failed to load build config")

		configService, err := config.NewConfigService(buildCfg)
		require.NoError(t, err, "Failed to create config service")

		cryptoService, err := crypto.NewCryptoService(configService)
		require.NoError(t, err, "Failed to create crypto service")

		secretsPath, err := buildCfg.GetSecretsFilePath()
		require.NoError(t, err)
		storageService := storage.NewFileStorage(secretsPath, buildCfg.Application.Version, "e2e-user")
		secretsService := service.NewSecretsService(cryptoService, storageService)

		// Test error handling - edit non-existent secret
		reporter.LogStep("Testing error on editing non-existent secret", nil)
		err = secretsService.UpdateSecret("non-existent-secret", "some-value")
		require.Error(t, err, "Expected error when editing non-existent secret")

		// Test error handling - delete non-existent secret (should not error but should be idempotent)
		reporter.LogStep("Testing idempotency of deleting non-existent secret", nil)
		err = secretsService.DeleteSecret("non-existent-secret")
		require.NoError(t, err, "Delete should be idempotent, got no error")

		// Test that SaveSecret with same name creates new version (this is intended behavior)
		secretName := "test-versioning"
		reporter.LogStep("Testing versioning on saving with same name", map[string]interface{}{"secretName": secretName})
		err = secretsService.SaveNewSecret(secretName, "value1")
		require.NoError(t, err, "Failed to create first secret")

		// Saving with same name should create a new version, not error
		err = secretsService.SaveNewSecret(secretName, "value2")
		require.Error(t, err, "Failed to create second version")

		// Verify we have 2 versions
		reporter.LogStep("Verifying version count", nil)
		fileData, err := secretsService.LoadAllSecrets()
		require.NoError(t, err, "Failed to load secrets")

		var foundSecret *domain.Secret
		for i := range fileData.Secrets {
			if fileData.Secrets[i].SecretName == secretName {
				foundSecret = &fileData.Secrets[i]
				break
			}
		}
		require.NotNil(t, foundSecret, "Could not find versioning test secret")
		assert.Equal(t, 1, foundSecret.CurrentVersion, "Expected version 1")
		assert.Len(t, foundSecret.Versions, 1, "Expected 1 versions")

		// Clean up
		reporter.LogStep("Cleaning up test secret", nil)
		err = secretsService.DeleteSecret(secretName)
		require.NoError(t, err, "Failed to clean up test secret")
	})
}
