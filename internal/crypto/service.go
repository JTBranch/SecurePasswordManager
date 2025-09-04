package crypto

import (
	"crypto/ed25519"
	"errors"
	"go-password-manager/internal/config/devicekeys"
	"go-password-manager/internal/logger"
)

// PEMProvider defines methods for PEM encoding/decoding X25519 keys
type PEMProvider interface {
	EncodeKeyToPEM(key []byte, keyType devicekeys.KeyType) ([]byte, error)
	DecodeKeyFromPEM(pemBytes []byte, keyType devicekeys.KeyType) ([]byte, error)
}

// CryptoService handles encryption and decryption operations.
type CryptoService struct {
	key         []byte
	PemProvider PEMProvider
}

type ConfigProvider interface {
	GetKeyUUID() string
}

// NewCryptoService creates a new CryptoService.
func NewCryptoService(configProvider ConfigProvider, pemProvider PEMProvider) (*CryptoService, error) {
	key, err := LoadOrCreateKey(configProvider)
	if err != nil {
		return nil, err
	}
	return &CryptoService{key: key, PemProvider: pemProvider}, nil
}

func NewCryptoServiceDefault(configProvider ConfigProvider) (*CryptoService, error) {
	return NewCryptoService(configProvider, &PemUtils{})
}

// GetKey returns the encryption key.
func (s *CryptoService) GetKey() []byte {
	return s.key
}

// Encrypt encrypts data symmetrically and returns base64-encoded ciphertext as bytes.
func (s *CryptoService) EncryptSymmetric(data, key []byte) ([]byte, []byte, error) {
	ciphertext, nonce, err := EncryptSymmetric(data, key)
	if err != nil {
		return nil, nil, err
	}
	return ciphertext, nonce, nil
}

// Decrypt implements the service.CryptoService interface
func (s *CryptoService) DecryptSymmetric(ciphertext, nonce, key []byte) ([]byte, error) {
	return DecryptSymmetric(ciphertext, nonce, key)
}

// GenerateX25519KeyPairPEM generates a new X25519 key pair and returns PEM-encoded keys
func (s *CryptoService) GenerateX25519KeyPairPEM() (pubPEM, privPEM []byte, err error) {
	pubKey, privKey, err := GenerateX25519KeyPair()
	if err != nil {
		return nil, nil, err
	}
	pubPEM, err = s.PemProvider.EncodeKeyToPEM(pubKey, devicekeys.KeyTypeX25519Public)
	if err != nil {
		return nil, nil, err
	}
	privPEM, err = s.PemProvider.EncodeKeyToPEM(privKey, devicekeys.KeyTypeX25519Private)
	if err != nil {
		return nil, nil, err
	}
	return pubPEM, privPEM, nil
}

// EncryptAsymmetricPEM encrypts data using a PEM-encoded public key
func (s *CryptoService) EncryptAsymmetric(plaintext []byte, pubPEM []byte) ([]byte, error) {
	pubKey, err := s.PemProvider.DecodeKeyFromPEM(pubPEM, devicekeys.KeyTypeX25519Public)
	if err != nil {
		return nil, err
	}
	result, err := EncryptAsymmetric(plaintext, pubKey)
	if err != nil {
		return nil, err
	}
	// If you need nonce, ephemeral keys, etc., use result.{Field}

	return result.Ciphertext, nil
}

// EncryptAsymmetricFull encrypts data using a PEM-encoded public key and returns the full result object
func (s *CryptoService) EncryptAsymmetricFull(plaintext []byte, pubPEM []byte) (AsymmetricEncryptResult, error) {
	pubKey, err := s.PemProvider.DecodeKeyFromPEM(pubPEM, devicekeys.KeyTypeX25519Public)
	if err != nil {
		return AsymmetricEncryptResult{}, err
	}
	return EncryptAsymmetric(plaintext, pubKey)
}

// DecryptAsymmetricPEM decrypts data using a PEM-encoded private key
func (s *CryptoService) DecryptAsymmetric(ciphertext, nonce, privPEM []byte) ([]byte, error) {
	privKey, err := s.PemProvider.DecodeKeyFromPEM(privPEM, devicekeys.KeyTypeX25519Private)
	if err != nil {
		return nil, err
	}
	return DecryptAsymmetric(ciphertext, nonce, privKey)
}

func (s *CryptoService) GenerateKey() ([]byte, error) {
	return CreateSymmetricKey(32)
}

func (s *CryptoService) ECDH(privateKey []byte, publicKey []byte) ([]byte, error) {
	return nil, nil
}

func (c *CryptoService) SignEd25519(data []byte, privKey []byte) ([]byte, error) {
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid Ed25519 private key length")
	}
	return ed25519.Sign(privKey, data), nil
}

func (c *CryptoService) VerifyEd25519(data []byte, sig []byte, pubKey []byte) (bool, error) {
	if len(pubKey) != ed25519.PublicKeySize {
		return false, errors.New("invalid Ed25519 public key length")
	}
	return ed25519.Verify(pubKey, data, sig), nil
}

func (c *CryptoService) GenerateX25519KeyPair() (publicKey []byte, privateKey []byte, err error) {
	logger.Debug("Generating X25519 key pair")
	return GenerateX25519KeyPair()
}

func (c *CryptoService) GenerateEd25519KeyPair() (publicKey []byte, privateKey []byte, err error) {
	logger.Debug("Generating X25519 key pair")
	return GenerateEd25519KeyPair()
}

// DeriveSymmetricKey logs and calls the HKDF key derivation.
func (c *CryptoService) DeriveSymmetricKey(sharedSecret, salt, info []byte) ([]byte, error) {
	logger.Info("Deriving symmetric key with HKDF")
	return DeriveSymmetricKeyHKDF(sharedSecret, salt, info)
}
