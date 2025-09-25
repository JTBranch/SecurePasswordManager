package bluetooth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PlatformAdapter is a minimal in-memory adapter useful for running tests
// and higher-level integration scenarios without real BLE hardware.
// It provides a mailbox per device ID. Calling ConnectToDevice("id")
// returns a Conn that reads from the mailbox for "id" and writes into
// the same mailbox (writes are visible to any reader of that mailbox).
type PlatformAdapter struct {
	mu         sync.RWMutex
	peers      map[string]*platformConn
	advertised map[string]bool
	// optional simulation parameters
	mtu            int
	failNextWrites int
}

// NewPlatformAdapter creates an empty in-memory adapter.
func NewPlatformAdapter() *PlatformAdapter {
	return &PlatformAdapter{peers: map[string]*platformConn{}, advertised: map[string]bool{}}
}

// CreateDevice registers a device mailbox with the given id.
func (a *PlatformAdapter) CreateDevice(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.peers[id]; !ok {
		a.peers[id] = &platformConn{ch: make(chan []byte, 32)}
	}
}

// AdvertiseDevice marks a device as advertised for Scan().
func (a *PlatformAdapter) AdvertiseDevice(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.advertised[id] = true
	if _, ok := a.peers[id]; !ok {
		a.peers[id] = &platformConn{ch: make(chan []byte, 32)}
	}
}

// StopAdvertise removes a device from advertised set.
func (a *PlatformAdapter) StopAdvertise(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.advertised, id)
}

// Scan returns up to limit advertised device IDs (order is unspecified).
func (a *PlatformAdapter) Scan(serviceUUID string, limit int) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]string, 0, len(a.advertised))
	for id := range a.advertised {
		out = append(out, id)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

// Test helpers to simulate MTU fragmentation and transient write failures.
func (a *PlatformAdapter) SetMTU(mtu int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.mtu = mtu
}

func (a *PlatformAdapter) FailNextWrites(n int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failNextWrites = n
}

// ConnectToDevice returns a Conn representing the mailbox for deviceID.
// If the mailbox does not exist, an error is returned.
func (a *PlatformAdapter) ConnectToDevice(ctx context.Context, deviceID string) (Conn, error) {
	a.mu.RLock()
	pc, ok := a.peers[deviceID]
	mtu := a.mtu
	fail := a.failNextWrites
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no such device: %s", deviceID)
	}
	// return a shallow copy that inherits the mailbox but has a snapshot of simulation params
	return &platformConn{ch: pc.ch, mtu: mtu, failCount: fail}, nil
}

// platformConn implements Conn backed by a channel for inbound messages.
type platformConn struct {
	ch        chan []byte
	pending   []byte
	mu        sync.Mutex
	mtu       int
	failCount int
}

func (c *platformConn) Write(b []byte) (int, error) {
	if c.failCount > 0 {
		c.failCount--
		return 0, fmt.Errorf("simulated write failure")
	}
	// If MTU is set and message is larger, fragment into multiple writes.
	if c.mtu > 0 && len(b) > c.mtu {
		sent := 0
		for sent < len(b) {
			end := sent + c.mtu
			if end > len(b) {
				end = len(b)
			}
			part := make([]byte, end-sent)
			copy(part, b[sent:end])
			// non-blocking send; if full return error
			select {
			case c.ch <- part:
			default:
				return 0, fmt.Errorf("platformConn: channel full")
			}
			sent = end
			// small pause to emulate asynchronous BLE fragments and give reader time to consume
			time.Sleep(5 * time.Millisecond)
		}
		return len(b), nil
	}
	nb := make([]byte, len(b))
	copy(nb, b)
	select {
	case c.ch <- nb:
	default:
		return 0, fmt.Errorf("platformConn: channel full")
	}
	return len(b), nil
}

func (c *platformConn) Read(b []byte) (int, error) {
	c.mu.Lock()
	if len(c.pending) > 0 {
		n := copy(b, c.pending)
		c.pending = c.pending[n:]
		c.mu.Unlock()
		return n, nil
	}
	c.mu.Unlock()
	part, ok := <-c.ch
	if !ok {
		return 0, fmt.Errorf("platformConn: channel closed")
	}
	n := copy(b, part)
	if n < len(part) {
		c.mu.Lock()
		c.pending = append(c.pending, part[n:]...)
		c.mu.Unlock()
	}
	return n, nil
}

func (c *platformConn) Close() error {
	return nil
}
