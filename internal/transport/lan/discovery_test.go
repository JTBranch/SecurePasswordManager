package lan

import (
	"go-password-manager/internal/transport"
	"net"
	"testing"

	"github.com/grandcat/zeroconf"
)

func TestBrowserProcessEntry(t *testing.T) {
	b := &Browser{entries: map[string]transport.DeviceDescriptor{}}
	entry := &zeroconf.ServiceEntry{Text: []string{"device_id=devXYZ", "device_name=Device X"}, Port: 4242}
	// Add a loopback IPv4 to simulate address resolution.
	entry.AddrIPv4 = []net.IP{net.ParseIP("127.0.0.1")}
	b.processEntry(entry)
	got := b.List()
	if len(got) != 1 {
		t.Fatalf("expected 1 entry got %d", len(got))
	}
	if got[0].DeviceID != "devXYZ" {
		t.Fatalf("unexpected device id: %s", got[0].DeviceID)
	}
	if got[0].LastAddr == "" {
		t.Fatalf("expected LastAddr populated")
	}
	// Simulate change in port (address update)
	entry.Port = 4343
	b.processEntry(entry)
	got2 := b.List()
	if got2[0].LastAddr == got[0].LastAddr {
		t.Fatalf("expected addr change detected")
	}
	if got2[0].LastSeenAt.Before(got[0].LastSeenAt) {
		t.Fatalf("expected updated LastSeenAt")
	}
}
