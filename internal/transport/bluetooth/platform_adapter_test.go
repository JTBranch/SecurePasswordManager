package bluetooth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlatformAdapterCreateAndScan(t *testing.T) {
	a := NewPlatformAdapter()
	a.CreateDevice("d1")
	a.CreateDevice("d2")
	a.AdvertiseDevice("d2")

	// scan should return advertised device
	ids := a.Scan("", 10)
	require.Len(t, ids, 1)
	require.Equal(t, "d2", ids[0])
}

func TestPlatformAdapterWriteRead(t *testing.T) {
	a := NewPlatformAdapter()
	a.CreateDevice("peer")
	conn, err := a.ConnectToDevice(context.Background(), "peer")
	require.NoError(t, err)
	defer conn.Close()

	// write and read a small message
	msg := []byte("hello")
	n, err := conn.Write(msg)
	require.NoError(t, err)
	require.Equal(t, len(msg), n)

	buf := make([]byte, 64)
	n, err = conn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, string(buf[:n]), string(msg))
}

func TestPlatformAdapterMTUFragmentation(t *testing.T) {
	a := NewPlatformAdapter()
	a.CreateDevice("peer")
	a.SetMTU(4)
	conn, err := a.ConnectToDevice(context.Background(), "peer")
	require.NoError(t, err)
	defer conn.Close()

	// send a message larger than MTU; fragments should be reassembled by Read
	payload := make([]byte, 100)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	_, err = conn.Write(payload)
	require.NoError(t, err)

	// read repeatedly and accumulate fragments until we have the full payload
	got := make([]byte, 0, len(payload))
	deadline := time.Now().Add(2 * time.Second)
	for len(got) < len(payload) && time.Now().Before(deadline) {
		// perform a read with timeout to avoid hanging the test
		readBuf := make([]byte, 64)
		type readResult struct {
			n   int
			err error
		}
		ch := make(chan readResult, 1)
		go func() {
			n, err := conn.Read(readBuf)
			ch <- readResult{n: n, err: err}
		}()
		select {
		case res := <-ch:
			require.NoError(t, res.err)
			require.Greater(t, res.n, 0)
			got = append(got, readBuf[:res.n]...)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout waiting for fragment; got %d/%d bytes", len(got), len(payload))
		}
	}
	require.Equal(t, len(payload), len(got))
	require.Equal(t, payload, got)
}

func TestPlatformAdapterFailNextWrites(t *testing.T) {
	a := NewPlatformAdapter()
	a.CreateDevice("peer")
	a.FailNextWrites(2)
	conn, err := a.ConnectToDevice(context.Background(), "peer")
	require.NoError(t, err)
	defer conn.Close()

	// first two writes should fail
	_, err = conn.Write([]byte("x"))
	require.Error(t, err)
	_, err = conn.Write([]byte("y"))
	require.Error(t, err)

	// third write should succeed
	_, err = conn.Write([]byte("z"))
	require.NoError(t, err)
	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "z", string(buf[:n]))
}
