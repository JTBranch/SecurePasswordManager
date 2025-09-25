//go:build !windows
// +build !windows

package bluetooth

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"time"

	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
)

// Config holds bluetooth transport runtime options.
type Config struct {
	// AdapterName optionally selects a platform adapter ("default" uses system default).
	AdapterName string
	// ServiceUUID is the GATT service UUID to advertise/scan.
	ServiceUUID string
	// MTU is the GATT MTU for chunking (0=use platform default).
	MTU int
	// ConnectTimeout for dialing a BLE peripheral.
	ConnectTimeout time.Duration
}

// Transport implements transport.BundleTransport over Bluetooth (skeleton).
type Transport struct {
	cfg     Config
	local   transport.DeviceDescriptor
	deps    transport.Dependencies
	adapter Adapter

	// internal fields for adapter/connection state will be added in implementation
}

// Ensure interface compliance at compile time.
var _ transport.BundleTransport = (*Transport)(nil)
var _ transport.DiscoverableTransport = (*Transport)(nil)

// Specific bluetooth transport errors for clearer diagnostics
var (
	errDiscoveryNotSupported = errors.New("bluetooth: discovery not supported by adapter")
)

func (t *Transport) ID() string { return "bluetooth" }

func (t *Transport) Send(ctx context.Context, bundle *sharing.SecretExportBundle, target transport.DeviceDescriptor) (*transport.TransportReceipt, error) {
	if bundle == nil {
		return nil, errors.New("bluetooth: nil bundle")
	}
	return t.sendUsingAdapter(ctx, t.adapter, bundle, target)
}

func (t *Transport) sendUsingAdapter(ctx context.Context, adapt Adapter, bundle *sharing.SecretExportBundle, target transport.DeviceDescriptor) (*transport.TransportReceipt, error) {
	started := time.Now()
	conn, err := adapt.ConnectToDevice(ctx, target.DeviceID)
	if err != nil {
		return &transport.TransportReceipt{TransportID: t.ID(), Target: target, BundleID: bundle.Payload.ID, StartedAt: started, Attempt: 1, Error: err}, err
	}
	defer conn.Close()

	body, mErr := json.Marshal(bundle)
	if mErr != nil {
		return &transport.TransportReceipt{TransportID: t.ID(), Target: target, BundleID: bundle.Payload.ID, StartedAt: started, Attempt: 1, Error: mErr}, mErr
	}
	hdr := make([]byte, 4)
	binary.BigEndian.PutUint32(hdr, uint32(len(body)))
	// Attempt write with a couple of retries for transient adapter errors.
	maxAttempts := 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if _, err := conn.Write(append(hdr, body...)); err != nil {
			lastErr = err
			// small backoff
			time.Sleep(time.Duration(attempt*10) * time.Millisecond)
			continue
		}
		// success
		return &transport.TransportReceipt{TransportID: t.ID(), Target: target, BundleID: bundle.Payload.ID, StartedAt: started, CompletedAt: time.Now(), Attempt: attempt}, nil
	}
	return &transport.TransportReceipt{TransportID: t.ID(), Target: target, BundleID: bundle.Payload.ID, StartedAt: started, Attempt: maxAttempts, Error: lastErr}, lastErr
}

func (t *Transport) Receive(ctx context.Context) (*transport.InboundBundle, error) {
	return t.receiveUsingAdapter(ctx, t.adapter)
}

func (t *Transport) receiveUsingAdapter(ctx context.Context, adapt Adapter) (*transport.InboundBundle, error) {
	conn, err := adapt.ConnectToDevice(ctx, t.local.DeviceID)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	hdr := make([]byte, 4)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return nil, err
	}
	ln := binary.BigEndian.Uint32(hdr)
	if ln == 0 {
		return nil, errors.New("bluetooth: zero-length frame")
	}
	body := make([]byte, ln)
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, err
	}
	var bundle sharing.SecretExportBundle
	if err := json.Unmarshal(body, &bundle); err != nil {
		return nil, err
	}
	ib := &transport.InboundBundle{From: transport.DeviceDescriptor{}, Bundle: &bundle, ReceivedAt: time.Now(), Envelope: map[string]any{"bluetooth": true}}
	return ib, nil
}

func (t *Transport) Close() error {
	return nil
}

func (t *Transport) Discover(ctx context.Context, limit int) ([]transport.DeviceDescriptor, error) {
	// discovery is optional on adapters; use type assertion to see if the adapter supports scanning
	type scanner interface {
		Scan(serviceUUID string, limit int) []string
	}
	if s, ok := t.adapter.(scanner); ok {
		ids := s.Scan(t.cfg.ServiceUUID, limit)
		out := make([]transport.DeviceDescriptor, 0, len(ids))
		for _, id := range ids {
			out = append(out, transport.DeviceDescriptor{DeviceID: id, DeviceName: id})
		}
		return out, nil
	}
	return nil, errDiscoveryNotSupported
}

// factory integrates with transport.Register
type factory struct{}

func (factory) New(ctx context.Context, cfg map[string]any, local transport.DeviceDescriptor, deps transport.Dependencies) (transport.BundleTransport, error) {
	// parse cfg into Config lazily in implementation
	// Use the same service UUID that the adapters advertise so discovery
	// looks for the correct GATT service. The adapters advertise
	// 00000000-0000-0000-0000-000000000001.
	c := Config{ServiceUUID: "00000000-0000-0000-0000-000000000001", MTU: 0, ConnectTimeout: 5 * time.Second}
	// adapter is mandatory for bluetooth transport
	if deps.BluetoothAdapter == nil {
		return nil, errors.New("bluetooth: adapter dependency required")
	}
	adapt, ok := deps.BluetoothAdapter.(Adapter)
	if !ok {
		return nil, errors.New("bluetooth: adapter dependency has wrong type")
	}
	t := &Transport{cfg: c, local: local, deps: deps, adapter: adapt}
	return t, nil
}

func init() {
	transport.Register("bluetooth", factory{})
}
