package env

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultSecretsFileName = "secrets.json"
	DefaultConfigFileName  = "app.config"
)

// Config holds all environment configuration
type Config struct {
	// Application
	Environment string
	AppVersion  string
	AppName     string

	// UI Configuration
	DefaultWindowWidth  int
	DefaultWindowHeight int

	// Development
	DebugLogging bool
	HotReload    bool

	// File Paths
	SecretsFilePath string
	ConfigFilePath  string

	// Testing
	TestDataDir    string
	E2ETestTimeout time.Duration

	// Security
	EncryptionKeySize int
}

var globalConfig *Config

// Load loads environment variables from .env files and system environment
func Load() (*Config, error) {
	// Load .env files in order of precedence (later files override earlier ones)
	envFiles := []string{
		".env.example", // Default values
		".env",         // General environment
		".env.local",   // Local overrides (git ignored)
	}

	// Load each file that exists, ignore missing files
	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			_ = godotenv.Load(file) // Ignore errors for optional files
		}
	}

	config := &Config{
		Environment:         getEnv("GO_PASSWORD_MANAGER_ENV", "dev"),
		AppVersion:          getEnv("APP_VERSION", "1.0.0"),
		AppName:             getEnv("APP_NAME", "GoPasswordManager"),
		DefaultWindowWidth:  getEnvInt("DEFAULT_WINDOW_WIDTH", 1600),
		DefaultWindowHeight: getEnvInt("DEFAULT_WINDOW_HEIGHT", 900),
		DebugLogging:        getEnvBool("DEBUG_LOGGING", true),
		HotReload:           getEnvBool("HOT_RELOAD", false),
		SecretsFilePath:     getEnv("SECRETS_FILE_PATH", ""),
		ConfigFilePath:      getEnv("CONFIG_FILE_PATH", ""),
		TestDataDir:         getEnv("TEST_DATA_DIR", ""),
		E2ETestTimeout:      getEnvDuration("E2E_TEST_TIMEOUT", 30*time.Second),
		EncryptionKeySize:   getEnvInt("ENCRYPTION_KEY_SIZE", 32),
	}

	globalConfig = config
	return config, nil
}

// Get returns the global config (must call Load() first)
func Get() *Config {
	if globalConfig == nil {
		// Auto-load if not already loaded
		Load()
	}
	return globalConfig
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "prod" || c.Environment == "production"
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "dev" || c.Environment == "development"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	return c.Environment == "test" || c.Environment == "integration-test" || c.Environment == "e2e-test"
}

// GetSecretsFilePath returns the secrets file path based on environment
func (c *Config) GetSecretsFilePath() string {
	if c.SecretsFilePath != "" {
		return c.SecretsFilePath
	}

	if c.IsTest() && c.TestDataDir != "" {
		return filepath.Join(c.TestDataDir, DefaultSecretsFileName)
	}

	if c.IsDevelopment() {
		return DefaultSecretsFileName // Current directory
	}

	// Production: use OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return DefaultSecretsFileName // Fallback
	}
	appConfigDir := filepath.Join(configDir, c.AppName)
	os.MkdirAll(appConfigDir, 0700)
	return filepath.Join(appConfigDir, DefaultSecretsFileName)
}

// GetConfigFilePath returns the config file path based on environment
func (c *Config) GetConfigFilePath() string {
	if c.ConfigFilePath != "" {
		return c.ConfigFilePath
	}

	if c.IsTest() && c.TestDataDir != "" {
		return filepath.Join(c.TestDataDir, DefaultConfigFileName)
	}

	if c.IsDevelopment() {
		return DefaultConfigFileName // Current directory
	}

	// Production: use OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return DefaultConfigFileName // Fallback
	}
	appConfigDir := filepath.Join(configDir, c.AppName)
	os.MkdirAll(appConfigDir, 0700)
	return filepath.Join(appConfigDir, DefaultConfigFileName)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
