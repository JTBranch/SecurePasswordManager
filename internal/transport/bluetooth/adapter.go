package bluetooth

import "context"

// Adapter abstracts a platform BLE adapter or a test mock.
type Adapter interface {
	// ConnectToDevice establishes a logical connection to the device identified by id.
	ConnectToDevice(ctx context.Context, deviceID string) (Conn, error)
}

// Conn represents a bidirectional byte stream used to exchange framed messages.
type Conn interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Close() error
}
