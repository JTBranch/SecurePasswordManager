package setup

import (
	"fmt"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/storage"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

const (
	TestVersion = "1.0.0-e2e"
	TestUser    = "e2e-test-user"
)

// E2ETestSuite holds the test environment setup
type E2ETestSuite struct {
	testDataDir    string
	originalEnv    string
	app            fyne.App
	Window         fyne.Window
	SecretsService *service.SecretsService
	t              *testing.T
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
		if err != nil {
			suite.t.Fatalf("Failed to create test directory: %v", err)
		}
		suite.testDataDir = testDir
		suite.t.Logf("E2E test environment created at: %s", testDir)
	} else {
		suite.t.Logf("E2E test environment reusing directory: %s", suite.testDataDir)
	}

	// Set environment to use test directory
	suite.originalEnv = os.Getenv("GO_PASSWORD_MANAGER_ENV")
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "e2e-test")
	os.Setenv("TEST_DATA_DIR", suite.testDataDir)

	// Reset global environment config to pick up test settings
	if _, err := buildconfig.Load(); err != nil {
		suite.t.Fatalf("Failed to load build configuration: %v", err)
	}

	// Create test application
	suite.app = test.NewApp()
	suite.Window = test.NewWindow(nil)
	suite.Window.Resize(fyne.NewSize(1200, 800))

	// Initialize services
	buildCfg, err := buildconfig.Load()
	if err != nil {
		suite.t.Fatalf("Failed to load build config: %v", err)
	}
	configService, err := config.NewConfigService(buildCfg)
	if err != nil {
		suite.t.Fatalf("Failed to create config service: %v", err)
	}
	cryptoService, err := crypto.NewCryptoService(configService)
	if err != nil {
		suite.t.Fatalf("Failed to create crypto service: %v", err)
	}
	secretsPath, err := buildCfg.GetSecretsFilePath()
	if err != nil {
		suite.t.Fatalf("Failed to get secrets file path: %v", err)
	}
	storageService := storage.NewFileStorage(secretsPath, buildCfg.Application.Version, "e2e-user")
	suite.SecretsService = service.NewSecretsService(cryptoService, storageService)
}

// SetTestDataDir sets the test data directory (for reusing existing test data)
func (suite *E2ETestSuite) SetTestDataDir(dataDir string) {
	suite.testDataDir = dataDir
	os.Setenv("TEST_DATA_DIR", dataDir)
	if _, err := buildconfig.Load(); err != nil {
		suite.t.Fatalf("Failed to load build configuration: %v", err)
	}
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
	buildconfig.Load()

	// Clean up test directory
	err := os.RemoveAll(suite.testDataDir)
	if err != nil {
		suite.t.Logf("Warning: Failed to clean up test directory %s: %v", suite.testDataDir, err)
	}
}

// WaitForUIUpdate provides a small delay for UI updates to complete
func (suite *E2ETestSuite) WaitForUIUpdate() {
	time.Sleep(50 * time.Millisecond)
}
