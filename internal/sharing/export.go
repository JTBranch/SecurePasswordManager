package sharing

import (
	"encoding/json"
	"fmt"
	"go-password-manager/internal/crypto"
	"time"

	"github.com/google/uuid"
)

type CryptoProvider interface {
	EncryptAsymmetricFull(plaintext []byte, pubKey []byte) (crypto.AsymmetricEncryptResult, error)
	DecryptAsymmetric(ciphertext, nonce, privKey []byte) ([]byte, error)
	EncryptSymmetric(plaintext []byte, key []byte) ([]byte, []byte, error)
	DecryptSymmetric(ciphertext, nonce, key []byte) ([]byte, error)
	ECDH(privateKey []byte, publicKey []byte) ([]byte, error)
	SignEd25519(data []byte, privKey []byte) ([]byte, error)
	VerifyEd25519(data []byte, sig []byte, pubKey []byte) (bool, error)
	DeriveSymmetricKey(sharedSecret, senderEphemeralPub, recipientEphemeralPubKey []byte) ([]byte, error)
}

// DeviceKeyManager handles long-term device key pair management
type DeviceKeyProvider interface {
	GetEncryptionDeviceKey() (*crypto.DeviceKey, error)
	GetSigningDeviceKey() (*crypto.DeviceKey, error)
	GetDeviceName() string
	GetAppUser() string
}

type ExportService struct {
	CryptoProvider    CryptoProvider
	DeviceKeyProvider DeviceKeyProvider
}

func NewExportService(cryptoProvider CryptoProvider, deviceKeyProvider DeviceKeyProvider) *ExportService {
	return &ExportService{
		CryptoProvider:    cryptoProvider,
		DeviceKeyProvider: deviceKeyProvider,
	}
}

// / ExportSecrets creates a secure export bundle for the given secrets and recipient.

func (s *ExportService) ExportSecrets(secrets []ExportSecret, recipientEphemeralPubKey []byte, expiryMinutes int, senderInfo SenderMetadata) (*SecretExportBundle, error) {
	// 1. Generate sender ephemeral key pair
	senderEphemeralPub, senderEphemeralPriv, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sender ephemeral key pair: %w", err)
	}

	// 2. Recipient ephemeral public key is provided (recipientEphemeralPubKey)
	// 3. ECDH: derive shared symmetric key
	sharedSecret, err := s.CryptoProvider.ECDH(senderEphemeralPriv, recipientEphemeralPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive shared secret: %w", err)
	}

	derivedKey, err := s.deriveSymmetricKey(sharedSecret, senderEphemeralPub, recipientEphemeralPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive symmetric key: %w", err)
	}
	encryptedSecrets, secretsNonce, err := s.encryptSecrets(secrets, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secrets: %w", err)
	}

	encryptionDeviceKey, err := s.DeviceKeyProvider.GetEncryptionDeviceKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get sender long-term public key: %w", err)
	}

	senderInfo.PublicKey = encryptionDeviceKey.PublicKey
	senderInfo.DeviceName = s.DeviceKeyProvider.GetDeviceName()
	senderInfo.UserID = s.DeviceKeyProvider.GetAppUser()

	encryptedSymmetricKey, keyNonce, err := s.encryptSymmetricKey(derivedKey, recipientEphemeralPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt symmetric key: %w", err)
	}

	// 5. Wipe ephemeral private keys (in Go, just let them go out of scope)

	id := uuid.NewString()
	shortID := id[:8]
	timestamp := time.Now().Unix()
	name := fmt.Sprintf("exported-%d-%s.pem", timestamp, shortID)

	signingDeviceKey, err := s.DeviceKeyProvider.GetSigningDeviceKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get sender long-term public key: %w", err)
	}
	senderInfo.SigningPublicKey = signingDeviceKey.PublicKey

	bundlePayload := &SecretExportPayload{
		ID:                    id,
		Name:                  name,
		EncryptedSecrets:      encryptedSecrets,
		SecretsNonce:          secretsNonce,
		KeyNonce:              keyNonce,
		Timestamp:             timestamp,
		ExpiresAt:             timestamp + int64(expiryMinutes*60),
		SenderInfo:            senderInfo,
		EncryptedSymmetricKey: encryptedSymmetricKey,
		EphemeralPublicKey:    senderEphemeralPub,
	}

	signature, err := s.SignBundle(bundlePayload, signingDeviceKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign export bundle: %w", err)
	}
	return &SecretExportBundle{
		Payload:   *bundlePayload,
		Signature: signature,
	}, nil
}

func (s *ExportService) deriveSymmetricKey(sharedSecret, senderEphemeralPub, recipientEphemeralPubKey []byte) ([]byte, error) {
	// Use senderEphemeralPub as salt, recipientEphemeralPubKey as info
	return s.CryptoProvider.DeriveSymmetricKey(sharedSecret, senderEphemeralPub, recipientEphemeralPubKey)
}

// encryptSecrets encrypts the secrets data with the symmetric key and returns ciphertext and nonce.

func (s *ExportService) encryptSecrets(secrets []ExportSecret, symmetricKey []byte) (ciphertext []byte, nonce []byte, err error) {
	// Serialize secrets to JSON
	secretsData, err := json.Marshal(secrets)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt with symmetric key using cryptoProvider
	encrypted, nonce, err := s.CryptoProvider.EncryptSymmetric(secretsData, symmetricKey)
	if err != nil {
		return nil, nil, err
	}
	return encrypted, nonce, nil
}

// encryptSymmetricKey encrypts the symmetric key with the recipient's public key.

func (s *ExportService) encryptSymmetricKey(derivedKey []byte, recipientEphemeralPubKey []byte) ([]byte, []byte, error) {
	encryptedSymmetricKeyResult, err := s.CryptoProvider.EncryptAsymmetricFull(derivedKey, recipientEphemeralPubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt symmetric key: %w", err)
	}
	encryptedSymmetricKey := append(encryptedSymmetricKeyResult.EphemeralPublicKey, encryptedSymmetricKeyResult.Ciphertext...)
	return encryptedSymmetricKey, encryptedSymmetricKeyResult.Nonce, nil
}

// signBundle signs the export bundle with the sender's private key.

func (s *ExportService) SignBundle(bundle *SecretExportPayload, senderPrivKey []byte) ([]byte, error) {
	data, err := json.Marshal(bundle)
	if err != nil {
		return nil, err
	}
	return s.CryptoProvider.SignEd25519(data, senderPrivKey)
}
