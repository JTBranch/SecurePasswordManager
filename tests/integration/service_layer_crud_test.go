package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-password-manager/internal/env"
	"go-password-manager/internal/service"
	"go-password-manager/tests/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite holds the test environment setup for service layer testing
type IntegrationTestSuite struct {
	testDataDir    string
	originalEnv    string
	SecretsService *service.SecretsService
	t              *testing.T
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{t: t}
	return suite
}

// SetupTestEnvironment creates an isolated test environment for integration testing
func (suite *IntegrationTestSuite) SetupTestEnvironment() {
	// Only create a new test directory if one hasn't been set
	if suite.testDataDir == "" {
		// Create isolated test environment
		testDir := filepath.Join(os.TempDir(), fmt.Sprintf("go-password-manager-integration-%d", time.Now().UnixNano()))
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			suite.t.Fatalf("Failed to create test directory: %v", err)
		}
		suite.testDataDir = testDir
		suite.t.Logf("Integration test environment created at: %s", testDir)
	} else {
		suite.t.Logf("Integration test environment reusing directory: %s", suite.testDataDir)
	}

	// Set environment to use test directory
	suite.originalEnv = os.Getenv("GO_PASSWORD_MANAGER_ENV")
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "integration-test")
	os.Setenv("TEST_DATA_DIR", suite.testDataDir)

	// Reset global environment config to pick up test settings
	env.Load()

	// Initialize secrets service with test configuration
	suite.SecretsService = service.NewSecretsService("1.0.0-integration", "integration-test-user")
}

// SetTestDataDir sets the test data directory (for reusing existing test data)
func (suite *IntegrationTestSuite) SetTestDataDir(dataDir string) {
	suite.testDataDir = dataDir
	os.Setenv("TEST_DATA_DIR", dataDir)
	env.Load()
}

// GetTestDataDir returns the test data directory path
func (suite *IntegrationTestSuite) GetTestDataDir() string {
	return suite.testDataDir
}

// GetSecretsFilePath returns the path to the secrets file
func (suite *IntegrationTestSuite) GetSecretsFilePath() string {
	return filepath.Join(suite.testDataDir, "secrets.json")
}

// Cleanup cleans up the integration test environment
func (suite *IntegrationTestSuite) Cleanup() {
	// Restore original environment
	if suite.originalEnv != "" {
		os.Setenv("GO_PASSWORD_MANAGER_ENV", suite.originalEnv)
	} else {
		os.Unsetenv("GO_PASSWORD_MANAGER_ENV")
	}
	os.Unsetenv("TEST_DATA_DIR")

	// Reload environment configuration to reset to defaults
	env.Load()

	// Clean up test directory
	err := os.RemoveAll(suite.testDataDir)
	if err != nil {
		suite.t.Logf("Warning: Failed to clean up test directory %s: %v", suite.testDataDir, err)
	}
}

func TestSecretCRUDOperationsServiceLayer(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	suite.SetupTestEnvironment()

	// Initialize test data manager
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate unique secret for consistent testing
	uniqueSecret, err := testDataManager.GenerateUniqueSimpleSecret("ServiceLayerCRUD")
	require.NoError(t, err, "Should generate unique secret")
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

	var testSecretName string

	t.Run("Create a secret on first load", func(t *testing.T) {
		// Verify no secrets exist initially
		secrets, err := suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		initialCount := len(secrets.Secrets)
		assert.Equal(t, 0, initialCount, "Should start with no secrets")

		// Create a new secret using unique data
		testSecretName = uniqueSecret.UniqueName
		err = suite.SecretsService.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, "key_value")
		require.NoError(t, err, "Should be able to create secret")

		// Verify secret was created
		secrets, err = suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		assert.Len(t, secrets.Secrets, 1, "Should have one secret after creation")
		assert.Equal(t, testSecretName, secrets.Secrets[0].SecretName)
		assert.Equal(t, 1, secrets.Secrets[0].CurrentVersion)
	})

	t.Run("Existing secrets show up on second load", func(t *testing.T) {
		// Simulate app restart by creating a new suite with the same data directory
		suite2 := NewIntegrationTestSuite(t)
		suite2.SetTestDataDir(suite.GetTestDataDir()) // Use same data directory
		suite2.SetupTestEnvironment()                 // Initialize with the shared data directory
		// Note: No cleanup here as we're sharing the test data directory

		// Verify the previously created secret is still there
		secrets, err := suite2.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		secretCount := len(secrets.Secrets)
		assert.Equal(t, 1, secretCount, "Secret should persist between app restarts")

		foundSecret := false
		for _, secret := range secrets.Secrets {
			if secret.SecretName == testSecretName {
				foundSecret = true
				break
			}
		}
		assert.True(t, foundSecret, "Previously created secret should be visible")

		// Verify we can decrypt the secret
		secret := secrets.Secrets[0]
		decryptedValue, err := suite2.SecretsService.DisplaySecret(secret)
		require.NoError(t, err)
		assert.Equal(t, uniqueSecret.Value, decryptedValue)
	})

	t.Run("Versioning", func(t *testing.T) {
		// Load current secrets from the original suite
		secrets, err := suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		require.Len(t, secrets.Secrets, 1, "Should have one secret")

		secret := secrets.Secrets[0]
		secretName := secret.SecretName

		// Verify initial version count
		initialVersions := len(secret.Versions)
		assert.Equal(t, 1, initialVersions, "Secret should start with 1 version")

		// Edit the secret to create a new version
		newSecretValue := "UpdatedSuperSecretPassword456"
		err = suite.SecretsService.EditSecret(secretName, newSecretValue)
		require.NoError(t, err, "Should be able to edit secret")

		// Verify version count increased
		secrets, err = suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)

		updatedSecret := secrets.Secrets[0]
		updatedVersions := len(updatedSecret.Versions)
		assert.Equal(t, 2, updatedVersions, "Secret should have 2 versions after edit")
		assert.Equal(t, 2, updatedSecret.CurrentVersion, "Current version should be 2")

		// Verify the latest version has the updated value
		latestVersion := updatedSecret.Versions[len(updatedSecret.Versions)-1]
		decryptedValue, err := suite.SecretsService.DecryptSecretVersion(latestVersion)
		require.NoError(t, err)
		assert.Equal(t, newSecretValue, decryptedValue, "Latest version should have updated value")

		// Verify we can still access the old version
		firstVersion := updatedSecret.Versions[0]
		oldDecryptedValue, err := suite.SecretsService.DecryptSecretVersion(firstVersion)
		require.NoError(t, err)
		assert.Equal(t, uniqueSecret.Value, oldDecryptedValue, "First version should have original value")
	})

	t.Run("Delete a secret", func(t *testing.T) {
		// Verify secret exists before deletion
		secrets, err := suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		initialCount := len(secrets.Secrets)
		assert.Equal(t, 1, initialCount, "Should have 1 secret before deletion")

		secretName := secrets.Secrets[0].SecretName

		// Delete the secret
		err = suite.SecretsService.DeleteSecret(secretName)
		require.NoError(t, err, "Should be able to delete secret")

		// Verify secret was deleted
		secrets, err = suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		finalCount := len(secrets.Secrets)
		assert.Equal(t, 0, finalCount, "Should have 0 secrets after deletion")

		// Verify the secrets file was actually updated
		secretsFilePath := suite.GetSecretsFilePath()
		_, err = os.Stat(secretsFilePath)
		if err == nil {
			// File exists, check it contains no secrets
			secrets, err := suite.SecretsService.LoadLatestSecrets()
			require.NoError(t, err)
			assert.Len(t, secrets.Secrets, 0, "Secrets file should contain no secrets")
		} else {
			// File doesn't exist, which is also valid for no secrets
			assert.True(t, os.IsNotExist(err), "Secrets file should either not exist or be empty")
		}
	})
}

func TestSecretCRUDOperationsErrorHandling(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	defer suite.Cleanup()

	suite.SetupTestEnvironment()

	// Initialize test data manager
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate unique secrets for multiple secret management
	uniqueSecrets, err := testDataManager.GenerateUniqueCRUDSet("ErrorHandling")
	require.NoError(t, err, "Should generate unique CRUD set")

	var secretNames []string
	for _, secret := range uniqueSecrets {
		secretNames = append(secretNames, secret.UniqueName)
	}
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, secretNames)

	t.Run("Error handling for non-existent secret operations", func(t *testing.T) {
		// Try to edit a non-existent secret
		err := suite.SecretsService.EditSecret("NonExistentSecret", "NewValue")
		assert.Error(t, err, "Should return error for non-existent secret")

		// Try to delete a non-existent secret
		err = suite.SecretsService.DeleteSecret("NonExistentSecret")
		assert.NoError(t, err, "Delete should not error for non-existent secret (idempotent)")

		// Verify no secrets were created during error scenarios
		secrets, err := suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		assert.Len(t, secrets.Secrets, 0, "No secrets should exist after error scenarios")
	})

	t.Run("Multiple secrets management", func(t *testing.T) {
		// Create multiple secrets using unique data

		for _, secret := range uniqueSecrets {
			err := suite.SecretsService.SaveSecret(secret.UniqueName, secret.Value, "key_value")
			require.NoError(t, err, "Should be able to create secret %s", secret.UniqueName)
		}

		// Verify all secrets were created
		secrets, err := suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		assert.Len(t, secrets.Secrets, 4, "Should have 4 secrets (create, read, update, delete)")

		// Verify each secret has correct content
		secretMap := make(map[string]string)
		for _, secret := range secrets.Secrets {
			decrypted, err := suite.SecretsService.DisplaySecret(secret)
			require.NoError(t, err)
			secretMap[secret.SecretName] = decrypted
		}

		for _, uniqueSecret := range uniqueSecrets {
			assert.Equal(t, uniqueSecret.Value, secretMap[uniqueSecret.UniqueName], "Secret %s should have correct value", uniqueSecret.UniqueName)
		}

		// Delete one secret and verify others remain
		var deletedSecretName string
		for _, secret := range uniqueSecrets {
			deletedSecretName = secret.UniqueName
			break // Get the first secret to delete
		}
		err = suite.SecretsService.DeleteSecret(deletedSecretName)
		require.NoError(t, err)

		secrets, err = suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		assert.Len(t, secrets.Secrets, 3, "Should have 3 secrets after deleting one")

		// Verify the correct secret was deleted
		remainingNames := make([]string, 0, 2)
		for _, secret := range secrets.Secrets {
			remainingNames = append(remainingNames, secret.SecretName)
		}

		// Verify we have the correct number of remaining secrets
		assert.NotContains(t, remainingNames, deletedSecretName, "Deleted secret should not be present")

		// Verify the other secrets still exist
		remainingCount := 0
		for _, uniqueSecret := range uniqueSecrets {
			if uniqueSecret.UniqueName != deletedSecretName {
				assert.Contains(t, remainingNames, uniqueSecret.UniqueName, "Remaining secret should be present")
				remainingCount++
			}
		}
		assert.Equal(t, 3, remainingCount, "Should have exactly 3 remaining secrets")
	})
}
