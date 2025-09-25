package helpers

import (
	"fmt"
	buildconfig "go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/config/devicekeys"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/config/secretkeymetadata"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/service"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/storage"
	"go-password-manager/tests/reporting"
	"os"

	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite holds the test environment setup for service layer testing
type IntegrationTestSuite struct {
	testDataDir          string
	originalEnv          string
	SecretsService       *service.SecretsService
	Reporter             *reporting.TestWrapper
	CryptoService        *crypto.CryptoService
	BuildConfig          *buildconfig.Config
	ConfigService        *config.ConfigService
	ExportService        *sharing.ExportService
	ImportService        *sharing.ImportService
	SharingService       *sharing.SharingService
	PemUtils             *crypto.PemUtils
	DeviceKeyManager     *crypto.DeviceKeyManager
	DeviceKeyFileService *devicekeys.DeviceKeyFileService
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

	// Set environment to use test directory (do not export TEST_DATA_DIR globally)
	suite.originalEnv = os.Getenv("GO_PASSWORD_MANAGER_ENV")
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "test")

	// Load test configuration
	var err error
	// Reset buildconfig cache so each suite gets a fresh load and then inject test data dir
	buildconfig.ResetCacheForTest()
	suite.BuildConfig, err = buildconfig.Load()
	require.NoError(suite.Reporter.T(), err, "Failed to load build config")
	suite.BuildConfig.Testing.DataDir = suite.testDataDir

	suite.ConfigService, err = config.NewConfigService(suite.BuildConfig)
	require.NoError(suite.Reporter.T(), err, "Failed to create config service")

	secretKeyMetadataService, err := secretkeymetadata.NewSecretKeyMetadataFileService(suite.BuildConfig)
	require.NoError(suite.Reporter.T(), err, "Failed to create secret key metadata service")

	secretsEncryptionKeyManager, err := crypto.NewSecretsEncryptionKeyManager(suite.ConfigService, secretKeyMetadataService)
	require.NoError(suite.Reporter.T(), err, "Failed to create secrets encryption key manager")

	suite.CryptoService, err = crypto.NewCryptoServiceDefault(suite.ConfigService, secretsEncryptionKeyManager)
	require.NoError(suite.Reporter.T(), err, "Failed to create crypto service")
	// Initialize secrets service with test configuration
	secretsPath, err := suite.BuildConfig.GetSecretsFilePath()
	fmt.Println("Secrets file path:", secretsPath)
	require.NoError(suite.Reporter.T(), err, "Failed to get secrets file path")
	storageService := storage.NewFileStorage(secretsPath, suite.BuildConfig.Application.Version, "integration-user")
	suite.SecretsService = service.NewSecretsService(suite.CryptoService, storageService)

	suite.DeviceKeyFileService = devicekeys.NewDeviceKeyFileService(suite.BuildConfig)

	suite.PemUtils = &crypto.PemUtils{}

	suite.DeviceKeyManager, err = crypto.NewDeviceKeyManager(suite.CryptoService, suite.PemUtils, suite.DeviceKeyFileService)
	require.NoError(suite.Reporter.T(), err, "Failed to create Device Key Manager")
	// In-memory behavior for keyring/device storage now handled within constructors based on build configuration.

	suite.ImportService = sharing.NewImportService(suite.CryptoService, suite.DeviceKeyManager, suite.SecretsService)

	suite.ExportService = sharing.NewExportService(suite.CryptoService, suite.DeviceKeyManager)

	suite.SharingService = sharing.NewSharingService(suite.ExportService, suite.ImportService, suite.SecretsService)
}

// SetTestDataDir sets the test data directory (for reusing existing test data)
func (suite *IntegrationTestSuite) SetTestDataDir(dataDir string) {
	suite.testDataDir = dataDir
	// Do not set global TEST_DATA_DIR; use suite-scoped value and reload
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
	require.NoError(suite.Reporter.T(), err, "Failed to get secrets file path")
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

	// Attempt to delete any device keys we created during the test run to avoid polluting the user's keychain.
	if suite.DeviceKeyManager != nil {
		_ = suite.DeviceKeyManager.DeleteEncryptionDeviceKey()
		_ = suite.DeviceKeyManager.DeleteSigningDeviceKey()
	}

	// The test temp directory is cleaned up automatically by the test framework
}
