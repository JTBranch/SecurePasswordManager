package lan

import (
	"context"
	"testing"
	"time"

	"go-password-manager/internal/transport"
)

func TestFactoryBuilds(t *testing.T) {
	local := transport.DeviceDescriptor{DeviceID: "local-1", UserID: "user", DeviceName: "Local"}
	tr, err := transport.Build(context.Background(), "lan", map[string]any{"listen_addr": ":0"}, local, transport.Dependencies{Registry: transport.NewInMemoryRegistry()})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if tr.ID() != "lan" {
		t.Fatalf("unexpected id %s", tr.ID())
	}
	_ = tr.Close()
}

func TestSendStub(t *testing.T) {
	local := transport.DeviceDescriptor{DeviceID: "local-1", UserID: "user", DeviceName: "Local"}
	tr, err := transport.Build(context.Background(), "lan", nil, local, transport.Dependencies{Registry: transport.NewInMemoryRegistry()})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	defer tr.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := tr.Send(ctx, nil, transport.DeviceDescriptor{DeviceID: "remote-1"})
	if err == nil {
		t.Fatalf("expected error from stub send")
	}
	if r == nil {
		t.Fatalf("expected receipt even on error")
	}
}
