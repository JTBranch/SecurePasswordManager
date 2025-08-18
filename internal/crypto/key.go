package crypto

import (
	"crypto/rand"
	"errors"
	"go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"os"
	"path/filepath"
)

func keyFilePath(keyUUID string) (string, error) {
	buildCfg, err := buildconfig.Load()
	if err != nil {
		return "", err
	}

	if buildCfg.IsTest() && buildCfg.Testing.DataDir != "" {
		// For tests, use test data directory
		keyDir := filepath.Join(buildCfg.Testing.DataDir, "keys")
		if err := os.MkdirAll(keyDir, 0700); err != nil {
			return "", err
		}
		return filepath.Join(keyDir, "."+keyUUID), nil
	}

	if buildCfg.IsDevelopment() {
		// For development, use current directory
		keyDir := "keys"
		if err := os.MkdirAll(keyDir, 0700); err != nil {
			return "", err
		}
		return filepath.Join(keyDir, "."+keyUUID), nil
	}

	// For production, use OS-specific config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appConfigDir := filepath.Join(configDir, buildCfg.Application.Name)
	if err := os.MkdirAll(appConfigDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(appConfigDir, "."+keyUUID), nil // Obfuscated file name
}

// LoadOrCreateKey loads an existing encryption key or creates a new one
func LoadOrCreateKey(cfgService *config.ConfigService) ([]byte, error) {
	buildCfg, err := buildconfig.Load()
	if err != nil {
		return nil, err
	}

	// Generate a default key UUID if config service is not available
	keyUUID := "default-key"

	// Try to get the actual key UUID from config service
	if cfgService != nil && cfgService.Config.KeyUUID != "" {
		keyUUID = cfgService.Config.KeyUUID
	}

	path, err := keyFilePath(keyUUID)
	if err != nil {
		return nil, err
	}

	// Check if key file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create new key
		keySize := buildCfg.Security.Encryption.KeySize
		if keySize == 0 {
			keySize = 32 // Default to AES-256
		}

		key := make([]byte, keySize)
		_, err := rand.Read(key)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(path, key, 0600); err != nil {
			return nil, err
		}

		return key, nil
	}

	// Load existing key
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Validate key size
	if len(key) == 0 {
		return nil, errors.New("encryption key is empty")
	}

	return key, nil
}
