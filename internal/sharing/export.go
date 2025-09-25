package sharing

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"go-password-manager/internal/config/devicekeys"
	"go-password-manager/internal/crypto"
	"time"

	"github.com/google/uuid"
)

type CryptoProvider interface {
	EncryptSymmetric(plaintext []byte, key []byte, aad []byte) ([]byte, []byte, error)
	DecryptSymmetric(ciphertext, nonce, key []byte, aad []byte) ([]byte, error) // used by import path
	SignEd25519(data []byte, privKey []byte) ([]byte, error)
	VerifyEd25519(data []byte, sig []byte, pubKey []byte) (bool, error)
}

// pemDecoder allows decoding PEM keys when the underlying provider supports it
type pemDecoder interface {
	DecodeKeyFromPEM(key []byte, keyType devicekeys.KeyType) ([]byte, error)
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

func (s *ExportService) ExportSecrets(secrets []ExportSecret, recipientEphemeralPubKey []byte, expirySeconds int, senderInfo SenderMetadata) (*SecretExportBundle, error) {
	// Prepare bundle ID early for AAD
	bundleID := uuid.NewString()
	// 1. Decode recipient public key (PEM) if provider supports it (raw X25519 key needed for key box wrapping)
	recipientPubDecoded := recipientEphemeralPubKey
	if dec, ok := s.CryptoProvider.(pemDecoder); ok {
		decoded, derr := dec.DecodeKeyFromPEM(recipientEphemeralPubKey, devicekeys.KeyTypeX25519Public)
		if derr != nil {
			return nil, fmt.Errorf("failed to decode recipient public key: %w", derr)
		}
		recipientPubDecoded = decoded
	}

	// 2. Generate a fresh 32-byte symmetric key for encrypting the secrets
	symKey := make([]byte, 32)
	if _, err := rand.Read(symKey); err != nil {
		return nil, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	// (Secrets encryption moved after AAD construction to bind same context)

	encryptionDeviceKey, err := s.DeviceKeyProvider.GetEncryptionDeviceKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get sender long-term public key: %w", err)
	}

	senderInfo.PublicKey = encryptionDeviceKey.PublicKey
	senderInfo.DeviceName = s.DeviceKeyProvider.GetDeviceName()
	senderInfo.UserID = s.DeviceKeyProvider.GetAppUser()

	// 4. Obtain signing key early to build AAD (bundleID || 0x00 || signingPub)
	signingDeviceKey, err := s.DeviceKeyProvider.GetSigningDeviceKey()
	if err != nil {
		crypto.Zeroize(symKey)
		return nil, fmt.Errorf("failed to get sender long-term public key: %w", err)
	}
	senderInfo.SigningPublicKey = signingDeviceKey.PublicKey
	// Key box AAD domain: bundleID || 0x00 || signingPub
	keyBoxAAD := buildKeyBoxAAD(bundleID, signingDeviceKey.PublicKey)
	secretsAAD := buildSecretsAAD(bundleID, signingDeviceKey.PublicKey)
	// 5. Encrypt secrets using secretsAAD and then wrap symmetric key with keyBoxAAD
	encryptedSecrets, secretsNonce, err := s.encryptSecrets(secrets, symKey, secretsAAD)
	if err != nil {
		crypto.Zeroize(symKey)
		return nil, fmt.Errorf("failed to encrypt secrets: %w", err)
	}
	// 6. Wrap the symmetric key in a single compact key box (version|ephemeral|nonce|ciphertext) with AAD
	wrappedKeyBox, err := crypto.WrapKeyBox(symKey, recipientPubDecoded, keyBoxAAD)
	if err != nil {
		crypto.Zeroize(symKey)
		return nil, fmt.Errorf("failed to wrap symmetric key: %w", err)
	}
	// Zeroize symmetric key after wrapping
	crypto.Zeroize(symKey)

	// 7. Wipe ephemeral private keys (in Go, just let them go out of scope)

	id := bundleID
	shortID := id[:8]
	timestamp := time.Now().Unix()
	name := fmt.Sprintf("exported-%d-%s.pem", timestamp, shortID)

	// signingDeviceKey already loaded above

	bundlePayload := &SecretExportPayload{ID: id, Name: name, EncryptedSecrets: encryptedSecrets, SecretsNonce: secretsNonce, Timestamp: timestamp, ExpiresAt: timestamp + int64(expirySeconds), SenderInfo: senderInfo, SymmetricKeyBox: wrappedKeyBox}

	signature, err := s.SignBundle(bundlePayload, signingDeviceKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign export bundle: %w", err)
	}
	return &SecretExportBundle{
		Payload:   *bundlePayload,
		Signature: signature,
	}, nil
}

// encryptSecrets encrypts the secrets data with the symmetric key and returns ciphertext and nonce.

func (s *ExportService) encryptSecrets(secrets []ExportSecret, symmetricKey []byte, aad []byte) (ciphertext []byte, nonce []byte, err error) {
	// Serialize secrets to JSON
	secretsData, err := json.Marshal(secrets)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt with symmetric key using cryptoProvider
	encrypted, nonce, err := s.CryptoProvider.EncryptSymmetric(secretsData, symmetricKey, aad)
	if err != nil {
		return nil, nil, err
	}
	return encrypted, nonce, nil
}

// encryptSymmetricKey encrypts the symmetric key with the recipient's public key.

// signBundle signs the export bundle with the sender's private key.

func (s *ExportService) SignBundle(bundle *SecretExportPayload, senderPrivKey []byte) ([]byte, error) {
	data, err := json.Marshal(bundle)
	if err != nil {
		return nil, err
	}
	return s.CryptoProvider.SignEd25519(data, senderPrivKey)
}
