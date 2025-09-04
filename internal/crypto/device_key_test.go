package crypto_test

import (
	"go-password-manager/internal/config/devicekeys"
	"go-password-manager/internal/crypto"
	"strings"

	"go-password-manager/internal/testHelpers/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testDeviceKeyConfig = &devicekeys.DeviceKeyConfig{
	CurrentKey: &devicekeys.DeviceKeyBundle{
		SigningKey: devicekeys.DeviceKey{
			KeyID:     "test-signing-id",
			PublicKey: []byte{1, 2, 3},
			KeyType:   devicekeys.KeyTypeEd25519,
			CreatedAt: time.Now().UTC(),
			Version:   1,
		},
		EncryptionKey: devicekeys.DeviceKey{
			KeyID:     "test-encryption-id",
			PublicKey: []byte{1, 2, 3},
			KeyType:   devicekeys.KeyTypeEd25519,
			CreatedAt: time.Now().UTC(),
			Version:   1,
		},
	},
	Archived: []devicekeys.DeviceKeyBundle{},
}

func TestGenerateEncryptionDeviceKey(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	expectedPEM := []byte("MOCK PEM")
	mockPEMProvider.On("EncodeKeyToPEM", mock.Anything, devicekeys.KeyTypeX25519Private).Return(expectedPEM, nil)

	// Set up mock keyring provider
	mockKeyringProvider.On("Set", "GoPasswordManager-encryption", mock.Anything, string(expectedPEM)).Return(nil)

	expectedPub := []byte("mock-public-key")
	expectedPriv := []byte("mock-private-key")
	mockKeyPairGenerator.On("GenerateX25519KeyPair").Return(expectedPub, expectedPriv, nil)

	mockKeyringProvider.On("Set", "GoPasswordManager-encryption", mock.Anything, mock.Anything).Return(nil)
	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)
	deviceKey, err := deviceKeyManager.GenerateEncryptionDeviceKey()

	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, deviceKey, "deviceKey should not be nil")
	assert.Equal(t, expectedPub, deviceKey.PublicKey, "PublicKey mismatch")
	assert.Equal(t, expectedPriv, deviceKey.PrivateKey, "PrivateKey mismatch")
	assert.NotNil(t, deviceKey.DeviceName, "DeviceName mismatch")
	assert.NotNil(t, deviceKey.UserID, "UserID mismatch")
	assert.Equal(t, devicekeys.KeyTypeX25519Private, deviceKey.KeyType, "KeyType mismatch")
	mockKeyringProvider.AssertCalled(t, "Set", "GoPasswordManager-encryption", mock.Anything, "MOCK PEM")
}

func TestGenerateSigningDeviceKey(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	expectedPEM := []byte("MOCK PEM")
	mockPEMProvider.On("EncodeKeyToPEM", mock.Anything, devicekeys.KeyTypeX25519Private).Return(expectedPEM, nil)

	// Set up mock keyring provider
	mockKeyringProvider.On("Set", "GoPasswordManager-signing", mock.Anything, string(expectedPEM)).Return(nil)

	expectedPub := []byte("mock-public-key")
	expectedPriv := []byte("mock-private-key")
	mockKeyPairGenerator.On("GenerateEd25519KeyPair").Return(expectedPub, expectedPriv, nil)
	mockPEMProvider.On("EncodeKeyToPEM", mock.Anything, devicekeys.KeyTypeEd25519Private).Return(expectedPEM, nil)
	mockKeyringProvider.On("Set", "GoPasswordManager-signing", mock.Anything, mock.Anything).Return(nil)
	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)
	deviceKey, err := deviceKeyManager.GenerateSigningDeviceKey()

	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, deviceKey, "deviceKey should not be nil")
	assert.Equal(t, expectedPub, deviceKey.PublicKey, "PublicKey mismatch")
	assert.Equal(t, expectedPriv, deviceKey.PrivateKey, "PrivateKey mismatch")
	assert.NotNil(t, deviceKey.DeviceName, "DeviceName mismatch")
	assert.NotNil(t, deviceKey.UserID, "UserID mismatch")
	assert.Equal(t, devicekeys.KeyTypeEd25519Private, deviceKey.KeyType, "KeyType mismatch")
	mockKeyringProvider.AssertCalled(t, "Set", "GoPasswordManager-signing", mock.Anything, "MOCK PEM")
}

func TestGetDeviceKeyReturnsWhenExists(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	// Simulate key exists in keyring
	existingPEM := []byte("EXISTING PEM")
	mockKeyringProvider.On("Get", "GoPasswordManager-encryption", mock.Anything).Return(string(existingPEM), nil)
	// DeviceKeyManager expects DecodeX25519PrivateKeyFromPEM to return the private key
	existingPriv := []byte("existing-private-key")
	mockPEMProvider.On("DecodeKeyFromPEM", existingPEM, devicekeys.KeyTypeX25519Private).Return(existingPriv, nil)

	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)
	deviceKey, err := deviceKeyManager.GetEncryptionDeviceKey()

	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, deviceKey, "deviceKey should not be nil")
	assert.Equal(t, existingPriv, deviceKey.PrivateKey, "PrivateKey mismatch")
	mockKeyringProvider.AssertCalled(t, "Get", "GoPasswordManager-encryption", mock.Anything)
	mockPEMProvider.AssertCalled(t, "DecodeKeyFromPEM", existingPEM, devicekeys.KeyTypeX25519Private)
}

func TestGetDeviceKeyCreatesNewWhenNotExists(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	// Simulate key not found in keyring
	mockKeyringProvider.On("Get", "GoPasswordManager-signing", mock.Anything).Return("", assert.AnError)

	newPub := []byte("new-public-key")
	newPriv := []byte("new-private-key")
	mockKeyPairGenerator.On("GenerateEd25519KeyPair").Return(newPub, newPriv, nil)

	newPEM := []byte("NEW PEM")
	mockPEMProvider.On("EncodeKeyToPEM", newPriv, devicekeys.KeyTypeEd25519Private).Return(newPEM, nil)
	mockKeyringProvider.On("Set", "GoPasswordManager-signing", mock.Anything, string(newPEM)).Return(nil)

	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)
	mockDeviceKeyProvider.On("SaveDeviceKeys", mock.Anything).Return(nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)
	deviceKey, err := deviceKeyManager.GetSigningDeviceKey()

	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, deviceKey, "deviceKey should not be nil")
	assert.Equal(t, newPriv, deviceKey.PrivateKey, "PrivateKey mismatch")
	assert.Equal(t, newPub, deviceKey.PublicKey, "PublicKey mismatch")
	mockKeyringProvider.AssertCalled(t, "Get", "GoPasswordManager-signing", mock.Anything)
	mockKeyPairGenerator.AssertCalled(t, "GenerateEd25519KeyPair")
	mockPEMProvider.AssertCalled(t, "EncodeKeyToPEM", newPriv, devicekeys.KeyTypeEd25519Private)
	mockKeyringProvider.AssertCalled(t, "Set", "GoPasswordManager-signing", mock.Anything, string(newPEM))
}

func TestDeleteEncryptionDeviceKey(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	mockKeyringProvider.On("Delete", "GoPasswordManager-encryption", mock.Anything).Return(nil)
	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)

	err = deviceKeyManager.DeleteEncryptionDeviceKey()
	assert.NoError(t, err, "unexpected error")
	mockKeyringProvider.AssertCalled(t, "Delete", "GoPasswordManager-encryption", mock.Anything)
}

func TestDeleteSigningDeviceKey(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	mockKeyringProvider.On("Delete", "GoPasswordManager-signing", mock.Anything).Return(nil)
	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)

	err = deviceKeyManager.DeleteSigningDeviceKey()
	assert.NoError(t, err, "unexpected error")
	mockKeyringProvider.AssertCalled(t, "Delete", "GoPasswordManager-signing", mock.Anything)
}

func TestRotateEncryptionDeviceKey(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	// Simulate existing key in keyring
	existingPEM := []byte("EXISTING PEM")
	mockKeyringProvider.On("Get", "GoPasswordManager-encryption", mock.Anything).Return(string(existingPEM), nil)

	// Archive call: expect Set with archived key name
	mockKeyringProvider.On("Set",
		mock.MatchedBy(func(service string) bool {
			return strings.HasPrefix(service, "GoPasswordManager-encryption-archived-")
		}),
		mock.Anything,
		string(existingPEM),
	).Return(nil)

	rotatedPub := []byte("rotated-public-key")
	rotatedPriv := []byte("rotated-private-key")
	mockKeyPairGenerator.On("GenerateX25519KeyPair").Return(rotatedPub, rotatedPriv, nil)

	rotatedPEM := []byte("ROTATED PEM")
	mockPEMProvider.On("EncodeKeyToPEM", rotatedPriv, devicekeys.KeyTypeX25519Private).Return(rotatedPEM, nil)
	mockKeyringProvider.On("Set", "GoPasswordManager-encryption", mock.Anything, string(rotatedPEM)).Return(nil)

	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)
	mockDeviceKeyProvider.On("SaveDeviceKeys", mock.Anything).Return(nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)

	rotatedKey, err := deviceKeyManager.RotateEncryptionDeviceKey("test-device", "test-user", "X25519")
	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, rotatedKey, "deviceKey should not be nil")
	assert.Equal(t, rotatedPriv, rotatedKey.PrivateKey, "PrivateKey mismatch")
	assert.Equal(t, rotatedPub, rotatedKey.PublicKey, "PublicKey mismatch")
	assert.Equal(t, devicekeys.KeyTypeX25519Private, rotatedKey.KeyType, "KeyType mismatch")
	mockKeyPairGenerator.AssertCalled(t, "GenerateX25519KeyPair")
	mockPEMProvider.AssertCalled(t, "EncodeKeyToPEM", rotatedPriv, devicekeys.KeyTypeX25519Private)
	mockKeyringProvider.AssertCalled(t, "Get", "GoPasswordManager-encryption", mock.Anything)
	mockKeyringProvider.AssertCalled(t, "Set", "GoPasswordManager-encryption", mock.Anything, string(rotatedPEM))
}

func TestRotateSigningDeviceKey(t *testing.T) {
	mockKeyPairGenerator := &mocks.CryptoProvider{}
	mockKeyringProvider := &mocks.MockKeyringProvider{}
	mockPEMProvider := &mocks.PEMProvider{}
	mockDeviceKeyProvider := &mocks.DeviceKeyFileProvidor{}

	// Simulate existing key in keyring
	existingPEM := []byte("EXISTING PEM")
	mockKeyringProvider.On("Get", "GoPasswordManager-signing", mock.Anything).Return(string(existingPEM), nil)

	// Archive call: expect Set with archived key name
	mockKeyringProvider.On("Set",
		mock.MatchedBy(func(service string) bool {
			return strings.HasPrefix(service, "GoPasswordManager-signing-archived-")
		}),
		mock.Anything,
		string(existingPEM),
	).Return(nil)

	rotatedPub := []byte("rotated-public-key")
	rotatedPriv := []byte("rotated-private-key")
	mockKeyPairGenerator.On("GenerateEd25519KeyPair").Return(rotatedPub, rotatedPriv, nil)

	rotatedPEM := []byte("ROTATED PEM")
	mockPEMProvider.On("EncodeKeyToPEM", rotatedPriv, devicekeys.KeyTypeEd25519Private).Return(rotatedPEM, nil)
	mockKeyringProvider.On("Set", "GoPasswordManager-signing", mock.Anything, string(rotatedPEM)).Return(nil)

	mockDeviceKeyProvider.On("LoadDeviceKeys").Return(testDeviceKeyConfig, nil)
	mockDeviceKeyProvider.On("SaveDeviceKeys", mock.Anything).Return(nil)

	deviceKeyManager, err := crypto.NewDeviceKeyManager(mockKeyPairGenerator, mockPEMProvider, mockDeviceKeyProvider)
	assert.NoError(t, err)
	deviceKeyManager.SetKeyringProvider(mockKeyringProvider)

	rotatedKey, err := deviceKeyManager.RotateSigningDeviceKey("test-device", "test-user", "X25519")
	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, rotatedKey, "deviceKey should not be nil")
	assert.Equal(t, rotatedPriv, rotatedKey.PrivateKey, "PrivateKey mismatch")
	assert.Equal(t, rotatedPub, rotatedKey.PublicKey, "PublicKey mismatch")
	assert.Equal(t, devicekeys.KeyTypeEd25519Private, rotatedKey.KeyType, "KeyType mismatch")
	mockKeyringProvider.AssertCalled(t, "Get", "GoPasswordManager-signing", mock.Anything)
	mockKeyPairGenerator.AssertCalled(t, "GenerateEd25519KeyPair")
	mockPEMProvider.AssertCalled(t, "EncodeKeyToPEM", rotatedPriv, devicekeys.KeyTypeEd25519Private)
	mockKeyringProvider.AssertCalled(t, "Set", "GoPasswordManager-signing", mock.Anything, string(rotatedPEM))
}
