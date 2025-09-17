package sharing

import (
	"encoding/json"
	"fmt"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"time"
)

// ImportService handles the import of shared secret bundles.
type ImportService struct {
	CryptoProvider    CryptoProvider
	DeviceKeyProvider DeviceKeyProvider
	SecretsProvider   SecretsProvider
	// Add other dependencies as needed (e.g., storage, logger)
}

// NewImportService creates a new ImportService with dependencies injected.
func NewImportService(cryptoProvider CryptoProvider, deviceKeyProvider DeviceKeyProvider, secretsProvidor SecretsProvider) *ImportService {
	return &ImportService{
		CryptoProvider:    cryptoProvider,
		DeviceKeyProvider: deviceKeyProvider,
		SecretsProvider:   secretsProvidor,
	}
}

// ImportSecrets imports and decrypts secrets from a SecretExportBundle using the recipient's private key.
func (s *ImportService) ImportSecrets(bundle *SecretExportBundle, recipientPrivateKey []byte, recipientEphemeralPubKey []byte) (*SecretImportResult, error) {

	// Check for bundle expiry
	if bundle.Payload.ExpiresAt > 0 && bundle.Payload.ExpiresAt < (time.Now().Unix()) {
		logger.Warn("Bundle has expired, ending import")
		return &SecretImportResult{
			ImportedSecretsCount: 0,
			Success:              false,
			Error:                fmt.Errorf("bundle has expired"),
		}, nil
	}
	logger.Debug("Bundle is not expired, proceeding with import")
	// Verify the bundle's signature
	valid, err := s.VerifyBundleSignature(bundle)
	if err != nil {
		logger.Error("signiture is not valid, ending import", err)
		return &SecretImportResult{
			ImportedSecretsCount: 0,
			Success:              false,
			Error:                err,
		}, nil
	}
	if !valid {
		return &SecretImportResult{
			ImportedSecretsCount: 0,
			Success:              false,
			Error:                fmt.Errorf("invalid bundle signature"),
		}, nil
	}
	logger.Debug("Signiture is valid, proceeding with import")

	// Decrypt the secrets
	secrets, err := s.DecryptSecrets(bundle, recipientPrivateKey, recipientEphemeralPubKey)
	if err != nil {
		logger.Error("One or more secrets failed to decrypt, ending import", err)
		return &SecretImportResult{
			ImportedSecretsCount: 0,
			Success:              false,
			Error:                err,
		}, nil
	}

	// save the secrets

	err = s.SecretsProvider.SaveOrUpdateSecrets(secrets)
	if err != nil {
		logger.Error("Failed to do bulk update in import secrets, ending import", err)
		return &SecretImportResult{
			ImportedSecretsCount: 0,
			Success:              false,
			Error:                err,
		}, nil
	}

	return &SecretImportResult{
		ImportedSecretsCount: len(secrets),
		Success:              true,
	}, nil
}

// ConvertExportSecretsToDomainSecrets converts []ExportSecret to []domain.Secret for bulk import.
func ConvertExportSecretsToDomainSecrets(exportSecrets []ExportSecret) []domain.Secret {
	domainSecrets := make([]domain.Secret, 0, len(exportSecrets))
	for _, es := range exportSecrets {
		sec := domain.Secret{
			SecretName: es.Name,
			Type:       es.Type,
			Versions: []domain.SecretVersion{
				{
					Version:        es.Version,
					SecretValueEnc: es.Value, // assumes already encrypted
					UpdatedAt:      es.UpdatedAt,
				},
			},
			CurrentVersion: es.Version,
		}
		domainSecrets = append(domainSecrets, sec)
	}
	return domainSecrets
}

// VerifyBundleSignature verifies the signature of the export bundle.
func (s *ImportService) VerifyBundleSignature(bundle *SecretExportBundle) (bool, error) {
	payloadString, err := json.Marshal(bundle.Payload)
	if err != nil {
		logger.Error("Failed to marshal bundle payload", err)
	}
	return s.CryptoProvider.VerifyEd25519(payloadString, bundle.Signature, bundle.Payload.SenderInfo.SigningPublicKey)
}

// DecryptSecrets decrypts the secrets from the bundle using the derived symmetric key.
func (s *ImportService) DecryptSecrets(bundle *SecretExportBundle, recipientPrivateKey []byte, recipientEphemeralPubKey []byte) ([]domain.Secret, error) {
	// 1. Derive shared secret using ECDH (recipient's private key, sender's ephemeral public key)
	sharedSecret, err := s.CryptoProvider.ECDH(recipientPrivateKey, bundle.Payload.EphemeralPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive shared secret: %w", err)
	}

	// 2. Derive symmetric key
	derivedKey, err := s.CryptoProvider.DeriveSymmetricKey(sharedSecret, bundle.Payload.EphemeralPublicKey, recipientEphemeralPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive symmetric key: %w", err)
	}

	// 3. Decrypt secrets
	decryptedBytes, err := s.CryptoProvider.DecryptSymmetric(bundle.Payload.EncryptedSecrets, bundle.Payload.SecretsNonce, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secrets: %w", err)
	}

	// 4. Unmarshal to []ExportSecret
	var exportSecrets []ExportSecret
	err = json.Unmarshal(decryptedBytes, &exportSecrets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted secrets: %w", err)
	}

	// 5. Convert to []domain.Secret
	domainSecrets := ConvertExportSecretsToDomainSecrets(exportSecrets)
	return domainSecrets, nil
}
