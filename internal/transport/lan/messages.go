//go:build lan_proto

package lan

import (
	"encoding/binary"
	"errors"
	"io"
)

// Protocol Versions
const (
	Version1 uint16 = 0x0001
)

// Message Type Codes (1 byte)
const (
	msgClientHello   = 0x01
	msgServerHello   = 0x02
	msgTransferStart = 0x10
	msgTransferChunk = 0x11
	msgTransferEnd   = 0x12
	msgAck           = 0x20
	msgError         = 0x21
)

// Max sizes / limits (initial conservative defaults)
const (
	maxHelloSize  = 256
	maxChunkSize  = 64 * 1024 // pre-encryption payload
	maxBundleSize = 8 * 1024 * 1024
)

// ClientHello initiates handshake: ephemeral public key + supported versions.
type ClientHello struct {
	Versions     []uint16
	EphemeralPub []byte
	Nonce        []byte // random 16 bytes to bind transcript
}

// ServerHello responds: chosen version, server ephemeral, signature over transcript.
type ServerHello struct {
	Version      uint16
	EphemeralPub []byte
	Signature    []byte // Ed25519 over hash(transcript)
}

// TransferStart conveys high-level bundle metadata (unencrypted header) before encrypted chunks.
type TransferStart struct {
	BundleID  string
	TotalSize uint32 // ciphertext size (for progress)
	ChunkSize uint32
}

// TransferChunk carries a slice of the encrypted bundle payload.
type TransferChunk struct {
	Index uint32
	Data  []byte
}

// TransferEnd signals completion and includes final MAC (redundant integrity / early close detect).
type TransferEnd struct {
	FinalTag []byte
}

// Ack acknowledges receipt of a message (by type + index) for retry/backoff strategy.
type Ack struct {
	RefType  byte
	RefIndex uint32
	Code     uint16 // 0 = OK, otherwise error class
}

// ErrorMsg sends an error and optional detail before closing connection.
type ErrorMsg struct {
	Code   uint16
	Detail string
}

// Frame header format:
// | 1 byte type | 3 bytes length (big endian) | payload bytes |
// Length is payload length only (0..16MB-1 theoretical). We use 3 bytes to keep header small.

var (
	errFrameTooLarge = errors.New("lan: frame too large")
	errShortHeader   = errors.New("lan: short frame header")
)

// writeFrame writes a single framed payload.
func writeFrame(w io.Writer, typ byte, payload []byte) error {
	if len(payload) > 0xFFFFFF {
		return errFrameTooLarge
	}
	hdr := []byte{typ, byte(len(payload) >> 16), byte(len(payload) >> 8), byte(len(payload))}
	if _, err := w.Write(hdr); err != nil {
		return err
	}
	if len(payload) > 0 {
		_, err := w.Write(payload)
		return err
	}
	return nil
}

// readFrame reads next frame returning type and payload.
func readFrame(r io.Reader) (byte, []byte, error) {
	hdr := make([]byte, 4)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return 0, nil, err
	}
	l := int(hdr[1])<<16 | int(hdr[2])<<8 | int(hdr[3])
	if l < 0 || l > 0xFFFFFF {
		return 0, nil, errFrameTooLarge
	}
	if l == 0 {
		return hdr[0], nil, nil
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, nil, err
	}
	return hdr[0], buf, nil
}

// encodeUint16Slice big-endian encodes a slice for inclusion in a payload.
func encodeUint16Slice(v []uint16) []byte {
	if len(v) == 0 {
		return nil
	}
	out := make([]byte, 2*len(v))
	for i, x := range v {
		binary.BigEndian.PutUint16(out[2*i:], x)
	}
	return out
}

// decodeUint16Slice decodes len(v)/2 items.
func decodeUint16Slice(b []byte) []uint16 {
	if len(b)%2 != 0 {
		return nil
	}
	out := make([]uint16, len(b)/2)
	for i := 0; i < len(out); i++ {
		out[i] = binary.BigEndian.Uint16(b[2*i:])
	}
	return out
}
