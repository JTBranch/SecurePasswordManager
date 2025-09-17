package devicekeys_test

import (
	"encoding/json"
	"go-password-manager/internal/config/devicekeys"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockBuildConfigProvider struct {
	filePath string
}

func (m *mockBuildConfigProvider) GetDeviceKeysFilePath() string {
	return m.filePath
}

func testDeviceKeyConfig() *devicekeys.DeviceKeyConfig {
	encryptionKey := &devicekeys.DeviceKey{
		KeyID:     "test-id",
		PublicKey: []byte{1, 2, 3},
		KeyType:   "X25519",
		CreatedAt: time.Now().UTC(),
		Version:   1,
	}
	signingKey := &devicekeys.DeviceKey{
		KeyID:     "test-id",
		PublicKey: []byte{1, 2, 3},
		KeyType:   "Ed25519",
		CreatedAt: time.Now().UTC(),
		Version:   1,
	}
	return &devicekeys.DeviceKeyConfig{
		CurrentKey: &devicekeys.DeviceKeyBundle{
			EncryptionKey: *encryptionKey,
			SigningKey:    *signingKey,
		},
		Archived: []devicekeys.DeviceKeyBundle{},
	}
}

func TestSaveDeviceKeys(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "device_keys_test_*.json")
	assert.NoError(t, err, "failed to create temp file")
	defer os.Remove(tmpFile.Name())

	provider := &mockBuildConfigProvider{filePath: tmpFile.Name()}
	svc := devicekeys.NewDeviceKeyFileService(provider)

	config := testDeviceKeyConfig()

	err = svc.SaveDeviceKeys(config)
	assert.NoError(t, err, "SaveDeviceKeys failed")

	// Read back and verify
	data, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err, "failed to read file")
	var loaded devicekeys.DeviceKeyConfig
	assert.NoError(t, json.Unmarshal(data, &loaded), "failed to unmarshal")
	assert.NotNil(t, loaded.CurrentKey, "CurrentKey not saved correctly")
	assert.NotNil(t, loaded.CurrentKey.EncryptionKey.CreatedAt, "CurrentKey EncryptionKey CreatedAt not saved correctly")
	assert.Equal(t, []byte{1, 2, 3}, loaded.CurrentKey.EncryptionKey.PublicKey, "CurrentKey EncryptionKey PublicKey not saved correctly")
	assert.Equal(t, "test-id", loaded.CurrentKey.EncryptionKey.KeyID, "CurrentKey EncryptionKey KeyID mismatch")
	assert.NotNil(t, loaded.CurrentKey.SigningKey.CreatedAt, "CurrentKey SigningKey CreatedAt not saved correctly")
	assert.Equal(t, []byte{1, 2, 3}, loaded.CurrentKey.SigningKey.PublicKey, "CurrentKey SigningKey PublicKey not saved correctly")
	assert.Equal(t, "test-id", loaded.CurrentKey.SigningKey.KeyID, "CurrentKey SigningKey KeyID mismatch")
	assert.Equal(t, 0, len(loaded.Archived), "Archived keys not saved correctly")
}

func TestLoadDeviceKeys(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "device_keys_test_load_*.json")
	assert.NoError(t, err, "failed to create temp file")
	defer os.Remove(tmpFile.Name())

	config := testDeviceKeyConfig()
	// Write config to file
	data, err := json.Marshal(config)
	assert.NoError(t, err, "failed to marshal config")
	err = os.WriteFile(tmpFile.Name(), data, 0600)
	assert.NoError(t, err, "failed to write file")

	provider := &mockBuildConfigProvider{filePath: tmpFile.Name()}
	svc := devicekeys.NewDeviceKeyFileService(provider)

	loaded, err := svc.LoadDeviceKeys()
	assert.NoError(t, err, "LoadDeviceKeys failed")
	assert.NotNil(t, loaded.CurrentKey, "CurrentKey not loaded")
	assert.NotNil(t, loaded.CurrentKey.EncryptionKey.CreatedAt, "CurrentKey EncryptionKey CreatedAt not loaded")
	assert.Equal(t, []byte{1, 2, 3}, loaded.CurrentKey.EncryptionKey.PublicKey, "CurrentKey EncryptionKey PublicKey not loaded correctly")
	assert.Equal(t, "test-id", loaded.CurrentKey.EncryptionKey.KeyID, "CurrentKey EncryptionKey KeyID mismatch")
	assert.NotNil(t, loaded.CurrentKey.SigningKey.CreatedAt, "CurrentKey SigningKey CreatedAt not loaded")
	assert.Equal(t, []byte{1, 2, 3}, loaded.CurrentKey.SigningKey.PublicKey, "CurrentKey SigningKey PublicKey not loaded correctly")
	assert.Equal(t, "test-id", loaded.CurrentKey.SigningKey.KeyID, "CurrentKey SigningKey KeyID mismatch")
	assert.Equal(t, 0, len(loaded.Archived), "Archived keys not loaded correctly")
}

func TestLoadDeviceKeysFileDoesNotExist(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "device_keys_test_missing_*.json")
	assert.NoError(t, err, "failed to create temp file")
	filePath := tmpFile.Name()
	tmpFile.Close()
	os.Remove(filePath) // Ensure file does not exist

	provider := &mockBuildConfigProvider{filePath: filePath}
	svc := devicekeys.NewDeviceKeyFileService(provider)

	config, err := svc.LoadDeviceKeys()
	assert.NoError(t, err, "should not error when file does not exist; should create default config")
	assert.NotNil(t, config, "config should not be nil when file does not exist")
	assert.Nil(t, config.CurrentKey, "CurrentKey should be nil in default config")
	assert.Equal(t, 0, len(config.Archived), "Archived should be empty in default config")
	// Confirm file was created
	_, statErr := os.Stat(filePath)
	assert.NoError(t, statErr, "file should be created if it did not exist")
}
