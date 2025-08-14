package service

import (
	"encoding/json"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"os"
)

// Legacy data structures for migration
type LegacySecret struct {
	SecretName     string            `json:"secretName"`
	SecretValueEnc string            `json:"secretValueEnc"`
	Type           domain.SecretType `json:"type"`
	Version        int               `json:"version"`
	UpdatedAt      string            `json:"updatedAt"`
	UpdatedBy      string            `json:"updatedBy,omitempty"`
}

type LegacySecretsFile struct {
	AppVersion  string         `json:"appVersion"`
	AppUser     string         `json:"appUser"`
	LastUpdated string         `json:"lastUpdated"`
	Secrets     []LegacySecret `json:"secrets"`
}

// MigrateToNewFormat migrates legacy secret files to the new versioned format
func MigrateToNewFormat(filePath string) error {
	logger.Debug("Starting migration to new format for file:", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Debug("File doesn't exist, no migration needed")
		return nil
	}

	// Read the current file
	file, err := os.Open(filePath)
	if err != nil {
		logger.Debug("Error opening file for migration:", err.Error())
		return err
	}
	defer file.Close()

	// Try to decode as legacy format first
	var legacyData LegacySecretsFile
	if err := json.NewDecoder(file).Decode(&legacyData); err != nil {
		logger.Debug("Error decoding legacy format, checking if already new format:", err.Error())

		// Try to decode as new format to see if already migrated
		file.Seek(0, 0)
		var newData domain.SecretsFile
		if err := json.NewDecoder(file).Decode(&newData); err != nil {
			logger.Debug("Error decoding as new format too:", err.Error())
			return err
		}

		// Already in new format
		logger.Debug("File is already in new format, no migration needed")
		return nil
	}

	// Convert to new format
	newData := convertLegacyToNew(legacyData)

	// Create backup of old file
	backupPath := filePath + ".backup"
	if err := os.Rename(filePath, backupPath); err != nil {
		logger.Error("Failed to create backup:", err.Error())
		return err
	}

	// Write new format
	jsonBytes, err := json.MarshalIndent(newData, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal new format:", err.Error())
		// Restore backup
		os.Rename(backupPath, filePath)
		return err
	}

	if err := os.WriteFile(filePath, jsonBytes, 0600); err != nil {
		logger.Error("Failed to write new format:", err.Error())
		// Restore backup
		os.Rename(backupPath, filePath)
		return err
	}

	logger.Debug("Migration completed successfully. Backup saved as:", backupPath)
	return nil
}

func convertLegacyToNew(legacy LegacySecretsFile) domain.SecretsFile {
	// Group legacy secrets by name
	secretGroups := make(map[string][]LegacySecret)

	for _, legacySecret := range legacy.Secrets {
		secretGroups[legacySecret.SecretName] = append(secretGroups[legacySecret.SecretName], legacySecret)
	}

	// Convert to new format
	var newSecrets []domain.Secret

	for secretName, versions := range secretGroups {
		// Find the latest version
		var latestVersion int
		var secretType domain.SecretType

		for _, version := range versions {
			if version.Version > latestVersion {
				latestVersion = version.Version
				secretType = version.Type
			}
		}

		// Create version entries
		var secretVersions []domain.SecretVersion
		for _, version := range versions {
			secretVersions = append(secretVersions, domain.SecretVersion{
				SecretValueEnc: version.SecretValueEnc,
				Version:        version.Version,
				UpdatedAt:      version.UpdatedAt,
				UpdatedBy:      version.UpdatedBy,
			})
		}

		// Create new secret
		newSecret := domain.Secret{
			SecretName:     secretName,
			Type:           secretType,
			CurrentVersion: latestVersion,
			Versions:       secretVersions,
		}

		newSecrets = append(newSecrets, newSecret)
	}

	return domain.SecretsFile{
		AppVersion:  legacy.AppVersion,
		AppUser:     legacy.AppUser,
		LastUpdated: legacy.LastUpdated,
		Secrets:     newSecrets,
	}
}
