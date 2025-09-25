//go:build !windows
// +build !windows

package adapters

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"go-password-manager/internal/logger"

	"github.com/go-ble/ble"
)

// service/characteristic UUIDs used for this app's mailbox
var (
	serviceUUID = ble.MustParse("00000000-0000-0000-0000-000000000001")
	charUUID    = ble.MustParse("00000000-0000-0000-0000-000000000002")
)

// btConn is a minimal Conn implementation used by the adapter. It implements
// simple Read/Write semantics mapped to characteristic operations.
type btConn struct {
	writeCh chan []byte
	readCh  chan []byte
	closed  chan struct{}
}

func (c *btConn) Write(b []byte) (int, error) {
	select {
	case c.writeCh <- append([]byte(nil), b...):
		return len(b), nil
	case <-c.closed:
		return 0, io.ErrClosedPipe
	}
}
func (c *btConn) Read(b []byte) (int, error) {
	select {
	case msg := <-c.readCh:
		n := copy(b, msg)
		return n, nil
	case <-c.closed:
		return 0, io.EOF
	}
}
func (c *btConn) Close() error { close(c.closed); return nil }

// newConnFromClient wires a ble.Client and characteristic into a simple
// read/write Conn backing used by the transport layer.
func newConnFromClient(cln ble.Client, foundChar *ble.Characteristic) (interface{}, error) {
	conn := &btConn{writeCh: make(chan []byte, 4), readCh: make(chan []byte, 4), closed: make(chan struct{})}

	logger.Debug(fmt.Sprintf("bluetooth.conn: newConnFromClient subscribe-capable=%v char=%v", reflect.ValueOf(cln).MethodByName("Subscribe").IsValid(), foundChar.UUID))
	// Helper: polling fallback reader
	startPolling := func() {
		go func() {
			for {
				data, err := cln.ReadCharacteristic(foundChar)
				if err != nil {
					return
				}
				select {
				case conn.readCh <- append([]byte(nil), data...):
				default:
				}
			}
		}()
	}

	// Try to use Subscribe if available; use reflection to support multiple
	// go-ble versions. If unavailable or signature mismatches, fall back to
	// polling.
	if sub := reflect.ValueOf(cln).MethodByName("Subscribe"); sub.IsValid() {
		handler := func(b []byte) {
			select {
			case conn.readCh <- append([]byte(nil), b...):
			default:
			}
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Debug("bluetooth.conn: Subscribe panic, falling back to polling")
					startPolling()
				}
			}()
			mt := sub.Type()
			hVal := reflect.ValueOf(handler)
			switch mt.NumIn() {
			case 1:
				sub.Call([]reflect.Value{hVal})
				return
			case 2:
				sub.Call([]reflect.Value{reflect.ValueOf(foundChar), hVal})
				return
			case 3:
				sub.Call([]reflect.Value{reflect.ValueOf(foundChar), reflect.ValueOf(false), hVal})
				return
			default:
				startPolling()
			}
		}()
	} else {
		startPolling()
	}

	logger.Debug("bluetooth.conn: starting write loop to characteristic")
	// write loop sends queued frames to the remote characteristic
	go func() {
		for {
			select {
			case b := <-conn.writeCh:
				_ = cln.WriteCharacteristic(foundChar, b, false)
			case <-conn.closed:
				_ = cln.CancelConnection()
				return
			}
		}
	}()

	return conn, nil
}

// scanWithBLE performs a short scan for devices advertising the given
// service UUID and returns up to `limit` device IDs.
func scanWithBLE(serviceUUID string, limit int) []string {
	out := make([]string, 0)
	ctx := context.Background()
	ch := make(chan string, 16)
	// Use a handler to capture advertisement addresses.
	handler := func(a ble.Advertisement) {
		if serviceUUID == "" {
			select {
			case ch <- a.Addr().String():
			default:
			}
			return
		}
		// check advertised service UUIDs
		want := ble.MustParse(serviceUUID)
		for _, s := range a.Services() {
			if s.Equal(want) {
				select {
				case ch <- a.Addr().String():
				default:
				}
				return
			}
		}
	}
	go func() {
		_ = ble.Scan(ctx, false, func(a ble.Advertisement) {
			handler(a)
		}, nil)
	}()
	timeout := time.After(2 * time.Second)
	for {
		select {
		case id := <-ch:
			// dedupe
			seen := false
			for _, e := range out {
				if e == id {
					seen = true
					break
				}
			}
			if !seen {
				out = append(out, id)
				if limit > 0 && len(out) >= limit {
					return out
				}
			}
		case <-timeout:
			return out
		}
	}
}
