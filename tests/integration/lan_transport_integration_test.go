package integration

import (
	"context"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	"go-password-manager/internal/transport/lan"
	rep "go-password-manager/tests/reporting"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func buildLan(t *testing.T, dev transport.DeviceDescriptor) transport.BundleTransport {
	tr, err := transport.Build(context.Background(), "lan", map[string]any{"listen_addr": ":0", "discovery": false}, dev, transport.Dependencies{Registry: transport.NewInMemoryRegistry()})
	if err != nil {
		require.NoError(t, err, "build %s: %v", dev.DeviceID, err)
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
		require.NoError(t, err, "send failed: %v", err)
		require.Equal(t, bundle.Payload.ID, receipt.BundleID, "expected receipt bundle id %s got %s", bundle.Payload.ID, receipt.BundleID)
		require.Equal(t, 1, receipt.Attempt, "expected delivery in 1 attempt, got %d", receipt.Attempt)

		select {
		case ib := <-recvCh:
			require.Equal(t, bundle.Payload.ID, ib.Bundle.Payload.ID, "mismatched bundle id")
			require.NotNil(t, ib.Envelope, "expected inbound envelope metadata")
		case <-ctx.Done():
			require.FailNow(t, "timeout waiting for receive")
		}
	})
}
