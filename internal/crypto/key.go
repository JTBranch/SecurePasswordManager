package crypto

import (
    "crypto/rand"
    "errors"
    "go-password-manager/internal/config"
    "os"
    "path/filepath"
)

const appName = "GoPasswordManager"

func keyFilePath(keyUUID string) (string, error) {
    configDir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    appConfigDir := filepath.Join(configDir, appName)
    if err := os.MkdirAll(appConfigDir, 0700); err != nil {
        return "", err
    }
    return filepath.Join(appConfigDir, "."+keyUUID), nil // Obfuscated file name
}

func LoadOrCreateKey() ([]byte, error) {
    cfgService, err := config.NewConfigService()
    if err != nil {
        return nil, err
    }
    keyUUID := cfgService.Config.KeyUUID
    path, err := keyFilePath(keyUUID)
    if err != nil {
        return nil, err
    }
    if _, err := os.Stat(path); os.IsNotExist(err) {
        key := make([]byte, 32) // AES-256
        _, err := rand.Read(key)
        if err != nil {
            return nil, err
        }
        if err := os.WriteFile(path, key, 0600); err != nil {
            return nil, err
        }
        return key, nil
    }
    key, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    if len(key) != 32 {
        return nil, errors.New("invalid key length")
    }
    return key, nil
}
