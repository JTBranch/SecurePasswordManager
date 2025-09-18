package service

import (
	"context"
	"errors"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	"time"
)

// ExportFlowState represents coarse export preparation steps.
type ExportFlowState string

const (
	ExportStatePreparing ExportFlowState = "preparing"
	ExportStateReady     ExportFlowState = "ready"
	ExportStateFailed    ExportFlowState = "failed"
	ExportStateCanceled  ExportFlowState = "canceled"
)

// ExportProgress conveys export preparation updates.
type ExportProgress struct {
	State       ExportFlowState
	BundleID    string
	Error       error
	StartedAt   time.Time
	CompletedAt time.Time
}

// ShareFlowState represents send lifecycle states.
type ShareFlowState string

const (
	ShareFlowQueued     ShareFlowState = "queued"
	ShareFlowConnecting ShareFlowState = "connecting"
	ShareFlowSending    ShareFlowState = "sending"
	ShareFlowWaitingAck ShareFlowState = "waiting_ack"
	ShareFlowSucceeded  ShareFlowState = "succeeded"
	ShareFlowFailed     ShareFlowState = "failed"
	ShareFlowCanceled   ShareFlowState = "canceled"
)

// TransferSendProgress reports send progress (distinct from legacy facade ShareProgress).
type TransferSendProgress struct {
	State       ShareFlowState
	BundleID    string
	TargetID    string
	Attempt     int
	Error       error
	StartedAt   time.Time
	CompletedAt time.Time
}

// InboundEventType enumerates inbound/import events.
type InboundEventType string

const (
	InboundBundleReceived  InboundEventType = "bundle_received"
	InboundImportSucceeded InboundEventType = "import_succeeded"
	InboundImportFailed    InboundEventType = "import_failed"
)

// InboundEvent communicates receive/import pipeline milestones.
type InboundEvent struct {
	Type      InboundEventType
	BundleID  string
	FromID    string
	Error     error
	Timestamp time.Time
}

// ExportProvider minimal dependency for preparing an export bundle.
type ExportProvider interface {
	// Export produces a signed/encrypted bundle ready for transport.
	Export(secrets []sharing.ExportSecret, recipientPubKey []byte, expiryMinutes int, meta sharing.SenderMetadata) (*sharing.SecretExportBundle, error)
}

// ImportProvider minimal dependency to import a received bundle.
type ImportProvider interface {
	ImportSecrets(bundle *sharing.SecretExportBundle, recipientPrivateKey []byte, recipientEphemeralPubKey []byte) (*sharing.SecretImportResult, error)
}

// DiscoverySession lists devices (externally managed / started).
type DiscoverySession interface {
	Devices() []transport.DeviceDescriptor
}

// TransportBuilder builds a transport by ID.
type TransportBuilder interface {
	Build(ctx context.Context, id string, cfg map[string]any, local transport.DeviceDescriptor, deps transport.Dependencies) (transport.BundleTransport, error)
}

// TransportBuilderFunc adapter.
type TransportBuilderFunc func(ctx context.Context, id string, cfg map[string]any, local transport.DeviceDescriptor, deps transport.Dependencies) (transport.BundleTransport, error)

func (f TransportBuilderFunc) Build(ctx context.Context, id string, cfg map[string]any, local transport.DeviceDescriptor, deps transport.Dependencies) (transport.BundleTransport, error) {
	return f(ctx, id, cfg, local, deps)
}

// SharingTransferService is a stateless orchestrator for export/send/receive.
// All long-running operations are controlled by contexts; no internal retained mutable state beyond its injected dependencies.
type SharingTransferService struct {
	local    transport.DeviceDescriptor
	deps     transport.Dependencies
	export   ExportProvider
	importp  ImportProvider
	discover DiscoverySession
	builder  TransportBuilder
}

// NewSharingTransferService constructs a stateless transfer service.
// Any dependency can be nil if the related feature is unused by the caller.
func NewSharingTransferService(local transport.DeviceDescriptor, deps transport.Dependencies, exportProv ExportProvider, importProv ImportProvider, discovery DiscoverySession, builder TransportBuilder) *SharingTransferService {
	if builder == nil {
		builder = TransportBuilderFunc(transport.Build)
	}
	return &SharingTransferService{local: local, deps: deps, export: exportProv, importp: importProv, discover: discovery, builder: builder}
}

// ListDevices returns currently discovered devices (empty if no discovery session injected).
func (s *SharingTransferService) ListDevices() []transport.DeviceDescriptor {
	if s.discover == nil {
		return nil
	}
	return s.discover.Devices()
}

// PrepareExport produces a bundle and emits coarse progress. Cancellation stops after current step.
func (s *SharingTransferService) PrepareExport(ctx context.Context, secrets []sharing.ExportSecret, expiryMinutes int, meta sharing.SenderMetadata) (*sharing.SecretExportBundle, <-chan ExportProgress, error) {
	ch := make(chan ExportProgress, 3)
	start := time.Now()
	if s.export == nil {
		close(ch)
		return nil, ch, errors.New("export provider not configured")
	}
	bundleID := "" // unknown until provider completes (if it sets ID internally)
	go func() {
		defer close(ch)
		ch <- ExportProgress{State: ExportStatePreparing, StartedAt: start}
		select {
		case <-ctx.Done():
			ch <- ExportProgress{State: ExportStateCanceled, StartedAt: start, CompletedAt: time.Now(), Error: ctx.Err()}
			return
		default:
		}
		b, err := s.export.Export(secrets, nil, expiryMinutes, meta)
		if err != nil {
			ch <- ExportProgress{State: ExportStateFailed, Error: err, StartedAt: start, CompletedAt: time.Now()}
			return
		}
		if b != nil {
			bundleID = b.Payload.ID
		}
		ch <- ExportProgress{State: ExportStateReady, BundleID: bundleID, StartedAt: start, CompletedAt: time.Now()}
	}()
	return nil, ch, nil // returning nil bundle here keeps flow simple; caller reads final event
}

// SendBundle transmits a prepared bundle; emits progress until completion or cancellation.
func (s *SharingTransferService) SendBundle(ctx context.Context, transportID string, bundle *sharing.SecretExportBundle, target transport.DeviceDescriptor) (<-chan TransferSendProgress, error) {
	ch := make(chan TransferSendProgress, 6)
	if bundle == nil {
		close(ch)
		return ch, errors.New("nil bundle")
	}
	if target.LastAddr == "" {
		close(ch)
		return ch, errors.New("target missing address")
	}
	if s.builder == nil {
		close(ch)
		return ch, errors.New("transport builder missing")
	}
	go func() {
		defer close(ch)
		start := time.Now()
		ch <- TransferSendProgress{State: ShareFlowQueued, BundleID: bundle.Payload.ID, TargetID: target.DeviceID, StartedAt: start}
		tr, err := s.builder.Build(ctx, transportID, map[string]any{"listen_addr": ":0", "discovery": false}, s.local, s.deps)
		if err != nil {
			ch <- TransferSendProgress{State: ShareFlowFailed, BundleID: bundle.Payload.ID, TargetID: target.DeviceID, Error: err, StartedAt: start, CompletedAt: time.Now()}
			return
		}
		defer tr.Close()
		receipt, err := tr.Send(ctx, bundle, target)
		if err != nil {
			ch <- TransferSendProgress{State: ShareFlowFailed, BundleID: bundle.Payload.ID, TargetID: target.DeviceID, Error: err, Attempt: receipt.Attempt, StartedAt: receipt.StartedAt, CompletedAt: time.Now()}
			return
		}
		ch <- TransferSendProgress{State: ShareFlowSucceeded, BundleID: bundle.Payload.ID, TargetID: target.DeviceID, Attempt: receipt.Attempt, StartedAt: receipt.StartedAt, CompletedAt: receipt.CompletedAt}
	}()
	return ch, nil
}

// ReceiveOnce waits for a single inbound bundle then (optionally) imports it.
// If autoImport is true and an ImportProvider is configured, import results are appended as events on the returned channel.
func (s *SharingTransferService) ReceiveOnce(ctx context.Context, transportID string, autoImport bool) (<-chan InboundEvent, error) {
	evCh := make(chan InboundEvent, 2)
	go func() {
		defer close(evCh)
		tr, err := s.builder.Build(ctx, transportID, map[string]any{"listen_addr": ":0", "discovery": false}, s.local, s.deps)
		if err != nil {
			evCh <- InboundEvent{Type: InboundImportFailed, Error: err, Timestamp: time.Now()}
			return
		}
		defer tr.Close()
		ib, err := tr.Receive(ctx)
		if err != nil || ib == nil {
			if err != nil {
				evCh <- InboundEvent{Type: InboundImportFailed, Error: err, Timestamp: time.Now()}
			}
			return
		}
		evCh <- InboundEvent{Type: InboundBundleReceived, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Timestamp: time.Now()}
		if autoImport && s.importp != nil {
			// Real key inputs to ImportSecrets omitted (placeholder nils) - replace with actual key management when integrated.
			if _, impErr := s.importp.ImportSecrets(ib.Bundle, nil, nil); impErr != nil {
				evCh <- InboundEvent{Type: InboundImportFailed, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Error: impErr, Timestamp: time.Now()}
			} else {
				evCh <- InboundEvent{Type: InboundImportSucceeded, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Timestamp: time.Now()}
			}
		}
	}()
	return evCh, nil
}

// ReceiveOnceAt is like ReceiveOnce but allows specifying a concrete listen address (useful for tests) instead of :0.
func (s *SharingTransferService) ReceiveOnceAt(ctx context.Context, transportID string, autoImport bool, listenAddr string) (<-chan InboundEvent, error) {
	if listenAddr == "" {
		listenAddr = ":0"
	}
	evCh := make(chan InboundEvent, 3)
	go func() {
		defer close(evCh)
		tr, err := s.builder.Build(ctx, transportID, map[string]any{"listen_addr": listenAddr, "discovery": false}, s.local, s.deps)
		if err != nil {
			evCh <- InboundEvent{Type: InboundImportFailed, Error: err, Timestamp: time.Now()}
			return
		}
		defer tr.Close()
		ib, err := tr.Receive(ctx)
		if err != nil {
			evCh <- InboundEvent{Type: InboundImportFailed, Error: err, Timestamp: time.Now()}
			return
		}
		if ib == nil {
			return
		}
		evCh <- InboundEvent{Type: InboundBundleReceived, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Timestamp: time.Now()}
		if !autoImport || s.importp == nil {
			return
		}
		if _, impErr := s.importp.ImportSecrets(ib.Bundle, nil, nil); impErr != nil {
			evCh <- InboundEvent{Type: InboundImportFailed, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Error: impErr, Timestamp: time.Now()}
			return
		}
		evCh <- InboundEvent{Type: InboundImportSucceeded, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Timestamp: time.Now()}
	}()
	return evCh, nil
}

// ReceiveOnceWithKeysAt extends ReceiveOnceAt allowing caller to supply recipient private/public keys for a real import.
func (s *SharingTransferService) ReceiveOnceWithKeysAt(ctx context.Context, transportID string, autoImport bool, listenAddr string, recipientPriv, recipientPub []byte) (<-chan InboundEvent, error) {
	if listenAddr == "" {
		listenAddr = ":0"
	}
	evCh := make(chan InboundEvent, 3)
	go func() {
		defer close(evCh)
		tr, err := s.builder.Build(ctx, transportID, map[string]any{"listen_addr": listenAddr, "discovery": false}, s.local, s.deps)
		if err != nil {
			evCh <- InboundEvent{Type: InboundImportFailed, Error: err, Timestamp: time.Now()}
			return
		}
		defer tr.Close()
		ib, err := tr.Receive(ctx)
		if err != nil {
			evCh <- InboundEvent{Type: InboundImportFailed, Error: err, Timestamp: time.Now()}
			return
		}
		if ib == nil {
			return
		}
		evCh <- InboundEvent{Type: InboundBundleReceived, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Timestamp: time.Now()}
		if !autoImport || s.importp == nil {
			return
		}
		if _, impErr := s.importp.ImportSecrets(ib.Bundle, recipientPriv, recipientPub); impErr != nil {
			evCh <- InboundEvent{Type: InboundImportFailed, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Error: impErr, Timestamp: time.Now()}
			return
		}
		evCh <- InboundEvent{Type: InboundImportSucceeded, BundleID: ib.Bundle.Payload.ID, FromID: ib.From.DeviceID, Timestamp: time.Now()}
	}()
	return evCh, nil
}
