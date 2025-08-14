# Integrated Unique Data Generation System

## Overview

The test data system has been enhanced with integrated unique data generation capabilities that ensure test independence while maintaining compatibility with the existing immutable test data structure.

## Key Components

### 1. UniqueDataGenerator (`unique.go`)

- Core unique data generation engine
- Generates timestamp-based unique names
- Creates unique secrets, scenarios, and version data based on templates

### 2. Enhanced TestDataManager (`helpers.go`)

- Integrated with `UniqueDataGenerator`
- Provides convenient methods for generating unique test data
- Maintains backward compatibility with existing test data access

### 3. New Data Types

- `UniqueTestSecret`: Extends `TestSecret` with unique naming
- `UniqueTestScenario`: Extends `TestScenario` with unique secrets
- `UniqueVersionTestData`: Extends `VersionTestData` with unique naming

## Migration Guide

### Before (Manual Unique Generation)

```go
// Old approach - manual unique name generation
uniqueSecretName := generateUniqueSecretName("TestSecretCreation")
defer cleanupTestSecrets(suite, []string{uniqueSecretName})

// Use test data template but with unique name
testSecretTemplate := testdata.TestSecrets.Simple
testSecretValue := testSecretTemplate.Value
```

### After (Integrated Unique Generation)

```go
// New approach - integrated unique data generation
uniqueSecret, err := testDataManager.GenerateUniqueSimpleSecret("TestSecretCreation")
require.NoError(t, err, "Should generate unique secret")
defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

// Access all properties through the unique secret
secretName := uniqueSecret.UniqueName
secretValue := uniqueSecret.Value
secretType := uniqueSecret.Type
```

## Usage Examples

### 1. Simple Unique Secret Generation

```go
func TestSecretCreation(t *testing.T) {
    // Initialize test data manager
    testDataManager := testdata.NewTestDataManager()

    // Generate unique secret
    uniqueSecret, err := testDataManager.GenerateUniqueSimpleSecret("TestSecretCreation")
    require.NoError(t, err)
    defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

    // Use the unique secret in your test
    err = suite.SecretsService.SaveSecret(uniqueSecret.UniqueName, uniqueSecret.Value, uniqueSecret.Type)
    require.NoError(t, err)
}
```

### 2. Versioning with Unique Data

```go
func TestSecretVersioning(t *testing.T) {
    testDataManager := testdata.NewTestDataManager()

    // Generate unique versioning data
    uniqueVersionData, err := testDataManager.GenerateUniqueSimpleVersioning("TestSecretVersioning")
    require.NoError(t, err)
    defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueVersionData.UniqueName})

    // Use versioning data
    initialValue := uniqueVersionData.GetInitialVersion()
    updatedValue := uniqueVersionData.GetLatestVersion()

    // Create and update secret...
}
```

### 3. Batch Operations with CRUD Set

```go
func TestSecretBatchOperations(t *testing.T) {
    testDataManager := testdata.NewTestDataManager()

    // Generate a complete CRUD set
    crudSet, err := testDataManager.GenerateUniqueCRUDSet("TestBatchOperations")
    require.NoError(t, err)

    // Extract names for cleanup
    var secretNames []string
    for _, secret := range crudSet {
        secretNames = append(secretNames, secret.UniqueName)
    }
    defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, secretNames)

    // Use different secret types for different operations
    createSecret := crudSet["create"]
    updateSecret := crudSet["update"]
    deleteSecret := crudSet["delete"]
    readSecret := crudSet["read"]
}
```

### 4. Custom Template Usage

```go
func TestCustomSecret(t *testing.T) {
    testDataManager := testdata.NewTestDataManager()

    // Generate unique secret from any template
    uniqueSecret, err := testDataManager.GenerateUniqueSecretByName("ComplexTestSecret", "TestCustom")
    require.NoError(t, err)
    defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})
}
```

## Available Generation Methods

### Individual Secret Generation

- `GenerateUniqueSimpleSecret(testName)` - Simple secrets
- `GenerateUniqueComplexSecret(testName)` - Complex secrets with special characters
- `GenerateUniqueTemporarySecret(testName)` - Temporary secrets for deletion testing
- `GenerateUniqueVersionedSecret(testName)` - Secrets for versioning tests
- `GenerateUniqueSecretByName(secretName, testName)` - Any template secret

### Batch Generation

- `GenerateUniqueSecretSet(testName)` - Comprehensive set (simple, complex, temporary, versioned)
- `GenerateUniqueCRUDSet(testName)` - CRUD-specific set (create, read, update, delete)

### Versioning Data

- `GenerateUniqueSimpleVersioning(testName)` - Simple 2-version scenario
- `GenerateUniqueVersionTestData(templateName, testName)` - Any version template

### Scenarios

- `GenerateUniqueScenario(scenarioName, testName)` - Complete scenarios with unique secrets

## Cleanup Methods

### Automatic Cleanup

```go
// Cleanup by unique secret names
defer testDataManager.CleanupUniqueSecretNames(suite.SecretsService, []string{uniqueSecret.UniqueName})

// Cleanup multiple secrets
defer testDataManager.CleanupUniqueSecrets(suite.SecretsService, []UniqueTestSecret{secret1, secret2})
```

### Helper Functions

```go
// Extract names from unique secrets for cleanup
secretNames := testdata.ExtractSecretNames(uniqueSecrets)

// Extract name from single unique secret
secretName := testdata.ExtractSecretName(uniqueSecret)

// Extract all names from a unique scenario
scenarioNames := testdata.ExtractScenarioSecretNames(uniqueScenario)
```

## Data Structure Access

### UniqueTestSecret Fields

```go
type UniqueTestSecret struct {
    TestSecret          // Embedded original test secret
    UniqueName string   // Unique name with timestamp
    Timestamp  int64    // Generation timestamp
}

// Access methods
secretName := uniqueSecret.UniqueName    // Unique name for test isolation
secretValue := uniqueSecret.Value        // Original template value
secretType := uniqueSecret.Type          // Original template type
description := uniqueSecret.Description  // Enhanced description with test name
```

### UniqueVersionTestData Fields

```go
type UniqueVersionTestData struct {
    VersionTestData     // Embedded original version data
    UniqueName string   // Unique secret name
    Timestamp  int64    // Generation timestamp
}

// Access methods
secretName := versionData.UniqueName          // Unique name
versions := versionData.GetAllVersions()     // All version values
initial := versionData.GetInitialVersion()   // First version
latest := versionData.GetLatestVersion()     // Last version
```

## Benefits

### 1. Test Independence

- Each test gets unique data that doesn't conflict with other tests
- Tests can run in any order (supports shuffle flag)
- Parallel test execution is safe

### 2. Consistency

- Uses the same high-quality test data templates
- Maintains data validation and type safety
- Preserves existing test data structure

### 3. Convenience

- One-line generation of complete test datasets
- Automatic cleanup integration
- Batch operations for complex scenarios

### 4. Maintainability

- Centralized unique data generation logic
- Backward compatibility with existing tests
- Easy migration path from manual unique generation

## Backward Compatibility

The existing functions are marked as deprecated but still work:

```go
// Deprecated but still functional
generateUniqueSecretName("TestName")  // Use testDataManager.GenerateUniqueSecretName() instead
cleanupTestSecrets(suite, names)      // Use testDataManager.CleanupUniqueSecretNames() instead
```

## Migration Steps

1. **Initialize TestDataManager**: Replace individual test data access with manager
2. **Replace Manual Generation**: Use integrated unique generation methods
3. **Update Cleanup**: Use manager cleanup methods instead of manual cleanup
4. **Remove Deprecated Calls**: Gradually replace deprecated functions

This integrated system provides a robust foundation for test independence while maintaining the quality and structure of the existing test data system.
