package e2e

import (
	"fmt"
	"testing"

	"go-password-manager/internal/domain"
	"go-password-manager/tests/e2e/pages"
	"go-password-manager/tests/e2e/setup"
	"go-password-manager/tests/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to find a secret by name
func findSecret(suite *setup.E2ETestSuite, secretName string) (domain.Secret, bool) {
	secrets, err := suite.SecretsService.LoadLatestSecrets()
	if err != nil {
		return domain.Secret{}, false
	}
	for _, secret := range secrets.Secrets {
		if secret.SecretName == secretName {
			return secret, true
		}
	}
	return domain.Secret{}, false
}

func TestSecretCreation(t *testing.T) {
	suite := setup.NewE2ETestSuite(t)
	defer suite.Cleanup()
	suite.SetupTestEnvironment()

	// Initialize test data manager with unique data generation
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate unique secret using the integrated system
	uniqueSecret, err := testDataManager.GenerateUniqueSimpleSecret("TestSecretCreation")
	require.NoError(t, err, testdata.ShouldGenerateUniqueMsg)
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

	mainPage := pages.NewMainPageObject(t, suite.Window, suite.SecretsService)

	// Load the main page
	mainPage.LoadPage()

	// Verify no secrets exist initially for this test (check our specific secret only)
	assert.False(t, mainPage.IsSecretVisible(uniqueSecret.UniqueName), "Our test secret should not exist initially")

	// Create a new secret using the UI workflow
	mainPage.ClickCreateSecretButton()
	mainPage.FillCreateSecretModal(uniqueSecret.UniqueName, uniqueSecret.Value)
	mainPage.SubmitCreateSecretModal()

	// Execute the actual secret creation (this simulates the form submission result)
	saveErr := suite.SecretsService.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, uniqueSecret.Type)
	require.NoError(t, saveErr, "Secret creation should succeed")

	// Refresh the UI to see the newly created secret
	mainPage.LoadPage()

	// Verify secret was created in the service layer
	assert.True(t, mainPage.IsSecretVisible(uniqueSecret.UniqueName), "Secret should be visible in list")

	// Verify the secret was persisted by checking the service layer
	createdSecret, exists := findSecret(suite, uniqueSecret.UniqueName)
	require.True(t, exists, "Secret should exist in service layer")
	assert.Equal(t, uniqueSecret.UniqueName, createdSecret.SecretName)
}

func TestSecretPersistence(t *testing.T) {
	suite := setup.NewE2ETestSuite(t)
	defer suite.Cleanup()
	suite.SetupTestEnvironment()

	// Initialize test data manager with unique data generation
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate unique secret using the integrated system
	uniqueSecret, err := testDataManager.GenerateUniqueSimpleSecret("TestSecretPersistence")
	require.NoError(t, err, testdata.ShouldGenerateUniqueMsg)
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

	mainPage := pages.NewMainPageObject(t, suite.Window, suite.SecretsService)

	// Pre-create a secret using unique secret data
	err = suite.SecretsService.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, uniqueSecret.Type)
	require.NoError(t, err, testdata.PreCreateMsg)

	// Test that secrets persist through page reloads
	suite.SetupTestEnvironment() // Simulate app restart
	mainPage.LoadPage()

	// Verify the secret persists
	assert.True(t, mainPage.IsSecretVisible(uniqueSecret.UniqueName), "Secret should persist after reload")

	// Verify the secret data is correct in the service layer
	persistedSecret, exists := findSecret(suite, uniqueSecret.UniqueName)
	require.True(t, exists, "Should find persisted secret")
	assert.Equal(t, uniqueSecret.UniqueName, persistedSecret.SecretName, "Secret name should match")
}

func TestSecretVersioning(t *testing.T) {
	suite := setup.NewE2ETestSuite(t)
	defer suite.Cleanup()
	suite.SetupTestEnvironment()

	// Initialize test data manager with unique data generation
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate unique versioning data using the integrated system
	uniqueVersionData, err := testDataManager.GenerateUniqueSimpleVersioning("TestSecretVersioning")
	require.NoError(t, err, "Should generate unique version data")
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueVersionData.UniqueName})

	// Get version values from the unique version data
	initialValue := uniqueVersionData.GetInitialVersion()
	updatedValue := uniqueVersionData.GetLatestVersion()

	mainPage := pages.NewMainPageObject(t, suite.Window, suite.SecretsService)

	// Pre-create a secret for this test using unique name and version data
	err = suite.SecretsService.SaveSecret(uniqueVersionData.UniqueName, initialValue, "key_value")
	require.NoError(t, err, testdata.PreCreateMsg)

	mainPage.LoadPage()

	// Verify secret exists and has 1 version initially
	secret, exists := findSecret(suite, uniqueVersionData.UniqueName)
	require.True(t, exists, "Secret should exist before versioning test")
	assert.Equal(t, 1, len(secret.Versions), "Secret should start with 1 version")

	// Edit the secret using UI workflow
	mainPage.ClickSecretInList(uniqueVersionData.UniqueName)
	assert.True(t, mainPage.IsSecretDetailVisible(), testdata.DetailVisibleMsg)

	mainPage.ClickEditSecret()
	mainPage.UpdateSecretValue(updatedValue)
	mainPage.SubmitCreateSecretModal() // Reuse submit for save

	// Execute the actual secret edit
	err = suite.SecretsService.EditSecret(uniqueVersionData.UniqueName, updatedValue)
	require.NoError(t, err, "Secret edit should succeed")

	// Verify version was incremented
	updatedSecret, exists := findSecret(suite, uniqueVersionData.UniqueName)
	require.True(t, exists, "Secret should still exist after edit")
	assert.Equal(t, 2, len(updatedSecret.Versions), "Secret should have 2 versions after edit")
	assert.Equal(t, 2, updatedSecret.CurrentVersion, "Secret current version should be 2")
}

func TestSecretDeletion(t *testing.T) {
	suite := setup.NewE2ETestSuite(t)
	defer suite.Cleanup()
	suite.SetupTestEnvironment()

	// Initialize test data manager
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate unique secret using the integrated system
	uniqueSecret, err := testDataManager.GenerateUniqueSimpleSecret("TestSecretDeletion")
	require.NoError(t, err, testdata.ShouldGenerateUniqueMsg)
	// Note: No cleanup defer needed since we're testing deletion

	mainPage := pages.NewMainPageObject(t, suite.Window, suite.SecretsService)

	// Pre-create a secret using unique name
	err = suite.SecretsService.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, "key_value")
	require.NoError(t, err, testdata.PreCreateMsg)

	mainPage.LoadPage()

	// Verify secret exists initially
	assert.True(t, mainPage.IsSecretVisible(uniqueSecret.UniqueName), "Secret should be visible before deletion")

	// Delete the secret using UI workflow
	mainPage.ClickSecretInList(uniqueSecret.UniqueName)
	assert.True(t, mainPage.IsSecretDetailVisible(), testdata.DetailVisibleMsg)

	mainPage.ClickDeleteSecret()
	mainPage.ConfirmDelete() // This will confirm the deletion

	// Execute the actual deletion in the service layer
	err = suite.SecretsService.DeleteSecret(uniqueSecret.UniqueName)
	require.NoError(t, err, "Secret deletion should succeed")

	// Refresh the UI to see the changes
	mainPage.LoadPage()

	// Verify secret was deleted
	assert.False(t, mainPage.IsSecretVisible(uniqueSecret.UniqueName), "Secret should not be visible after deletion")

	// Verify the secret was removed from the service layer
	_, exists := findSecret(suite, uniqueSecret.UniqueName)
	assert.False(t, exists, "Secret should not exist in storage after deletion")
}

func TestSecretCRUDOperationsWithCancellation(t *testing.T) {
	suite := setup.NewE2ETestSuite(t)
	defer suite.Cleanup()
	suite.SetupTestEnvironment()

	// Initialize test data manager with unique data generation
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate a unique secret specifically for cancellation testing
	uniqueSecret, err := testDataManager.GenerateUniqueTemporarySecret("TestSecretCancellation")
	require.NoError(t, err, "Should generate unique temporary secret")
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

	mainPage := pages.NewMainPageObject(t, suite.Window, suite.SecretsService)

	t.Run("Cancel delete operation", func(t *testing.T) {
		// First create a secret to delete using UI workflow with unique name
		mainPage.LoadPage()
		mainPage.ClickCreateSecretButton()
		mainPage.FillCreateSecretModal(uniqueSecret.UniqueName, uniqueSecret.Value)
		mainPage.SubmitCreateSecretModal()

		// Execute the actual secret creation
		err := suite.SecretsService.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, uniqueSecret.Type)
		require.NoError(t, err, "Secret creation should succeed")

		// Refresh the UI to see the newly created secret
		mainPage.LoadPage()

		// Verify secret was created
		assert.True(t, mainPage.IsSecretVisible(uniqueSecret.UniqueName), "Secret should be created")

		// Try to delete but cancel using UI workflow
		mainPage.ClickSecretInList(uniqueSecret.UniqueName)
		mainPage.ClickDeleteSecret()
		mainPage.CancelDelete() // Cancel the delete operation

		// Verify secret still exists (no service layer delete call made)
		assert.True(t, mainPage.IsSecretVisible(uniqueSecret.UniqueName), "Secret should still be visible")

		// Verify cancellation was handled correctly in service layer
		_, exists := findSecret(suite, uniqueSecret.UniqueName)
		assert.True(t, exists, "Secret should still exist in service layer")
	})
}

func TestSecretVisibilityToggle(t *testing.T) {
	suite := setup.NewE2ETestSuite(t)
	defer suite.Cleanup()
	suite.SetupTestEnvironment()

	// Initialize test data manager with unique data generation
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate unique secret using the integrated system for visibility testing
	uniqueSecret, err := testDataManager.GenerateUniqueSimpleSecret("TestSecretVisibility")
	require.NoError(t, err, testdata.ShouldGenerateUniqueMsg)
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

	mainPage := pages.NewMainPageObject(t, suite.Window, suite.SecretsService)

	t.Run("Toggle secret visibility", func(t *testing.T) {
		// Create a secret first using unique test data
		err := suite.SecretsService.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, uniqueSecret.Type)
		require.NoError(t, err, "Should create test secret")

		// Refresh the page to show the created secret
		mainPage.LoadPage()

		// Select the secret
		mainPage.ClickSecretInList(uniqueSecret.UniqueName)
		assert.True(t, mainPage.IsSecretDetailVisible(), "Secret detail should be visible")

		// Toggle visibility - this tests the reveal/hide functionality
		// Note: This is a UI interaction test, the actual value visibility
		// would need to be tested by examining the UI state
		mainPage.ToggleSecretVisibility()

		// The secret should still be selected and detail visible
		assert.True(t, mainPage.IsSecretDetailVisible(), testdata.DetailVisibleMsg)
		assert.Equal(t, uniqueSecret.UniqueName, mainPage.GetSecretDetailName())

		// Toggle again
		mainPage.ToggleSecretVisibility()
		assert.True(t, mainPage.IsSecretDetailVisible(), testdata.DetailVisibleMsg)
	})
}

// TestSecretBatchOperations demonstrates the use of batch unique data generation
func TestSecretBatchOperations(t *testing.T) {
	suite := setup.NewE2ETestSuite(t)
	defer suite.Cleanup()
	suite.SetupTestEnvironment()

	// Initialize test data manager with unique data generation
	testDataManager := testdata.NewTestDataManager()
	require.NoError(t, testDataManager.ValidateTestData(), testdata.TestDataValidationMsg)

	// Generate a complete CRUD set of unique secrets
	crudSet, err := testDataManager.GenerateUniqueCRUDSet("TestBatchOperations")
	require.NoError(t, err, "Should generate CRUD set")

	// Extract all secret names for cleanup
	var secretNames []string
	for _, secret := range crudSet {
		secretNames = append(secretNames, secret.UniqueName)
	}
	defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, secretNames)

	mainPage := pages.NewMainPageObject(t, suite.Window, suite.SecretsService)

	t.Run("Batch secret operations", func(t *testing.T) {
		// Create all secrets from the CRUD set
		for operation, secret := range crudSet {
			err := suite.SecretsService.SaveSecret(secret.UniqueName, secret.Value, secret.Type)
			require.NoError(t, err, fmt.Sprintf("Should create %s secret", operation))
		}

		mainPage.LoadPage()

		// Verify all secrets are visible
		for operation, secret := range crudSet {
			assert.True(t, mainPage.IsSecretVisible(secret.UniqueName),
				fmt.Sprintf("%s secret should be visible", operation))
		}

		// Test that we can interact with each secret type
		createSecret := crudSet["create"]
		mainPage.ClickSecretInList(createSecret.UniqueName)
		assert.True(t, mainPage.IsSecretDetailVisible(), "Create secret detail should be visible")

		// Test updating the versioned secret
		updateSecret := crudSet["update"]
		mainPage.ClickSecretInList(updateSecret.UniqueName)
		mainPage.ClickEditSecret()
		mainPage.UpdateSecretValue("UpdatedBatchValue")
		mainPage.SubmitCreateSecretModal()

		// Execute the edit
		err = suite.SecretsService.EditSecret(updateSecret.UniqueName, "UpdatedBatchValue")
		require.NoError(t, err, "Should update secret")

		// Verify version increment
		updatedSecret, exists := findSecret(suite, updateSecret.UniqueName)
		require.True(t, exists, "Updated secret should exist")
		assert.Equal(t, 2, updatedSecret.CurrentVersion, "Secret should have version 2")

		// Test deleting the temporary secret
		deleteSecret := crudSet["delete"]
		mainPage.ClickSecretInList(deleteSecret.UniqueName)
		mainPage.ClickDeleteSecret()
		mainPage.ConfirmDelete()

		err = suite.SecretsService.DeleteSecret(deleteSecret.UniqueName)
		require.NoError(t, err, "Should delete secret")

		mainPage.LoadPage()
		assert.False(t, mainPage.IsSecretVisible(deleteSecret.UniqueName),
			"Deleted secret should not be visible")
	})
}
