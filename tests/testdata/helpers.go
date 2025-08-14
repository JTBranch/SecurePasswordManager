// Package helpers provides utility functions for working with immutable test data
package testdata

import (
	"fmt"
	"go-password-manager/internal/service"
)

// Error constants
const (
	ErrNotInitialized = "test data manager not initialized"
)

// TestDataManager provides safe access to test data with validation
type TestDataManager struct {
	// Private fields to prevent external mutation
	initialized     bool
	uniqueGenerator *UniqueDataGenerator
}

// NewTestDataManager creates a new test data manager
func NewTestDataManager() *TestDataManager {
	return &TestDataManager{
		initialized:     true,
		uniqueGenerator: NewUniqueDataGenerator(),
	}
}

// Validation methods

// ValidateTestData performs comprehensive validation of all test data
func (tm *TestDataManager) ValidateTestData() error {
	if !tm.initialized {
		return fmt.Errorf(ErrNotInitialized)
	}

	// Validate users
	if err := tm.validateUsers(); err != nil {
		return err
	}

	// Validate secrets
	if err := tm.validateSecrets(); err != nil {
		return err
	}

	// Validate scenarios
	return tm.validateScenarios()
}

// validateUsers validates all test users
func (tm *TestDataManager) validateUsers() error {
	users := []struct {
		user TestUser
		name string
	}{
		{TestUsers.E2EUser, "E2EUser"},
		{TestUsers.IntegrationUser, "IntegrationUser"},
		{TestUsers.UnitTestUser, "UnitTestUser"},
	}

	for _, u := range users {
		if !u.user.IsValid() {
			return fmt.Errorf("%s test data is invalid", u.name)
		}
	}
	return nil
}

// validateSecrets validates all test secrets
func (tm *TestDataManager) validateSecrets() error {
	secrets := []struct {
		secret TestSecret
		name   string
	}{
		{TestSecrets.Simple, "Simple"},
		{TestSecrets.Complex, "Complex"},
		{TestSecrets.Long, "Long"},
		{TestSecrets.Special, "Special"},
		{TestSecrets.Versioned, "Versioned"},
		{TestSecrets.Temporary, "Temporary"},
	}

	for _, s := range secrets {
		if !s.secret.IsValid() {
			return fmt.Errorf("%s test secret is invalid", s.name)
		}
	}
	return nil
}

// validateScenarios validates all test scenarios
func (tm *TestDataManager) validateScenarios() error {
	scenarios := []struct {
		scenario TestScenario
		name     string
	}{
		{TestScenarios.Basic, "Basic"},
		{TestScenarios.CRUD, "CRUD"},
		{TestScenarios.Versioning, "Versioning"},
		{TestScenarios.Visibility, "Visibility"},
		{TestScenarios.Search, "Search"},
		{TestScenarios.Persistence, "Persistence"},
	}

	for _, sc := range scenarios {
		if !sc.scenario.IsValid() {
			return fmt.Errorf("%s test scenario is invalid", sc.name)
		}
	}
	return nil
}

// Safe accessors that return clones

// GetTestSecret returns a clone of a test secret by name
func (tm *TestDataManager) GetTestSecret(name string) (TestSecret, error) {
	if !tm.initialized {
		return TestSecret{}, fmt.Errorf(ErrNotInitialized)
	}

	switch name {
	case TestSecrets.Simple.Name:
		return TestSecrets.Simple.CloneSecret(), nil
	case TestSecrets.Complex.Name:
		return TestSecrets.Complex.CloneSecret(), nil
	case TestSecrets.Long.Name:
		return TestSecrets.Long.CloneSecret(), nil
	case TestSecrets.Special.Name:
		return TestSecrets.Special.CloneSecret(), nil
	case TestSecrets.Versioned.Name:
		return TestSecrets.Versioned.CloneSecret(), nil
	case TestSecrets.Temporary.Name:
		return TestSecrets.Temporary.CloneSecret(), nil
	default:
		return TestSecret{}, fmt.Errorf("test secret '%s' not found", name)
	}
}

// GetTestScenario returns a clone of a test scenario by name
func (tm *TestDataManager) GetTestScenario(name string) (TestScenario, error) {
	if !tm.initialized {
		return TestScenario{}, fmt.Errorf(ErrNotInitialized)
	}

	switch name {
	case TestScenarios.Basic.Name:
		return TestScenarios.Basic.CloneScenario(), nil
	case TestScenarios.CRUD.Name:
		return TestScenarios.CRUD.CloneScenario(), nil
	case TestScenarios.Versioning.Name:
		return TestScenarios.Versioning.CloneScenario(), nil
	case TestScenarios.Visibility.Name:
		return TestScenarios.Visibility.CloneScenario(), nil
	case TestScenarios.Search.Name:
		return TestScenarios.Search.CloneScenario(), nil
	case TestScenarios.Persistence.Name:
		return TestScenarios.Persistence.CloneScenario(), nil
	default:
		return TestScenario{}, fmt.Errorf("test scenario '%s' not found", name)
	}
}

// GetTestUser returns a clone of a test user by environment
func (tm *TestDataManager) GetTestUser(environment string) (TestUser, error) {
	if !tm.initialized {
		return TestUser{}, fmt.Errorf(ErrNotInitialized)
	}

	switch environment {
	case "e2e-test":
		return TestUsers.E2EUser.CloneUser(), nil
	case "integration-test":
		return TestUsers.IntegrationUser.CloneUser(), nil
	case "test":
		return TestUsers.UnitTestUser.CloneUser(), nil
	default:
		return TestUser{}, fmt.Errorf("test user for environment '%s' not found", environment)
	}
}

// Service integration helpers

// CreateTestSecret creates a test secret in the service using test data
func (tm *TestDataManager) CreateTestSecret(service *service.SecretsService, secretName string) error {
	if !tm.initialized {
		return fmt.Errorf(ErrNotInitialized)
	}

	testSecret, err := tm.GetTestSecret(secretName)
	if err != nil {
		return fmt.Errorf("failed to get test secret: %w", err)
	}

	return service.SaveSecret(testSecret.Name, testSecret.Value, testSecret.Type)
}

// CreateTestSecrets creates multiple test secrets in the service
func (tm *TestDataManager) CreateTestSecrets(service *service.SecretsService, secretNames []string) error {
	if !tm.initialized {
		return fmt.Errorf(ErrNotInitialized)
	}

	for _, name := range secretNames {
		if err := tm.CreateTestSecret(service, name); err != nil {
			return fmt.Errorf("failed to create test secret '%s': %w", name, err)
		}
	}

	return nil
}

// CreateScenarioSecrets creates all secrets from a test scenario
func (tm *TestDataManager) CreateScenarioSecrets(service *service.SecretsService, scenarioName string) error {
	if !tm.initialized {
		return fmt.Errorf(ErrNotInitialized)
	}

	scenario, err := tm.GetTestScenario(scenarioName)
	if err != nil {
		return fmt.Errorf("failed to get test scenario: %w", err)
	}

	for _, secret := range scenario.Secrets {
		if err := service.SaveSecret(secret.Name, secret.Value, secret.Type); err != nil {
			return fmt.Errorf("failed to create secret '%s' from scenario: %w", secret.Name, err)
		}
	}

	return nil
}

// Utility methods

// ListAvailableSecrets returns a list of all available test secret names
func (tm *TestDataManager) ListAvailableSecrets() []string {
	if !tm.initialized {
		return nil
	}

	return []string{
		TestSecrets.Simple.Name,
		TestSecrets.Complex.Name,
		TestSecrets.Long.Name,
		TestSecrets.Special.Name,
		TestSecrets.Versioned.Name,
		TestSecrets.Temporary.Name,
	}
}

// ListAvailableScenarios returns a list of all available test scenario names
func (tm *TestDataManager) ListAvailableScenarios() []string {
	if !tm.initialized {
		return nil
	}

	return []string{
		TestScenarios.Basic.Name,
		TestScenarios.CRUD.Name,
		TestScenarios.Versioning.Name,
		TestScenarios.Visibility.Name,
		TestScenarios.Search.Name,
		TestScenarios.Persistence.Name,
	}
}

// GetTestSecretsByType returns all test secrets of a specific type
func (tm *TestDataManager) GetTestSecretsByType(secretType string) []TestSecret {
	if !tm.initialized {
		return nil
	}

	var secrets []TestSecret
	allSecrets := []TestSecret{
		TestSecrets.Simple,
		TestSecrets.Complex,
		TestSecrets.Long,
		TestSecrets.Special,
		TestSecrets.Versioned,
		TestSecrets.Temporary,
	}

	for _, secret := range allSecrets {
		if secret.Type == secretType {
			secrets = append(secrets, secret.CloneSecret())
		}
	}

	return secrets
}

// Unique data generation methods

// GenerateUniqueSecretName creates a unique secret name for test independence
func (tm *TestDataManager) GenerateUniqueSecretName(testName string) string {
	if !tm.initialized {
		return ""
	}
	return tm.uniqueGenerator.GenerateUniqueSecretName(testName)
}

// GenerateUniqueSecret creates a unique test secret based on a template
func (tm *TestDataManager) GenerateUniqueSecret(template TestSecret, testName string) (UniqueTestSecret, error) {
	if !tm.initialized {
		return UniqueTestSecret{}, fmt.Errorf(ErrNotInitialized)
	}
	return tm.uniqueGenerator.GenerateUniqueSecret(template, testName), nil
}

// GenerateUniqueSecretByName creates a unique test secret based on a predefined secret name
func (tm *TestDataManager) GenerateUniqueSecretByName(secretName, testName string) (UniqueTestSecret, error) {
	if !tm.initialized {
		return UniqueTestSecret{}, fmt.Errorf(ErrNotInitialized)
	}

	template, err := tm.GetTestSecret(secretName)
	if err != nil {
		return UniqueTestSecret{}, fmt.Errorf("failed to get template secret: %w", err)
	}

	return tm.uniqueGenerator.GenerateUniqueSecret(template, testName), nil
}

// GenerateUniqueScenario creates a unique test scenario with unique secrets
func (tm *TestDataManager) GenerateUniqueScenario(scenarioName, testName string) (UniqueTestScenario, error) {
	if !tm.initialized {
		return UniqueTestScenario{}, fmt.Errorf(ErrNotInitialized)
	}

	template, err := tm.GetTestScenario(scenarioName)
	if err != nil {
		return UniqueTestScenario{}, fmt.Errorf("failed to get template scenario: %w", err)
	}

	return tm.uniqueGenerator.GenerateUniqueScenario(template, testName), nil
}

// Predefined unique generators for convenience

// GenerateUniqueSimpleSecret creates a unique simple secret
func (tm *TestDataManager) GenerateUniqueSimpleSecret(testName string) (UniqueTestSecret, error) {
	return tm.GenerateUniqueSecretByName(TestSecrets.Simple.Name, testName)
}

// GenerateUniqueComplexSecret creates a unique complex secret
func (tm *TestDataManager) GenerateUniqueComplexSecret(testName string) (UniqueTestSecret, error) {
	return tm.GenerateUniqueSecretByName(TestSecrets.Complex.Name, testName)
}

// GenerateUniqueTemporarySecret creates a unique temporary secret
func (tm *TestDataManager) GenerateUniqueTemporarySecret(testName string) (UniqueTestSecret, error) {
	return tm.GenerateUniqueSecretByName(TestSecrets.Temporary.Name, testName)
}

// GenerateUniqueVersionedSecret creates a unique versioned secret
func (tm *TestDataManager) GenerateUniqueVersionedSecret(testName string) (UniqueTestSecret, error) {
	return tm.GenerateUniqueSecretByName(TestSecrets.Versioned.Name, testName)
}

// Versioning support

// GenerateUniqueVersionTestData creates unique version test data
func (tm *TestDataManager) GenerateUniqueVersionTestData(templateName, testName string) (UniqueVersionTestData, error) {
	if !tm.initialized {
		return UniqueVersionTestData{}, fmt.Errorf(ErrNotInitialized)
	}

	var template VersionTestData
	switch templateName {
	case "SimpleVersioning":
		template = VersioningTestData.SimpleVersioning
	case "MultipleVersions":
		template = VersioningTestData.MultipleVersions
	case "LongVersionHistory":
		template = VersioningTestData.LongVersionHistory
	default:
		return UniqueVersionTestData{}, fmt.Errorf("version template '%s' not found", templateName)
	}

	return tm.uniqueGenerator.GenerateUniqueVersionTestData(template, testName), nil
}

// GenerateUniqueSimpleVersioning creates unique simple versioning data
func (tm *TestDataManager) GenerateUniqueSimpleVersioning(testName string) (UniqueVersionTestData, error) {
	return tm.GenerateUniqueVersionTestData("SimpleVersioning", testName)
}

// Batch generation methods

// GenerateUniqueSecretSet generates a set of unique secrets for comprehensive testing
func (tm *TestDataManager) GenerateUniqueSecretSet(testName string) (map[string]UniqueTestSecret, error) {
	if !tm.initialized {
		return nil, fmt.Errorf(ErrNotInitialized)
	}
	return tm.uniqueGenerator.GenerateUniqueSecretSet(testName), nil
}

// GenerateUniqueCRUDSet generates secrets specifically for CRUD testing
func (tm *TestDataManager) GenerateUniqueCRUDSet(testName string) (map[string]UniqueTestSecret, error) {
	if !tm.initialized {
		return nil, fmt.Errorf(ErrNotInitialized)
	}
	return tm.uniqueGenerator.GenerateUniqueCRUDSet(testName), nil
}

// Service integration with unique data

// CreateUniqueTestSecret creates a unique test secret in the service
func (tm *TestDataManager) CreateUniqueTestSecret(service *service.SecretsService, template TestSecret, testName string) (UniqueTestSecret, error) {
	if !tm.initialized {
		return UniqueTestSecret{}, fmt.Errorf(ErrNotInitialized)
	}

	uniqueSecret := tm.uniqueGenerator.GenerateUniqueSecret(template, testName)

	err := service.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, uniqueSecret.Type)
	if err != nil {
		return UniqueTestSecret{}, fmt.Errorf("failed to create unique test secret: %w", err)
	}

	return uniqueSecret, nil
}

// CreateUniqueTestSecretByName creates a unique test secret in the service using a predefined template
func (tm *TestDataManager) CreateUniqueTestSecretByName(service *service.SecretsService, secretName, testName string) (UniqueTestSecret, error) {
	template, err := tm.GetTestSecret(secretName)
	if err != nil {
		return UniqueTestSecret{}, fmt.Errorf("failed to get template secret: %w", err)
	}

	return tm.CreateUniqueTestSecret(service, template, testName)
}

// CleanupUniqueSecrets removes unique test secrets from the service
func (tm *TestDataManager) CleanupUniqueSecrets(service *service.SecretsService, uniqueSecrets []UniqueTestSecret) error {
	if !tm.initialized {
		return fmt.Errorf(ErrNotInitialized)
	}

	for _, secret := range uniqueSecrets {
		if err := service.DeleteSecret(secret.UniqueName); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Warning: failed to cleanup secret '%s': %v\n", secret.UniqueName, err)
		}
	}

	return nil
}

// CleanupUniqueSecretNames removes unique test secrets by names from the service
func (tm *TestDataManager) CleanupUniqueSecretNames(service *service.SecretsService, secretNames []string) error {
	if !tm.initialized {
		return fmt.Errorf(ErrNotInitialized)
	}

	for _, name := range secretNames {
		if err := service.DeleteSecret(name); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Warning: failed to cleanup secret '%s': %v\n", name, err)
		}
	}

	return nil
}
