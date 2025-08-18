package crypto

import config "go-password-manager/internal/config/runtimeconfig"

// CryptoService handles encryption and decryption operations.
type CryptoService struct {
	key []byte
}

// NewCryptoService creates a new CryptoService.
func NewCryptoService(configService *config.ConfigService) (*CryptoService, error) {
	key, err := LoadOrCreateKey(configService)
	if err != nil {
		return nil, err
	}
	return &CryptoService{key: key}, nil
}

// GetKey returns the encryption key.
func (s *CryptoService) GetKey() []byte {
	return s.key
}

// Encrypt implements the service.CryptoService interface
func (s *CryptoService) Encrypt(data, key []byte) ([]byte, error) {
	// This seems incorrect, the free function Encrypt returns a string.
	// For now, I will just call it and convert.
	// We can refactor this later.
	encryptedString, err := Encrypt(data, key)
	if err != nil {
		return nil, err
	}
	return []byte(encryptedString), nil
}

// Decrypt implements the service.CryptoService interface
func (s *CryptoService) Decrypt(data, key []byte) ([]byte, error) {
	return Decrypt(string(data), key)
}
