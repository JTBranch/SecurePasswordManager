package integration

import (
	"context"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	"go-password-manager/internal/transport/lan"
	rep "go-password-manager/tests/reporting"
	"testing"
	"time"
)

func buildLan(t *testing.T, dev transport.DeviceDescriptor) transport.BundleTransport {
	tr, err := transport.Build(context.Background(), "lan", map[string]any{"listen_addr": ":0", "discovery": false}, dev, transport.Dependencies{Registry: transport.NewInMemoryRegistry()})
	if err != nil {
		t.Fatalf("build %s: %v", dev.DeviceID, err)
	}
	return tr
}

// TestLanSendReceive verifies LAN transport can deliver a bundle A->B with ACK.
func TestLanSendReceive(t *testing.T) {
	rep.WithReporting(t, "TestLanSendReceive", func(r *rep.TestWrapper) {
		devA := transport.DeviceDescriptor{DeviceID: "devA", DeviceName: "A"}
		devB := transport.DeviceDescriptor{DeviceID: "devB", DeviceName: "B"}

		trA := buildLan(t, devA)
		defer trA.Close()
		trB := buildLan(t, devB)
		defer trB.Close()

		lb := trB.(*lan.Transport).Addr()
		targetB := devB
		targetB.LastAddr = lb

		bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "bundle-1", Timestamp: time.Now().Unix()}}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		recvCh := make(chan *transport.InboundBundle, 1)
		go func() {
			ib, err := trB.Receive(ctx)
			if err == nil && ib != nil {
				recvCh <- ib
			}
		}()

		receipt, err := trA.Send(ctx, bundle, targetB)
		if err != nil {
			t.Fatalf("send failed: %v", err)
		}
		if receipt.BundleID != bundle.Payload.ID {
			t.Fatalf("expected receipt bundle id %s got %s", bundle.Payload.ID, receipt.BundleID)
		}
		if receipt.Attempt != 1 {
			t.Fatalf("expected delivery in 1 attempt, got %d", receipt.Attempt)
		}

		select {
		case ib := <-recvCh:
			if ib.Bundle.Payload.ID != bundle.Payload.ID {
				t.Fatalf("mismatched bundle id")
			}
			if ib.Envelope == nil {
				t.Fatalf("expected inbound envelope metadata")
			}
		case <-ctx.Done():
			t.Fatalf("timeout waiting for receive")
		}
	})
}
