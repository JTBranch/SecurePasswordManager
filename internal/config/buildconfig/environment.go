package buildconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// EnvironmentConfig represents environment-specific configuration loaded from YAML
type EnvironmentConfig struct {
	Application ApplicationConfig `yaml:"application"`
	UI          UIConfig          `yaml:"ui"`
	Logging     LoggingConfig     `yaml:"logging"`
	Security    SecurityConfig    `yaml:"security"`
	Storage     StorageConfig     `yaml:"storage"`
	Development DevelopmentConfig `yaml:"development"`
	Testing     TestingConfig     `yaml:"testing"`
}

var globalEnvConfig *EnvironmentConfig

// LoadEnvironmentConfig loads configuration from YAML files and environment variables
func LoadEnvironmentConfig() (*EnvironmentConfig, error) {
	// Determine environment
	env := os.Getenv("GO_PASSWORD_MANAGER_ENV")
	if env == "" {
		env = "development" // Default to development
	}

	// Load default config first
	config, err := loadYAMLConfigFile("configs/default.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	// Load environment-specific config and merge
	envConfigPath := fmt.Sprintf("configs/%s.yaml", env)
	if _, err := os.Stat(envConfigPath); err == nil {
		envConfig, err := loadYAMLConfigFile(envConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s config: %w", env, err)
		}
		config = mergeEnvironmentConfig(config, envConfig)
	}

	// Apply environment variable overrides
	applyEnvironmentOverrides(config)

	globalEnvConfig = config
	return config, nil
}

// GetEnvironmentConfig returns the global environment config (loads if not already loaded)
func GetEnvironmentConfig() *EnvironmentConfig {
	if globalEnvConfig == nil {
		// Auto-load if not already loaded
		LoadEnvironmentConfig()
	}
	return globalEnvConfig
}

// loadYAMLConfigFile loads a YAML configuration file
func loadYAMLConfigFile(filepath string) (*EnvironmentConfig, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config EnvironmentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// mergeEnvironmentConfig merges environment-specific config into base config
func mergeEnvironmentConfig(base, override *EnvironmentConfig) *EnvironmentConfig {
	// Create a copy of base config
	result := *base

	// Merge each section only if values are non-empty/non-zero
	if override.Application.Name != "" {
		result.Application.Name = override.Application.Name
	}
	if override.Application.Version != "" {
		result.Application.Version = override.Application.Version
	}
	if override.Application.Environment != "" {
		result.Application.Environment = override.Application.Environment
	}

	if override.UI.Window.Width != 0 {
		result.UI.Window.Width = override.UI.Window.Width
	}
	if override.UI.Window.Height != 0 {
		result.UI.Window.Height = override.UI.Window.Height
	}
	if override.UI.Theme != "" {
		result.UI.Theme = override.UI.Theme
	}

	if override.Logging.Level != "" {
		result.Logging.Level = override.Logging.Level
	}
	if override.Logging.Format != "" {
		result.Logging.Format = override.Logging.Format
	}
	// Override debug setting regardless of value
	result.Logging.Debug = override.Logging.Debug

	if override.Security.Encryption.KeySize != 0 {
		result.Security.Encryption.KeySize = override.Security.Encryption.KeySize
	}
	if override.Security.Encryption.Algorithm != "" {
		result.Security.Encryption.Algorithm = override.Security.Encryption.Algorithm
	}

	if override.Storage.SecretsFile != "" {
		result.Storage.SecretsFile = override.Storage.SecretsFile
	}
	if override.Storage.ConfigFile != "" {
		result.Storage.ConfigFile = override.Storage.ConfigFile
	}

	// Override development settings
	result.Development.HotReload = override.Development.HotReload
	result.Development.AutoSave = override.Development.AutoSave

	if override.Testing.Timeout != "" {
		result.Testing.Timeout = override.Testing.Timeout
	}
	if override.Testing.DataDir != "" {
		result.Testing.DataDir = override.Testing.DataDir
	}
	result.Testing.Parallel = override.Testing.Parallel
	result.Testing.Cleanup = override.Testing.Cleanup

	return &result
}

// applyEnvironmentOverrides applies environment variable overrides
func applyEnvironmentOverrides(config *EnvironmentConfig) {
	// Application overrides
	if env := os.Getenv("APP_NAME"); env != "" {
		config.Application.Name = env
	}
	if env := os.Getenv("APP_VERSION"); env != "" {
		config.Application.Version = env
	}
	if env := os.Getenv("GO_PASSWORD_MANAGER_ENV"); env != "" {
		config.Application.Environment = env
	}

	// UI overrides
	if env := os.Getenv("DEFAULT_WINDOW_WIDTH"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			config.UI.Window.Width = val
		}
	}
	if env := os.Getenv("DEFAULT_WINDOW_HEIGHT"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			config.UI.Window.Height = val
		}
	}

	// Logging overrides
	if env := os.Getenv("DEBUG_LOGGING"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			config.Logging.Debug = val
		}
	}
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		config.Logging.Level = env
	}

	// Security overrides
	if env := os.Getenv("ENCRYPTION_KEY_SIZE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			config.Security.Encryption.KeySize = val
		}
	}

	// Storage overrides
	if env := os.Getenv("SECRETS_FILE_PATH"); env != "" {
		config.Storage.SecretsFile = env
	}
	if env := os.Getenv("CONFIG_FILE_PATH"); env != "" {
		config.Storage.ConfigFile = env
	}

	// Development overrides
	if env := os.Getenv("HOT_RELOAD"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			config.Development.HotReload = val
		}
	}

	// Testing overrides
	if env := os.Getenv("TEST_DATA_DIR"); env != "" {
		config.Testing.DataDir = env
	}
	if env := os.Getenv("E2E_TEST_TIMEOUT"); env != "" {
		config.Testing.Timeout = env
	}
}

// Helper methods for common configuration needs

// IsProduction returns true if running in production mode
func (c *EnvironmentConfig) IsProduction() bool {
	return c.Application.Environment == "production" || c.Application.Environment == "prod"
}

// IsDevelopment returns true if running in development mode
func (c *EnvironmentConfig) IsDevelopment() bool {
	return c.Application.Environment == "development" || c.Application.Environment == "dev"
}

// IsTest returns true if running in test mode
func (c *EnvironmentConfig) IsTest() bool {
	env := c.Application.Environment
	return env == "test" || env == "integration-test" || env == "e2e-test"
}

// GetSecretsFilePath returns the secrets file path based on environment
func (c *EnvironmentConfig) GetSecretsFilePath() string {
	if c.Storage.SecretsFile != "" && filepath.IsAbs(c.Storage.SecretsFile) {
		return c.Storage.SecretsFile
	}

	// Use test data directory for tests
	if c.IsTest() && c.Testing.DataDir != "" {
		return filepath.Join(c.Testing.DataDir, c.Storage.SecretsFile)
	}

	// Use current directory for development
	if c.IsDevelopment() {
		return c.Storage.SecretsFile
	}

	// Production: use OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return c.Storage.SecretsFile // Fallback
	}
	appConfigDir := filepath.Join(configDir, c.Application.Name)
	os.MkdirAll(appConfigDir, 0700)
	return filepath.Join(appConfigDir, c.Storage.SecretsFile)
}

// GetConfigFilePath returns the config file path based on environment
func (c *EnvironmentConfig) GetConfigFilePath() string {
	if c.Storage.ConfigFile != "" && filepath.IsAbs(c.Storage.ConfigFile) {
		return c.Storage.ConfigFile
	}

	// Use test data directory for tests
	if c.IsTest() && c.Testing.DataDir != "" {
		return filepath.Join(c.Testing.DataDir, c.Storage.ConfigFile)
	}

	// Use current directory for development
	if c.IsDevelopment() {
		return c.Storage.ConfigFile
	}

	// Production: use OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return c.Storage.ConfigFile // Fallback
	}
	appConfigDir := filepath.Join(configDir, c.Application.Name)
	os.MkdirAll(appConfigDir, 0700)
	return filepath.Join(appConfigDir, c.Storage.ConfigFile)
}

// GetTestTimeout returns the test timeout as a duration
func (c *EnvironmentConfig) GetTestTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Testing.Timeout)
	if err != nil {
		return 30 * time.Second // Default fallback
	}
	return duration
}

// GetWindowSize returns the configured window dimensions
func (c *EnvironmentConfig) GetWindowSize() (int, int) {
	return c.UI.Window.Width, c.UI.Window.Height
}
