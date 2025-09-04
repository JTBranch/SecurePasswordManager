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
	mockProvider.On("GetKeyUUID").Return("test-uuid-12345")

	// Act: Create the CryptoService with the mock provider
	cryptoService, err := crypto.NewCryptoService(mockProvider, &mocks.PEMProvider{})

	// Assert: Check that the service was created successfully
	assert.NoError(t, err, "NewCryptoService should not return an error with a valid mock provider")
	assert.NotNil(t, cryptoService, "NewCryptoService should return a non-nil service instance")
}

func TestGenerateX25519KeyPairPEMPositive(t *testing.T) {
	svc := &crypto.CryptoService{
		PemProvider: &crypto.PemUtils{},
	}
	pubPEM, privPEM, err := svc.GenerateX25519KeyPairPEM()
	assert.NoError(t, err, errKeyGenShouldNotError)
	assert.NotNil(t, pubPEM, errPubPEMShouldNotBeNil)
	assert.NotNil(t, privPEM, errPrivPEMShouldNotBeNil)
}

func TestEncryptAsymmetricPEMPositive(t *testing.T) {
	svc := &crypto.CryptoService{
		PemProvider: &crypto.PemUtils{},
	}
	pubPEM, _, err := svc.GenerateX25519KeyPairPEM()
	assert.NoError(t, err, errKeyGenShouldNotError)
	symmetricKey := []byte("Some Symetric Key")
	ciphertext, err := svc.EncryptAsymmetric(symmetricKey, pubPEM)
	assert.NoError(t, err, errEncryptionShouldNotError)
	assert.NotNil(t, ciphertext, errCiphertextShouldNotBeNil)
}

func TestEncryptDecryptAsymmetricPEMPositive(t *testing.T) {
	svc := &crypto.CryptoService{
		PemProvider: &crypto.PemUtils{},
	}
	pubPEM, privPEM, err := svc.GenerateX25519KeyPairPEM()
	assert.NoError(t, err, errKeyGenShouldNotError)
	symmetricKey := []byte("Some Symmetric Key")
	ciphertext, err := svc.EncryptAsymmetricFull(symmetricKey, pubPEM)
	assert.NoError(t, err, errEncryptionShouldNotError)
	assert.NotNil(t, ciphertext, errCiphertextShouldNotBeNil)
	fullCipher := append(ciphertext.EphemeralPublicKey, ciphertext.Ciphertext...)
	decrypted, err := svc.DecryptAsymmetric(fullCipher, ciphertext.Nonce, privPEM)
	assert.NoError(t, err, errDecryptionShouldNotError)
	assert.Equal(t, symmetricKey, decrypted, errDecryptedShouldMatchOriginal)
}

func TestEncryptAsymmetricFullPEMPositive(t *testing.T) {
	svc := &crypto.CryptoService{
		PemProvider: &crypto.PemUtils{},
	}
	pubPEM, privPEM, err := svc.GenerateX25519KeyPairPEM()
	assert.NoError(t, err, errKeyGenShouldNotError)
	symmetricKey := []byte("Some Symmetric Key")
	result, err := svc.EncryptAsymmetricFull(symmetricKey, pubPEM)
	assert.NoError(t, err, errEncryptionShouldNotError)
	assert.NotNil(t, result.Ciphertext, errCiphertextShouldNotBeNil)
	assert.NotNil(t, result.Nonce, "Nonce should not be nil")
	assert.NotNil(t, result.EphemeralPublicKey, "Ephemeral public key should not be nil")
	assert.Equal(t, 32, len(result.EphemeralPublicKey), "Ephemeral public key should be 32 bytes")
	// Compose full ciphertext for decryption
	fullCipher := append(result.EphemeralPublicKey, result.Ciphertext...)
	decrypted, err := svc.DecryptAsymmetric(fullCipher, result.Nonce, privPEM)
	assert.NoError(t, err, errDecryptionShouldNotError)
	assert.Equal(t, symmetricKey, decrypted, errDecryptedShouldMatchOriginal)
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
