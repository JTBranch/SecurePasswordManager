package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"time"
)

// CryptoProvider defines the contract for cryptographic operations that the SecretsService needs.
type CryptoProvider interface {
	EncryptSymmetric(plaintext, key []byte) ([]byte, []byte, error)
	DecryptSymmetric(ciphertext, nonce, key []byte) ([]byte, error)
	GetKey() []byte
}

// StorageProvider defines the contract for storing and retrieving secrets.
type StorageProvider interface {
	ReadSecrets() (domain.SecretsFile, error)
	WriteSecrets(secrets domain.SecretsFile) error
}

// SecretsService manages secret operations with encryption
type SecretsService struct {
	crypto  CryptoProvider
	storage StorageProvider
}

// NewSecretsService creates a new secrets service
func NewSecretsService(crypto CryptoProvider, storage StorageProvider) *SecretsService {
	return &SecretsService{
		crypto:  crypto,
		storage: storage,
	}
}

// LoadAllSecrets loads all secrets with nested versions from the file
func (s *SecretsService) LoadAllSecrets() (domain.SecretsFile, error) {
	logger.Debug("Loading all secrets")
	return s.storage.ReadSecrets()
}

func (s *SecretsService) GetSecret(name string) (*domain.Secret, error) {
	secrets, err := s.storage.ReadSecrets()
	if err != nil {
		return nil, err
	}

	for i, secret := range secrets.Secrets {
		if secret.SecretName == name {
			return &secrets.Secrets[i], nil
		}
	}

	return nil, fmt.Errorf("secret not found: %s", name)
}

func (s *SecretsService) SaveOrUpdateSecrets(secrets []domain.Secret) error {
	logger.Debug("Doing bulk save/update for secrets")
	for _, secret := range secrets {
		if err := s.SaveOrUpdateSecret(secret.SecretName, secret.GetCurrentVersion().SecretValueEnc); err != nil {
			return err
		}
	}
	return nil
}

func (s *SecretsService) SaveOrUpdateSecret(name, value string) error {
	logger.Debug("updating or creating secret " + name)
	secretsData, err := s.storage.ReadSecrets()
	if err != nil {
		return err
	}

	var secretToUpdate *domain.Secret
	for i := range secretsData.Secrets {
		if secretsData.Secrets[i].SecretName == name {
			secretToUpdate = &secretsData.Secrets[i]
			break
		}
	}

	encryptedValue, nonce, err := s.crypto.EncryptSymmetric([]byte(value), s.crypto.GetKey())
	if err != nil {
		return err
	}

	if secretToUpdate == nil {
		// New secret
		newSecret := domain.Secret{
			SecretName: name,
			Versions: []domain.SecretVersion{
				{
					Version:        1,
					SecretValueEnc: base64.StdEncoding.EncodeToString(encryptedValue),
					Nonce:          nonce,
					UpdatedAt:      time.Now().Format(time.RFC3339),
				},
			},
			CurrentVersion: 1,
		}
		secretsData.Secrets = append(secretsData.Secrets, newSecret)
	} else {
		// Update existing secret
		newVersion := domain.SecretVersion{
			Version:        secretToUpdate.CurrentVersion + 1,
			SecretValueEnc: base64.StdEncoding.EncodeToString(encryptedValue),
			Nonce:          nonce,
			UpdatedAt:      time.Now().Format(time.RFC3339),
		}
		secretToUpdate.Versions = append(secretToUpdate.Versions, newVersion)
		secretToUpdate.CurrentVersion++
	}

	return s.storage.WriteSecrets(secretsData)
}

func (s *SecretsService) SaveNewSecret(name, value string) error {
	secretsData, err := s.storage.ReadSecrets()
	if err != nil {
		return err
	}

	// Check if secret with the same name already exists
	for _, secret := range secretsData.Secrets {
		if secret.SecretName == name {
			return fmt.Errorf("secret '%s' already exists", name)
		}
	}

	encryptedValue, nonce, err := s.crypto.EncryptSymmetric([]byte(value), s.crypto.GetKey())
	if err != nil {
		return err
	}

	newSecret := domain.Secret{
		SecretName: name,
		Versions: []domain.SecretVersion{
			{
				Version:        1,
				SecretValueEnc: base64.StdEncoding.EncodeToString(encryptedValue),
				Nonce:          nonce,
				UpdatedAt:      time.Now().Format(time.RFC3339),
			},
		},
		CurrentVersion: 1,
	}

	secretsData.Secrets = append(secretsData.Secrets, newSecret)
	return s.storage.WriteSecrets(secretsData)
}

func (s *SecretsService) UpdateSecret(name, newValue string) error {
	secretsData, err := s.storage.ReadSecrets()
	if err != nil {
		return err
	}

	var secretToUpdate *domain.Secret
	for i := range secretsData.Secrets {
		if secretsData.Secrets[i].SecretName == name {
			secretToUpdate = &secretsData.Secrets[i]
			break
		}
	}

	if secretToUpdate == nil {
		return fmt.Errorf("secret '%s' not found", name)
	}

	encryptedValue, nonce, err := s.crypto.EncryptSymmetric([]byte(newValue), s.crypto.GetKey())
	if err != nil {
		return err
	}

	newVersion := domain.SecretVersion{
		Version:        secretToUpdate.CurrentVersion + 1,
		SecretValueEnc: base64.StdEncoding.EncodeToString(encryptedValue),
		Nonce:          nonce,
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}

	secretToUpdate.Versions = append(secretToUpdate.Versions, newVersion)
	secretToUpdate.CurrentVersion++

	return s.storage.WriteSecrets(secretsData)
}

func (s *SecretsService) DeleteSecret(name string) error {
	data, err := s.storage.ReadSecrets()
	if err != nil {
		return err
	}

	var newSecrets []domain.Secret
	var found bool
	for _, secret := range data.Secrets {
		if secret.SecretName != name {
			newSecrets = append(newSecrets, secret)
		} else {
			found = true
		}
	}

	if !found {
		return nil // Idempotent delete
	}
	data.Secrets = newSecrets

	return s.storage.WriteSecrets(data)
}

func (s *SecretsService) GetSecretValue(secret *domain.Secret) (string, error) {
	var currentVersion *domain.SecretVersion
	for i := range secret.Versions {
		if secret.Versions[i].Version == secret.CurrentVersion {
			currentVersion = &secret.Versions[i]
			break
		}
	}

	if currentVersion == nil {
		return "", fmt.Errorf("no current version found for secret '%s'", secret.SecretName)
	}

	encryptedData, err := base64.StdEncoding.DecodeString(currentVersion.SecretValueEnc)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	plainBytes, err := s.crypto.DecryptSymmetric(encryptedData, currentVersion.Nonce, s.crypto.GetKey())
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to decrypt secret '%s': ", secret.SecretName), err)
		return "", err
	}
	logger.Debug("Decrypted secret value:", string(plainBytes))
	return string(plainBytes), nil
}

func (s *SecretsService) GetSecretValueByVersion(secret *domain.Secret, versionNumber int) (string, error) {
	for _, version := range secret.Versions {
		if version.Version == versionNumber {
			logger.Debug("Decrypting secret version:", fmt.Sprintf("%d", version.Version))

			encryptedData, err := base64.StdEncoding.DecodeString(version.SecretValueEnc)
			if err != nil {
				return "", fmt.Errorf("failed to decode encrypted data for version %d: %v", version.Version, err)
			}

			plainBytes, err := s.crypto.DecryptSymmetric(encryptedData, version.Nonce, s.crypto.GetKey())
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to decrypt secret version %d for '%s': ", version.Version, secret.SecretName), err)
				return "", err
			}

			return string(plainBytes), nil
		}
	}
	return "", fmt.Errorf("version %d not found for secret '%s'", versionNumber, secret.SecretName)
}

func (s *SecretsService) RevertToVersion(secretName string, version int) error {
	secrets, err := s.storage.ReadSecrets()
	if err != nil {
		return err
	}

	for i, sec := range secrets.Secrets {
		if sec.SecretName == secretName {
			secrets.Secrets[i].CurrentVersion = version
			return s.storage.WriteSecrets(secrets)
		}
	}

	return fmt.Errorf("secret not found: %s", secretName)
}

func (s *SecretsService) GetCurrentVersionValue(name string) (string, error) {
	secret, err := s.GetSecret(name)
	if err != nil {
		return "", err
	}
	return s.GetSecretValue(secret)
}

func (s *SecretsService) GetRaw() ([]byte, error) {
	secretsFile, err := s.storage.ReadSecrets()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(secretsFile, "", "  ")
}

func (s *SecretsService) GetTotalSecrets() (int, error) {
	secretsFile, err := s.storage.ReadSecrets()
	if err != nil {
		return 0, err
	}
	return len(secretsFile.Secrets), nil
}

func (s *SecretsService) GetTotalVersions() (int, error) {
	secretsFile, err := s.storage.ReadSecrets()
	if err != nil {
		return 0, err
	}
	totalVersions := 0
	for _, secret := range secretsFile.Secrets {
		totalVersions += len(secret.Versions)
	}
	return totalVersions, nil
}

func (s *SecretsService) GetLastUpdated() (string, error) {
	secretsFile, err := s.storage.ReadSecrets()
	if err != nil {
		return "", err
	}
	return secretsFile.LastUpdated, nil
}

func (s *SecretsService) GetAppVersion() (string, error) {
	secretsFile, err := s.storage.ReadSecrets()
	if err != nil {
		return "", err
	}
	return secretsFile.AppVersion, nil
}

func (s *SecretsService) GetAppUser() (string, error) {
	secretsFile, err := s.storage.ReadSecrets()
	if err != nil {
		return "", err
	}
	return secretsFile.AppUser, nil
}
