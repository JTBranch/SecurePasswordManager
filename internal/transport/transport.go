package transport

import (
	"context"
	"errors"
	"sync"
	"time"

	"go-password-manager/internal/sharing"
)

// BundleTransport defines the minimal behavior a transport must provide
// to move a sharing.SecretExportBundle from a sender to a recipient.
// Implementations SHOULD be stateless or encapsulate their connection
// management behind the interface. All methods must be safe for use
// by concurrent goroutines.
type BundleTransport interface {
	// ID returns a short stable identifier (e.g. "lan", "qr", "email").
	ID() string
	// Send transmits an exported bundle to the target device. Implementations
	// should perform any required handshake / channel setup implicitly if not
	// already established. A TransportReceipt is returned describing the
	// attempt (even on some errors when partial progress info is useful).
	Send(ctx context.Context, bundle *sharing.SecretExportBundle, target DeviceDescriptor) (*TransportReceipt, error)
	// Receive blocks (until ctx done) waiting for the next inbound bundle
	// directed at the local device, returning its metadata and the bundle.
	// Transports that are inherently push-less (e.g. QR) may return
	// ErrReceiveNotSupported.
	Receive(ctx context.Context) (*InboundBundle, error)
	// Close releases any underlying resources (listeners, file handles, etc.).
	Close() error
}

// DiscoverableTransport optionally supports discovering nearby devices via
// network broadcast / directory mechanisms. Implementations should return
// rapidly with a snapshot (best-effort). Continuous discovery can be modeled
// by polling from the UI layer when needed.
type DiscoverableTransport interface {
	BundleTransport
	Discover(ctx context.Context, limit int) ([]DeviceDescriptor, error)
}

// InboundBundle represents a received bundle plus the sending device metadata.
type InboundBundle struct {
	From   DeviceDescriptor
	Bundle *sharing.SecretExportBundle
	// Raw or transport-specific envelope details (opaque to core logic)
	Envelope   map[string]any
	ReceivedAt time.Time
}

// TransportReceipt captures the outcome of a Send attempt.
type TransportReceipt struct {
	TransportID string
	Target      DeviceDescriptor
	BundleID    string
	StartedAt   time.Time
	CompletedAt time.Time
	Attempt     int
	Error       error
	// Implementation specific diagnostics (latency, retries, etc.)
	Meta map[string]any
}

// Error values returned by transports.
var (
	ErrReceiveNotSupported = errors.New("transport: receive not supported")
)

// Factory allows runtime DI-friendly construction of transports.
type Factory interface {
	New(ctx context.Context, cfg map[string]any, local DeviceDescriptor, deps Dependencies) (BundleTransport, error)
}

// Dependencies aggregates services required by transports (expanded as needed).
type Dependencies struct {
	Crypto   interface{} // loosely typed for DI; concrete aware transports can type assert
	Registry DeviceRegistry
	// Advertisement allows injecting an mDNS advertisement starter for LAN transport testing.
	Advertisement AdvertisementStarter
	// KeyGen provides Ed25519 key generation explicitly (preferred over type assertion on Crypto).
	KeyGen Ed25519KeyGenerator
}

// AdvertisementStarter abstracts starting an mDNS advertisement returning a closable handle.
type AdvertisementStarter interface {
	Start(local *DeviceDescriptor, addr string, port int, txt []string) (AdvertisementHandle, error)
}

// AdvertisementHandle represents an active advertisement that can be stopped.
type AdvertisementHandle interface{ Close() error }

// Ed25519KeyGenerator explicit interface for key generation.
type Ed25519KeyGenerator interface {
	GenerateEd25519KeyPair() (publicKey []byte, privateKey []byte, err error)
}

// registry of transport factories keyed by ID.
var (
	regMu     sync.RWMutex
	factories = map[string]Factory{}
)

// Register makes a transport factory available. Panics on duplicate IDs.
func Register(id string, f Factory) {
	regMu.Lock()
	defer regMu.Unlock()
	if _, exists := factories[id]; exists {
		panic("transport: duplicate factory id: " + id)
	}
	factories[id] = f
}

// Build constructs a transport using a previously registered factory.
func Build(ctx context.Context, id string, cfg map[string]any, local DeviceDescriptor, deps Dependencies) (BundleTransport, error) {
	regMu.RLock()
	f := factories[id]
	regMu.RUnlock()
	if f == nil {
		return nil, errors.New("transport: unknown id: " + id)
	}
	return f.New(ctx, cfg, local, deps)
}
