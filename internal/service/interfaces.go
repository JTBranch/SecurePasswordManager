package service

import "go-password-manager/internal/domain"

// CryptoService defines the interface for crypto operations.
type CryptoService interface {
	GetKey() []byte
	Encrypt(data, key []byte) ([]byte, error)
	Decrypt(data, key []byte) ([]byte, error)
}

// StorageService defines the interface for secret persistence.
type StorageService interface {
	ReadSecrets() (domain.SecretsFile, error)
	WriteSecrets(data domain.SecretsFile) error
}
