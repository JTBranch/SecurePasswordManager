package devicekeys

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type KeyType string

const (
	KeyTypeX25519         KeyType = "X25519"
	KeyTypeEd25519        KeyType = "Ed25519"
	KeyTypeX25519Private  KeyType = "X25519 PRIVATE KEY"
	KeyTypeX25519Public   KeyType = "X25519 PUBLIC KEY"
	KeyTypeEd25519Private KeyType = "ED25519 PRIVATE KEY"
	KeyTypeEd25519Public  KeyType = "ED25519 PUBLIC KEY"
	KeyTypeSymmetric      KeyType = "SYMMETRIC KEY"
)

type DeviceKeyConfig struct {
	CurrentKey *DeviceKeyBundle  `json:"currentKey"`
	Archived   []DeviceKeyBundle `json:"archived"`
}

type DeviceKeyBundle struct {
	EncryptionKey DeviceKey
	SigningKey    DeviceKey
	// ...other key types if needed
}

// DeviceKey holds non-sensitive device key metadata for runtime config
type DeviceKey struct {
	KeyID     string    `json:"keyId"`
	PublicKey []byte    `json:"publicKey"`
	KeyType   KeyType   `json:"keyType"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int       `json:"version"`
}

type BuildConfigProvider interface {
	GetDeviceKeysFilePath() string
}

type DeviceKeyFileService struct {
	buildCfg BuildConfigProvider
	filePath string
}

func NewDeviceKeyFileService(buildCfg BuildConfigProvider) (*DeviceKeyFileService, error) {
	filePath := buildCfg.GetDeviceKeysFilePath()
	return &DeviceKeyFileService{
		buildCfg: buildCfg,
		filePath: filePath,
	}, nil
}

func (svc *DeviceKeyFileService) SaveDeviceKeys(config *DeviceKeyConfig) error {
	if err := os.MkdirAll(filepath.Dir(svc.filePath), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(svc.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(config)
}

func (svc *DeviceKeyFileService) LoadDeviceKeys() (*DeviceKeyConfig, error) {
	data, err := os.ReadFile(svc.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config and file
			defaultConfig := &DeviceKeyConfig{
				CurrentKey: nil,
				Archived:   []DeviceKeyBundle{},
			}
			jsonData, marshalErr := json.MarshalIndent(defaultConfig, "", "  ")
			if marshalErr != nil {
				return nil, marshalErr
			}
			writeErr := os.WriteFile(svc.filePath, jsonData, 0600)
			if writeErr != nil {
				return nil, writeErr
			}
			return defaultConfig, nil
		}
		return nil, err
	}
	var config DeviceKeyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
