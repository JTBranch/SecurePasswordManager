package service

import (
	"encoding/json"
	"fmt"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"os"
	"path/filepath"
	"time"
)

const secretsFile = "secrets.json"

// SecretsService manages secret operations with encryption
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

// NewSecretsService creates a new secrets service
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

// LoadAllSecrets loads all secrets with nested versions from the file
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

// LoadLatestSecrets loads only the latest versions of secrets for UI display
func (s *SecretsService) LoadLatestSecrets() (domain.SecretsFile, error) {
	allSecrets, err := s.LoadAllSecrets()
	if err != nil {
		return domain.SecretsFile{}, err
	}

	// The data structure already contains only latest versions in the UI representation
	// Each secret has its versions nested, so we can return it as-is
	return allSecrets, nil
}

// 1b. Get all versions of a specific secret for history display
func (s *SecretsService) GetSecretVersions(secretName string) ([]domain.SecretVersion, error) {
	allSecrets, err := s.LoadAllSecrets()
	if err != nil {
		return nil, err
	}

	// Find the secret and return its versions
	for _, secret := range allSecrets.Secrets {
		if secret.SecretName == secretName {
			return secret.GetVersionsSorted(), nil
		}
	}

	return []domain.SecretVersion{}, nil
}

// 2. Save a secret (encrypted) - creates new secret or adds version to existing
func (s *SecretsService) SaveSecret(name, value, secretType string) error {
	logger.Debug("Saving secret:", name)
	secretsData, err := s.LoadAllSecrets()
	if err != nil {
		logger.Error("Failed to load secrets:", err.Error())
		return err
	}

	encryptedValue, err := crypto.Encrypt([]byte(value), s.encryptionKey)
	if err != nil {
		logger.Error("Failed to encrypt secret:", err.Error())
		return err
	}

	// Find if secret already exists
	var existingSecret *domain.Secret
	for i := range secretsData.Secrets {
		if secretsData.Secrets[i].SecretName == name {
			existingSecret = &secretsData.Secrets[i]
			break
		}
	}

	if existingSecret != nil {
		// Add new version to existing secret
		newVersion := domain.SecretVersion{
			SecretValueEnc: encryptedValue,
			Version:        existingSecret.CurrentVersion + 1,
			UpdatedAt:      time.Now().Format(time.RFC3339),
		}
		existingSecret.Versions = append(existingSecret.Versions, newVersion)
		existingSecret.CurrentVersion = newVersion.Version
	} else {
		// Create new secret with first version
		newSecret := domain.Secret{
			SecretName:     name,
			Type:           domain.SecretType(secretType),
			CurrentVersion: 1,
			Versions: []domain.SecretVersion{
				{
					SecretValueEnc: encryptedValue,
					Version:        1,
					UpdatedAt:      time.Now().Format(time.RFC3339),
				},
			},
		}
		secretsData.Secrets = append(secretsData.Secrets, newSecret)
	}

	secretsData.LastUpdated = time.Now().Format(time.RFC3339)
	return s.saveFile(secretsData)
}

// 2a. Edit a secret (creates new version, keeps historical versions)
func (s *SecretsService) EditSecret(name, newValue string) error {
	logger.Debug("Editing secret: ", name, " value: ", newValue)
	secretsData, err := s.LoadAllSecrets()
	if err != nil {
		logger.Error("Failed to load secrets:", err.Error())
		return err
	}

	// Find the secret to edit
	var secretToEdit *domain.Secret
	for i := range secretsData.Secrets {
		if secretsData.Secrets[i].SecretName == name {
			secretToEdit = &secretsData.Secrets[i]
			break
		}
	}

	if secretToEdit == nil {
		logger.Error("Secret not found for editing:", name)
		return fmt.Errorf("secret '%s' not found", name)
	}

	// Create a new version of the secret (keep old versions)
	encryptedValue, err := crypto.Encrypt([]byte(newValue), s.encryptionKey)
	if err != nil {
		logger.Error("Failed to encrypt secret:", err.Error())
		return err
	}

	// Create a new version
	newVersion := domain.SecretVersion{
		SecretValueEnc: encryptedValue,
		Version:        secretToEdit.CurrentVersion + 1,
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}

	// Add the new version to the secret
	secretToEdit.Versions = append(secretToEdit.Versions, newVersion)
	secretToEdit.CurrentVersion = newVersion.Version
	secretsData.LastUpdated = time.Now().Format(time.RFC3339)

	return s.saveFile(secretsData)
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

// 4. Display (decrypt) a secret value from current version
func (s *SecretsService) DisplaySecret(secret domain.Secret) (string, error) {
	logger.Debug("Decrypting secret:", secret.SecretName)

	currentVersion := secret.GetCurrentVersion()
	if currentVersion == nil {
		return "", fmt.Errorf("no current version found for secret '%s'", secret.SecretName)
	}

	plainBytes, err := crypto.Decrypt(currentVersion.SecretValueEnc, s.encryptionKey)
	if err != nil {
		logger.Debug("Error decrypting secret:", err.Error())
		return "", err
	}
	logger.Debug("Decrypted secret value for name:", secret.SecretName, " value:", string(plainBytes))
	return string(plainBytes), nil
}

// 4a. Decrypt a specific secret version
func (s *SecretsService) DecryptSecretVersion(version domain.SecretVersion) (string, error) {
	logger.Debug("Decrypting secret version:", fmt.Sprintf("%d", version.Version))

	plainBytes, err := crypto.Decrypt(version.SecretValueEnc, s.encryptionKey)
	if err != nil {
		logger.Debug("Error decrypting secret version:", err.Error())
		return "", err
	}
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
		// Create a new secret with first version
		entries = append(entries, domain.Secret{
			SecretName:     us.SecretName,
			Type:           domain.SecretType(us.Type),
			CurrentVersion: 1,
			Versions: []domain.SecretVersion{
				{
					SecretValueEnc: us.SecretValueEnc,
					Version:        1,
					UpdatedAt:      time.Now().Format(time.RFC3339),
				},
			},
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

// Map backend secrets to UI secrets (using current version)
func (s *SecretsService) ToUISecrets(entries []domain.Secret) []domain.SecretView {
	var uiSecrets []domain.SecretView
	for _, e := range entries {
		currentVersion := e.GetCurrentVersion()
		if currentVersion != nil {
			uiSecrets = append(uiSecrets, domain.SecretView{
				SecretName:     e.SecretName,
				Type:           string(e.Type),
				SecretValueEnc: currentVersion.SecretValueEnc,
				Revealed:       false,
			})
		}
	}
	return uiSecrets
}
