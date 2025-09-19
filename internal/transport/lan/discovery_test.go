package lan

import (
	"go-password-manager/internal/transport"
	"net"
	"testing"

	"github.com/grandcat/zeroconf"
	"github.com/stretchr/testify/require"
)

func TestBrowserProcessEntry(t *testing.T) {
	b := &Browser{entries: map[string]transport.DeviceDescriptor{}}
	entry := &zeroconf.ServiceEntry{Text: []string{"device_id=devXYZ", "device_name=Device X"}, Port: 4242}
	// Add a loopback IPv4 to simulate address resolution.
	entry.AddrIPv4 = []net.IP{net.ParseIP("127.0.0.1")}
	b.processEntry(entry)
	got := b.List()
	require.Len(t, got, 1, "expected 1 entry got %d", len(got))
	require.Equal(t, "devXYZ", got[0].DeviceID, "unexpected device id: %s", got[0].DeviceID)
	require.NotEmpty(t, got[0].LastAddr, "expected LastAddr populated")
	// Simulate change in port (address update)
	entry.Port = 4343
	b.processEntry(entry)
	got2 := b.List()
	require.NotEqual(t, got2[0].LastAddr, got[0].LastAddr, "expected addr change detected")
	require.False(t, got2[0].LastSeenAt.Before(got[0].LastSeenAt), "expected updated LastSeenAt")
}
