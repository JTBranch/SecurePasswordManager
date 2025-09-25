package lan

import (
	"context"
	"testing"
	"time"

	"go-password-manager/internal/transport"

	"github.com/stretchr/testify/require"
)

func TestFactoryBuilds(t *testing.T) {
	local := transport.DeviceDescriptor{DeviceID: "local-1", UserID: "user", DeviceName: "Local"}
	tr, err := transport.Build(context.Background(), "lan", map[string]any{"listen_addr": ":0"}, local, transport.Dependencies{Registry: transport.NewInMemoryRegistry()})
	require.NoError(t, err, "build: %v", err)
	require.Equal(t, "lan", tr.ID(), "unexpected id %s", tr.ID())
	_ = tr.Close()
}

func TestSendStub(t *testing.T) {
	local := transport.DeviceDescriptor{DeviceID: "local-1", UserID: "user", DeviceName: "Local"}
	tr, err := transport.Build(context.Background(), "lan", nil, local, transport.Dependencies{Registry: transport.NewInMemoryRegistry()})
	require.NoError(t, err, "build: %v", err)
	defer tr.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := tr.Send(ctx, nil, transport.DeviceDescriptor{DeviceID: "remote-1"})
	require.Error(t, err, "expected error from stub send")
	require.NotNil(t, r, "expected receipt even on error")
}
