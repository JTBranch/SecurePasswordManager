package secrets

import (
	"errors"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/crypto"
	"sync"
	"time"
)

type SecretsManager struct {
	secrets      map[string]*domain.Secret
	mu           sync.RWMutex
	encryptionKey []byte
}

func NewSecretsManager(encryptionKey string) *SecretsManager {
	return &SecretsManager{
		secrets:      make(map[string]*domain.Secret),
		encryptionKey: []byte(encryptionKey),
	}
}

func (sm *SecretsManager) AddSecret(key, value, createdBy string, secretType domain.SecretType) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	encryptedValue, err := crypto.Encrypt([]byte(value), sm.encryptionKey)
	if err != nil {
		return err
	}

	version := domain.SecretVersion{
		Value:     string(encryptedValue),
		Version:   1,
		CreatedAt: time.Now().Format(time.RFC3339), // Set to current time
		CreatedBy: createdBy,
	}

	secret := &domain.Secret{
		Key:      key,
		Type:     secretType,
		Versions: []domain.SecretVersion{version},
	}

	sm.secrets[key] = secret
	return nil
}

func (sm *SecretsManager) UpdateSecret(key, value, createdBy string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	secret, exists := sm.secrets[key]
	if !exists {
		return errors.New("secret not found")
	}

	encryptedValue, err := crypto.Encrypt([]byte(value), sm.encryptionKey)
	if err != nil {
		return err
	}

	latestVersion := len(secret.Versions) + 1
	version := domain.SecretVersion{
		Value:     string(encryptedValue),
		Version:   latestVersion,
		CreatedAt: time.Now().Format(time.RFC3339), // Set to current time
		CreatedBy: createdBy,
	}

	secret.Versions = append(secret.Versions, version)
	return nil
}

func (sm *SecretsManager) GetSecret(key string) (string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	secret, exists := sm.secrets[key]
	if !exists || len(secret.Versions) == 0 {
		return "", errors.New("secret not found")
	}
	plainBytes, err := crypto.Decrypt(secret.Versions[len(secret.Versions)-1].Value, sm.encryptionKey)
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}

func (sm *SecretsManager) DeleteSecret(key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.secrets[key]; !exists {
		return errors.New("secret not found")
	}

	delete(sm.secrets, key)
	return nil
}

func (sm *SecretsManager) ListSecrets() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	keys := make([]string, 0, len(sm.secrets))
	for key := range sm.secrets {
		keys = append(keys, key)
	}
	return keys
}