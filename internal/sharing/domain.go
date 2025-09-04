package sharing

import "go-password-manager/internal/domain"

type SecretExportPayload struct {
	ID                    string         // Unique identifier for the bundle
	Name                  string         // Human-readable name for the bundle
	EncryptedSecrets      []byte         // The encrypted secrets data
	EncryptedSymmetricKey []byte         // Symmetric key, encrypted with recipient's public key
	SecretsNonce          []byte         // Nonce used for encryption of the secrets
	KeyNonce              []byte         // Nonce used for encryption of the shared key
	Timestamp             int64          // Unix timestamp when bundle was created
	ExpiresAt             int64          // Unix timestamp when bundle expires
	SenderInfo            SenderMetadata // Info about the sender (optional)
	EphemeralPublicKey    []byte         // Sender's ephemeral public key
}

// SecretExportBundle represents the encrypted package to be shared.
type SecretExportBundle struct {
	Payload   SecretExportPayload // Main bundle payload
	Signature []byte              // Digital signature for integrity (optional)
}

// SenderMetadata contains info about the sender for verification/audit.
type SenderMetadata struct {
	DeviceName       string
	UserID           string
	PublicKey        []byte
	SigningPublicKey []byte
}

// SharingSession represents a sharing session between two devices.
type SharingSession struct {
	SessionID          string
	SenderPublicKey    []byte
	RecipientPublicKey []byte
	Status             SharingSessionStatus // e.g., "pending", "confirmed", "completed"
	StartedAt          int64                // Unix timestamp
}

// SecretImportResult is the result of importing a shared bundle.
type SecretImportResult struct {
	ImportedSecretsCount int
	VaultName            string
	Success              bool
	Error                error
}

// TransferMetadata contains protocol-level info (optional).
type TransferMetadata struct {
	Transport           string // e.g., "bluetooth", "local_network"
	ExpiresAt           int64
	PassphraseProtected bool
}

type SharingSessionStatus string

const (
	SessionStatusPending   SharingSessionStatus = "pending"
	SessionStatusConfirmed SharingSessionStatus = "confirmed"
	SessionStatusCompleted SharingSessionStatus = "completed"
	SessionStatusExpired   SharingSessionStatus = "expired"
	SessionStatusFailed    SharingSessionStatus = "failed"
)

// SharingLogEntry records an import or export event for auditing.
type SharingLogEntry struct {
	Timestamp int64          // Unix timestamp of the event
	Action    string         // "import" or "export"
	BundleID  string         // Unique identifier for the shared bundle (optional)
	FileName  string         // Name of the file involved (optional)
	PeerInfo  SenderMetadata // Info about the sender/recipient
	Status    string         // "success", "failed", "expired"
	Details   string         // Optional error message or notes
}

type ExportSecret struct {
	Name      string
	Type      domain.SecretType
	Value     string // decrypted
	UpdatedAt string
	Version   int
	// Add other metadata as needed (UpdatedBy, etc.)
}
