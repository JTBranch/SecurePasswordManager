package sharing_test

import (
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/sharing"
	testhelpers "go-password-manager/internal/testHelpers/mocks"
	"strings"
	"testing"

	"crypto/ed25519"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestSecretExportBundle() *sharing.SecretExportBundle {
	return &sharing.SecretExportBundle{
		Payload: sharing.SecretExportPayload{
			EncryptedSecrets:      []byte("mock-ciphertext"),
			ID:                    "test-id",
			Name:                  "test-bundle",
			EncryptedSymmetricKey: []byte("mock-asymmetric-ciphertextmock-ciphertext"),
			SecretsNonce:          []byte("mock-secrets-nonce"),
			KeyNonce:              []byte("mock-key-nonce"),
			Timestamp:             1234567890,
			ExpiresAt:             1234567890 + 60*60,
			SenderInfo:            testSenderMetadata,
		},
		Signature: []byte("mock-signature"),
		// Add other expected fields as needed
	}
}

var testSenderMetadata = sharing.SenderMetadata{
	// Fill with expected fields for your tests
	DeviceName: "Test-device",
	UserID:     "test@sender.com",
	PublicKey:  []byte("test-public-key"),
	// Add other fields as needed
}

func TestExportSecretsCreatesValidBundle(t *testing.T) {
	// Arrange
	mockCrypto := &testhelpers.CryptoProvider{}
	mockDeviceKeyProvider := &testhelpers.DeviceKeyProvider{}
	mockCrypto.On("ECDH", mock.Anything, mock.Anything).Return([]byte("mock-shared-secret"), nil)
	mockCrypto.On("EncryptSymmetric", mock.Anything, mock.Anything).Return(
		[]byte("mock-ciphertext"),    // ciphertext
		[]byte("mock-secrets-nonce"), // nonce (matches expected)
		nil,
	)
	expectedSender := sharing.SenderMetadata{
		DeviceName:       "Test-device",
		UserID:           "test@sender.com",
		PublicKey:        []byte("test-public-key"),
		SigningPublicKey: []byte("test-public-key"), // or the actual test key used
	}
	mockCrypto.On("SignEd25519", mock.Anything, mock.Anything).Return([]byte("mock-signature"), nil)
	mockCrypto.On("EncryptAsymmetricFull", mock.Anything, mock.Anything).Return(crypto.AsymmetricEncryptResult{
		EphemeralPublicKey: []byte("mock-asymmetric-ciphertext"), // example
		Ciphertext:         []byte("mock-ciphertext"),
		Nonce:              []byte("mock-nonce"),
	}, nil)
	mockDeviceKeyProvider.On("GetEncryptionDeviceKey").Return(&crypto.DeviceKey{
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("mock-private-key"),
		// ...other fields as needed
	}, nil)
	mockDeviceKeyProvider.On("GetSigningDeviceKey").Return(&crypto.DeviceKey{
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("mock-private-key"),
		// ...other fields as needed
	}, nil)
	mockDeviceKeyProvider.On("GetDeviceName").Return("Test-device")
	mockDeviceKeyProvider.On("GetAppUser").Return("test@sender.com")
	mockCrypto.On("DeriveSymmetricKey", mock.Anything, mock.Anything, mock.Anything).Return([]byte("0123456789abcdef0123456789abcdef"), nil)

	exportService := sharing.NewExportService(mockCrypto, mockDeviceKeyProvider)
	secrets := []sharing.ExportSecret{
		{Name: "test-secret", Value: "test-value"},
	}
	recipientPubKey := []byte("recipient-public-key")
	expiry := 60
	senderInfo := testSenderMetadata

	mockDeviceKeyProvider.On("GetDeviceKey").Return(&crypto.DeviceKey{
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("mock-private-key"),
		// ...other fields as needed
	}, nil)
	mockCrypto.On("DeriveSymmetricKey", mock.Anything, mock.Anything, mock.Anything).Return([]byte("0123456789abcdef0123456789abcdef"), nil)

	expected := newTestSecretExportBundle()

	// Act
	result, err := exportService.ExportSecrets(secrets, recipientPubKey, expiry, senderInfo)

	// Assert
	assert.NoError(t, err, "ExportSecrets should not return error")
	assert.NotNil(t, result, "ExportSecrets should not return nil bundle")
	assert.Equal(t, string(expected.Payload.EncryptedSecrets), string(result.Payload.EncryptedSecrets), "EncryptedSecrets mismatch")
	assert.Equal(t, string(expected.Payload.EncryptedSymmetricKey), string(result.Payload.EncryptedSymmetricKey), "EncryptedSymmetricKey mismatch")
	assert.Equal(t, expected.Payload.SecretsNonce, result.Payload.SecretsNonce, "SecretsNonce mismatch")
	assert.NotNil(t, result.Payload.Timestamp, "Timestamp mismatch")
	assert.NotNil(t, result.Payload.ExpiresAt, "ExpiresAt mismatch")
	assert.Equal(t, string(expected.Signature), string(result.Signature), "Signature mismatch")
	assert.NotNil(t, result.Payload.ID)
	assert.True(t, strings.HasSuffix(result.Payload.Name, ".pem"), "Bundle name should end with .pem")
	assert.NotEmpty(t, result.Payload.ID, "ID should not be empty")
	assert.Equal(t, expectedSender, result.Payload.SenderInfo, "SenderInfo mismatch")
}

func TestExportSecretsUsesExpiryCorrectly(t *testing.T) {
	// TODO: Implement test for expiry logic in bundle
}

func TestExportSecretsHandlesEmptySecrets(t *testing.T) {
	// TODO: Test that ExportSecrets handles empty secrets gracefully
}

func TestExportSecretsInvalidRecipientKey(t *testing.T) {
	// TODO: Test that ExportSecrets returns error for invalid recipient public key
}

func TestExportSecretsExpiryIsSetCorrectly(t *testing.T) {
	// TODO: Test that expiry is set and enforced in the bundle
}

func TestExportSecretsSignatureIsGenerated(t *testing.T) {
	// TODO: Test that the bundle is signed correctly
}

func TestExportSecretsErrorOnEncryptionFailure(t *testing.T) {
	// TODO: Test error handling if encryption fails
}

func TestSignBundleReturnsExpectedSignature(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(nil)
	bundlePayload := &sharing.SecretExportPayload{
		ID:                    "test-id",
		Name:                  "test-bundle",
		EncryptedSecrets:      []byte("mock-encrypt-symmetric"),
		EncryptedSymmetricKey: []byte("mock-encrypt-symmetric-key"),
		SecretsNonce:          []byte("mock-secrets-nonce"),
		KeyNonce:              []byte("mock-key-nonce"),
		Timestamp:             1234567890,
		ExpiresAt:             1234567890 + 60*60,
		SenderInfo:            testSenderMetadata,
	}
	mockCrypto := &testhelpers.CryptoProvider{}
	mockDeviceKeyProvider := &testhelpers.DeviceKeyProvider{}
	mockCrypto.On("SignEd25519", mock.Anything, mock.Anything).Return(make([]byte, 64), nil)
	exportService := sharing.NewExportService(mockCrypto, mockDeviceKeyProvider)

	signature, err := exportService.SignBundle(bundlePayload, priv)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, ed25519.SignatureSize, len(signature))
}
