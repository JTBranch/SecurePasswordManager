package lan

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"go-password-manager/internal/transport"

	"github.com/grandcat/zeroconf"
)

// Browser discovers LAN devices advertising the password manager service.
type Browser struct {
	mu      sync.RWMutex
	entries map[string]transport.DeviceDescriptor
	cancel  context.CancelFunc
}

// StartBrowser begins asynchronous discovery. Repeated calls replace the prior session.
func StartBrowser(ctx context.Context) *Browser {
	b := &Browser{entries: map[string]transport.DeviceDescriptor{}}
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return b
	}
	results := make(chan *zeroconf.ServiceEntry, 16)
	go b.runBrowse(ctx, resolver, results)
	go b.consumeResults(ctx, results)
	return b
}

func (b *Browser) runBrowse(ctx context.Context, resolver *zeroconf.Resolver, results chan *zeroconf.ServiceEntry) {
	defer close(results)
	_ = resolver.Browse(ctx, "_vibes-pass._tcp", "local.", results)
}

func (b *Browser) consumeResults(ctx context.Context, results chan *zeroconf.ServiceEntry) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-results:
			if !ok {
				return
			}
			b.processEntry(e)
		}
	}
}

func (b *Browser) processEntry(e *zeroconf.ServiceEntry) {
	if e == nil {
		return
	}
	id := txtValue(e, "device_id")
	if id == "" {
		return
	}
	dd := transport.DeviceDescriptor{DeviceID: id, DeviceName: txtValue(e, "device_name")}
	if pkStr := txtValue(e, "ed25519"); pkStr != "" {
		if pk, err := base64.StdEncoding.DecodeString(pkStr); err == nil {
			dd.Ed25519Pub = pk
		}
	}
	if len(e.AddrIPv4) > 0 {
		dd.LastAddr = e.AddrIPv4[0].String() + ":" + itoa(e.Port)
	} else if len(e.AddrIPv6) > 0 {
		dd.LastAddr = "[" + e.AddrIPv6[0].String() + "]:" + itoa(e.Port)
	}
	b.mu.Lock()
	prior, exists := b.entries[dd.DeviceID]
	if !exists || prior.LastAddr != dd.LastAddr {
		dd.LastSeenAt = time.Now()
		b.entries[dd.DeviceID] = dd
	}
	b.mu.Unlock()
}

// Stop terminates discovery.
func (b *Browser) Stop() {
	if b == nil || b.cancel == nil {
		return
	}
	b.cancel()
}

// List returns a snapshot of discovered devices.
func (b *Browser) List() []transport.DeviceDescriptor {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]transport.DeviceDescriptor, 0, len(b.entries))
	for _, v := range b.entries {
		out = append(out, v)
	}
	return out
}

func txtValue(e *zeroconf.ServiceEntry, key string) string {
	prefix := key + "="
	for _, t := range e.Text {
		if len(t) >= len(prefix) && t[:len(prefix)] == prefix {
			return t[len(prefix):]
		}
	}
	return ""
}

// small int -> string without allocating fmt machinery.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [6]byte // ports < 65536
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + (i % 10))
		i /= 10
	}
	return string(buf[pos:])
}
