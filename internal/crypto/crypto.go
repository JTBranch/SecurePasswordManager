package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"go-password-manager/internal/logger"
	"io"

	"golang.org/x/crypto/curve25519"
)

// Legacy asymmetric encryption helpers removed (unused after key box v2).

const (
	keyBoxVersion      byte = 2
	x25519PubSize           = 32
	errSharedSecretFmt      = "failed to derive shared secret: %w"
)

// WrapKeyBox encrypts a symmetric key using X25519 + AES-GCM and returns a single box:
// version(1) | ephemeralPub(32) | nonce(12) | ciphertext(var)
func WrapKeyBox(symKey []byte, recipientPub []byte, aad []byte) ([]byte, error) {
	// Generate ephemeral key pair
	ephPriv := make([]byte, curve25519.ScalarSize)
	if _, err := rand.Read(ephPriv); err != nil {
		return nil, fmt.Errorf("failed to gen ephemeral priv: %w", err)
	}
	ephPub, err := curve25519.X25519(ephPriv, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to gen ephemeral pub: %w", err)
	}
	// Derive shared secret
	shared, err := curve25519.X25519(ephPriv, recipientPub)
	if err != nil {
		return nil, fmt.Errorf(errSharedSecretFmt, err)
	}
	// zero ephemeral priv ASAP
	Zeroize(ephPriv)
	block, err := aes.NewCipher(shared)
	if err != nil {
		Zeroize(shared)
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Zeroize(shared)
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		Zeroize(shared)
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, symKey, aad)
	// wipe shared secret now that ciphertext derived
	Zeroize(shared)
	box := make([]byte, 1+len(ephPub)+len(nonce)+len(ct))
	off := 0
	box[off] = keyBoxVersion
	off++
	copy(box[off:], ephPub)
	off += len(ephPub)
	copy(box[off:], nonce)
	off += len(nonce)
	copy(box[off:], ct)
	return box, nil
}

// UnwrapKeyBox decrypts a box produced by WrapKeyBox using recipient private key.
func UnwrapKeyBox(box []byte, recipientPriv []byte, aad []byte) ([]byte, error) {
	if len(box) < 1+x25519PubSize+12 { // minimal size check
		return nil, fmt.Errorf("key box too short")
	}
	ver := box[0]
	if ver != keyBoxVersion {
		return nil, fmt.Errorf("unsupported key box version %d", ver)
	}
	off := 1
	ephPub := box[off : off+x25519PubSize]
	off += x25519PubSize
	// Derive shared
	shared, err := curve25519.X25519(recipientPriv, ephPub)
	if err != nil {
		return nil, fmt.Errorf(errSharedSecretFmt, err)
	}
	block, err := aes.NewCipher(shared)
	if err != nil {
		Zeroize(shared)
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Zeroize(shared)
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(box) < 1+x25519PubSize+nonceSize+1 {
		Zeroize(shared)
		return nil, fmt.Errorf("key box malformed")
	}
	nonce := box[off : off+nonceSize]
	off += nonceSize
	ct := box[off:]
	pt, err := gcm.Open(nil, nonce, ct, aad)
	// zero shared regardless of success
	Zeroize(shared)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

// Encrypt encrypts the plaintext using the provided key.
func EncryptSymmetric(plaintext []byte, key []byte, aad []byte) ([]byte, []byte, error) {
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

	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)
	return ciphertext, nonce, nil
}

// Decrypt decrypts the ciphertext using the provided key.
func DecryptSymmetric(ciphertext, nonce, key []byte, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("nonce key set as %x", nonce))

	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
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
// Legacy asymmetric + HKDF code removed.
