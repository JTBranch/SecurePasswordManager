package crypto

import (
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
)

type DeviceKeyManager struct {
	KeyPairGenerator      KeyPairGenerator
	KeyringProvider       KeyringProvider
	PEMProvider           PEMProvider
	DeviceName            string
	UserID                string
	DeviceKeyConfig       *devicekeys.DeviceKeyConfig // injected config
	DeviceKeyFileProvidor DeviceKeyFileProvidor
}
type KeyringProvider interface {
	Set(service, key, value string) error
	Get(service, key string) (string, error)
	Delete(service, key string) error
}

type DefaultKeyringProvider struct{}

func (k *DefaultKeyringProvider) Set(service, key, value string) error {
	return keyring.Set(service, key, value)
}
func (k *DefaultKeyringProvider) Get(service, key string) (string, error) {
	return keyring.Get(service, key)
}
func (k *DefaultKeyringProvider) Delete(service, key string) error {
	return keyring.Delete(service, key)
}

type DeviceKeyFileProvidor interface {
	SaveDeviceKeys(config *devicekeys.DeviceKeyConfig) error
	LoadDeviceKeys() (*devicekeys.DeviceKeyConfig, error)
}

func NewDeviceKeyManager(keyPairGenerator KeyPairGenerator,
	pemProvider PEMProvider,
	deviceKeyFileProvidor DeviceKeyFileProvidor,
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
	deviceKeyConfig, err := deviceKeyFileProvidor.LoadDeviceKeys()
	if err != nil {
		logger.Error("failed to load device key config: ", err)
		return nil, err
	}
	return &DeviceKeyManager{
		KeyPairGenerator:      keyPairGenerator,
		KeyringProvider:       &DefaultKeyringProvider{},
		PEMProvider:           pemProvider,
		DeviceName:            deviceName,
		UserID:                u.Username,
		DeviceKeyConfig:       deviceKeyConfig,
		DeviceKeyFileProvidor: deviceKeyFileProvidor,
	}, nil
}

func (m *DeviceKeyManager) SetKeyringProvider(keyringProvider KeyringProvider) {
	m.KeyringProvider = keyringProvider
}

func (m *DeviceKeyManager) SetDeviceKeyFileProvidor(provider DeviceKeyFileProvidor) {
	m.DeviceKeyFileProvidor = provider
}

// GenerateDeviceKey creates a new device key pair and returns the DeviceKey struct.
func (m *DeviceKeyManager) GenerateEncryptionDeviceKey() (*DeviceKey, error) {
	pub, priv, err := m.KeyPairGenerator.GenerateX25519KeyPair()
	if err != nil {
		logger.Error("failed to generate key pair: ", err)
		return nil, err
	}
	return m.GenerateDeviceKey(pub, priv, devicekeys.KeyTypeX25519Private, KeyringServiceEncryption)
}

// GenerateDeviceKey creates a new device key pair and returns the DeviceKey struct.
func (m *DeviceKeyManager) GenerateSigningDeviceKey() (*DeviceKey, error) {
	pub, priv, err := m.KeyPairGenerator.GenerateEd25519KeyPair()
	if err != nil {
		logger.Error("failed to generate key pair: ", err)
		return nil, err
	}
	return m.GenerateDeviceKey(pub, priv, devicekeys.KeyTypeEd25519Private, KeyringServiceSigning)
}

func (m *DeviceKeyManager) GenerateDeviceKey(pub, priv []byte, keyType devicekeys.KeyType, serviceName string) (*DeviceKey, error) {

	pemBytes, err := m.PEMProvider.EncodeKeyToPEM(priv, keyType)
	if err != nil {
		logger.Error("failed to encode private key to PEM: ", err)
		return nil, err
	}

	err = m.KeyringProvider.Set(serviceName, m.DeviceName, string(pemBytes))
	if err != nil {
		logger.Error("failed to set keyring: ", err)
		return nil, err
	}

	return &DeviceKey{
		ID:         uuid.New().String(),
		DeviceName: m.DeviceName,
		CreatedAt:  time.Now(),
		PublicKey:  pub,
		PrivateKey: priv,
		UserID:     m.UserID,
		KeyType:    keyType,
	}, nil
}

// GetDeviceKey loads the device key from secure storage.
func (m *DeviceKeyManager) GetEncryptionDeviceKey() (*DeviceKey, error) {

	key, err := m.KeyringProvider.Get(KeyringServiceEncryption, m.DeviceName)
	if err != nil {
		logger.Error("failed to get keyring: ", err)
		newKey, genErr := m.GenerateEncryptionDeviceKey()
		if genErr != nil {
			logger.Error("failed to save device keys: ", genErr)
			return nil, genErr
		}
		m.DeviceKeyConfig.CurrentKey.EncryptionKey = *cryptoToConfigKey(newKey, 1)
		m.DeviceKeyFileProvidor.SaveDeviceKeys(m.DeviceKeyConfig)
		return newKey, nil
	}

	decodedPriv, err := m.PEMProvider.DecodeKeyFromPEM([]byte(key), devicekeys.KeyTypeX25519Private)
	if err != nil {
		logger.Error("failed to decode private key from PEM: ", err)
		return nil, err
	}

	return &DeviceKey{
		ID:         uuid.New().String(),
		DeviceName: m.DeviceName,
		CreatedAt:  time.Now(),
		PublicKey:  m.DeviceKeyConfig.CurrentKey.EncryptionKey.PublicKey,
		PrivateKey: decodedPriv,
		UserID:     m.UserID,
		KeyType:    devicekeys.KeyTypeX25519,
	}, nil
}

// GetDeviceKey loads the device key from secure storage.
func (m *DeviceKeyManager) GetSigningDeviceKey() (*DeviceKey, error) {

	key, err := m.KeyringProvider.Get(KeyringServiceSigning, m.DeviceName)
	if err != nil {
		logger.Error("failed to get keyring: ", err)
		newKey, genErr := m.GenerateSigningDeviceKey()
		if genErr != nil {
			logger.Error("failed to save device keys: ", genErr)
			return nil, genErr
		}
		m.DeviceKeyConfig.CurrentKey.SigningKey = *cryptoToConfigKey(newKey, 1)
		m.DeviceKeyFileProvidor.SaveDeviceKeys(m.DeviceKeyConfig)
		return newKey, nil
	}

	decodedPriv, err := m.PEMProvider.DecodeKeyFromPEM([]byte(key), devicekeys.KeyTypeEd25519Private)
	if err != nil {
		logger.Error("failed to decode private key from PEM: ", err)
		return nil, err
	}

	return &DeviceKey{
		ID:         uuid.New().String(),
		DeviceName: m.DeviceName,
		CreatedAt:  time.Now(),
		PublicKey:  m.DeviceKeyConfig.CurrentKey.SigningKey.PublicKey,
		PrivateKey: decodedPriv,
		UserID:     m.UserID,
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
	return m.KeyringProvider.Delete(KeyringServiceSigning, m.DeviceName)
}

// DeleteDeviceKey deletes the device key from secure storage (for reset/compromise).
func (m *DeviceKeyManager) DeleteEncryptionDeviceKey() error {
	return m.KeyringProvider.Delete(KeyringServiceEncryption, m.DeviceName)
}

// RotateDeviceKey generates a new device key, archives the old one, and returns the new key.
func (m *DeviceKeyManager) RotateEncryptionDeviceKey(deviceName, userID, keyType string) (*DeviceKey, error) {
	oldKey, err := m.KeyringProvider.Get(KeyringServiceEncryption, m.DeviceName)
	if err != nil {
		return m.GetEncryptionDeviceKey()
	}
	oldKeyID := m.DeviceKeyConfig.CurrentKey.EncryptionKey.KeyID
	archivedService := KeyringServiceEncryption + "-archived-" + m.DeviceName
	err = m.KeyringProvider.Set(archivedService, oldKeyID, oldKey)
	if err != nil {
		logger.Error("failed to store old private key in keychain: ", err)
		return nil, err
	}
	newKey, genErr := m.GenerateEncryptionDeviceKey()

	if genErr != nil {
		logger.Error("failed to save device keys: ", genErr)
		return nil, genErr
	}

	m.DeviceKeyConfig.Archived = append(m.DeviceKeyConfig.Archived, *m.DeviceKeyConfig.CurrentKey)
	version := m.DeviceKeyConfig.CurrentKey.EncryptionKey.Version

	m.DeviceKeyConfig.CurrentKey.EncryptionKey = *cryptoToConfigKey(newKey, version+1)
	m.DeviceKeyFileProvidor.SaveDeviceKeys(m.DeviceKeyConfig)

	return newKey, nil
}

// RotateDeviceKey generates a new device key, archives the old one, and returns the new key.
func (m *DeviceKeyManager) RotateSigningDeviceKey(deviceName, userID, keyType string) (*DeviceKey, error) {
	oldKey, err := m.KeyringProvider.Get(KeyringServiceSigning, m.DeviceName)
	if err != nil {
		return m.GetSigningDeviceKey()
	}
	oldKeyID := m.DeviceKeyConfig.CurrentKey.SigningKey.KeyID
	archivedService := KeyringServiceSigning + "-archived-" + m.DeviceName
	err = m.KeyringProvider.Set(archivedService, oldKeyID, oldKey)
	if err != nil {
		logger.Error("failed to store old private key in keychain: ", err)
		return nil, err
	}
	newKey, genErr := m.GenerateSigningDeviceKey()

	if genErr != nil {
		logger.Error("failed to save device keys: ", genErr)
		return nil, genErr
	}

	m.DeviceKeyConfig.Archived = append(m.DeviceKeyConfig.Archived, *m.DeviceKeyConfig.CurrentKey)
	version := m.DeviceKeyConfig.CurrentKey.SigningKey.Version

	m.DeviceKeyConfig.CurrentKey.SigningKey = *cryptoToConfigKey(newKey, version+1)
	m.DeviceKeyFileProvidor.SaveDeviceKeys(m.DeviceKeyConfig)

	return newKey, nil
}

func (m *DeviceKeyManager) GetDeviceName() string {
	return m.DeviceName
}

func (m *DeviceKeyManager) GetAppUser() string {
	return m.UserID
}
