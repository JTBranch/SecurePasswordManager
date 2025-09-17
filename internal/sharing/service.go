// NOTE: Only methods on SharingService are exported and form the public API for sharing.
// Internal helpers and logic should be kept unexported in other files (e.g., export.go).
package sharing

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"time"
)

type ExportProvider interface {
	ExportSecrets(secrets []ExportSecret, recipientPubKey []byte, expiryMinutes int, senderInfo SenderMetadata) (*SecretExportBundle, error)
}

type SecretsProvider interface {
	GetSecretValue(secret *domain.Secret) (string, error)
	SaveOrUpdateSecrets(secrets []domain.Secret) error
}

type ImportProvider interface {
	ImportSecrets(bundle *SecretExportBundle, recipientPrivateKey []byte, recipientEphemeralPubKey []byte) (*SecretImportResult, error)
}

type LoggerProvider interface {
	Info(message string)
	Error(message string)
}

type SharingService struct {
	exportProvider  ExportProvider
	importProvider  ImportProvider
	secretsProvider SecretsProvider
}

func NewSharingService(exportProvider ExportProvider, importProvider ImportProvider, secretsProvider SecretsProvider) *SharingService {
	return &SharingService{
		exportProvider:  exportProvider,
		importProvider:  importProvider,
		secretsProvider: secretsProvider,
	}
}

// ExportSecrets exports selected secrets to a secure bundle.
func (s *SharingService) ExportSecrets(secrets []domain.Secret, recipientPubKey []byte, expiryMinutes int) (*SecretExportBundle, error) {
	logger.Info("Starting export of secrets")

	if len(secrets) == 0 {
		return nil, errors.New("no secrets to export")
	}
	if len(recipientPubKey) == 0 {
		return nil, errors.New("recipient public key is missing")
	}
	if expiryMinutes <= 0 {
		return nil, errors.New("expiry must be positive")
	}

	exportSecrets, err := mapToExportSecrets(secrets, s.secretsProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to map secrets for export: %w", err)
	}

	result, err := s.exportProvider.ExportSecrets(exportSecrets, recipientPubKey, expiryMinutes, SenderMetadata{
		// Set sender metadata fields
	})
	entry := SharingLogEntry{
		Action:  "export",
		Status:  "failed",
		Details: "",
	}
	if result != nil {
		entry.Timestamp = result.Payload.Timestamp
		entry.BundleID = result.Payload.ID
		entry.FileName = result.Payload.Name
		entry.PeerInfo = result.Payload.SenderInfo
		entry.Status = "success"
	}
	if err != nil {
		entry.Status = "failed"
		entry.Details = err.Error()
	}
	s.LogEvent(entry)
	return result, err
}

// ImportSecrets imports a bundle and returns the result.
func (s *SharingService) ImportSecrets(bundle *SecretExportBundle, recipientPrivateKey []byte, recipientEphemeralPubKey []byte) (*SecretImportResult, error) {

	logger.Info("Beginning import of secrets")

	result, err := s.importProvider.ImportSecrets(bundle, recipientPrivateKey, recipientEphemeralPubKey)

	entry := SharingLogEntry{Action: "import", Status: "failed"}

	if result != nil {
		entry.Timestamp = time.Now().Unix()
		entry.BundleID = bundle.Payload.ID
		entry.FileName = bundle.Payload.Name
		entry.PeerInfo = bundle.Payload.SenderInfo
		// Mark success only if the provider indicated success and no embedded error
		if result.Success && result.Error == nil {
			entry.Status = "success"
		} else if result.Error != nil {
			entry.Status = "failed"
			entry.Details = result.Error.Error()
		}
	}
	if err != nil {
		// Underlying provider returned an operational error (unexpected)
		entry.Status = "failed"
		entry.Details = err.Error()
	}
	s.LogEvent(entry)
	// propagate provider error (tests expect method-level err to be nil; import result carries logical errors)
	return result, err
}

// LogEvent records an import/export event.
func (s *SharingService) LogEvent(entry SharingLogEntry) {
	data, _ := json.Marshal(entry)
	logger.Info("Sharing event logged:")
	logger.Info(string(data)) // for file logging
}

// GetLog returns the import/export history.
func (s *SharingService) GetLog() ([]SharingLogEntry, error) {
	// this can be added in future if need for it
	return nil, nil
}

func mapToExportSecrets(secrets []domain.Secret, secretsProvider SecretsProvider) ([]ExportSecret, error) {
	exportSecrets := make([]ExportSecret, 0, len(secrets))
	for _, secret := range secrets {
		value, err := secretsProvider.GetSecretValue(&secret)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret %s: %w", secret.SecretName, err)
		}
		current := secret.GetCurrentVersion()
		if current == nil {
			return nil, fmt.Errorf("secret %s has no current version", secret.SecretName)
		}
		exportSecrets = append(exportSecrets, ExportSecret{
			Name:      secret.SecretName,
			Type:      secret.Type,
			Value:     value,
			UpdatedAt: current.UpdatedAt,
			Version:   current.Version})
	}
	return exportSecrets, nil
}
