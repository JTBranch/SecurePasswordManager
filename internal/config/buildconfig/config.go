package buildconfig

import (
	"errors"
	"fmt"
	"os" // Changed from io/ioutil
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents environment-specific configuration loaded from YAML
type Config struct {
	Application ApplicationConfig `yaml:"application"`
	UI          UIConfig          `yaml:"ui"`
	Logging     LoggingConfig     `yaml:"logging"`
	Security    SecurityConfig    `yaml:"security"`
	Storage     StorageConfig     `yaml:"storage"`
	Development DevelopmentConfig `yaml:"development"`
	Testing     TestingConfig     `yaml:"testing"`
}

type ApplicationConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

type UIConfig struct {
	Window WindowConfig `yaml:"window"`
	Theme  string       `yaml:"theme"`
}

type WindowConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type LoggingConfig struct {
	Debug  bool   `yaml:"debug"`
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type SecurityConfig struct {
	Encryption EncryptionConfig `yaml:"encryption"`
}

type EncryptionConfig struct {
	KeySize   int    `yaml:"key_size"`
	Algorithm string `yaml:"algorithm"`
}

type StorageConfig struct {
	SecretsFile string `yaml:"secrets_file"`
	ConfigFile  string `yaml:"config_file"`
}

type DevelopmentConfig struct {
	HotReload bool `yaml:"hot_reload"`
	AutoSave  bool `yaml:"auto_save"`
}

type TestingConfig struct {
	Timeout  string `yaml:"timeout"`
	DataDir  string `yaml:"data_dir"`
	Parallel bool   `yaml:"parallel"`
	Cleanup  bool   `yaml:"cleanup"`
}

// findProjectRoot finds the root of the project by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}

var globalConfig *Config

// Load loads configuration from YAML files and environment variables
func Load() (*Config, error) {
	// Determine environment
	env := os.Getenv("GO_PASSWORD_MANAGER_ENV")
	if env == "" {
		env = "development" // Default to development
	}

	root, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// Try to load environment-specific config first
	envConfigPath := filepath.Join(root, "configs", fmt.Sprintf("%s.yaml", env))
	var config *Config

	if _, statErr := os.Stat(envConfigPath); statErr == nil {
		// if it exists, load it
		config, err = loadConfigFile(envConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s config: %w", env, err)
		}
	} else {
		// otherwise, fall back to default
		defaultConfigPath := filepath.Join(root, "configs", "default.yaml")
		config, err = loadConfigFile(defaultConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load default config: %w", err)
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(config)

	globalConfig = config
	return config, nil
}

// Get returns the loaded configuration
func Get() *Config {
	return globalConfig
}

// loadConfigFile loads a configuration file from the given path
func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// applyEnvOverrides applies environment variable overrides in smaller functions
func applyEnvOverrides(config *Config) {
	applyApplicationOverrides(config)
	applyUIOverrides(config)
	applyLoggingOverrides(config)
	applySecurityOverrides(config)
	applyStorageOverrides(config)
	applyDevelopmentOverrides(config)
	applyTestingOverrides(config)
}

func applyApplicationOverrides(config *Config) {
	if env := os.Getenv("APP_NAME"); env != "" {
		config.Application.Name = env
	}
	if env := os.Getenv("APP_VERSION"); env != "" {
		config.Application.Version = env
	}
	if env := os.Getenv("GO_PASSWORD_MANAGER_ENV"); env != "" {
		config.Application.Environment = env
	}
}

func applyUIOverrides(config *Config) {
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
}

func applyLoggingOverrides(config *Config) {
	if env := os.Getenv("DEBUG_LOGGING"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			config.Logging.Debug = val
		}
	}
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		config.Logging.Level = env
	}
}

func applySecurityOverrides(config *Config) {
	if env := os.Getenv("ENCRYPTION_KEY_SIZE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			config.Security.Encryption.KeySize = val
		}
	}
}

func applyStorageOverrides(config *Config) {
	if env := os.Getenv("SECRETS_FILE_PATH"); env != "" {
		config.Storage.SecretsFile = env
	}
	if env := os.Getenv("CONFIG_FILE_PATH"); env != "" {
		config.Storage.ConfigFile = env
	}
}

func applyDevelopmentOverrides(config *Config) {
	if env := os.Getenv("HOT_RELOAD"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			config.Development.HotReload = val
		}
	}
}

func applyTestingOverrides(config *Config) {
	if env := os.Getenv("TEST_DATA_DIR"); env != "" {
		config.Testing.DataDir = env
	}
	if env := os.Getenv("E2E_TEST_TIMEOUT"); env != "" {
		config.Testing.Timeout = env
	}
}

// Helper methods for common configuration needs

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Application.Environment == "production" || c.Application.Environment == "prod"
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Application.Environment == "development" || c.Application.Environment == "dev"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	env := c.Application.Environment
	return env == "test" || env == "integration-test" || env == "e2e-test"
}

// GetConfigFilePath returns the config file path based on environment
func (c *Config) GetConfigFilePath() (string, error) {
	if c.Storage.ConfigFile != "" && filepath.IsAbs(c.Storage.ConfigFile) {
		return c.Storage.ConfigFile, nil
	}

	// Use test data directory for tests
	if c.IsTest() && c.Testing.DataDir != "" {
		return filepath.Join(c.Testing.DataDir, c.Storage.ConfigFile), nil
	}

	// Use current directory for development
	if c.IsDevelopment() {
		return c.Storage.ConfigFile, nil
	}

	// Production: use OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	appConfigDir := filepath.Join(configDir, c.Application.Name)
	if err := os.MkdirAll(appConfigDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create application config directory: %w", err)
	}
	return filepath.Join(appConfigDir, c.Storage.ConfigFile), nil
}

// GetSecretsFilePath returns the secrets file path based on environment
func (c *Config) GetSecretsFilePath() (string, error) {
	if c.Storage.SecretsFile != "" && filepath.IsAbs(c.Storage.SecretsFile) {
		return c.Storage.SecretsFile, nil
	}

	// Use test data directory for tests
	if c.IsTest() && c.Testing.DataDir != "" {
		return filepath.Join(c.Testing.DataDir, c.Storage.SecretsFile), nil
	}

	// Use current directory for development
	if c.IsDevelopment() {
		return c.Storage.SecretsFile, nil
	}

	// Production: use OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	appConfigDir := filepath.Join(configDir, c.Application.Name)
	if err := os.MkdirAll(appConfigDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create application config directory: %w", err)
	}
	return filepath.Join(appConfigDir, c.Storage.SecretsFile), nil
}

// GetTestTimeout returns the test timeout as a duration
func (c *Config) GetTestTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Testing.Timeout)
	if err != nil {
		return 30 * time.Second // Default fallback
	}
	return duration
}

// GetWindowSize returns the configured window dimensions
func (c *Config) GetWindowSize() (int, int) {
	return c.UI.Window.Width, c.UI.Window.Height
}

// GetAppVersion returns the application version.
func (c *Config) GetAppVersion() string {
	return c.Application.Version
}

func (c *Config) IsDebug() bool {
	return c.Logging.Debug
}

func (c *Config) GetLogLevel() string {
	return c.Logging.Level
}

func (c *Config) GetUiConfig() UIConfig {
	return c.UI
}
