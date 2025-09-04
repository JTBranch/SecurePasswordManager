package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"go-password-manager/internal/logger"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// AsymmetricEncryptResult holds the result of asymmetric encryption
type AsymmetricEncryptResult struct {
	Ciphertext         []byte
	Nonce              []byte
	EphemeralPublicKey []byte
}

var encryptionKey []byte

// Encrypt encrypts the plaintext using the provided key.
func EncryptSymmetric(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	encryptionKey = key
	return ciphertext, nonce, nil
}

// Decrypt decrypts the ciphertext using the provided key.
func DecryptSymmetric(ciphertext, nonce, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("nonce key set as %x", nonce))

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func EncryptAsymmetric(plaintext []byte, pubKey []byte) (AsymmetricEncryptResult, error) {
	var result AsymmetricEncryptResult
	ephemeralPriv := make([]byte, curve25519.ScalarSize)
	_, err := rand.Read(ephemeralPriv)
	if err != nil {
		return AsymmetricEncryptResult{}, fmt.Errorf("failed to generate ephemeral private key: %w", err)
	}
	ephemeralPub, err := curve25519.X25519(ephemeralPriv, curve25519.Basepoint)
	if err != nil {
		return AsymmetricEncryptResult{}, fmt.Errorf("failed to generate ephemeral public key: %w", err)
	}
	sharedSecret, err := curve25519.X25519(ephemeralPriv, pubKey)
	if err != nil {
		return AsymmetricEncryptResult{}, fmt.Errorf("failed to derive shared secret: %w", err)
	}
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return AsymmetricEncryptResult{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return AsymmetricEncryptResult{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return AsymmetricEncryptResult{}, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	result.Ciphertext = ciphertext
	result.Nonce = nonce
	result.EphemeralPublicKey = ephemeralPub
	return result, nil
}

func DecryptAsymmetric(ciphertext, nonce, privKey []byte) ([]byte, error) {
	if len(ciphertext) < curve25519.PointSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	ephemeralPub := ciphertext[:curve25519.PointSize]
	encrypted := ciphertext[curve25519.PointSize:]
	// Derive shared secret
	sharedSecret, err := curve25519.X25519(privKey, ephemeralPub)
	if err != nil {
		return nil, fmt.Errorf("failed to derive shared secret: %w", err)
	}
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// GenerateX25519KeyPair generates a new X25519 key pair
func GenerateX25519KeyPair() (publicKey []byte, privateKey []byte, err error) {
	privateKey = make([]byte, curve25519.ScalarSize)
	_, err = rand.Read(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate public key: %w", err)
	}
	return publicKey, privateKey, nil
}

func GenerateEd25519KeyPair() (publicKey []byte, privateKey []byte, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}
	return pub, priv, nil
}

// DeriveSymmetricKeyHKDF derives a 32-byte symmetric key using HKDF-SHA256 from the shared secret, salt, and info.
func DeriveSymmetricKeyHKDF(sharedSecret, salt, info []byte) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, sharedSecret, salt, info)
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}
	return key, nil
}
