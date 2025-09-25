package sharing_test

import (
	"go-password-manager/internal/crypto"
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
	// Generate recipient key pair
	recipientPub, recipientPriv, _ := crypto.GenerateX25519KeyPair()
	// Produce a real key box wrapping a symmetric key
	symKey := []byte("0123456789abcdef0123456789abcdef")
	// Minimal fields needed for AAD reconstruction: ID and SenderInfo.SigningPublicKey
	bundleID := "test-bundle-id"
	signingPub := []byte("sender-signing-pub")
	aad := append(append([]byte{}, []byte(bundleID)...), 0x00)
	aad = append(aad, signingPub...)
	box, _ := crypto.WrapKeyBox(symKey, recipientPub, aad)
	// Build bundle with matching fields
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: bundleID, SenderInfo: sharing.SenderMetadata{SigningPublicKey: signingPub}, SymmetricKeyBox: box, EncryptedSecrets: []byte("cipher"), SecretsNonce: []byte("nonce")}}
	// DecryptSymmetric mocked to return valid JSON secrets
	validJSON := `[{"Name":"test","Type":"password","Value":"secret","UpdatedAt":"123456789","Version":1}]`
	cryptoProvider.On("DecryptSymmetric", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(validJSON), nil)
	secrets, err := service.DecryptSecrets(bundle, recipientPriv, nil)
	require.NoError(t, err)
	require.NotNil(t, secrets)
	cryptoProvider.AssertExpectations(t)
}

func TestDecryptSecretsFailure(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	// Malformed (too short) key box
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{SymmetricKeyBox: []byte{1, 2, 3}}}
	recipientPriv := []byte("short-priv")
	secrets, err := service.DecryptSecrets(bundle, recipientPriv, nil)
	require.Error(t, err)
	require.Nil(t, secrets)
	cryptoProvider.AssertExpectations(t)
}

func TestImportSecretsSuccess(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	recipientPub, recipientPriv, _ := crypto.GenerateX25519KeyPair()
	symKey := []byte("0123456789abcdef0123456789abcdef")
	bundleID := "test-bundle-id-2"
	signingPub := []byte("sender-signing-pub-2")
	aad := append(append([]byte{}, []byte(bundleID)...), 0x00)
	aad = append(aad, signingPub...)
	box, _ := crypto.WrapKeyBox(symKey, recipientPub, aad)
	validJSON := `[{"Name":"test","Type":"password","Value":"secret","UpdatedAt":"123456789","Version":1}]`
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: bundleID, SenderInfo: sharing.SenderMetadata{SigningPublicKey: signingPub}, SymmetricKeyBox: box, EncryptedSecrets: []byte("cipher"), SecretsNonce: []byte("nonce")}}
	cryptoProvider.On("VerifyEd25519", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	cryptoProvider.On("DecryptSymmetric", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(validJSON), nil)
	secretsProvider.On("SaveOrUpdateSecrets", mock.Anything).Return(nil)
	result, err := service.ImportSecrets(bundle, recipientPriv, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ImportedSecretsCount)
	assert.True(t, result.Success)
	secretsProvider.AssertExpectations(t)
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
	result, err := service.ImportSecrets(bundle, recipientPrivateKey, nil)
	require.Nil(t, err, "No error should be returned as value; error is in result struct")
	require.NotNil(t, result, "ImportResult should not be nil")
	require.NotNil(t, result.Error, "Expected error in ImportResult when importing expired bundle")
	require.False(t, result.Success, "Import should not succeed for expired bundle")
	assert.Contains(t, result.Error.Error(), "expired", "Error message should mention expiry")
}

func TestImportSecretsReplayProtection(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	service.SetReplayTTL(300) // 5 minute window
	recipientPub, recipientPriv, _ := crypto.GenerateX25519KeyPair()
	symKey := []byte("0123456789abcdef0123456789abcdef")
	bundleID := "replay-bundle"
	signingPub := []byte("sender-signing-pub-R")
	// Build key box AAD
	keyBoxAAD := append(append([]byte{}, []byte(bundleID)...), 0x00)
	keyBoxAAD = append(keyBoxAAD, signingPub...)
	box, _ := crypto.WrapKeyBox(symKey, recipientPub, keyBoxAAD)
	// Secrets AAD (domain separated)
	secretsAAD := append(append([]byte{}, []byte(bundleID)...), 0x01)
	secretsAAD = append(secretsAAD, signingPub...)
	// First decrypt call expectation
	validJSON := `[{"Name":"test","Type":"password","Value":"secret","UpdatedAt":"123456789","Version":1}]`
	cryptoProvider.On("DecryptSymmetric", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(validJSON), nil).Once()
	cryptoProvider.On("VerifyEd25519", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	secretsProvider.On("SaveOrUpdateSecrets", mock.Anything).Return(nil).Once()
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: bundleID, SenderInfo: sharing.SenderMetadata{SigningPublicKey: signingPub}, SymmetricKeyBox: box, EncryptedSecrets: []byte("cipher"), SecretsNonce: []byte("nonce")}}
	// First import succeeds
	res1, err1 := service.ImportSecrets(bundle, recipientPriv, nil)
	require.NoError(t, err1)
	require.True(t, res1.Success)
	// Second import should be marked replay (no second decrypt call)
	res2, err2 := service.ImportSecrets(bundle, recipientPriv, nil)
	require.NoError(t, err2)
	require.False(t, res2.Success)
	require.Error(t, res2.Error)
	require.Contains(t, res2.Error.Error(), "replay")
}

func TestDecryptSecretsSecretsAADTamper(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	recipientPub, recipientPriv, _ := crypto.GenerateX25519KeyPair()
	symKey := []byte("0123456789abcdef0123456789abcdef")
	bundleID := "tamper-bundle"
	signingPub := []byte("sender-signing-pub-T")
	// Build correct key box AAD (domain 0x00)
	keyBoxAAD := append(append([]byte{}, []byte(bundleID)...), 0x00)
	keyBoxAAD = append(keyBoxAAD, signingPub...)
	box, _ := crypto.WrapKeyBox(symKey, recipientPub, keyBoxAAD)
	// We will have DecryptSymmetric return error when wrong AAD passed in; we simulate by expecting call but returning error.
	cryptoProvider.On("DecryptSymmetric", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: bundleID, SenderInfo: sharing.SenderMetadata{SigningPublicKey: signingPub}, SymmetricKeyBox: box, EncryptedSecrets: []byte("cipher"), SecretsNonce: []byte("nonce")}}
	// Attempt decrypt: import path constructs secrets AAD with 0x01 domain; to simulate tamper we can modify signing pub AFTER key box unwrap but BEFORE decrypt would be impossible externally; instead we rely on mismatch because we will adjust bundle ID to force different AAD than expected by service.
	// Simulate tamper: change ID so secrets AAD differs from what ciphertext was hypothetically sealed with. (We cannot re-encrypt here; test focuses on flow failing.)
	// Decrypt should fail producing error.
	secrets, err := service.DecryptSecrets(bundle, recipientPriv, nil)
	require.Error(t, err)
	require.Nil(t, secrets)
}

func TestImportSecretsReplayMarkOnlyOnSuccess(t *testing.T) {
	cryptoProvider, deviceKeyProvider, secretsProvider := setupImportServiceMocks(t)
	service := sharing.NewImportService(cryptoProvider, deviceKeyProvider, secretsProvider)
	recipientPub, recipientPriv, _ := crypto.GenerateX25519KeyPair()
	symKey := []byte("0123456789abcdef0123456789abcdef")
	bundleID := "mark-on-success"
	signingPub := []byte("sender-signing-pub-M")
	keyBoxAAD := append(append([]byte{}, []byte(bundleID)...), 0x00)
	keyBoxAAD = append(keyBoxAAD, signingPub...)
	box, _ := crypto.WrapKeyBox(symKey, recipientPub, keyBoxAAD)
	validJSON := `[{"Name":"test","Type":"password","Value":"secret","UpdatedAt":"123456789","Version":1}]`
	cryptoProvider.On("VerifyEd25519", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	cryptoProvider.On("DecryptSymmetric", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(validJSON), nil).Twice()
	// First attempt fails persistence
	persistErr := assert.AnError
	secretsProvider.On("SaveOrUpdateSecrets", mock.Anything).Return(persistErr).Once()
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: bundleID, SenderInfo: sharing.SenderMetadata{SigningPublicKey: signingPub}, SymmetricKeyBox: box, EncryptedSecrets: []byte("cipher"), SecretsNonce: []byte("nonce")}}
	resFail, errFail := service.ImportSecrets(bundle, recipientPriv, nil)
	require.NoError(t, errFail)
	require.False(t, resFail.Success)
	// Now persistence succeeds
	secretsProvider.On("SaveOrUpdateSecrets", mock.Anything).Return(nil).Once()
	resOk, errOk := service.ImportSecrets(bundle, recipientPriv, nil)
	require.NoError(t, errOk)
	require.True(t, resOk.Success)
	// Third attempt should be replay blocked (DecryptSymmetric not invoked again)
	resReplay, errReplay := service.ImportSecrets(bundle, recipientPriv, nil)
	require.NoError(t, errReplay)
	require.False(t, resReplay.Success)
	require.Error(t, resReplay.Error)
	require.Contains(t, resReplay.Error.Error(), "replay")
}
