package helpers

import (
	"fmt"
	"go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
	"go-password-manager/tests/reporting"
	"os"
)

// IntegrationTestSuite holds the test environment setup for service layer testing
type IntegrationTestSuite struct {
	testDataDir    string
	originalEnv    string
	SecretsService *service.SecretsService
	Reporter       *reporting.TestWrapper
	CryptoService  *crypto.CryptoService
	BuildConfig    *buildconfig.Config
	ConfigService  *config.ConfigService
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(reporter *reporting.TestWrapper) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{Reporter: reporter}
	return suite
}

// SetupTestEnvironment creates an isolated test environment for integration testing
func (suite *IntegrationTestSuite) SetupTestEnvironment() {
	// Only create a new test directory if one hasn't been set
	if suite.testDataDir == "" {
		// Create isolated test environment
		suite.testDataDir = suite.Reporter.T().TempDir()
		suite.Reporter.T().Logf("Integration test environment created at: %s", suite.testDataDir)
	} else {
		suite.Reporter.T().Logf("Integration test environment reusing directory: %s", suite.testDataDir)
	}

	// Set environment to use test directory
	suite.originalEnv = os.Getenv("GO_PASSWORD_MANAGER_ENV")
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "integration-test")
	os.Setenv("TEST_DATA_DIR", suite.testDataDir)

	// Load test configuration
	var err error
	suite.BuildConfig, err = buildconfig.Load()
	if err != nil {
		suite.Reporter.T().Fatalf("Failed to load build config: %v", err)
	}

	suite.ConfigService, err = config.NewConfigService(suite.BuildConfig)
	if err != nil {
		suite.Reporter.T().Fatalf("Failed to create config service: %v", err)
	}

	suite.CryptoService, err = crypto.NewCryptoService(suite.ConfigService)
	if err != nil {
		suite.Reporter.T().Fatalf("Failed to create crypto service: %v", err)
	}

	// Initialize secrets service with test configuration
	secretsPath, err := suite.BuildConfig.GetSecretsFilePath()
	fmt.Println("Secrets file path:", secretsPath)
	if err != nil {
		suite.Reporter.T().Fatalf("Failed to get secrets file path: %v", err)
	}
	storageService := storage.NewFileStorage(secretsPath, suite.BuildConfig.Application.Version, "integration-user")
	suite.SecretsService = service.NewSecretsService(suite.CryptoService, storageService)
}

// SetTestDataDir sets the test data directory (for reusing existing test data)
func (suite *IntegrationTestSuite) SetTestDataDir(dataDir string) {
	suite.testDataDir = dataDir
	os.Setenv("TEST_DATA_DIR", dataDir)
	// Reload configuration to use the new data directory
	suite.SetupTestEnvironment()
}

// GetTestDataDir returns the test data directory path
func (suite *IntegrationTestSuite) GetTestDataDir() string {
	return suite.testDataDir
}

// GetSecretsFilePath returns the path to the secrets file
func (suite *IntegrationTestSuite) GetSecretsFilePath() string {
	secretsPath, err := suite.BuildConfig.GetSecretsFilePath()
	if err != nil {
		suite.Reporter.T().Fatalf("Failed to get secrets file path: %v", err)
	}
	return secretsPath
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

	// The test temp directory is cleaned up automatically by the test framework
}
