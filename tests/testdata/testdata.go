// Package testdata provides immutable test data for all test suites.
// All data structures are read-only to prevent corruption during test execution.
package testdata

import (
	"go-password-manager/internal/domain"
	"time"
)

// TestSecret represents an immutable test secret
type TestSecret struct {
	Name        string
	Value       string
	Type        string
	Description string
}

// TestUser represents an immutable test user
type TestUser struct {
	Name        string
	Environment string
	AppVersion  string
}

// TestScenario represents a complete test scenario with immutable data
type TestScenario struct {
	Name        string
	Description string
	User        TestUser
	Secrets     []TestSecret
}

// TestSecretsCollection holds all predefined test secrets
type TestSecretsCollection struct {
	Simple    TestSecret
	Complex   TestSecret
	Long      TestSecret
	Special   TestSecret
	Versioned TestSecret
	Temporary TestSecret
}

// TestScenariosCollection holds all predefined test scenarios
type TestScenariosCollection struct {
	Basic       TestScenario
	CRUD        TestScenario
	Versioning  TestScenario
	Visibility  TestScenario
	Search      TestScenario
	Persistence TestScenario
}

const (
	TestEncryptionKey      = "12345678901234567890123456789012"
	DifferentEncryptionKey = "abcdefghijklmnopqrstuvwxyz123456"
)

var (
	// TestUsers provides immutable test user data
	TestUsers = struct {
		E2EUser         TestUser
		IntegrationUser TestUser
		UnitTestUser    TestUser
	}{
		E2EUser: TestUser{
			Name:        "e2e-test-user",
			Environment: "e2e-test",
			AppVersion:  "1.0.0-e2e",
		},
		IntegrationUser: TestUser{
			Name:        "integration-test-user",
			Environment: "integration-test",
			AppVersion:  "1.0.0-integration",
		},
		UnitTestUser: TestUser{
			Name:        "unit-test-user",
			Environment: "test",
			AppVersion:  "1.0.0-test",
		},
	}

	// TestSecrets provides immutable test secret data
	TestSecrets = TestSecretsCollection{
		Simple: TestSecret{
			Name:        "SimpleTestSecret",
			Value:       "SimplePassword123",
			Type:        "key_value",
			Description: "A basic test secret for simple scenarios",
		},
		Complex: TestSecret{
			Name:        "ComplexTestSecret",
			Value:       "C0mpl3x_P@ssw0rd_W1th_Sp3c1@l_Ch@rs!",
			Type:        "key_value",
			Description: "A complex test secret with special characters",
		},
		Long: TestSecret{
			Name:        "LongTestSecret",
			Value:       "ThisIsAVeryLongPasswordThatContainsMultipleWordsAndIsDesignedToTestLongValueHandling",
			Type:        "key_value",
			Description: "A long test secret for testing value length limits",
		},
		Special: TestSecret{
			Name:        "SpecialCharsSecret",
			Value:       "!@#$%^&*()_+-=[]{}|;':\",./<>?`~",
			Type:        "key_value",
			Description: "A secret containing all special characters",
		},
		Versioned: TestSecret{
			Name:        "VersionedTestSecret",
			Value:       "InitialVersion1",
			Type:        "key_value",
			Description: "A secret designed for version testing",
		},
		Temporary: TestSecret{
			Name:        "TempTestSecret",
			Value:       "TemporaryValue",
			Type:        "key_value",
			Description: "A temporary secret for deletion testing",
		},
	}

	// TestScenarios provides immutable test scenario data
	TestScenarios = TestScenariosCollection{
		Basic: TestScenario{
			Name:        "BasicSecretCreation",
			Description: "Basic secret creation and verification",
			User:        TestUsers.E2EUser,
			Secrets: []TestSecret{
				TestSecrets.Simple,
			},
		},
		CRUD: TestScenario{
			Name:        "CRUDOperations",
			Description: "Complete Create, Read, Update, Delete operations",
			User:        TestUsers.E2EUser,
			Secrets: []TestSecret{
				TestSecrets.Simple,
				TestSecrets.Complex,
				TestSecrets.Temporary,
			},
		},
		Versioning: TestScenario{
			Name:        "VersionManagement",
			Description: "Secret version creation and management",
			User:        TestUsers.E2EUser,
			Secrets: []TestSecret{
				TestSecrets.Versioned,
			},
		},
		Visibility: TestScenario{
			Name:        "VisibilityToggling",
			Description: "Secret visibility show/hide functionality",
			User:        TestUsers.E2EUser,
			Secrets: []TestSecret{
				TestSecrets.Simple,
			},
		},
		Search: TestScenario{
			Name:        "SearchFunctionality",
			Description: "Secret search and filtering capabilities",
			User:        TestUsers.E2EUser,
			Secrets: []TestSecret{
				TestSecrets.Simple,
				TestSecrets.Complex,
				TestSecrets.Long,
				TestSecrets.Special,
			},
		},
		Persistence: TestScenario{
			Name:        "DataPersistence",
			Description: "Data persistence across application restarts",
			User:        TestUsers.E2EUser,
			Secrets: []TestSecret{
				TestSecrets.Simple,
				TestSecrets.Complex,
			},
		},
	}
)

// Clone methods to ensure immutability - these return copies, not references

// CloneSecret returns a deep copy of a TestSecret to prevent mutations
func (ts TestSecret) CloneSecret() TestSecret {
	return TestSecret{
		Name:        ts.Name,
		Value:       ts.Value,
		Type:        ts.Type,
		Description: ts.Description,
	}
}

// CloneUser returns a deep copy of a TestUser to prevent mutations
func (tu TestUser) CloneUser() TestUser {
	return TestUser{
		Name:        tu.Name,
		Environment: tu.Environment,
		AppVersion:  tu.AppVersion,
	}
}

// CloneScenario returns a deep copy of a TestScenario to prevent mutations
func (ts TestScenario) CloneScenario() TestScenario {
	secrets := make([]TestSecret, len(ts.Secrets))
	for i, secret := range ts.Secrets {
		secrets[i] = secret.CloneSecret()
	}

	return TestScenario{
		Name:        ts.Name,
		Description: ts.Description,
		User:        ts.User.CloneUser(),
		Secrets:     secrets,
	}
}

// Conversion utilities to domain objects

// ToDomainSecret converts a TestSecret to a domain.Secret
func (ts TestSecret) ToDomainSecret() domain.Secret {
	now := time.Now()
	return domain.Secret{
		SecretName:     ts.Name,
		Type:           domain.SecretType(ts.Type),
		CurrentVersion: 1,
		Versions: []domain.SecretVersion{
			{
				SecretValueEnc: "", // Will be encrypted by service layer
				Version:        1,
				UpdatedAt:      now.Format(time.RFC3339),
			},
		},
	}
}

// Validation methods

// IsValid checks if a TestSecret has all required fields
func (ts TestSecret) IsValid() bool {
	return ts.Name != "" && ts.Value != "" && ts.Type != ""
}

// IsValid checks if a TestUser has all required fields
func (tu TestUser) IsValid() bool {
	return tu.Name != "" && tu.Environment != "" && tu.AppVersion != ""
}

// IsValid checks if a TestScenario has all required fields and valid data
func (ts TestScenario) IsValid() bool {
	if ts.Name == "" || ts.Description == "" || !ts.User.IsValid() {
		return false
	}

	for _, secret := range ts.Secrets {
		if !secret.IsValid() {
			return false
		}
	}

	return true
}

// Utility functions for test scenarios

// GetSecretByName returns a copy of a secret by name from a scenario
func (ts TestScenario) GetSecretByName(name string) (TestSecret, bool) {
	for _, secret := range ts.Secrets {
		if secret.Name == name {
			return secret.CloneSecret(), true
		}
	}
	return TestSecret{}, false
}

// GetSecretsCount returns the number of secrets in a scenario
func (ts TestScenario) GetSecretsCount() int {
	return len(ts.Secrets)
}

// GetSecretNames returns a slice of all secret names in a scenario
func (ts TestScenario) GetSecretNames() []string {
	names := make([]string, len(ts.Secrets))
	for i, secret := range ts.Secrets {
		names[i] = secret.Name
	}
	return names
}
