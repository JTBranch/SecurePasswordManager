package sharing_test

import (
	"go-password-manager/internal/domain"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/testHelpers/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupImportServiceMocks(t *testing.T) (*mocks.CryptoProvider, *mocks.DeviceKeyProvider, *mocks.SecretsProvider) {
	cryptoProvider := mocks.NewCryptoProvider(t)
	deviceKeyProvider := mocks.NewDeviceKeyProvider(t)
	secretsProvider := mocks.NewSecretsProvider(t)
	return cryptoProvider, deviceKeyProvider, secretsProvider
}

func TestNewImportService(t *testing.T) {
	service := sharing.NewImportService(setupImportServiceMocks(t))
	require.NotNil(t, service, "ImportService should be created")
}

func TestVerifyBundleSignatureValid(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	bundle := &sharing.SecretExportBundle{
		Payload:   sharing.SecretExportPayload{},
		Signature: []byte("valid-signature"),
	}
	// Setup mock expectation for valid signature
	cryptoProvider.On("VerifyEd25519", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	valid, err := service.VerifyBundleSignature(bundle)
	require.NoError(t, err)
	require.True(t, valid, "Signature should be valid")
	cryptoProvider.AssertExpectations(t)
}

func TestVerifyBundleSignatureInvalid(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	bundle := &sharing.SecretExportBundle{
		Payload:   sharing.SecretExportPayload{},
		Signature: []byte("invalid-signature"),
	}
	// Setup mock expectation for invalid signature
	cryptoProvider.On("VerifyEd25519", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

	valid, err := service.VerifyBundleSignature(bundle)
	require.NoError(t, err)
	require.False(t, valid, "Signature should be invalid")
	cryptoProvider.AssertExpectations(t)
}

func TestDecryptSecretsSuccess(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	bundle := &sharing.SecretExportBundle{
		Payload: sharing.SecretExportPayload{},
	}
	recipientPrivateKey := []byte("test-private-key")
	recipientEphemeralPubKey := []byte("test-ephemeral-public-key")

	// Setup mock expectation for decryption
	cryptoProvider.On("DecryptAsymmetric", mock.Anything, mock.Anything, mock.Anything).Return([]byte("decrypted-key"), nil).Maybe()
	cryptoProvider.On("DecryptAsymmetric", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return([]byte("decrypted-key"), nil).Maybe()
	cryptoProvider.On("DecryptAsymmetric", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]byte("decrypted-key"), nil).Maybe()
	cryptoProvider.On("ECDH", mock.Anything, mock.Anything).Return([]byte("shared-secret"), nil)
	cryptoProvider.On("DeriveSymmetricKey", mock.Anything, mock.Anything, mock.Anything).Return([]byte("derived-key"), nil)
	// Prepare a valid JSON array for the expected secrets
	validJSON := `[{"Name":"test","Type":"password","Value":"secret","UpdatedAt":"123456789","Version":1}]`
	cryptoProvider.On("DecryptSymmetric", mock.Anything, mock.Anything, mock.Anything).Return([]byte(validJSON), nil)

	secrets, err := service.DecryptSecrets(bundle, recipientPrivateKey, recipientEphemeralPubKey)
	require.NoError(t, err)
	require.NotNil(t, secrets, "Decrypted secrets should not be nil")
	cryptoProvider.AssertExpectations(t)
}

func TestDecryptSecretsFailure(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	bundle := &sharing.SecretExportBundle{
		Payload: sharing.SecretExportPayload{},
	}
	recipientPrivateKey := []byte("bad-private-key")
	recipientEphemeralPubKey := []byte("test-ephemeral-public-key")
	// Setup mock expectation for decryption failure
	cryptoProvider.On("ECDH", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	cryptoProvider.On("DecryptAsymmetric", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError).Maybe()
	secrets, err := service.DecryptSecrets(bundle, recipientPrivateKey, recipientEphemeralPubKey)
	require.Error(t, err)
	require.Nil(t, secrets, "Decrypted secrets should be nil on failure")
	cryptoProvider.AssertExpectations(t)
}

func TestImportSecretsSuccess(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	bundle := &sharing.SecretExportBundle{
		Payload: sharing.SecretExportPayload{},
	}
	recipientPrivateKey := []byte("test-private-key")
	recipientEphemeralPubKey := []byte("test-ephemeral-public-key")
	// Setup mock expectations for full import flow as needed
	cryptoProvider.On("VerifyEd25519", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	cryptoProvider.On("ECDH", mock.Anything, mock.Anything).Return([]byte("shared-secret"), nil)
	cryptoProvider.On("DeriveSymmetricKey", mock.Anything, mock.Anything, mock.Anything).Return([]byte("derived-key"), nil)
	validJSON := `[{"Name":"test","Type":"password","Value":"secret","UpdatedAt":"123456789","Version":1}]`
	cryptoProvider.On("DecryptSymmetric", mock.Anything, mock.Anything, mock.Anything).Return([]byte(validJSON), nil)
	secretsProvider.On("SaveOrUpdateSecrets", mock.Anything).Return(nil)
	result, err := service.ImportSecrets(bundle, recipientPrivateKey, recipientEphemeralPubKey)
	require.NoError(t, err)
	require.NotNil(t, result, "ImportSecrets should return non-nil result")
	assert.Equal(t, 1, result.ImportedSecretsCount, "should import one secret")
	assert.True(t, result.Success, "ImportSecrets should succeed")

	// Assert SaveOrUpdateSecrets was called with the expected secrets
	secretsProvider.AssertCalled(t, "SaveOrUpdateSecrets", mock.MatchedBy(func(secrets interface{}) bool {
		secretsSlice, ok := secrets.([]domain.Secret)
		if !ok || len(secretsSlice) != 1 {
			return false
		}
		s := secretsSlice[0]
		// Check the main fields of the imported secret
		return s.SecretName == "test" && string(s.Type) == "password" && s.CurrentVersion == 1 &&
			len(s.Versions) == 1 && s.Versions[0].SecretValueEnc == "secret" && s.Versions[0].UpdatedAt == "123456789" && s.Versions[0].Version == 1
	}))

	cryptoProvider.AssertExpectations(t)
}

func TestImportSecretsExpiredBundle(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	// Set ExpiresAt to a time in the past
	bundle := &sharing.SecretExportBundle{
		Payload: sharing.SecretExportPayload{
			ExpiresAt: 1, // Unix time in the past
		},
	}
	recipientPrivateKey := []byte("test-private-key")
	recipientEphemeralPubKey := []byte("test-ephemeral-public-key")

	result, err := service.ImportSecrets(bundle, recipientPrivateKey, recipientEphemeralPubKey)
	require.Nil(t, err, "No error should be returned as value; error is in result struct")
	require.NotNil(t, result, "ImportResult should not be nil")
	require.NotNil(t, result.Error, "Expected error in ImportResult when importing expired bundle")
	require.False(t, result.Success, "Import should not succeed for expired bundle")
	assert.Contains(t, result.Error.Error(), "expired", "Error message should mention expiry")
}
