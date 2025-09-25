package btmock

import (
	"context"
	"errors"
	"time"

	bt "go-password-manager/internal/transport/bluetooth"
)

// testConn implements bt.Conn for in-process testing and supports MTU/failure simulation.
type testConn struct {
	ch        chan []byte
	mtu       int
	failCount int
	// pending stores bytes that have been read from the channel but not yet
	// consumed by the caller's Read buffer.
	pending []byte
}

func (c *testConn) Write(b []byte) (int, error) {
	if c.failCount > 0 {
		c.failCount--
		return 0, errors.New("simulated write failure")
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
			c.ch <- part
			sent = end
			// small pause to emulate asynchronous BLE fragments
			time.Sleep(5 * time.Millisecond)
		}
		return len(b), nil
	}
	nb := make([]byte, len(b))
	copy(nb, b)
	c.ch <- nb
	return len(b), nil
}

func (c *testConn) Read(b []byte) (int, error) {
	// If we already have pending data, consume that first.
	if len(c.pending) > 0 {
		n := copy(b, c.pending)
		c.pending = c.pending[n:]
		return n, nil
	}

	// Aggregate fragments until we have at least a full length-prefixed frame.
	buf := make([]byte, 0, 1024)
	for {
		part := <-c.ch
		buf = append(buf, part...)
		// Need at least 4 bytes for the length header
		if len(buf) >= 4 {
			ln := (uint32(buf[0]) << 24) | (uint32(buf[1]) << 16) | (uint32(buf[2]) << 8) | uint32(buf[3])
			if int(4+ln) <= len(buf) {
				break
			}
		}
	}

	// Return assembled frame and store no leftovers since transport expects full frame.
	n := copy(b, buf)
	if n < len(buf) {
		c.pending = append(c.pending, buf[n:]...)
	}
	return n, nil
}

func (c *testConn) Close() error { return nil }

// Adapter is a simple in-memory pairing adapter for tests. It supports MTU simulation
// and transient write failures, as well as a minimal advertise/scan simulation.
type Adapter struct {
	peers          map[string]*testConn
	mtu            int
	failNextWrites int
	advertised     map[string]bool
}

func NewTestAdapter() *Adapter {
	return &Adapter{peers: map[string]*testConn{}, mtu: 0, failNextWrites: 0, advertised: map[string]bool{}}
}

func (a *Adapter) PairDevices(x, y string) {
	// create separate incoming queues for each endpoint
	chA := make(chan []byte, 4)
	chB := make(chan []byte, 4)
	// each peer's connection exposes their incoming queue
	cx := &testConn{ch: chA}
	cy := &testConn{ch: chB}
	a.peers[x] = cx
	a.peers[y] = cy
}

func (a *Adapter) ConnectToDevice(ctx context.Context, deviceID string) (bt.Conn, error) {
	if c, ok := a.peers[deviceID]; ok {
		// create a shallow copy to avoid races on per-connection state
		tc := &testConn{ch: c.ch, mtu: a.mtu, failCount: a.failNextWrites}
		return tc, nil
	}
	return nil, &NoSuchDeviceError{}
}

type NoSuchDeviceError struct{}

func (NoSuchDeviceError) Error() string { return "no such device" }

// Test helpers
func (a *Adapter) SetMTU(mtu int)       { a.mtu = mtu }
func (a *Adapter) FailNextWrites(n int) { a.failNextWrites = n }

// Advertise/Scan helpers for discovery tests
func (a *Adapter) AdvertiseDevice(deviceID string) { a.advertised[deviceID] = true }
func (a *Adapter) StopAdvertise(deviceID string)   { delete(a.advertised, deviceID) }
func (a *Adapter) Scan(serviceUUID string, limit int) []string {
	var res []string
	for id := range a.advertised {
		res = append(res, id)
		if limit > 0 && len(res) >= limit {
			break
		}
	}
	return res
}
