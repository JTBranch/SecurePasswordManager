package service

import (
	"encoding/json"
	"fmt"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"time"
)

// SecretsService manages secret operations with encryption
type SecretsService struct {
	crypto  CryptoService
	storage StorageService
}

// NewSecretsService creates a new secrets service
func NewSecretsService(crypto CryptoService, storage StorageService) *SecretsService {
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

	encryptedValue, err := s.crypto.Encrypt([]byte(value), s.crypto.GetKey())
	if err != nil {
		return err
	}

	newSecret := domain.Secret{
		SecretName: name,
		Versions: []domain.SecretVersion{
			{
				Version:        1,
				SecretValueEnc: string(encryptedValue),
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

	encryptedValue, err := s.crypto.Encrypt([]byte(newValue), s.crypto.GetKey())
	if err != nil {
		return err
	}

	newVersion := domain.SecretVersion{
		Version:        secretToUpdate.CurrentVersion + 1,
		SecretValueEnc: string(encryptedValue),
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
	plainBytes, err := s.crypto.Decrypt([]byte(currentVersion.SecretValueEnc), s.crypto.GetKey())
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}

func (s *SecretsService) GetSecretValueByVersion(secret *domain.Secret, versionNumber int) (string, error) {
	for _, version := range secret.Versions {
		if version.Version == versionNumber {
			logger.Debug("Decrypting secret version:", fmt.Sprintf("%d", version.Version))

			plainBytes, err := s.crypto.Decrypt([]byte(version.SecretValueEnc), s.crypto.GetKey())
			if err != nil {
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
