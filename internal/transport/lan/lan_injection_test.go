package lan

import (
	"context"
	"go-password-manager/internal/transport"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeKeyGen struct{ calls int }

func (f *fakeKeyGen) GenerateEd25519KeyPair() ([]byte, []byte, error) {
	f.calls++
	// 32-byte pub, 64-byte priv (ed25519 sizes)
	pub := make([]byte, 32)
	priv := make([]byte, 64)
	for i := 0; i < 32; i++ {
		pub[i] = byte(i)
	}
	for i := 0; i < 64; i++ {
		priv[i] = byte(100 + i)
	}
	return pub, priv, nil
}

type stubAdv struct{ starts int }

func (s *stubAdv) Start(local *transport.DeviceDescriptor, addr string, port int, txt []string) (transport.AdvertisementHandle, error) {
	s.starts++
	return advHandleFunc(func() error { return nil }), nil
}

type advHandleFunc func() error

func (f advHandleFunc) Close() error { return f() }

// TestKeyGenInjectionEnsuresDevicePub set by injected key generator.
func TestKeyGenInjectionEnsuresDevicePub(t *testing.T) {
	kg := &fakeKeyGen{}
	local := transport.DeviceDescriptor{DeviceID: "dev-keygen", DeviceName: "KeyGen"}
	trGeneric, err := transport.Build(context.Background(), "lan", map[string]any{"listen_addr": ":0"}, local, transport.Dependencies{Registry: transport.NewInMemoryRegistry(), KeyGen: kg})
	require.NoError(t, err, "build failed: %v", err)
	tr := trGeneric.(*Transport)
	require.Equal(t, 32, len(tr.local.Ed25519Pub), "expected ed25519 pub set, got %d", len(tr.local.Ed25519Pub))
	require.NotZero(t, kg.calls, "expected keygen to be called")
	_ = tr.Close()
}

// TestAdvertisementInjectionStartCalled ensures injected advertisement starter is invoked when discovery enabled.
func TestAdvertisementInjectionStartCalled(t *testing.T) {
	adv := &stubAdv{}
	local := transport.DeviceDescriptor{DeviceID: "dev-adv", DeviceName: "Adv"}
	trGeneric, err := transport.Build(context.Background(), "lan", map[string]any{"listen_addr": ":0", "discovery": true}, local, transport.Dependencies{Registry: transport.NewInMemoryRegistry(), Advertisement: adv})
	require.NoError(t, err, "build failed: %v", err)
	require.Equal(t, 1, adv.starts, "expected advertisement start once, got %d", adv.starts)
	_ = trGeneric.Close()
}
