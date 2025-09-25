package bluetooth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
)

func TestFactoryRegistration(t *testing.T) {
	// Ensure the transport.Register in init() executed and the factory can be created.
	deps := transport.Dependencies{BluetoothAdapter: newTestAdapter()}
	local := transport.DeviceDescriptor{DeviceID: "test-local"}
	f := factory{}
	tpt, err := f.New(context.Background(), nil, local, deps)
	if err != nil {
		t.Fatalf("factory.New returned error: %v", err)
	}
	if tpt == nil {
		t.Fatalf("factory.New returned nil transport")
	}
}

func TestTransportInterface(t *testing.T) {
	deps := transport.Dependencies{BluetoothAdapter: newTestAdapter()}
	local := transport.DeviceDescriptor{DeviceID: "test-local"}
	f := factory{}
	tpt, err := f.New(context.Background(), nil, local, deps)
	if err != nil {
		t.Fatalf("factory.New returned error: %v", err)
	}

	if tpt.ID() != "bluetooth" {
		t.Fatalf("expected ID bluetooth, got %s", tpt.ID())
	}
}

// --- Test-only in-package mock adapter to avoid import cycles with integration mocks ---
type testConn struct {
	ch      chan []byte
	pending []byte
}

func (c *testConn) Write(b []byte) (int, error) {
	nb := make([]byte, len(b))
	copy(nb, b)
	c.ch <- nb
	return len(b), nil
}

func (c *testConn) Read(b []byte) (int, error) {
	if len(c.pending) > 0 {
		n := copy(b, c.pending)
		c.pending = c.pending[n:]
		return n, nil
	}
	part := <-c.ch
	n := copy(b, part)
	if n < len(part) {
		c.pending = append(c.pending, part[n:]...)
	}
	return n, nil
}

func (c *testConn) Close() error { return nil }

type testAdapter struct {
	peers      map[string]*testConn
	advertised map[string]bool
}

func newTestAdapter() *testAdapter {
	return &testAdapter{peers: map[string]*testConn{}, advertised: map[string]bool{}}
}

func (a *testAdapter) PairDevices(aid, bid string) {
	chA := make(chan []byte, 4)
	chB := make(chan []byte, 4)
	a.peers[aid] = &testConn{ch: chA}
	a.peers[bid] = &testConn{ch: chB}
	// Each peer writes into the other's channel
	// To simulate two endpoints, swap channels on write by wrapping ConnectToDevice
}

func (a *testAdapter) ConnectToDevice(ctx context.Context, deviceID string) (AdapterConn, error) {
	// AdapterConn is local alias to bluetooth.Conn interface; define expected type
	if c, ok := a.peers[deviceID]; ok {
		return c, nil
	}
	return nil, &noDeviceErr{}
}

type noDeviceErr struct{}

func (noDeviceErr) Error() string { return "no such device" }

// Make testAdapter implement an optional scanner used by Discover
func (a *testAdapter) Scan(serviceUUID string, limit int) []string {
	var out []string
	for id := range a.advertised {
		out = append(out, id)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func (a *testAdapter) AdvertiseDevice(id string) { a.advertised[id] = true }
func (a *testAdapter) StopAdvertise(id string)   { delete(a.advertised, id) }

// AdapterConn aliases the Conn interface from adapter.go so types align in tests.
type AdapterConn = Conn

// --- Unit tests for bluetooth transport ---

func TestBluetoothSendReceiveBundle(t *testing.T) {
	mock := newTestAdapter()
	// Pairing: create channels such that writes are visible to the other side
	chA := make(chan []byte, 4)
	chB := make(chan []byte, 4)
	mock.peers["sender"] = &testConn{ch: chB} // sender writes go to receiver chB
	mock.peers["receiver"] = &testConn{ch: chA}

	devA := transport.DeviceDescriptor{DeviceID: "sender", DeviceName: "sender"}
	devB := transport.DeviceDescriptor{DeviceID: "receiver", DeviceName: "receiver"}
	depsA := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}
	depsB := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}

	trA, err := transport.Build(context.Background(), "bluetooth", nil, devA, depsA)
	if err != nil {
		t.Fatalf("build sender: %v", err)
	}
	defer trA.Close()
	trB, err := transport.Build(context.Background(), "bluetooth", nil, devB, depsB)
	if err != nil {
		t.Fatalf("build receiver: %v", err)
	}
	defer trB.Close()

	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "bid", Timestamp: time.Now().Unix()}, Signature: []byte("sig")}
	// Start receiver
	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		ib, err := trB.Receive(ctx)
		if err != nil {
			done <- err
			return
		}
		if ib == nil {
			done <- nil
			return
		}
		if ib.Bundle.Payload.ID != "bid" {
			done <- fmt.Errorf("bad id: %s", ib.Bundle.Payload.ID)
			return
		}
		done <- nil
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	receipt, err := trA.Send(ctx, bundle, devB)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if receipt.BundleID != "bid" {
		t.Fatalf("receipt id mismatch")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("receive path failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for receive")
	}
}

func TestBluetoothSendNilBundle(t *testing.T) {
	mock := newTestAdapter()
	dev := transport.DeviceDescriptor{DeviceID: "d1"}
	deps := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}
	tr, err := transport.Build(context.Background(), "bluetooth", nil, dev, deps)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	defer tr.Close()
	ctx := context.Background()
	_, err = tr.Send(ctx, nil, transport.DeviceDescriptor{DeviceID: "other"})
	if err == nil {
		t.Fatalf("expected error sending nil bundle")
	}
}

func TestBluetoothReceiveZeroLengthFrame(t *testing.T) {
	mock := newTestAdapter()
	// Prepare channels so writing to sender channel is seen by receiver
	chA := make(chan []byte, 4)
	chB := make(chan []byte, 4)
	mock.peers["sender"] = &testConn{ch: chB}
	mock.peers["receiver"] = &testConn{ch: chA}

	dev := transport.DeviceDescriptor{DeviceID: "receiver"}
	deps := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}
	tr, err := transport.Build(context.Background(), "bluetooth", nil, dev, deps)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	defer tr.Close()

	// send zero-length header into the receiver's incoming channel by writing to the same conn the receiver will use
	conn, err := mock.ConnectToDevice(context.Background(), "receiver")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	zero := make([]byte, 4)
	if _, err := conn.Write(zero); err != nil {
		t.Fatalf("write: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ib, err := tr.Receive(ctx)
	if err == nil {
		t.Fatalf("expected error on zero-length frame")
	}
	if ib != nil {
		t.Fatalf("expected nil inbound bundle on error")
	}
}

func TestBluetoothDiscoverUsesAdapterScan(t *testing.T) {
	mock := newTestAdapter()
	mock.AdvertiseDevice("dev1")
	local := transport.DeviceDescriptor{DeviceID: "local"}
	deps := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}
	trIfc, err := new(factory).New(context.Background(), nil, local, deps)
	if err != nil {
		t.Fatalf("new transport: %v", err)
	}
	tr, ok := trIfc.(*Transport)
	if !ok {
		t.Fatalf("expected *Transport")
	}
	devs, err := tr.Discover(context.Background(), 10)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(devs) == 0 {
		t.Fatalf("expected discovered devices")
	}
}

func TestBluetoothCloseNoop(t *testing.T) {
	deps := transport.Dependencies{BluetoothAdapter: newTestAdapter()}
	local := transport.DeviceDescriptor{DeviceID: "l"}
	f := factory{}
	tpt, err := f.New(context.Background(), nil, local, deps)
	if err != nil {
		t.Fatalf("factory.New: %v", err)
	}
	if err := tpt.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}
