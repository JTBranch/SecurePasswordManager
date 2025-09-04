package sharing_test

import (
	"fmt"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/sharing"
	"strings"
	"testing"
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

func setupExportSecretsTest() (*sharing.SharingService, *MockSharingProvider, *MockSecretsService, domain.Secret) {
	const (
		secretValueEnc = "enc-value"
		updatedAt      = "2025-08-27"
	)
	provider := &MockSharingProvider{}
	mockSecrets := &MockSecretsService{Values: map[string]string{"test": "decrypted-value"}}
	service := sharing.NewSharingService(provider, mockSecrets)
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
	service, provider, _, validSecret := setupExportSecretsTest()
	secrets := []domain.Secret{validSecret}
	recipientPubKey := []byte("recipient-key")
	expiry := 60

	bundle, err := service.ExportSecrets(secrets, recipientPubKey, expiry)
	if err != nil {
		t.Fatalf("ExportSecrets returned error: %v", err)
	}
	if bundle == nil {
		t.Fatalf("ExportSecrets returned nil bundle")
	}
	if !provider.EncryptCalled {
		t.Errorf("EncryptAsymmetricPEM should be called")
	}
	if string(bundle.Payload.EncryptedSecrets) != "mock-ciphertext" {
		t.Errorf("EncryptedSecrets should match mock ciphertext")
	}
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
	mockSecrets.Values["bad"] = "any-value" // ensure GetSecretValue succeeds
	badSecret := domain.Secret{
		SecretName:     "bad",
		Type:           domain.SecretTypeKeyValue,
		CurrentVersion: 0,
		Versions: []domain.SecretVersion{{
			SecretValueEnc: "enc-value",
			Version:        1,
			UpdatedAt:      "2025-08-27",
		}},
	}
	_, err := service.ExportSecrets([]domain.Secret{badSecret}, []byte("recipient-key"), 60)
	if err == nil || err.Error() != "failed to map secrets for export: secret bad has no current version" {
		t.Errorf("Expected error for missing current version, got: %v", err)
	}
	delete(mockSecrets.Values, "bad") // cleanup
}

func TestExportSecretsFailsWithGetSecretValueError(t *testing.T) {
	service, _, mockSecrets, validSecret := setupExportSecretsTest()
	mockSecrets.Values = map[string]string{} // simulate missing value
	_, err := service.ExportSecrets([]domain.Secret{validSecret}, []byte("recipient-key"), 60)
	if err == nil || err.Error() == "" || !contains(err.Error(), "failed to decrypt secret test") {
		t.Errorf("Expected error for GetSecretValue failure, got: %v", err)
	}
	mockSecrets.Values = map[string]string{"test": "decrypted-value"} // restore
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
