package storage

import (
	"encoding/json"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	os "os"
	"time"
)

type FileStorage struct {
	filePath   string
	appVersion string
	appUser    string
}

func NewFileStorage(filePath, appVersion, appUser string) service.StorageService {
	return &FileStorage{
		filePath:   filePath,
		appVersion: appVersion,
		appUser:    appUser,
	}
}

func (fs *FileStorage) ReadSecrets() (domain.SecretsFile, error) {
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return domain.SecretsFile{
			AppVersion: fs.appVersion,
			AppUser:    fs.appUser,
			Secrets:    []domain.Secret{},
		}, nil
	}

	file, err := os.Open(fs.filePath)
	if err != nil {
		return domain.SecretsFile{}, err
	}
	defer file.Close()

	var data domain.SecretsFile
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return domain.SecretsFile{}, err
	}
	return data, nil
}

func (fs *FileStorage) WriteSecrets(data domain.SecretsFile) error {
	data.LastUpdated = time.Now().Format(time.RFC3339)
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filePath, jsonBytes, 0600)
}
