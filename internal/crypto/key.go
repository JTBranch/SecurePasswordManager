package crypto

import (
	"crypto/rand"
	"fmt"
	buildconfig "go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/config/secretkeymetadata"
	"go-password-manager/internal/logger"
	"time"

	"github.com/google/uuid"
)

const (
	KeyringSecretsEncryption = AppName + "secrets-encryption"
)

type SecretsKeyMetadataProvider interface {
	UpdateCurrentKey(newKey secretkeymetadata.SecretsEncryptionKeyMetadata) *secretkeymetadata.SecretsKeyMetadataBundle
}

type SecretsEncryptionKeyManager struct {
	configProvider             ConfigProvider
	keyringProvider            KeyringProvider
	secretsKeyMetadataProvider SecretsKeyMetadataProvider
	keyUUID                    string
	keySize                    int
}

// UsingInMemoryKeyring reports whether the manager is currently configured to use the in-memory keyring provider.
func (m *SecretsEncryptionKeyManager) UsingInMemoryKeyring() bool {
	_, ok := m.keyringProvider.(*InMemoryKeyringProvider)
	return ok
}

func NewSecretsEncryptionKeyManager(configProvider ConfigProvider,
	secretsKeyMetadataProvider SecretsKeyMetadataProvider) (*SecretsEncryptionKeyManager, error) {
	keyUUID := configProvider.GetKeyUUID()
	buildCfg, err := buildconfig.Load()
	if err != nil {
		return nil, err
	}
	mgr := &SecretsEncryptionKeyManager{
		configProvider:             configProvider,
		keyringProvider:            &DefaultkeyringProvider{},
		secretsKeyMetadataProvider: secretsKeyMetadataProvider,
		keyUUID:                    keyUUID,
		keySize:                    buildCfg.Security.Encryption.KeySize,
	}
	// Honor build configuration for in-memory keyring usage (tests/dev isolation)
	if buildCfg.Security.Keyring.InMemory {
		logger.Debug("Using in-memory keyring")
		mgr.keyringProvider = NewInMemoryKeyring()
	}
	return mgr, nil
}

func (m *SecretsEncryptionKeyManager) SetKeyringProvider(keyringProvider KeyringProvider) {
	m.keyringProvider = keyringProvider
}

func (m *SecretsEncryptionKeyManager) CreateSymmetricKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating new encryption key: ", string(key))
	return key, nil
}

// LoadOrCreateKey loads an existing encryption key or creates a new one
func (m *SecretsEncryptionKeyManager) LoadOrCreateKey() ([]byte, error) {
	key, err := m.keyringProvider.Get(KeyringSecretsEncryption, m.keyUUID)

	if err != nil {
		return m.createKey()
	} else {
		return []byte(key), nil
	}
}

func (m *SecretsEncryptionKeyManager) createKey() ([]byte, error) {
	// Create new key
	keySize := m.keySize
	if keySize == 0 {
		keySize = 32 // Default to AES-256
	}
	createdKey, err := m.CreateSymmetricKey(keySize)
	if err != nil {
		logger.Error("Failed to create symetric key", err)
		return nil, err
	}
	err = m.keyringProvider.Set(KeyringSecretsEncryption, m.keyUUID, string(createdKey))
	now := time.Now()
	m.secretsKeyMetadataProvider.UpdateCurrentKey(secretkeymetadata.SecretsEncryptionKeyMetadata{
		UUID:         m.keyUUID,
		DateCreated:  now,
		DateModified: now,
	})
	return createdKey, err

}

func (m *SecretsEncryptionKeyManager) RotateKey() ([]byte, error) {
	oldKey, err := m.keyringProvider.Get(KeyringSecretsEncryption, m.keyUUID)
	if err != nil {
		return m.createKey()
	} else {
		newKeyUUID, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("failed to generate KeyUUID: %w", err)
		}
		oldUUID := m.keyUUID
		m.keyUUID = newKeyUUID.String()
		newKey, err := m.createKey()
		if err != nil {
			logger.Error("Failed to generate a new Encryption key", err)
			return nil, err
		}
		err = m.keyringProvider.Set(KeyringSecretsEncryption+"-archived", oldUUID, string(oldKey))
		return newKey, err
	}
}
