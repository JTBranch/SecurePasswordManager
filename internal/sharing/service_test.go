package sharing_test

import (
	"fmt"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/testHelpers/mocks"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockSharingProvider struct {
	EncryptCalled bool
	LastSecrets   []sharing.ExportSecret
	LastPubKey    []byte
	LastExpiry    int
	LastSender    sharing.SenderMetadata
	ReturnError   error
}

func (m *MockSharingProvider) ExportSecrets(secrets []sharing.ExportSecret, recipientPubKey []byte, expiryMinutes int, senderInfo sharing.SenderMetadata) (*sharing.SecretExportBundle, error) {
	m.EncryptCalled = true
	m.LastSecrets = secrets
	m.LastPubKey = recipientPubKey
	m.LastExpiry = expiryMinutes
	m.LastSender = senderInfo
	if m.ReturnError != nil {
		return nil, m.ReturnError
	}
	return &sharing.SecretExportBundle{
		Payload: sharing.SecretExportPayload{
			EncryptedSecrets: []byte("mock-ciphertext"),
		},
	}, nil
}

// MockSecretsService implements only GetSecretValue for testing
type MockSecretsService struct {
	Values map[string]string
}

func (m *MockSecretsService) GetSecretValue(secret *domain.Secret) (string, error) {
	if val, ok := m.Values[secret.SecretName]; ok {
		return val, nil
	}
	return "", fmt.Errorf("not found")
}

type MockConfigProvider struct{}

type MockLoggerProvider struct {
	Infos  []string
	Errors []string
}

func (m *MockLoggerProvider) Info(message string) {
	m.Infos = append(m.Infos, message)
}
func (m *MockLoggerProvider) Error(message string) {
	m.Errors = append(m.Errors, message)
}

func setupExportSecretsTest() (*sharing.SharingService, *MockSharingProvider, *mocks.SecretsProvider, domain.Secret) {
	const (
		secretValueEnc = "enc-value"
		updatedAt      = "2025-08-27"
	)
	provider := &MockSharingProvider{}
	importProvider := &mocks.ImportProvider{}
	mockSecrets := &mocks.SecretsProvider{}
	service := sharing.NewSharingService(provider, importProvider, mockSecrets)
	validSecret := domain.Secret{
		SecretName:     "test",
		Type:           domain.SecretTypeKeyValue,
		CurrentVersion: 1,
		Versions: []domain.SecretVersion{{
			SecretValueEnc: secretValueEnc,
			Version:        1,
			UpdatedAt:      updatedAt,
		}},
	}
	return service, provider, mockSecrets, validSecret
}

func TestExportSecretsEncryptsAndBundlesCorrectly(t *testing.T) {
	service, provider, mockSecrets, validSecret := setupExportSecretsTest()
	secrets := []domain.Secret{validSecret}
	recipientPubKey := []byte("recipient-key")
	expiry := 60

	mockSecrets.On("GetSecretValue", &validSecret).Return("decrypted-value", nil)

	bundle, err := service.ExportSecrets(secrets, recipientPubKey, expiry)
	require.NoError(t, err)
	require.NotNil(t, bundle)
	require.True(t, provider.EncryptCalled)
	require.Equal(t, "mock-ciphertext", string(bundle.Payload.EncryptedSecrets))
	mockSecrets.AssertExpectations(t)
}

func TestExportSecretsFailsWithEmptySecrets(t *testing.T) {
	service, _, _, _ := setupExportSecretsTest()
	_, err := service.ExportSecrets([]domain.Secret{}, []byte("recipient-key"), 60)
	if err == nil || err.Error() != "no secrets to export" {
		t.Errorf("Expected error for empty secrets, got: %v", err)
	}
}

func TestExportSecretsFailsWithEmptyRecipientKey(t *testing.T) {
	service, _, _, validSecret := setupExportSecretsTest()
	_, err := service.ExportSecrets([]domain.Secret{validSecret}, []byte{}, 60)
	if err == nil || err.Error() != "recipient public key is missing" {
		t.Errorf("Expected error for empty recipient key, got: %v", err)
	}
}

func TestExportSecretsFailsWithInvalidExpiry(t *testing.T) {
	service, _, _, validSecret := setupExportSecretsTest()
	_, err := service.ExportSecrets([]domain.Secret{validSecret}, []byte("recipient-key"), 0)
	if err == nil || err.Error() != "expiry must be positive" {
		t.Errorf("Expected error for invalid expiry, got: %v", err)
	}
}

func TestExportSecretsFailsWithMissingCurrentVersion(t *testing.T) {
	service, _, mockSecrets, _ := setupExportSecretsTest()
	badSecret := domain.Secret{
		SecretName:     "bad",
		Type:           domain.SecretTypeKeyValue,
		CurrentVersion: 0, // No current version
		Versions: []domain.SecretVersion{{
			SecretValueEnc: "enc-value",
			Version:        1,
			UpdatedAt:      "2025-08-27",
		}},
	}
	secrets := []domain.Secret{badSecret}
	mockSecrets.On("GetSecretValue", &badSecret).Return("any-value", nil)
	_, err := service.ExportSecrets(secrets, []byte("recipient-key"), 60)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no current version")
	mockSecrets.AssertExpectations(t)
}

func TestExportSecretsFailsWithGetSecretValueError(t *testing.T) {
	service, _, mockSecrets, validSecret := setupExportSecretsTest()
	// Set up the mock to return an error for GetSecretValue
	mockSecrets.On("GetSecretValue", &validSecret).Return("", fmt.Errorf("failed to decrypt secret test"))
	secrets := []domain.Secret{validSecret}
	_, err := service.ExportSecrets(secrets, []byte("recipient-key"), 60)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt secret test")
	mockSecrets.AssertExpectations(t)
}

func TestSharingServiceImportSecretsSuccess(t *testing.T) {
	provider := &MockSharingProvider{}
	importProvider := &mocks.ImportProvider{}
	mockSecrets := &mocks.SecretsProvider{}
	service := sharing.NewSharingService(provider, importProvider, mockSecrets)
	bundle := &sharing.SecretExportBundle{} // Fill with test data as needed
	recipientPrivKey := []byte("recipient-private-key")
	recipientEphemeralPubKey := []byte("test-ephemeral-public-key")

	expectedResult := &sharing.SecretImportResult{
		Success:              true,
		ImportedSecretsCount: 2,
		VaultName:            "TestVault",
		Error:                nil,
	}

	importProvider.On("ImportSecrets", bundle, recipientPrivKey, recipientEphemeralPubKey).Return(expectedResult, nil)

	result, err := service.ImportSecrets(bundle, recipientPrivKey, recipientEphemeralPubKey)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Equal(t, 2, result.ImportedSecretsCount)
	require.Equal(t, "TestVault", result.VaultName)

	require.Nil(t, result.Error, "Error field should be nil on success")
	require.GreaterOrEqual(t, result.ImportedSecretsCount, 1, "Should import at least one secret")
	require.NotEmpty(t, result.VaultName, "VaultName should not be empty")
	importProvider.AssertExpectations(t)

}
