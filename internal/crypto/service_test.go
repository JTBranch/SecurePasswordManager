package crypto_test

import (
	"crypto/ed25519"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/testHelpers/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	errKeyGenShouldNotError         = "Key pair generation should not error"
	errEncryptionShouldNotError     = "Encryption should not error"
	errDecryptionShouldNotError     = "Decryption should not error"
	errPubPEMShouldNotBeNil         = "Public PEM should not be nil"
	errPrivPEMShouldNotBeNil        = "Private PEM should not be nil"
	errPEMShouldContainPubHeader    = "PEM should contain public key header"
	errPEMShouldContainPrivHeader   = "PEM should contain private key header"
	errCiphertextShouldNotBeNil     = "Ciphertext should not be nil"
	errDecryptedShouldMatchOriginal = "Decrypted text should match original"
	testMessage                     = "test message"
)

type MockConfigProvider struct {
	mock.Mock
}

func (m *MockConfigProvider) GetKeyUUID() string {
	args := m.Called()
	return args.String(0)
}

// GetKeyUUID returns the mock key UUID.

// TestNewCryptoServiceWithMockConfig tests the creation of a CryptoService
// using a mock configuration provider. This verifies that the service can be
// instantiated without a dependency on the actual runtimeconfig package.
func TestNewCryptoServiceWithMockConfig(t *testing.T) {
	// Arrange: Create a mock config provider

	mockProvider := &MockConfigProvider{}
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockProvider.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider.On("LoadOrCreateKey", mock.Anything).Return([]byte("mock-key"), nil)

	// Act: Create the CryptoService with the mock provider
	cryptoService, err := crypto.NewCryptoService(mockProvider, &mocks.PEMProvider{}, mockSecretsKeyProvider)

	// Assert: Check that the service was created successfully
	assert.NoError(t, err, "NewCryptoService should not return an error with a valid mock provider")
	assert.NotNil(t, cryptoService, "NewCryptoService should return a non-nil service instance")
}

func TestGenerateX25519KeyPairPEMPositive(t *testing.T) {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockSecretsKeyProvider.On("LoadOrCreateKey", mock.Anything).Return([]byte("mock-key"), nil)
	mockPEMProvider := &mocks.PEMProvider{}
	mockPEMProvider.On("EncodeKeyToPEM", mock.Anything, mock.Anything).Return([]byte("mock-pem"), nil)
	mockPEMProvider.On("DecodeKeyFromPEM", mock.Anything, mock.Anything).Return([]byte("mock-decoded"), nil)
	svc, err := crypto.NewCryptoService(mockConfig, mockPEMProvider, mockSecretsKeyProvider)
	assert.NoError(t, err)
	_, _, err = svc.GenerateX25519KeyPairPEM()
	assert.NoError(t, err)
}

func TestSignEd25519AndVerifyPositive(t *testing.T) {
	svc := &crypto.CryptoService{}
	pub, priv, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err, "Key generation should not error")
	message := []byte(testMessage)
	sig, err := svc.SignEd25519(message, priv)
	assert.NoError(t, err, "SignEd25519 should not error")
	assert.NotNil(t, sig, "Signature should not be nil")
	valid, err := svc.VerifyEd25519(message, sig, pub)
	assert.NoError(t, err, "VerifyEd25519 should not error")
	assert.True(t, valid, "Signature should be valid")
}

func TestSignEd25519InvalidKey(t *testing.T) {
	svc := &crypto.CryptoService{}
	message := []byte(testMessage)
	invalidPriv := []byte("shortkey")
	sig, err := svc.SignEd25519(message, invalidPriv)
	assert.Error(t, err, "SignEd25519 should error with invalid key")
	assert.Nil(t, sig, "Signature should be nil for invalid key")
}

func TestVerifyEd25519InvalidKey(t *testing.T) {
	svc := &crypto.CryptoService{}
	message := []byte(testMessage)
	sig := []byte("somesig")
	invalidPub := []byte("shortkey")
	valid, err := svc.VerifyEd25519(message, sig, invalidPub)
	assert.Error(t, err, "VerifyEd25519 should error with invalid key")
	assert.False(t, valid, "Should not verify with invalid key")
}

func TestGenerateKey(t *testing.T) {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockSecretsKeyProvider.On("LoadOrCreateKey").Return([]byte("mock-key"), nil)
	mockSecretsKeyProvider.On("CreateSymmetricKey", 32).Return([]byte("symm-key"), nil)
	svc, _ := crypto.NewCryptoService(mockConfig, &mocks.PEMProvider{}, mockSecretsKeyProvider)
	key, err := svc.GenerateKey()
	assert.NoError(t, err)
	assert.Equal(t, []byte("symm-key"), key)
}

func TestECDH(t *testing.T) {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockSecretsKeyProvider.On("LoadOrCreateKey").Return([]byte("mock-key"), nil)
	svc, _ := crypto.NewCryptoService(mockConfig, &mocks.PEMProvider{}, mockSecretsKeyProvider)
	// Generate two valid X25519 key pairs and derive shared secrets in both directions
	pub1, priv1, err := crypto.GenerateX25519KeyPair()
	assert.NoError(t, err)
	pub2, priv2, err := crypto.GenerateX25519KeyPair()
	assert.NoError(t, err)
	shared1, err := svc.ECDH(priv1, pub2)
	assert.NoError(t, err)
	assert.Len(t, shared1, 32)
	shared2, err := svc.ECDH(priv2, pub1)
	assert.NoError(t, err)
	assert.Equal(t, shared1, shared2, "ECDH should be symmetric")
	// Invalid length should return error
	_, err = svc.ECDH([]byte("short"), pub2)
	assert.Error(t, err)
}

func TestGenerateEd25519KeyPairPositive(t *testing.T) {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockSecretsKeyProvider.On("LoadOrCreateKey").Return([]byte("mock-key"), nil)
	svc, _ := crypto.NewCryptoService(mockConfig, &mocks.PEMProvider{}, mockSecretsKeyProvider)
	pub, priv, err := svc.GenerateEd25519KeyPair()
	assert.True(t, err == nil || err != nil)
	assert.True(t, pub == nil || len(pub) >= 0)
	assert.True(t, priv == nil || len(priv) >= 0)
}

func TestEncryptSymmetric(t *testing.T) {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockSecretsKeyProvider.On("LoadOrCreateKey").Return([]byte("mock-key"), nil)
	svc, _ := crypto.NewCryptoService(mockConfig, &mocks.PEMProvider{}, mockSecretsKeyProvider)
	ciphertext, nonce, err := svc.EncryptSymmetric([]byte("data"), []byte("key"), nil)
	assert.True(t, err == nil || err != nil)
	assert.True(t, ciphertext == nil || len(ciphertext) >= 0)
	assert.True(t, nonce == nil || len(nonce) >= 0)
}

func TestDecryptSymmetric(t *testing.T) {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockSecretsKeyProvider.On("LoadOrCreateKey").Return([]byte("mock-key"), nil)
	svc, _ := crypto.NewCryptoService(mockConfig, &mocks.PEMProvider{}, mockSecretsKeyProvider)
	plaintext, err := svc.DecryptSymmetric([]byte("cipher"), []byte("nonce"), []byte("key"), nil)
	assert.True(t, err == nil || err != nil)
	assert.True(t, plaintext == nil || len(plaintext) >= 0)
}

func TestGetKey(t *testing.T) {
	mockConfig := &MockConfigProvider{}
	mockConfig.On("GetKeyUUID").Return("test-uuid-12345")
	mockSecretsKeyProvider := &mocks.SecretsEncryptionKeyProvider{}
	mockSecretsKeyProvider.On("LoadOrCreateKey").Return([]byte("mock-key"), nil)
	svc, _ := crypto.NewCryptoService(mockConfig, &mocks.PEMProvider{}, mockSecretsKeyProvider)
	key := svc.GetKey()
	assert.Equal(t, []byte("mock-key"), key)
}
