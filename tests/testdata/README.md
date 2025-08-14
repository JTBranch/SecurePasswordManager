# Test Data Package

This package provides a comprehensive, immutable test data system for all test suites in the password manager application.

## 🔒 **Key Principles: Immutability & Read-Only**

- **All test data is IMMUTABLE** - original data cannot be modified
- **Clone methods ensure safety** - always returns copies, never references
- **Validation ensures integrity** - comprehensive data validation
- **Type safety** - strongly typed data structures
- **Consistent across all tests** - shared data prevents duplication

## 📁 **Package Structure**

```
tests/testdata/
├── testdata.go     # Core test data definitions
├── versioning.go   # Version-specific test data
├── ui.go          # UI workflow test data
├── helpers.go     # Test data manager and utilities
└── README.md      # This documentation
```

## 🏗 **Core Components**

### 1. **TestDataManager** - Safe Access Controller

```go
testDataManager := testdata.NewTestDataManager()

// Always validate before use
require.NoError(t, testDataManager.ValidateTestData())

// Get immutable copies
secret, err := testDataManager.GetTestSecret("SimpleTestSecret")
scenario, err := testDataManager.GetTestScenario("CRUDOperations")
user, err := testDataManager.GetTestUser("e2e-test")
```

### 2. **TestSecrets** - Predefined Secret Data

```go
// Available test secrets (all immutable)
testdata.TestSecrets.Simple      // Basic secret for simple tests
testdata.TestSecrets.Complex     // Complex secret with special chars
testdata.TestSecrets.Long        // Long value testing
testdata.TestSecrets.Special     // All special characters
testdata.TestSecrets.Versioned   // For version testing
testdata.TestSecrets.Temporary   // For deletion testing
```

### 3. **TestScenarios** - Complete Test Scenarios

```go
// Available scenarios (all immutable)
testdata.TestScenarios.Basic       // Basic secret creation
testdata.TestScenarios.CRUD        // Complete CRUD operations
testdata.TestScenarios.Versioning  // Version management
testdata.TestScenarios.Visibility  // Show/hide functionality
testdata.TestScenarios.Search      // Search and filtering
testdata.TestScenarios.Persistence // Data persistence testing
```

### 4. **TestUsers** - Environment-Specific Users

```go
// Available users (all immutable)
testdata.TestUsers.E2EUser          // For e2e tests
testdata.TestUsers.IntegrationUser  // For integration tests
testdata.TestUsers.UnitTestUser     // For unit tests
```

## 🛠 **Usage Examples**

### **Example 1: Basic Secret Creation Test**

```go
func TestSecretCreation(t *testing.T) {
    // Initialize test data manager
    testDataManager := testdata.NewTestDataManager()
    require.NoError(t, testDataManager.ValidateTestData())

    // Get immutable test secret
    testSecret, err := testDataManager.GetTestSecret(testdata.TestSecrets.Simple.Name)
    require.NoError(t, err)

    // Use the test data (always safe - it's a copy)
    secretName := testSecret.Name   // "SimpleTestSecret"
    secretValue := testSecret.Value // "SimplePassword123"

    // Create secret using test data
    err = service.SaveSecret(testSecret.Name, testSecret.Value, testSecret.Type)
    require.NoError(t, err)
}
```

### **Example 2: Version Testing with Test Data**

```go
func TestSecretVersioning(t *testing.T) {
    testDataManager := testdata.NewTestDataManager()
    require.NoError(t, testDataManager.ValidateTestData())

    // Get versioning test data
    versionData := testdata.VersioningTestData.SimpleVersioning.CloneVersionTestData()

    // Use initial version
    initialValue := versionData.GetInitialVersion() // "InitialValue"

    // Create secret with initial version
    err := service.SaveSecret(versionData.SecretName, initialValue, "key_value")
    require.NoError(t, err)

    // Update to latest version
    latestValue := versionData.GetLatestVersion() // "UpdatedValue"
    err = service.EditSecret(versionData.SecretName, latestValue)
    require.NoError(t, err)
}
```

### **Example 3: Complete Scenario Testing**

```go
func TestCRUDOperations(t *testing.T) {
    testDataManager := testdata.NewTestDataManager()
    require.NoError(t, testDataManager.ValidateTestData())

    // Get complete CRUD scenario
    scenario, err := testDataManager.GetTestScenario("CRUDOperations")
    require.NoError(t, err)

    // Create all secrets from scenario
    err = testDataManager.CreateScenarioSecrets(service, scenario.Name)
    require.NoError(t, err)

    // Verify all secrets were created
    assert.Equal(t, scenario.GetSecretsCount(), 3)
}
```

### **Example 4: Service Integration Helper**

```go
func TestServiceIntegration(t *testing.T) {
    testDataManager := testdata.NewTestDataManager()

    // Create multiple test secrets at once
    secretNames := []string{
        testdata.TestSecrets.Simple.Name,
        testdata.TestSecrets.Complex.Name,
    }

    err := testDataManager.CreateTestSecrets(service, secretNames)
    require.NoError(t, err)
}
```

## 🔐 **Safety Guarantees**

### **1. Immutability**

```go
// ✅ SAFE - Always returns a copy
secret := testdata.TestSecrets.Simple.CloneSecret()
secret.Value = "Modified" // Only affects the copy, not original

// ✅ SAFE - Original data unchanged
original := testdata.TestSecrets.Simple.Value // Still "SimplePassword123"
```

### **2. Validation**

```go
// ✅ SAFE - Always validate before use
testDataManager := testdata.NewTestDataManager()
require.NoError(t, testDataManager.ValidateTestData())

// ✅ SAFE - Individual validation
secret := testdata.TestSecrets.Simple
assert.True(t, secret.IsValid())
```

### **3. Type Safety**

```go
// ✅ SAFE - Strongly typed
var secret testdata.TestSecret = testdata.TestSecrets.Simple.CloneSecret()
var scenario testdata.TestScenario = testdata.TestScenarios.Basic.CloneScenario()
```

## 📋 **Available Test Data**

### **Test Secrets**

| Name        | Description                | Value Pattern                          |
| ----------- | -------------------------- | -------------------------------------- |
| `Simple`    | Basic test secret          | `SimplePassword123`                    |
| `Complex`   | Complex with special chars | `C0mpl3x_P@ssw0rd_W1th_Sp3c1@l_Ch@rs!` |
| `Long`      | Long value testing         | `ThisIsAVeryLong...`                   |
| `Special`   | All special characters     | `!@#$%^&*()_+-=...`                    |
| `Versioned` | For version testing        | `InitialVersion1`                      |
| `Temporary` | For deletion testing       | `TemporaryValue`                       |

### **Test Scenarios**

| Name          | Purpose                  | Secrets Included               |
| ------------- | ------------------------ | ------------------------------ |
| `Basic`       | Simple secret creation   | Simple                         |
| `CRUD`        | Complete CRUD operations | Simple, Complex, Temporary     |
| `Versioning`  | Version management       | Versioned                      |
| `Visibility`  | Show/hide functionality  | Simple                         |
| `Search`      | Search and filtering     | Simple, Complex, Long, Special |
| `Persistence` | Data persistence         | Simple, Complex                |

### **Versioning Test Data**

| Name                 | Versions    | Description        |
| -------------------- | ----------- | ------------------ |
| `SimpleVersioning`   | 2 versions  | Basic version test |
| `MultipleVersions`   | 4 versions  | Multi-version test |
| `LongVersionHistory` | 10 versions | Long history test  |

## 🚫 **What NOT to Do**

```go
// ❌ WRONG - Don't modify original data
testdata.TestSecrets.Simple.Value = "Modified" // This would corrupt data

// ❌ WRONG - Don't use without validation
secret := testdata.TestSecrets.Simple // No validation

// ❌ WRONG - Don't hardcode test values
secretValue := "HardcodedPassword123" // Use test data instead
```

## ✅ **Best Practices**

```go
// ✅ CORRECT - Always use test data manager
testDataManager := testdata.NewTestDataManager()
require.NoError(t, testDataManager.ValidateTestData())

// ✅ CORRECT - Always get clones
secret := testdata.TestSecrets.Simple.CloneSecret()

// ✅ CORRECT - Use helper methods
err := testDataManager.CreateTestSecret(service, testdata.TestSecrets.Simple.Name)

// ✅ CORRECT - Use scenarios for complex tests
scenario, err := testDataManager.GetTestScenario("CRUDOperations")
```

## 🔄 **Migration Guide**

### **Before (Hardcoded Values)**

```go
func TestSecretCreation(t *testing.T) {
    secretName := "MyTestSecret"        // ❌ Hardcoded
    secretValue := "MyPassword123"      // ❌ Hardcoded
    secretType := "key_value"           // ❌ Hardcoded
}
```

### **After (Using Test Data)**

```go
func TestSecretCreation(t *testing.T) {
    testDataManager := testdata.NewTestDataManager()
    require.NoError(t, testDataManager.ValidateTestData())

    testSecret, err := testDataManager.GetTestSecret(testdata.TestSecrets.Simple.Name)
    require.NoError(t, err)

    secretName := testSecret.Name       // ✅ From test data
    secretValue := testSecret.Value     // ✅ From test data
    secretType := testSecret.Type       // ✅ From test data
}
```

## 🎯 **Benefits**

1. **Consistency** - All tests use the same reliable data
2. **Immutability** - No risk of data corruption between tests
3. **Maintainability** - Single source of truth for test data
4. **Type Safety** - Compile-time validation of data usage
5. **Validation** - Runtime validation ensures data integrity
6. **Reusability** - Same data can be used across multiple test types
7. **Documentation** - Self-documenting test scenarios

## 🏁 **Quick Start**

1. **Import the package**:

   ```go
   import "go-password-manager/tests/testdata"
   ```

2. **Initialize manager**:

   ```go
   testDataManager := testdata.NewTestDataManager()
   require.NoError(t, testDataManager.ValidateTestData())
   ```

3. **Use test data**:

   ```go
   secret, err := testDataManager.GetTestSecret(testdata.TestSecrets.Simple.Name)
   require.NoError(t, err)
   ```

4. **Create secrets in service**:
   ```go
   err := testDataManager.CreateTestSecret(service, testdata.TestSecrets.Simple.Name)
   require.NoError(t, err)
   ```

This test data system ensures your tests are reliable, maintainable, and immune to data corruption! 🔒
