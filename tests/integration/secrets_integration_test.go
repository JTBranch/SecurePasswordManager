package integration

import (
	"go-password-manager/internal/domain"
	"go-password-manager/internal/env"
	"go-password-manager/internal/service"
	"go-password-manager/tests/testdata"
	"os"
	"testing"
)

func setupTestEnvironment(t *testing.T) {
	// Set integration test environment variables
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "integration-test")
	os.Setenv("TEST_DATA_DIR", t.TempDir())
	os.Setenv("APP_NAME", "GoPasswordManager-Integration")

	// Reset global config to pick up test environment
	env.Load()
}

func TestSecretsWorkflowIntegration(t *testing.T) {
	setupTestEnvironment(t)
	testCreateEditDeleteWorkflow(t)
}

func testCreateEditDeleteWorkflow(t *testing.T) {
	// Create secrets service using the standard constructor
	secretsService := service.NewSecretsService("1.0.0", "integration-test-user")

	// Initialize test data manager
	testDataManager := testdata.NewTestDataManager()
	if err := testDataManager.ValidateTestData(); err != nil {
		t.Fatalf("Test data validation failed: %v", err)
	}

	// Test 1: Create a secret using test data
	testSecret := testdata.TestSecrets.Simple
	secretName := testSecret.Name
	secretValue := testSecret.Value

	err := testDataManager.CreateTestSecret(secretsService, testSecret.Name)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Test 2: Load and verify secret
	fileData, err := secretsService.LoadLatestSecrets()
	if err != nil {
		t.Fatalf("Failed to load secrets: %v", err)
	}

	if len(fileData.Secrets) == 0 {
		t.Fatal("Expected at least 1 secret, got 0")
	}

	// Find our test secret
	var foundSecret *domain.Secret
	for i := range fileData.Secrets {
		if fileData.Secrets[i].SecretName == secretName {
			foundSecret = &fileData.Secrets[i]
			break
		}
	}

	if foundSecret == nil {
		t.Fatalf("Could not find test secret '%s'", secretName)
	}

	// Test 3: Display secret (decrypt)
	decrypted, err := secretsService.DisplaySecret(*foundSecret)
	if err != nil {
		t.Fatalf("Failed to decrypt secret: %v", err)
	}

	if decrypted != secretValue {
		t.Errorf("Expected secret value '%s', got '%s'", secretValue, decrypted)
	}

	testEditSecretWorkflow(t, secretsService, secretName)
	testDeleteSecretWorkflow(t, secretsService, secretName)
}

func testEditSecretWorkflow(t *testing.T, secretsService *service.SecretsService, secretName string) {
	// Initialize test data manager for unique data
	testDataManager := testdata.NewTestDataManager()

	// Generate unique complex secret for editing
	uniqueComplexSecret, err := testDataManager.GenerateUniqueSimpleSecret("EditWorkflow")
	if err != nil {
		t.Fatalf("Failed to generate unique complex value: %v", err)
	}

	// Test 4: Edit secret (create new version) using unique data
	newValue := uniqueComplexSecret.Value

	err = secretsService.EditSecret(secretName, newValue)
	if err != nil {
		t.Fatalf("Failed to edit secret: %v", err)
	}

	// Test 5: Verify edit created new version
	fileData, err := secretsService.LoadLatestSecrets()
	if err != nil {
		t.Fatalf("Failed to reload secrets after edit: %v", err)
	}

	// Find and verify our secret
	var foundSecret *domain.Secret
	for i := range fileData.Secrets {
		if fileData.Secrets[i].SecretName == secretName {
			foundSecret = &fileData.Secrets[i]
			break
		}
	}

	if foundSecret == nil {
		t.Fatalf("Could not find secret after edit: %s", secretName)
	}

	if foundSecret.CurrentVersion < 2 {
		t.Errorf("Expected version >= 2 after edit, got %d", foundSecret.CurrentVersion)
	}

	// Test 6: Verify current value is updated
	currentValue, err := secretsService.DisplaySecret(*foundSecret)
	if err != nil {
		t.Fatalf("Failed to display current secret value: %v", err)
	}

	if currentValue != newValue {
		t.Errorf("Expected current value '%s', got '%s'", newValue, currentValue)
	}
}

func testDeleteSecretWorkflow(t *testing.T, secretsService *service.SecretsService, secretName string) {
	// Test 7: Delete secret
	err := secretsService.DeleteSecret(secretName)
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	// Test 8: Verify deletion
	fileData, err := secretsService.LoadLatestSecrets()
	if err != nil {
		t.Fatalf("Failed to reload secrets after deletion: %v", err)
	}

	// Ensure the deleted secret is not present
	for _, secret := range fileData.Secrets {
		if secret.SecretName == secretName {
			t.Errorf("Found deleted secret '%s' in file", secretName)
		}
	}
}

func TestErrorHandlingIntegration(t *testing.T) {
	setupTestEnvironment(t)

	// Setup
	secretsService := service.NewSecretsService("1.0.0", "integration-test-user")

	// Test error handling - edit non-existent secret
	err := secretsService.EditSecret("non-existent-secret", "some-value")
	if err == nil {
		t.Error("Expected error when editing non-existent secret")
	}

	// Test error handling - delete non-existent secret (should not error but should be idempotent)
	err = secretsService.DeleteSecret("non-existent-secret")
	if err != nil {
		t.Errorf("Delete should be idempotent, got error: %v", err)
	}

	// Test that SaveSecret with same name creates new version (this is intended behavior)
	testDataManager := testdata.NewTestDataManager()

	// Generate unique secret for versioning test
	uniqueVersionSecret, err := testDataManager.GenerateUniqueSimpleSecret("ErrorHandlingVersioning")
	if err != nil {
		t.Fatalf("Failed to generate unique versioning secret: %v", err)
	}
	defer testDataManager.CleanupUniqueSecretNames(secretsService, []string{uniqueVersionSecret.UniqueName})

	secretName := uniqueVersionSecret.UniqueName

	// Generate unique values for the different versions
	firstVersionSecret, err := testDataManager.GenerateUniqueSimpleSecret("ErrorHandlingV1")
	if err != nil {
		t.Fatalf("Failed to generate first version value: %v", err)
	}

	secondVersionSecret, err := testDataManager.GenerateUniqueSimpleSecret("ErrorHandlingV2")
	if err != nil {
		t.Fatalf("Failed to generate second version value: %v", err)
	}

	err = secretsService.SaveSecret(secretName, firstVersionSecret.Value, "key_value")
	if err != nil {
		t.Fatalf("Failed to create first secret: %v", err)
	}

	// Saving with same name should create a new version, not error
	err = secretsService.SaveSecret(secretName, secondVersionSecret.Value, "key_value")
	if err != nil {
		t.Fatalf("Failed to create second version: %v", err)
	}

	// Verify we have 2 versions
	fileData, err := secretsService.LoadLatestSecrets()
	if err != nil {
		t.Fatalf("Failed to load secrets: %v", err)
	}

	var foundSecret *domain.Secret
	for i := range fileData.Secrets {
		if fileData.Secrets[i].SecretName == secretName {
			foundSecret = &fileData.Secrets[i]
			break
		}
	}

	if foundSecret == nil {
		t.Fatal("Could not find versioning test secret")
	}

	if foundSecret.CurrentVersion != 2 {
		t.Errorf("Expected version 2, got %d", foundSecret.CurrentVersion)
	}

	if len(foundSecret.Versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(foundSecret.Versions))
	}

	// Clean up
	err = secretsService.DeleteSecret(secretName)
	if err != nil {
		t.Fatalf("Failed to clean up test secret: %v", err)
	}
}
