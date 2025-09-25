package sharing

import (
	"encoding/json"
	"fmt"
	"go-password-manager/internal/config/devicekeys"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"time"
)

// ImportService handles the import of shared secret bundles.
type ImportService struct {
	CryptoProvider    CryptoProvider
	DeviceKeyProvider DeviceKeyProvider
	SecretsProvider   SecretsProvider
	replayStore       ReplayStore
	// Add other dependencies as needed (e.g., storage, logger)
}

// NewImportService creates a new ImportService with dependencies injected.
func NewImportService(cryptoProvider CryptoProvider, deviceKeyProvider DeviceKeyProvider, secretsProvider SecretsProvider) *ImportService {
	return &ImportService{
		CryptoProvider:    cryptoProvider,
		DeviceKeyProvider: deviceKeyProvider,
		SecretsProvider:   secretsProvider,
		replayStore:       NewInMemoryReplayStore(time.Hour),
	}
}

// NewImportServiceWithReplayConfig builds an ImportService using a ReplayConfig (memory by default).
func NewImportServiceWithReplayConfig(cryptoProvider CryptoProvider, deviceKeyProvider DeviceKeyProvider, secretsProvider SecretsProvider, cfg ReplayConfig) (*ImportService, error) {
	store, err := NewReplayStoreFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &ImportService{CryptoProvider: cryptoProvider, DeviceKeyProvider: deviceKeyProvider, SecretsProvider: secretsProvider, replayStore: store}, nil
}

// GetReplayStore returns the underlying ReplayStore (read-only usage recommended).
func (s *ImportService) GetReplayStore() ReplayStore { return s.replayStore }

// SetReplayTTL updates TTL on underlying replay store if supported.
func (s *ImportService) SetReplayTTL(seconds int64) {
	if store, ok := s.replayStore.(*InMemoryReplayStore); ok {
		store.SetTTL(time.Duration(seconds) * time.Second)
	}
	if store, ok := s.replayStore.(*FileReplayStore); ok {
		store.SetTTL(time.Duration(seconds) * time.Second)
	}
}

// UseFileReplayStore swaps in a persistent file-based replay store (path relative or absolute).
func (s *ImportService) UseFileReplayStore(path string) error {
	store, err := NewFileReplayStore(path, time.Hour)
	if err != nil {
		return err
	}
	s.replayStore = store
	return nil
}

// ImportSecrets imports and decrypts secrets from a SecretExportBundle using the recipient's private key.
func (s *ImportService) ImportSecrets(bundle *SecretExportBundle, recipientPrivateKey []byte, recipientEphemeralPubKey []byte) (*SecretImportResult, error) {
	if bundle == nil {
		logger.Warn("ImportService.ImportSecrets called with nil bundle")
		return &SecretImportResult{ImportedSecretsCount: 0, Success: false, Error: fmt.Errorf("nil bundle")}, nil
	}
	// Log entry for observability during runtime/share flows
	logger.Info(fmt.Sprintf("ImportService.ImportSecrets invoked: bundleID=%s senderSigningPub=%x", bundle.Payload.ID, bundle.Payload.SenderInfo.SigningPublicKey))

	// Check for bundle expiry
	if bundle.Payload.ExpiresAt > 0 && bundle.Payload.ExpiresAt < (time.Now().Unix()) {
		logger.Warn("Bundle has expired, ending import")
		return &SecretImportResult{
			ImportedSecretsCount: 0,
			Success:              false,
			Error:                fmt.Errorf("bundle has expired"),
		}, nil
	}
	// Replay protection
	if s.replayStore != nil && s.replayStore.Seen(bundle.Payload.ID) {
		return &SecretImportResult{ImportedSecretsCount: 0, Success: false, Error: fmt.Errorf("bundle already processed (replay)")}, nil
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

	// save (will encrypt locally)
	err = s.SecretsProvider.SaveOrUpdateSecrets(secrets)
	if err != nil {
		logger.Error("Failed to do bulk update in import secrets, ending import", err)
		return &SecretImportResult{
			ImportedSecretsCount: 0,
			Success:              false,
			Error:                err,
		}, nil
	}

	// Mark as seen after successful persistence
	if s.replayStore != nil {
		s.replayStore.Mark(bundle.Payload.ID, 0)
	}
	return &SecretImportResult{ImportedSecretsCount: len(secrets), Success: true}, nil
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
func (s *ImportService) DecryptSecrets(bundle *SecretExportBundle, recipientPrivateKey []byte, _ []byte) ([]domain.Secret, error) {
	// Decode recipient private key PEM if supported
	if dec, ok := s.CryptoProvider.(interface {
		DecodeKeyFromPEM([]byte, devicekeys.KeyType) ([]byte, error)
	}); ok {
		decodedPriv, derr := dec.DecodeKeyFromPEM(recipientPrivateKey, devicekeys.KeyTypeX25519Private)
		if derr != nil {
			return nil, fmt.Errorf("failed to decode recipient private key: %w", derr)
		}
		recipientPrivateKey = decodedPriv
	}

	// 1. Reconstruct key box AAD: bundleID || 0x00 || senderSigningPub
	keyBoxAAD := buildKeyBoxAAD(bundle.Payload.ID, bundle.Payload.SenderInfo.SigningPublicKey)
	// 2. Unwrap symmetric key from key box using keyBoxAAD
	symKey, err := crypto.UnwrapKeyBox(bundle.Payload.SymmetricKeyBox, recipientPrivateKey, keyBoxAAD)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap symmetric key: %w", err)
	}
	// 3. Decrypt secrets
	// 3. Reconstruct secrets AAD: bundleID || 0x01 || senderSigningPub
	secretsAAD := buildSecretsAAD(bundle.Payload.ID, bundle.Payload.SenderInfo.SigningPublicKey)
	decryptedBytes, err := s.CryptoProvider.DecryptSymmetric(bundle.Payload.EncryptedSecrets, bundle.Payload.SecretsNonce, symKey, secretsAAD)
	if err != nil {
		crypto.Zeroize(symKey)
		return nil, fmt.Errorf("failed to decrypt secrets: %w", err)
	}
	// Zeroize symmetric key after decryption
	crypto.Zeroize(symKey)
	// 4. Unmarshal to []ExportSecret
	var exportSecrets []ExportSecret
	if err := json.Unmarshal(decryptedBytes, &exportSecrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted secrets: %w", err)
	}

	// 4. Convert to []domain.Secret
	domainSecrets := ConvertExportSecretsToDomainSecrets(exportSecrets)
	return domainSecrets, nil
}
