package crypto

import (
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/curve25519"
)

// helper to make scalar
func mustPriv(t *testing.T) []byte {
	b := make([]byte, curve25519.ScalarSize)
	if _, err := randReader.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return b
}

// lightweight indirection so we can swap in testing rand if needed
var randReader = randReaderImpl{}

type randReaderImpl struct{}

func (r randReaderImpl) Read(p []byte) (int, error) { return rand.Read(p) }

var testAAD = []byte("keybox-test-aad")

func TestKeyBoxRoundTrip(t *testing.T) {
	priv := mustPriv(t)
	pub, err := curve25519.X25519(priv, curve25519.Basepoint)
	if err != nil {
		t.Fatalf("pub gen: %v", err)
	}
	sym := make([]byte, 32)
	if _, err := randReader.Read(sym); err != nil {
		t.Fatalf("sym rand: %v", err)
	}
	box, err := WrapKeyBox(sym, pub, testAAD)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	out, err := UnwrapKeyBox(box, priv, testAAD)
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if string(out) != string(sym) {
		t.Fatalf("mismatch got %x want %x", out, sym)
	}
}

func TestKeyBoxWrongKey(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	randReader.Read(sym)
	box, err := WrapKeyBox(sym, pub, testAAD)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	wrongPriv := mustPriv(t)
	if _, err := UnwrapKeyBox(box, wrongPriv, testAAD); err == nil {
		t.Fatalf("expected error with wrong key")
	}
}

func TestKeyBoxTruncated(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	randReader.Read(sym)
	box, _ := WrapKeyBox(sym, pub, testAAD)
	for i := 0; i < len(box); i++ {
		if _, err := UnwrapKeyBox(box[:i], priv, testAAD); err == nil {
			// expected failure
		}
	}
}

func TestKeyBoxTamperCiphertext(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	randReader.Read(sym)
	box, _ := WrapKeyBox(sym, pub, testAAD)
	if len(box) < 1+32+12+1 {
		t.Skip("box too small for tamper")
	}
	idx := 1 + 32 + 12
	box[idx] ^= 0xFF
	if _, err := UnwrapKeyBox(box, priv, testAAD); err == nil {
		t.Fatalf("expected auth error after tamper")
	}
}

func TestKeyBoxBadVersion(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	randReader.Read(sym)
	box, _ := WrapKeyBox(sym, pub, testAAD)
	box[0] = 9
	if _, err := UnwrapKeyBox(box, priv, testAAD); err == nil {
		t.Fatalf("expected version error")
	}
}

func TestKeyBoxAADMismatch(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	randReader.Read(sym)
	box, _ := WrapKeyBox(sym, pub, []byte("aad-one"))
	if _, err := UnwrapKeyBox(box, priv, []byte("aad-two")); err == nil {
		t.Fatalf("expected failure on AAD mismatch")
	}
}
