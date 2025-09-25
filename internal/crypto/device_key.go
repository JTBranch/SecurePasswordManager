package crypto

import (
	buildconfig "go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/config/devicekeys"
	"go-password-manager/internal/logger"
	"os"
	"os/user"
	"time"

	"github.com/google/uuid"
	"github.com/zalando/go-keyring"
)

// DeviceKey holds the device's asymmetric key pair and metadata.
type DeviceKey struct {
	ID         string             // UUID for this device key
	DeviceName string             // Human-readable device name
	CreatedAt  time.Time          // Timestamp of key creation
	PublicKey  []byte             // Device public key (e.g., X25519 or Ed25519)
	PrivateKey []byte             // Device private key (store securely)
	UserID     string             // (Optional) User ID associated with device
	KeyType    devicekeys.KeyType // (Optional) "X25519", "Ed25519", etc.
}
type KeyPairGenerator interface {
	GenerateX25519KeyPair() (publicKey []byte, privateKey []byte, err error)
	GenerateEd25519KeyPair() (publicKey []byte, privateKey []byte, err error)
}

const (
	AppName                  = "GoPasswordManager"
	KeyringServiceEncryption = AppName + "-encryption"
	KeyringServiceSigning    = AppName + "-signing"
	logFailedSaveKeys        = "failed to save device keys: "
)

type DeviceKeyManager struct {
	keyPairGenerator      KeyPairGenerator
	keyringProvider       KeyringProvider
	pemProvider           PEMProvider
	deviceName            string
	userID                string
	deviceKeyConfig       *devicekeys.DeviceKeyConfig // injected config
	deviceKeyFileProvider DeviceKeyFileProvider
}

// UsingInMemoryProviders reports whether both keyring and device key file providers are in-memory implementations.
func (m *DeviceKeyManager) UsingInMemoryProviders() bool {
	if m == nil {
		return false
	}
	_, keyringIsMem := m.keyringProvider.(*InMemoryKeyringProvider)
	_, fileIsMem := m.deviceKeyFileProvider.(*InMemoryDeviceKeyFileProvider)
	return keyringIsMem && fileIsMem
}

type KeyringProvider interface {
	Set(service, key, value string) error
	Get(service, key string) (string, error)
	Delete(service, key string) error
}

type DefaultkeyringProvider struct{}

func (k *DefaultkeyringProvider) Set(service, key, value string) error {
	return keyring.Set(service, key, value)
}
func (k *DefaultkeyringProvider) Get(service, key string) (string, error) {
	return keyring.Get(service, key)
}
func (k *DefaultkeyringProvider) Delete(service, key string) error {
	return keyring.Delete(service, key)
}

// InMemoryKeyringProvider is a test/dev implementation storing keys in-memory only.
type InMemoryKeyringProvider struct{ store map[string]map[string]string }

func NewInMemoryKeyring() *InMemoryKeyringProvider {
	return &InMemoryKeyringProvider{store: map[string]map[string]string{}}
}

func (m *InMemoryKeyringProvider) Set(service, key, value string) error {
	if _, ok := m.store[service]; !ok {
		m.store[service] = map[string]string{}
	}
	m.store[service][key] = value
	return nil
}
func (m *InMemoryKeyringProvider) Get(service, key string) (string, error) {
	if svc, ok := m.store[service]; ok {
		if v, ok2 := svc[key]; ok2 {
			return v, nil
		}
	}
	return "", os.ErrNotExist
}
func (m *InMemoryKeyringProvider) Delete(service, key string) error {
	if svc, ok := m.store[service]; ok {
		delete(svc, key)
	}
	return nil
}

type DeviceKeyFileProvider interface {
	SaveDeviceKeys(config *devicekeys.DeviceKeyConfig) error
	LoadDeviceKeys() (*devicekeys.DeviceKeyConfig, error)
}

// InMemoryDeviceKeyFileProvider is a test-only provider storing keys in memory.
type InMemoryDeviceKeyFileProvider struct{ cfg *devicekeys.DeviceKeyConfig }

func NewInMemoryDeviceKeyFileProvider() *InMemoryDeviceKeyFileProvider {
	return &InMemoryDeviceKeyFileProvider{cfg: &devicekeys.DeviceKeyConfig{CurrentKey: &devicekeys.DeviceKeyBundle{}, Archived: []devicekeys.DeviceKeyBundle{}}}
}
func (p *InMemoryDeviceKeyFileProvider) SaveDeviceKeys(c *devicekeys.DeviceKeyConfig) error {
	p.cfg = c
	return nil
}
func (p *InMemoryDeviceKeyFileProvider) LoadDeviceKeys() (*devicekeys.DeviceKeyConfig, error) {
	return p.cfg, nil
}

func NewDeviceKeyManager(keyPairGenerator KeyPairGenerator,
	pemProvider PEMProvider,
	deviceKeyFileProvider DeviceKeyFileProvider,
) (*DeviceKeyManager, error) {
	u, err := user.Current()
	if err != nil {
		logger.Error("failed to get current user: ", err)
		return nil, err
	}
	deviceName, err := os.Hostname()
	if err != nil {
		logger.Error("failed to get hostname: ", err)
		return nil, err
	}
	deviceKeyConfig, err := deviceKeyFileProvider.LoadDeviceKeys()
	if err != nil {
		logger.Error("failed to load device key config: ", err)
		return nil, err
	}
	// Allocate config defensively if provider returned nil (should not normally happen) and ensure CurrentKey bundle exists.
	if deviceKeyConfig == nil {
		deviceKeyConfig = &devicekeys.DeviceKeyConfig{CurrentKey: &devicekeys.DeviceKeyBundle{}, Archived: []devicekeys.DeviceKeyBundle{}}
	} else if deviceKeyConfig.CurrentKey == nil {
		deviceKeyConfig.CurrentKey = &devicekeys.DeviceKeyBundle{}
	}
	mgr := &DeviceKeyManager{
		keyPairGenerator:      keyPairGenerator,
		keyringProvider:       &DefaultkeyringProvider{},
		pemProvider:           pemProvider,
		deviceName:            deviceName,
		userID:                u.Username,
		deviceKeyConfig:       deviceKeyConfig,
		deviceKeyFileProvider: deviceKeyFileProvider,
	}
	// Respect build configuration for in-memory keyring/device key storage.
	if buildCfg, err := buildconfig.Load(); err == nil {
		if buildCfg.Security.Keyring.InMemory {
			mgr.keyringProvider = NewInMemoryKeyring()
			mgr.deviceKeyFileProvider = NewInMemoryDeviceKeyFileProvider()
		}
	}
	return mgr, nil
}

// test detection logic removed: selection of in-memory key handling is controlled exclusively by build configuration.

func (m *DeviceKeyManager) SetKeyringProvider(keyringProvider KeyringProvider) {
	m.keyringProvider = keyringProvider
}

func (m *DeviceKeyManager) SetDeviceKeyFileProvider(provider DeviceKeyFileProvider) {
	m.deviceKeyFileProvider = provider
}

// GenerateDeviceKey creates a new device key pair and returns the DeviceKey struct.
func (m *DeviceKeyManager) GenerateEncryptionDeviceKey() (*DeviceKey, error) {
	pub, priv, err := m.keyPairGenerator.GenerateX25519KeyPair()
	if err != nil {
		logger.Error("failed to generate key pair: ", err)
		return nil, err
	}
	return m.GenerateDeviceKey(pub, priv, devicekeys.KeyTypeX25519Private, KeyringServiceEncryption)
}

// GenerateDeviceKey creates a new device key pair and returns the DeviceKey struct.
func (m *DeviceKeyManager) GenerateSigningDeviceKey() (*DeviceKey, error) {
	pub, priv, err := m.keyPairGenerator.GenerateEd25519KeyPair()
	if err != nil {
		logger.Error("failed to generate key pair: ", err)
		return nil, err
	}
	return m.GenerateDeviceKey(pub, priv, devicekeys.KeyTypeEd25519Private, KeyringServiceSigning)
}

func (m *DeviceKeyManager) GenerateDeviceKey(pub, priv []byte, keyType devicekeys.KeyType, serviceName string) (*DeviceKey, error) {

	pemBytes, err := m.pemProvider.EncodeKeyToPEM(priv, keyType)
	if err != nil {
		logger.Error("failed to encode private key to PEM: ", err)
		return nil, err
	}

	err = m.keyringProvider.Set(serviceName, m.deviceName, string(pemBytes))
	if err != nil {
		logger.Error("failed to set keyring: ", err)
		return nil, err
	}

	return &DeviceKey{
		ID:         uuid.New().String(),
		DeviceName: m.deviceName,
		CreatedAt:  time.Now(),
		PublicKey:  pub,
		PrivateKey: priv,
		UserID:     m.userID,
		KeyType:    keyType,
	}, nil
}

// GetDeviceKey loads the device key from secure storage.
func (m *DeviceKeyManager) GetEncryptionDeviceKey() (*DeviceKey, error) {
	// Defensive: guarantee config and CurrentKey bundle exist
	if m.deviceKeyConfig == nil {
		m.deviceKeyConfig = &devicekeys.DeviceKeyConfig{CurrentKey: &devicekeys.DeviceKeyBundle{}, Archived: []devicekeys.DeviceKeyBundle{}}
	} else if m.deviceKeyConfig.CurrentKey == nil {
		m.deviceKeyConfig.CurrentKey = &devicekeys.DeviceKeyBundle{}
	}

	key, err := m.keyringProvider.Get(KeyringServiceEncryption, m.deviceName)
	if err != nil {
		logger.Error("failed to get keyring: ", err)
		newKey, genErr := m.GenerateEncryptionDeviceKey()
		if genErr != nil {
			logger.Error(logFailedSaveKeys, genErr)
			return nil, genErr
		}
		if m.deviceKeyConfig.CurrentKey == nil { // second guard in unlikely race
			m.deviceKeyConfig.CurrentKey = &devicekeys.DeviceKeyBundle{}
		}
		m.deviceKeyConfig.CurrentKey.EncryptionKey = *cryptoToConfigKey(newKey, 1)
		err = m.deviceKeyFileProvider.SaveDeviceKeys(m.deviceKeyConfig)
		return newKey, err
	}

	decodedPriv, err := m.pemProvider.DecodeKeyFromPEM([]byte(key), devicekeys.KeyTypeX25519Private)
	if err != nil {
		logger.Error("failed to decode private key from PEM: ", err)
		return nil, err
	}

	// Reuse persisted key ID if present to provide stable device identity.
	persistedID := m.deviceKeyConfig.CurrentKey.EncryptionKey.KeyID
	if persistedID == "" {
		persistedID = uuid.New().String()
		m.deviceKeyConfig.CurrentKey.EncryptionKey.KeyID = persistedID
		// Save immediately to persist stable device identity for future sessions.
		if err := m.deviceKeyFileProvider.SaveDeviceKeys(m.deviceKeyConfig); err != nil {
			logger.Error("failed to persist new encryption key ID: ", err)
		}
	}
	return &DeviceKey{
		ID:         persistedID,
		DeviceName: m.deviceName,
		CreatedAt:  time.Now(),
		PublicKey:  m.deviceKeyConfig.CurrentKey.EncryptionKey.PublicKey,
		PrivateKey: decodedPriv,
		UserID:     m.userID,
		KeyType:    devicekeys.KeyTypeX25519,
	}, nil
}

// GetDeviceKey loads the device key from secure storage.
func (m *DeviceKeyManager) GetSigningDeviceKey() (*DeviceKey, error) {
	if m.deviceKeyConfig == nil {
		m.deviceKeyConfig = &devicekeys.DeviceKeyConfig{CurrentKey: &devicekeys.DeviceKeyBundle{}, Archived: []devicekeys.DeviceKeyBundle{}}
	} else if m.deviceKeyConfig.CurrentKey == nil {
		m.deviceKeyConfig.CurrentKey = &devicekeys.DeviceKeyBundle{}
	}

	key, err := m.keyringProvider.Get(KeyringServiceSigning, m.deviceName)
	if err != nil {
		logger.Error("failed to get keyring: ", err)
		newKey, genErr := m.GenerateSigningDeviceKey()
		if genErr != nil {
			logger.Error(logFailedSaveKeys, genErr)
			return nil, genErr
		}
		if m.deviceKeyConfig.CurrentKey == nil {
			m.deviceKeyConfig.CurrentKey = &devicekeys.DeviceKeyBundle{}
		}
		m.deviceKeyConfig.CurrentKey.SigningKey = *cryptoToConfigKey(newKey, 1)
		err = m.deviceKeyFileProvider.SaveDeviceKeys(m.deviceKeyConfig)
		return newKey, err
	}

	decodedPriv, err := m.pemProvider.DecodeKeyFromPEM([]byte(key), devicekeys.KeyTypeEd25519Private)
	if err != nil {
		logger.Error("failed to decode private key from PEM: ", err)
		return nil, err
	}

	return &DeviceKey{
		ID:         uuid.New().String(),
		DeviceName: m.deviceName,
		CreatedAt:  time.Now(),
		PublicKey:  m.deviceKeyConfig.CurrentKey.SigningKey.PublicKey,
		PrivateKey: decodedPriv,
		UserID:     m.userID,
		KeyType:    devicekeys.KeyTypeEd25519,
	}, nil
}

func cryptoToConfigKey(cryptoKey *DeviceKey, version int) *devicekeys.DeviceKey {
	return &devicekeys.DeviceKey{
		KeyID:     cryptoKey.ID,
		PublicKey: cryptoKey.PublicKey,
		KeyType:   cryptoKey.KeyType,
		CreatedAt: cryptoKey.CreatedAt,
		Version:   version,
	}
}

// DeleteDeviceKey deletes the device key from secure storage (for reset/compromise).
func (m *DeviceKeyManager) DeleteSigningDeviceKey() error {
	return m.keyringProvider.Delete(KeyringServiceSigning, m.deviceName)
}

// DeleteDeviceKey deletes the device key from secure storage (for reset/compromise).
func (m *DeviceKeyManager) DeleteEncryptionDeviceKey() error {
	return m.keyringProvider.Delete(KeyringServiceEncryption, m.deviceName)
}

// RotateDeviceKey generates a new device key, archives the old one, and returns the new key.
func (m *DeviceKeyManager) RotateEncryptionDeviceKey(deviceName, userID, keyType string) (*DeviceKey, error) {
	oldKey, err := m.keyringProvider.Get(KeyringServiceEncryption, m.deviceName)
	if err != nil {
		return m.GetEncryptionDeviceKey()
	}
	oldKeyID := m.deviceKeyConfig.CurrentKey.EncryptionKey.KeyID
	archivedService := KeyringServiceEncryption + "-archived-" + m.deviceName
	err = m.keyringProvider.Set(archivedService, oldKeyID, oldKey)
	if err != nil {
		logger.Error("failed to store old private key in keychain: ", err)
		return nil, err
	}
	newKey, genErr := m.GenerateEncryptionDeviceKey()

	if genErr != nil {
		logger.Error(logFailedSaveKeys, genErr)
		return nil, genErr
	}

	m.deviceKeyConfig.Archived = append(m.deviceKeyConfig.Archived, *m.deviceKeyConfig.CurrentKey)
	version := m.deviceKeyConfig.CurrentKey.EncryptionKey.Version

	m.deviceKeyConfig.CurrentKey.EncryptionKey = *cryptoToConfigKey(newKey, version+1)
	m.deviceKeyFileProvider.SaveDeviceKeys(m.deviceKeyConfig)

	return newKey, nil
}

// RotateDeviceKey generates a new device key, archives the old one, and returns the new key.
func (m *DeviceKeyManager) RotateSigningDeviceKey(deviceName, userID, keyType string) (*DeviceKey, error) {
	oldKey, err := m.keyringProvider.Get(KeyringServiceSigning, m.deviceName)
	if err != nil {
		return m.GetSigningDeviceKey()
	}
	oldKeyID := m.deviceKeyConfig.CurrentKey.SigningKey.KeyID
	archivedService := KeyringServiceSigning + "-archived-" + m.deviceName
	err = m.keyringProvider.Set(archivedService, oldKeyID, oldKey)
	if err != nil {
		logger.Error("failed to store old private key in keychain: ", err)
		return nil, err
	}
	newKey, genErr := m.GenerateSigningDeviceKey()

	if genErr != nil {
		logger.Error(logFailedSaveKeys, genErr)
		return nil, genErr
	}

	m.deviceKeyConfig.Archived = append(m.deviceKeyConfig.Archived, *m.deviceKeyConfig.CurrentKey)
	version := m.deviceKeyConfig.CurrentKey.SigningKey.Version

	m.deviceKeyConfig.CurrentKey.SigningKey = *cryptoToConfigKey(newKey, version+1)
	m.deviceKeyFileProvider.SaveDeviceKeys(m.deviceKeyConfig)

	return newKey, nil
}

func (m *DeviceKeyManager) GetDeviceName() string {
	return m.deviceName
}

func (m *DeviceKeyManager) GetAppUser() string {
	return m.userID
}
