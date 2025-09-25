package service

import "go-password-manager/internal/domain"

// StorageService defines the interface for secret persistence.
type StorageService interface {
	ReadSecrets() (domain.SecretsFile, error)
	WriteSecrets(data domain.SecretsFile) error
}
