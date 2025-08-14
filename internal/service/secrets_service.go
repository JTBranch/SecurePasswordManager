package service

import (
	"encoding/json"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"os"
	"path/filepath"
	"time"
)

const secretsFile = "secrets.json"

type SecretsService struct {
	AppVersion    string
	AppUser       string
	filePath      string
	encryptionKey []byte
}

func ensureSecretsFileExists(filePath, appVersion, appUser string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Debug("secrets.json not found, creating new file")
		initial := domain.SecretsFile{
			AppVersion:  appVersion,
			AppUser:     appUser,
			LastUpdated: time.Now().Format(time.RFC3339),
			Secrets:     []domain.Secret{},
		}
		jsonBytes, _ := json.MarshalIndent(initial, "", "  ")
		_ = os.WriteFile(filePath, jsonBytes, 0600)
	}
}

func NewSecretsService(appVersion, appUser string) *SecretsService {
	key, err := crypto.LoadOrCreateKey()
	if err != nil {
		logger.Error("Failed to load encryption key:", err.Error())
	}

	envService := NewEnvironmentService()
	var filePath string
	if envService.IsProduction() {
		configDir, err := os.UserConfigDir()
		if err != nil {
			logger.Error("Failed to get user config dir:", err.Error())
			filePath = filepath.Join(".", secretsFile) // fallback
		} else {
			appConfigDir := filepath.Join(configDir, "GoPasswordManager")
			if err := os.MkdirAll(appConfigDir, 0700); err != nil {
				logger.Error("Failed to create app config dir:", err.Error())
				filePath = filepath.Join(".", secretsFile) // fallback
			} else {
				filePath = filepath.Join(appConfigDir, secretsFile)
			}
		}
	} else {
		filePath = filepath.Join(".", secretsFile)
	}
	ensureSecretsFileExists(filePath, appVersion, appUser)
	return &SecretsService{
		AppVersion:    appVersion,
		AppUser:       appUser,
		filePath:      filePath,
		encryptionKey: key,
	}
}

// 1. Load all secrets (no decryption)
func (s *SecretsService) LoadAllSecrets() (domain.SecretsFile, error) {
	logger.Debug("Loading all secrets from file:", s.filePath)
	file, err := os.Open(s.filePath)
	if err != nil {
		logger.Debug("Error opening secrets file:", err.Error())
		return domain.SecretsFile{}, err
	}
	defer file.Close()
	var data domain.SecretsFile
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		logger.Debug("Error decoding secrets file:", err.Error())
		return domain.SecretsFile{}, err
	}
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	logger.Debug("Loaded secrets file JSON:", string(jsonBytes))
	return data, nil
}

// 2. Save or update a secret
func (s *SecretsService) SaveSecret(secretName, secretValue string, secretType domain.SecretType) error {
	logger.Debug("Encrypting and saving secret:", secretName, secretValue)
	encryptedValue, err := crypto.Encrypt([]byte(secretValue), s.encryptionKey)
	if err != nil {
		logger.Debug("Encryption failed:", err.Error())
		return err
	}
	secret := domain.Secret{
		SecretName:     secretName,
		SecretValueEnc: encryptedValue,
		Type:           secretType,
	}
	data, err := s.LoadAllSecrets()
	if err != nil {
		logger.Debug("No existing secrets file, creating new one")
		data = domain.SecretsFile{
			AppVersion:  s.AppVersion,
			AppUser:     s.AppUser,
			LastUpdated: time.Now().Format(time.RFC3339),
			Secrets:     []domain.Secret{},
		}
	}
	found := false
	for i, sec := range data.Secrets {
		if sec.SecretName == secret.SecretName {
			secret.Version = sec.Version + 1
			secret.UpdatedAt = time.Now().Format(time.RFC3339)
			data.Secrets[i] = secret
			found = true
			break
		}
	}
	if !found {
		secret.Version = 1
		secret.UpdatedAt = time.Now().Format(time.RFC3339)
		data.Secrets = append(data.Secrets, secret)
	}
	data.LastUpdated = time.Now().Format(time.RFC3339)
	logger.Debug("Saving secrets file after update")
	return s.saveFile(data)
}

// 3. Delete a secret (hard delete)
func (s *SecretsService) DeleteSecret(secretName string) error {
	logger.Debug("Deleting secret:", secretName)
	data, err := s.LoadAllSecrets()
	if err != nil {
		logger.Debug("Error loading secrets for delete:", err.Error())
		return err
	}
	newSecrets := []domain.Secret{}
	for _, sec := range data.Secrets {
		if sec.SecretName != secretName {
			newSecrets = append(newSecrets, sec)
		}
	}
	data.Secrets = newSecrets
	data.LastUpdated = time.Now().Format(time.RFC3339)
	logger.Debug("Saving secrets file after delete")
	return s.saveFile(data)
}

// 4. Display (decrypt) a secret value
func (s *SecretsService) DisplaySecret(secret domain.Secret) (string, error) {
	logger.Debug("Decrypting secret:", secret.SecretName)
	plainBytes, err := crypto.Decrypt(secret.SecretValueEnc, s.encryptionKey)
	if err != nil {
		logger.Debug("Error decrypting secret:", err.Error())
		return "", err
	}
	logger.Debug("Decrypted secret value for name:", secret.SecretName, " value:", string(plainBytes))
	return string(plainBytes), nil
}

// Helper to save file
func (s *SecretsService) saveFile(data domain.SecretsFile) error {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	logger.Debug("Saving secrets file JSON:", string(jsonBytes))
	file, err := os.Create(s.filePath)
	if err != nil {
		logger.Debug("Error creating secrets file:", err.Error())
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(data)
}

// Map UI secrets to backend secrets
func (s *SecretsService) ToBackendSecrets(uiSecrets []domain.SecretView) []domain.Secret {
	var entries []domain.Secret
	for _, us := range uiSecrets {
		entries = append(entries, domain.Secret{
			SecretName:     us.SecretName,
			Type:           domain.SecretType(us.Type),
			SecretValueEnc: us.SecretValueEnc,
		})
	}
	return entries
}

// Create backend secrets file from UI secrets
func (s *SecretsService) CreateSecretsFile(uiSecrets []domain.SecretView) domain.SecretsFile {
	return domain.SecretsFile{
		AppVersion:  s.AppVersion,
		AppUser:     s.AppUser,
		LastUpdated: time.Now().Format(time.RFC3339),
		Secrets:     s.ToBackendSecrets(uiSecrets),
	}
}

// Map backend secrets to UI secrets
func (s *SecretsService) ToUISecrets(entries []domain.Secret) []domain.SecretView {
	var uiSecrets []domain.SecretView
	for _, e := range entries {
		uiSecrets = append(uiSecrets, domain.SecretView{
			SecretName:     e.SecretName,
			Type:           string(e.Type),
			SecretValueEnc: e.SecretValueEnc,
			Revealed:       false,
		})
	}
	return uiSecrets
}
