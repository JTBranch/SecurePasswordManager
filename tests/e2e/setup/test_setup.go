package setup

import (
	"fmt"
	"go-password-manager/internal/config/devicekeys"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/config/secretkeymetadata"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/storage"
	"os"
	"path/filepath"
	"testing"
	"time"

	buildconfig "go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/service"
	"go-password-manager/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/require"
)

const (
	TestVersion = "1.0.0-e2e"
	TestUser    = "e2e-test-user"
)

// E2ETestSuite holds the test environment setup
type E2ETestSuite struct {
	testDataDir      string
	originalEnv      string
	app              fyne.App
	Window           fyne.Window
	SecretsService   *service.SecretsService
	DeviceKeyManager *crypto.DeviceKeyManager
	t                *testing.T
}

// NewE2ETestSuite creates a new E2E test suite
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	suite := &E2ETestSuite{t: t}
	return suite
}

// SetupTestEnvironment creates an isolated test environment for E2E testing
func (suite *E2ETestSuite) SetupTestEnvironment() {
	// Only create a new test directory if one hasn't been set
	if suite.testDataDir == "" {
		// Create isolated test environment
		testDir := filepath.Join(os.TempDir(), fmt.Sprintf("go-password-manager-e2e-%d", time.Now().UnixNano()))
		err := os.MkdirAll(testDir, 0755)
		require.NoError(suite.t, err, "Failed to create test directory")
		suite.testDataDir = testDir
		suite.t.Logf("E2E test environment created at: %s", testDir)
	} else {
		suite.t.Logf("E2E test environment reusing directory: %s", suite.testDataDir)
	}

	// Set environment to use test directory
	suite.originalEnv = os.Getenv("GO_PASSWORD_MANAGER_ENV")
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "test")
	os.Setenv("TEST_DATA_DIR", suite.testDataDir)

	// Reset cached build config so tests pick up the test environment and in-memory flags
	buildconfig.ResetCacheForTest()

	// Create test application (headless) and window for UI interaction
	suite.app = test.NewApp()
	testWin := test.NewWindow(nil)
	testWin.Resize(fyne.NewSize(1200, 800))

	// Initialize services
	buildCfg, err := buildconfig.Load()
	require.NoError(suite.t, err, "Failed to load build config")
	configService, err := config.NewConfigService(buildCfg)
	require.NoError(suite.t, err, "Failed to create config service")
	secretsKeyProvider, err := secretkeymetadata.NewSecretKeyMetadataFileService(buildCfg)
	require.NoError(suite.t, err, "Failed to create secrets key provider")
	secretsEncryptionKeyManager, err := crypto.NewSecretsEncryptionKeyManager(configService, secretsKeyProvider)
	require.NoError(suite.t, err, "Failed to create secrets encryption key manager")
	cryptoService, err := crypto.NewCryptoServiceDefault(configService, secretsEncryptionKeyManager)
	require.NoError(suite.t, err, "Failed to create crypto service")
	secretsPath, err := buildCfg.GetSecretsFilePath()
	require.NoError(suite.t, err, "Failed to get secrets file path")
	storageService := storage.NewFileStorage(secretsPath, buildCfg.Application.Version, "e2e-user")
	suite.SecretsService = service.NewSecretsService(cryptoService, storageService)
	deviceKeyFileSvc := devicekeys.NewDeviceKeyFileService(buildCfg)
	suite.DeviceKeyManager, err = crypto.NewDeviceKeyManager(cryptoService, &crypto.PemUtils{}, deviceKeyFileSvc)
	require.NoError(suite.t, err, "Failed to create device key manager")
	// Build the UI app backed by the test app and window to allow real UI interactions via fyne/test
	uiApp := ui.NewAppWithFyne(suite.app, testWin, buildCfg, suite.SecretsService, nil)
	uiApp.Start()
	suite.Window = testWin
	// In-memory keyring & device key storage now configured within constructors based on build configuration.
}

// SetTestDataDir sets the test data directory (for reusing existing test data)
func (suite *E2ETestSuite) SetTestDataDir(dataDir string) {
	suite.testDataDir = dataDir
	os.Setenv("TEST_DATA_DIR", dataDir)
}

// GetTestDataDir returns the test data directory path
func (suite *E2ETestSuite) GetTestDataDir() string {
	return suite.testDataDir
}

// GetSecretsFilePath returns the path to the secrets file
func (suite *E2ETestSuite) GetSecretsFilePath() string {
	return filepath.Join(suite.testDataDir, "secrets.json")
}

// Cleanup cleans up the E2E test environment
func (suite *E2ETestSuite) Cleanup() {
	// Application cleanup is handled automatically by test framework

	// Restore original environment
	if suite.originalEnv != "" {
		os.Setenv("GO_PASSWORD_MANAGER_ENV", suite.originalEnv)
	} else {
		os.Unsetenv("GO_PASSWORD_MANAGER_ENV")
	}
	os.Unsetenv("TEST_DATA_DIR")

	// Reload environment configuration to reset to defaults
	buildconfig.ResetCacheForTest()
	if _, err := buildconfig.Load(); err != nil {
		suite.t.Logf("Warning: failed to reload build config during cleanup: %v", err)
	}

	// Clean up test directory
	err := os.RemoveAll(suite.testDataDir)
	if err != nil {
		suite.t.Logf("Warning: Failed to clean up test directory %s: %v", suite.testDataDir, err)
	}

	// Attempt to delete any device keys we created during the test run to avoid polluting the user's keychain.
	if suite.DeviceKeyManager != nil {
		_ = suite.DeviceKeyManager.DeleteEncryptionDeviceKey()
		_ = suite.DeviceKeyManager.DeleteSigningDeviceKey()
	}
}

// WaitForUIUpdate provides a small delay for UI updates to complete
func (suite *E2ETestSuite) WaitForUIUpdate() {
	time.Sleep(50 * time.Millisecond)
}
