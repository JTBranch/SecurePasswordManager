package crypto

import (
	"crypto/ed25519"
	"errors"
	"go-password-manager/internal/config/devicekeys"
	"go-password-manager/internal/logger"

	"golang.org/x/crypto/curve25519"
)

// PEMProvider defines methods for PEM encoding/decoding X25519 keys
type PEMProvider interface {
	EncodeKeyToPEM(key []byte, keyType devicekeys.KeyType) ([]byte, error)
	DecodeKeyFromPEM(pemBytes []byte, keyType devicekeys.KeyType) ([]byte, error)
}

type SecretsEncryptionKeyProvider interface {
	LoadOrCreateKey() ([]byte, error)
	CreateSymmetricKey(keySize int) ([]byte, error)
}

// CryptoService handles encryption and decryption operations.
type CryptoService struct {
	key                          []byte
	pemProvider                  PEMProvider
	secretsEncryptionKeyProvider SecretsEncryptionKeyProvider
}

type ConfigProvider interface {
	GetKeyUUID() string
}

// NewCryptoService creates a new CryptoService.
func NewCryptoService(configProvider ConfigProvider, pemProvider PEMProvider, secretsEncryptionKeyProvider SecretsEncryptionKeyProvider) (*CryptoService, error) {
	key, err := secretsEncryptionKeyProvider.LoadOrCreateKey()
	if err != nil {
		return nil, err
	}
	return &CryptoService{key: key,
		pemProvider:                  pemProvider,
		secretsEncryptionKeyProvider: secretsEncryptionKeyProvider,
	}, nil
}

func NewCryptoServiceDefault(configProvider ConfigProvider,
	secretsEncryptionKeyProvider SecretsEncryptionKeyProvider) (*CryptoService, error) {
	return NewCryptoService(configProvider, &PemUtils{}, secretsEncryptionKeyProvider)
}

// GetKey returns the encryption key.
func (s *CryptoService) GetKey() []byte {
	return s.key
}

// Encrypt encrypts data symmetrically and returns base64-encoded ciphertext as bytes.
func (s *CryptoService) EncryptSymmetric(data, key []byte, aad []byte) ([]byte, []byte, error) {
	ciphertext, nonce, err := EncryptSymmetric(data, key, aad)
	if err != nil {
		return nil, nil, err
	}
	return ciphertext, nonce, nil
}

// Decrypt implements the service.CryptoService interface
func (s *CryptoService) DecryptSymmetric(ciphertext, nonce, key []byte, aad []byte) ([]byte, error) {
	return DecryptSymmetric(ciphertext, nonce, key, aad)
}

// GenerateX25519KeyPairPEM generates a new X25519 key pair and returns PEM-encoded keys
func (s *CryptoService) GenerateX25519KeyPairPEM() (pubPEM, privPEM []byte, err error) {
	pubKey, privKey, err := GenerateX25519KeyPair()
	if err != nil {
		return nil, nil, err
	}
	pubPEM, err = s.pemProvider.EncodeKeyToPEM(pubKey, devicekeys.KeyTypeX25519Public)
	if err != nil {
		return nil, nil, err
	}
	privPEM, err = s.pemProvider.EncodeKeyToPEM(privKey, devicekeys.KeyTypeX25519Private)
	if err != nil {
		return nil, nil, err
	}
	return pubPEM, privPEM, nil
}

// DecodeKeyFromPEM exposes PEM decoding for consumers needing raw keys
func (s *CryptoService) DecodeKeyFromPEM(pemBytes []byte, keyType devicekeys.KeyType) ([]byte, error) {
	return s.pemProvider.DecodeKeyFromPEM(pemBytes, keyType)
}

func (s *CryptoService) GenerateKey() ([]byte, error) {
	return s.secretsEncryptionKeyProvider.CreateSymmetricKey(32)
}

func (s *CryptoService) ECDH(privateKey []byte, publicKey []byte) ([]byte, error) {
	if len(privateKey) == 64 { // tolerate extended private key, use first 32 bytes
		privateKey = privateKey[:curve25519.ScalarSize]
	}
	if len(privateKey) != curve25519.ScalarSize {
		return nil, errors.New("invalid private key length for X25519 ECDH")
	}
	if len(publicKey) != curve25519.PointSize {
		return nil, errors.New("invalid public key length for X25519 ECDH")
	}
	shared, err := curve25519.X25519(privateKey, publicKey)
	if err != nil {
		return nil, err
	}
	return shared, nil
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
