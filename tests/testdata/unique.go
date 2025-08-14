// Package unique provides dynamic unique data generation for test independence
package testdata

import (
	"fmt"
	"time"
)

// Constants for unique data generation
const (
	UniqueDescriptionFormat = "%s (Unique: %s)"
	FallbackNameFormat      = "%s_fallback_%d"
	UniqueNameFormat        = "%s_%d"
	ScenarioNameFormat      = "%s_%s_%d"
)

// UniqueDataGenerator provides methods to generate unique test data
// while maintaining compatibility with the existing test data system
type UniqueDataGenerator struct {
	initialized bool
	baseTime    time.Time
}

// NewUniqueDataGenerator creates a new unique data generator
func NewUniqueDataGenerator() *UniqueDataGenerator {
	return &UniqueDataGenerator{
		initialized: true,
		baseTime:    time.Now(),
	}
}

// UniqueTestSecret extends TestSecret with unique naming capabilities
type UniqueTestSecret struct {
	TestSecret
	UniqueName string
	Timestamp  int64
}

// UniqueTestScenario extends TestScenario with unique naming capabilities
type UniqueTestScenario struct {
	TestScenario
	UniqueName    string
	Timestamp     int64
	UniqueSecrets []UniqueTestSecret
}

// GenerateUniqueSecretName creates a unique secret name for a test
func (udg *UniqueDataGenerator) GenerateUniqueSecretName(testName string) string {
	if !udg.initialized {
		return fmt.Sprintf(FallbackNameFormat, testName, time.Now().UnixNano())
	}

	timestamp := time.Now().UnixNano()
	return fmt.Sprintf(UniqueNameFormat, testName, timestamp)
}

// GenerateUniqueSecret creates a unique test secret based on a template
func (udg *UniqueDataGenerator) GenerateUniqueSecret(template TestSecret, testName string) UniqueTestSecret {
	timestamp := time.Now().UnixNano()
	uniqueName := udg.GenerateUniqueSecretName(testName)

	return UniqueTestSecret{
		TestSecret: TestSecret{
			Name:        uniqueName,
			Value:       template.Value,
			Type:        template.Type,
			Description: fmt.Sprintf(UniqueDescriptionFormat, template.Description, testName),
		},
		UniqueName: uniqueName,
		Timestamp:  timestamp,
	}
}

// GenerateUniqueScenario creates a unique test scenario with unique secrets
func (udg *UniqueDataGenerator) GenerateUniqueScenario(template TestScenario, testName string) UniqueTestScenario {
	timestamp := time.Now().UnixNano()
	uniqueName := fmt.Sprintf(ScenarioNameFormat, template.Name, testName, timestamp)

	// Generate unique secrets for the scenario
	uniqueSecrets := make([]UniqueTestSecret, len(template.Secrets))
	for i, secret := range template.Secrets {
		secretTestName := fmt.Sprintf("%s_Secret%d", testName, i)
		uniqueSecrets[i] = udg.GenerateUniqueSecret(secret, secretTestName)
	}

	// Create the unique scenario
	uniqueScenario := TestScenario{
		Name:        uniqueName,
		Description: fmt.Sprintf(UniqueDescriptionFormat, template.Description, testName),
		User:        template.User.CloneUser(),
		Secrets:     make([]TestSecret, len(uniqueSecrets)),
	}

	// Convert unique secrets to regular test secrets for the scenario
	for i, uniqueSecret := range uniqueSecrets {
		uniqueScenario.Secrets[i] = uniqueSecret.TestSecret
	}

	return UniqueTestScenario{
		TestScenario:  uniqueScenario,
		UniqueName:    uniqueName,
		Timestamp:     timestamp,
		UniqueSecrets: uniqueSecrets,
	}
}

// Predefined unique secret generators for common test patterns

// GenerateUniqueSimpleSecret creates a unique simple secret
func (udg *UniqueDataGenerator) GenerateUniqueSimpleSecret(testName string) UniqueTestSecret {
	return udg.GenerateUniqueSecret(TestSecrets.Simple, testName)
}

// GenerateUniqueComplexSecret creates a unique complex secret
func (udg *UniqueDataGenerator) GenerateUniqueComplexSecret(testName string) UniqueTestSecret {
	return udg.GenerateUniqueSecret(TestSecrets.Complex, testName)
}

// GenerateUniqueTemporarySecret creates a unique temporary secret
func (udg *UniqueDataGenerator) GenerateUniqueTemporarySecret(testName string) UniqueTestSecret {
	return udg.GenerateUniqueSecret(TestSecrets.Temporary, testName)
}

// GenerateUniqueVersionedSecret creates a unique versioned secret
func (udg *UniqueDataGenerator) GenerateUniqueVersionedSecret(testName string) UniqueTestSecret {
	return udg.GenerateUniqueSecret(TestSecrets.Versioned, testName)
}

// Versioning support for unique secrets

// UniqueVersionTestData extends VersionTestData with unique naming
type UniqueVersionTestData struct {
	VersionTestData
	UniqueName string
	Timestamp  int64
}

// GenerateUniqueVersionTestData creates unique version test data
func (udg *UniqueDataGenerator) GenerateUniqueVersionTestData(template VersionTestData, testName string) UniqueVersionTestData {
	timestamp := time.Now().UnixNano()
	uniqueName := udg.GenerateUniqueSecretName(testName)

	return UniqueVersionTestData{
		VersionTestData: VersionTestData{
			SecretName:  uniqueName,
			Versions:    append([]string{}, template.Versions...), // Deep copy
			Description: fmt.Sprintf(UniqueDescriptionFormat, template.Description, testName),
		},
		UniqueName: uniqueName,
		Timestamp:  timestamp,
	}
}

// GenerateUniqueSimpleVersioning creates unique simple versioning data
func (udg *UniqueDataGenerator) GenerateUniqueSimpleVersioning(testName string) UniqueVersionTestData {
	return udg.GenerateUniqueVersionTestData(VersioningTestData.SimpleVersioning, testName)
}

// GenerateUniqueMultipleVersions creates unique multiple versions data
func (udg *UniqueDataGenerator) GenerateUniqueMultipleVersions(testName string) UniqueVersionTestData {
	return udg.GenerateUniqueVersionTestData(VersioningTestData.MultipleVersions, testName)
}

// Cleanup helper methods

// ExtractSecretNames extracts secret names from unique secrets for cleanup
func ExtractSecretNames(uniqueSecrets []UniqueTestSecret) []string {
	names := make([]string, len(uniqueSecrets))
	for i, secret := range uniqueSecrets {
		names[i] = secret.UniqueName
	}
	return names
}

// ExtractSecretName extracts the secret name from a unique secret for cleanup
func ExtractSecretName(uniqueSecret UniqueTestSecret) string {
	return uniqueSecret.UniqueName
}

// ExtractScenarioSecretNames extracts all secret names from a unique scenario for cleanup
func ExtractScenarioSecretNames(uniqueScenario UniqueTestScenario) []string {
	return ExtractSecretNames(uniqueScenario.UniqueSecrets)
}

// Batch generation methods for common test patterns

// GenerateUniqueSecretSet generates a set of unique secrets for comprehensive testing
func (udg *UniqueDataGenerator) GenerateUniqueSecretSet(testName string) map[string]UniqueTestSecret {
	return map[string]UniqueTestSecret{
		"simple":    udg.GenerateUniqueSimpleSecret(fmt.Sprintf("%s_Simple", testName)),
		"complex":   udg.GenerateUniqueComplexSecret(fmt.Sprintf("%s_Complex", testName)),
		"temporary": udg.GenerateUniqueTemporarySecret(fmt.Sprintf("%s_Temp", testName)),
		"versioned": udg.GenerateUniqueVersionedSecret(fmt.Sprintf("%s_Versioned", testName)),
	}
}

// GenerateUniqueCRUDSet generates secrets specifically for CRUD testing
func (udg *UniqueDataGenerator) GenerateUniqueCRUDSet(testName string) map[string]UniqueTestSecret {
	return map[string]UniqueTestSecret{
		"create": udg.GenerateUniqueSecret(TestSecrets.Simple, fmt.Sprintf("%s_Create", testName)),
		"read":   udg.GenerateUniqueSecret(TestSecrets.Complex, fmt.Sprintf("%s_Read", testName)),
		"update": udg.GenerateUniqueSecret(TestSecrets.Versioned, fmt.Sprintf("%s_Update", testName)),
		"delete": udg.GenerateUniqueSecret(TestSecrets.Temporary, fmt.Sprintf("%s_Delete", testName)),
	}
}

// Validation methods for unique data

// IsValid checks if a UniqueTestSecret has all required fields
func (uts UniqueTestSecret) IsValid() bool {
	return uts.TestSecret.IsValid() && uts.UniqueName != "" && uts.Timestamp > 0
}

// IsValid checks if a UniqueTestScenario has all required fields
func (uts UniqueTestScenario) IsValid() bool {
	if !uts.TestScenario.IsValid() || uts.UniqueName == "" || uts.Timestamp <= 0 {
		return false
	}

	for _, uniqueSecret := range uts.UniqueSecrets {
		if !uniqueSecret.IsValid() {
			return false
		}
	}

	return true
}

// IsValid checks if a UniqueVersionTestData has all required fields
func (uvtd UniqueVersionTestData) IsValid() bool {
	return uvtd.UniqueName != "" && uvtd.Timestamp > 0 && len(uvtd.Versions) > 0
}
