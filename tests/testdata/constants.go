package testdata

// Shared test constants to avoid duplication across test files

// Test Messages - used by both E2E and integration tests
const (
	DetailVisibleMsg         = "Secret detail should be visible"
	PreCreateMsg             = "Pre-creating secret should succeed"
	TestDataValidationMsg    = "Test data validation should pass"
	ShouldGetTestSecretMsg   = "Should get test secret"
	ShouldGenerateUniqueMsg  = "Should generate unique secret"
	SecretCreationSuccessMsg = "Secret creation should succeed"
	SecretDeletionSuccessMsg = "Secret deletion should succeed"
	SecretUpdateSuccessMsg   = "Secret update should succeed"
	SecretNotFoundMsg        = "Secret should not be found after deletion"
	UniqueDataGenerationMsg  = "Should generate unique test data"
	CleanupSuccessMsg        = "Cleanup should complete successfully"
)

// Test Timeouts (in seconds)
const (
	DefaultTestTimeout   = 30
	ShortTestTimeout     = 5
	LongTestTimeout      = 60
	UIInteractionTimeout = 2
	FileOperationTimeout = 3
)

// Test Data Validation
const (
	MinSecretNameLength  = 1
	MaxSecretNameLength  = 100
	MinSecretValueLength = 1
	MaxSecretValueLength = 10000
	ValidSecretTypes     = "key_value,note,file"
)

// Environment-specific test constants
const (
	E2ETestEnvironment         = "e2e-test"
	IntegrationTestEnvironment = "integration-test"
	TestDataDirPrefix          = "/tmp/go-password-manager"
	TestSecretsFileName        = "secrets.json"
	TestConfigFileName         = "app.config"
)

// Error Messages for consistent testing
const (
	EmptySecretNameError    = "Secret name cannot be empty"
	EmptySecretValueError   = "Secret value cannot be empty"
	DuplicateSecretError    = "Secret with this name already exists"
	InvalidSecretNameError  = "Secret name contains invalid characters"
	SecretNotFoundError     = "Secret not found"
	ServiceUnavailableError = "Service temporarily unavailable"
	ValidationFailedError   = "Validation failed"
)

// UI Test Constants (for E2E tests)
const (
	CreateButtonText  = "Create Secret"
	SaveButtonText    = "Save"
	CancelButtonText  = "Cancel"
	EditButtonText    = "Edit"
	DeleteButtonText  = "Delete"
	ConfirmButtonText = "Confirm"
	RevealButtonText  = "Reveal"
	HideButtonText    = "Hide"
	SearchButtonText  = "Search"
	MenuButtonText    = "☰"
)

// Test Categories for organizing test data
const (
	TestCategorySmoke       = "smoke"
	TestCategoryRegression  = "regression"
	TestCategoryPerformance = "performance"
	TestCategoryIntegration = "integration"
	TestCategoryE2E         = "e2e"
	TestCategoryUnit        = "unit"
)
