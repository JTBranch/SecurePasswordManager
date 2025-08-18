package integration

import (
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/pkg/reporting"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretsWorkflowIntegration(t *testing.T) {
	reporting.WithReporting(t, "TestSecretsWorkflowIntegration", func(reporter *reporting.TestWrapper) {
		suite := helpers.NewIntegrationTestSuite(reporter)
		suite.SetupTestEnvironment()
		defer suite.Cleanup()
		testCreateEditDeleteWorkflow(reporter, suite.SecretsService)
	})
}

func testCreateEditDeleteWorkflow(reporter *reporting.TestWrapper, secretsService *service.SecretsService) {
	// Initialize test data manager
	testDataManager := testdata.NewTestDataManager()
	require.NoError(reporter.T(), testDataManager.ValidateTestData(), "Test data validation failed")

	// Test 1: Create a secret using test data
	testSecret := testdata.TestSecrets.Simple
	secretName := testSecret.Name
	secretValue := testSecret.Value

	err := testDataManager.CreateTestSecret(secretsService, testSecret.Name)
	require.NoError(reporter.T(), err, "Failed to create secret")

	// Test 2: Load and verify secret
	fileData, err := secretsService.LoadAllSecrets()
	require.NoError(reporter.T(), err, "Failed to load secrets")

	require.NotEmpty(reporter.T(), fileData.Secrets, "Expected at least 1 secret, got 0")

	// Find our test secret
	var foundSecret *domain.Secret
	for i := range fileData.Secrets {
		if fileData.Secrets[i].SecretName == secretName {
			foundSecret = &fileData.Secrets[i]
			break
		}
	}

	require.NotNil(reporter.T(), foundSecret, "Could not find test secret '%s'", secretName)

	// Test 3: Display secret (decrypt)
	decrypted, err := secretsService.GetSecretValue(foundSecret)
	require.NoError(reporter.T(), err, "Failed to decrypt secret")

	assert.Equal(reporter.T(), secretValue, decrypted, "Expected secret value '%s', got '%s'", secretValue, decrypted)

	testEditSecretWorkflow(reporter, secretsService, secretName)
	testDeleteSecretWorkflow(reporter, secretsService, secretName)
}

func testEditSecretWorkflow(reporter *reporting.TestWrapper, secretsService *service.SecretsService, secretName string) {
	// Initialize test data manager for unique data
	testDataManager := testdata.NewTestDataManager()

	// Generate unique complex secret for editing
	uniqueComplexSecret, err := testDataManager.GenerateUniqueSimpleSecret("EditWorkflow")
	require.NoError(reporter.T(), err, "Failed to generate unique complex value")

	// Test 4: Edit secret (create new version) using unique data
	newValue := uniqueComplexSecret.Value

	err = secretsService.UpdateSecret(secretName, newValue)
	require.NoError(reporter.T(), err, "Failed to edit secret")

	// Test 5: Verify edit created new version
	fileData, err := secretsService.LoadAllSecrets()
	require.NoError(reporter.T(), err, "Failed to reload secrets after edit")

	// Find and verify our secret
	var foundSecret *domain.Secret
	for i := range fileData.Secrets {
		if fileData.Secrets[i].SecretName == secretName {
			foundSecret = &fileData.Secrets[i]
			break
		}
	}

	require.NotNil(reporter.T(), foundSecret, "Could not find secret after edit: %s", secretName)

	assert.GreaterOrEqual(reporter.T(), foundSecret.CurrentVersion, 2, "Expected version >= 2 after edit, got %d", foundSecret.CurrentVersion)

	// Test 6: Verify current value is updated
	currentValue, err := secretsService.GetSecretValue(foundSecret)
	require.NoError(reporter.T(), err, "Failed to display current secret value")

	assert.Equal(reporter.T(), newValue, currentValue, "Expected current value '%s', got '%s'", newValue, currentValue)
}

func testDeleteSecretWorkflow(reporter *reporting.TestWrapper, secretsService *service.SecretsService, secretName string) {
	// Test 7: Delete secret
	err := secretsService.DeleteSecret(secretName)
	require.NoError(reporter.T(), err, "Failed to delete secret")

	// Test 8: Verify deletion
	fileData, err := secretsService.LoadAllSecrets()
	require.NoError(reporter.T(), err, "Failed to reload secrets after deletion")

	// Ensure the deleted secret is not present
	for _, secret := range fileData.Secrets {
		assert.NotEqual(reporter.T(), secretName, secret.SecretName, "Found deleted secret '%s' in file", secretName)
	}
}

func TestErrorHandlingIntegration(t *testing.T) {
	reporting.WithReporting(t, "TestErrorHandlingIntegration", func(reporter *reporting.TestWrapper) {
		suite := helpers.NewIntegrationTestSuite(reporter)
		suite.SetupTestEnvironment()
		defer suite.Cleanup()
		testErrorHandling(reporter, suite.SecretsService)
	})
}

func testErrorHandling(reporter *reporting.TestWrapper, secretsService *service.SecretsService) {
	// Test 1: Load secrets when none exist
	fileData, err := secretsService.LoadAllSecrets()
	require.NoError(reporter.T(), err, "Should not fail when no secrets file exists")
	assert.Empty(reporter.T(), fileData.Secrets, "Expected no secrets, but found some")

	// Test 2: Delete a secret that doesn't exist
	err = secretsService.DeleteSecret("non-existent-secret")
	require.NoError(reporter.T(), err, "Deleting a non-existent secret should not produce an error")
}
