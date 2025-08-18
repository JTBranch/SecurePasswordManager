package crypto

// CryptoService handles encryption and decryption operations.
type CryptoService struct {
	key []byte
}

type ConfigProvider interface {
	GetKeyUUID() string
}

// NewCryptoService creates a new CryptoService.
func NewCryptoService(configProvider ConfigProvider) (*CryptoService, error) {
	key, err := LoadOrCreateKey(configProvider)
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
