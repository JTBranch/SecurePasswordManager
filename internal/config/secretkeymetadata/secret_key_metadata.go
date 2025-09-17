package secretkeymetadata

import (
	"encoding/json"
	"go-password-manager/internal/logger"
	"os"
	"time"
)

type SecretsKeyMetadataBundle struct {
	CurrentKey *SecretsEncryptionKeyMetadata  `json:"current_key"`
	Archived   []SecretsEncryptionKeyMetadata `json:"archived"`
}

type SecretsEncryptionKeyMetadata struct {
	UUID         string    `json:"uuid"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
}

type BuildConfigProvider interface {
	GetSecretKeyMetadataFilePath() string
}

type SecretKeyMetadataFileService struct {
	secretsKeyMetadataBundle SecretsKeyMetadataBundle
	filePath                 string
}

func NewSecretKeyMetadataFileService(buildCfg BuildConfigProvider) (*SecretKeyMetadataFileService, error) {

	filePath := buildCfg.GetSecretKeyMetadataFilePath()
	metadata, err := getSecretKeyMetadataFromFile(filePath)

	if err != nil {
		logger.Error("Error getting the secret key metadata from file", err)
		return nil, err
	}

	return &SecretKeyMetadataFileService{
		secretsKeyMetadataBundle: *metadata,
		filePath:                 filePath,
	}, nil

}

func getSecretKeyMetadataFromFile(filePath string) (*SecretsKeyMetadataBundle, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config and file
			defaultMetadata := &SecretsKeyMetadataBundle{
				CurrentKey: nil,
				Archived:   []SecretsEncryptionKeyMetadata{},
			}
			jsonData, marshalErr := json.MarshalIndent(defaultMetadata, "", "  ")
			if marshalErr != nil {
				return nil, marshalErr
			}
			writeErr := os.WriteFile(filePath, jsonData, 0600)
			if writeErr != nil {
				return nil, writeErr
			}
			return defaultMetadata, nil
		}
		return nil, err
	}
	var config SecretsKeyMetadataBundle
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (svc *SecretKeyMetadataFileService) saveMetadataToFile() error {
	jsonData, marshalErr := json.MarshalIndent(svc.secretsKeyMetadataBundle, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}
	writeErr := os.WriteFile(svc.filePath, jsonData, 0600)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func (svc *SecretKeyMetadataFileService) GetSecretKeyMetadata() *SecretsKeyMetadataBundle {
	return &svc.secretsKeyMetadataBundle
}

func (svc *SecretKeyMetadataFileService) UpdateCurrentKey(newKey SecretsEncryptionKeyMetadata) *SecretsKeyMetadataBundle {
	if svc.secretsKeyMetadataBundle.CurrentKey != nil {
		svc.secretsKeyMetadataBundle.Archived = append(
			svc.secretsKeyMetadataBundle.Archived,
			*svc.secretsKeyMetadataBundle.CurrentKey,
		)
	}
	svc.secretsKeyMetadataBundle.CurrentKey = &newKey
	svc.saveMetadataToFile()
	return &svc.secretsKeyMetadataBundle
}
