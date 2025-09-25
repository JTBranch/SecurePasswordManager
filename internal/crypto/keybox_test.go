package crypto

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/curve25519"
)

// helper to make scalar
func mustPriv(t *testing.T) []byte {
	b := make([]byte, curve25519.ScalarSize)
	if _, err := randReader.Read(b); err != nil {
		require.NoError(t, err, "rand: %v", err)
	}
	return b
}

// lightweight indirection so we can swap in testing rand if needed
var randReader = randReaderImpl{}

type randReaderImpl struct{}

func (r randReaderImpl) Read(p []byte) (int, error) { return rand.Read(p) }

var testAAD = []byte("keybox-test-aad")

const randErrMsg = "sym rand: %v"

func TestKeyBoxRoundTrip(t *testing.T) {
	priv := mustPriv(t)
	pub, err := curve25519.X25519(priv, curve25519.Basepoint)
	require.NoError(t, err, "pub gen: %v", err)
	sym := make([]byte, 32)
	if _, err := randReader.Read(sym); err != nil {
		require.NoError(t, err, randErrMsg, err)
	}
	box, err := WrapKeyBox(sym, pub, testAAD)
	require.NoError(t, err, "wrap: %v", err)
	out, err := UnwrapKeyBox(box, priv, testAAD)
	require.NoError(t, err, "unwrap: %v", err)
	require.Equal(t, string(sym), string(out), "mismatch got %x want %x", out, sym)
}

func TestKeyBoxWrongKey(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	if _, err := randReader.Read(sym); err != nil {
		require.NoError(t, err, "sym rand: %v", err)
	}
	box, err := WrapKeyBox(sym, pub, testAAD)
	require.NoError(t, err, "wrap: %v", err)
	wrongPriv := mustPriv(t)
	_, err = UnwrapKeyBox(box, wrongPriv, testAAD)
	require.Error(t, err, "expected error with wrong key")
}

func TestKeyBoxTruncated(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	if _, err := randReader.Read(sym); err != nil {
		require.NoError(t, err, "sym rand: %v", err)
	}
	box, _ := WrapKeyBox(sym, pub, testAAD)
	for i := 0; i < len(box); i++ {
		_, err := UnwrapKeyBox(box[:i], priv, testAAD)
		require.Error(t, err, "expected error with truncated box")
	}
}

func TestKeyBoxTamperCiphertext(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	if _, err := randReader.Read(sym); err != nil {
		require.NoError(t, err, "sym rand: %v", err)
	}
	box, _ := WrapKeyBox(sym, pub, testAAD)
	if len(box) < 1+32+12+1 {
		t.Skip("box too small for tamper")
	}
	idx := 1 + 32 + 12
	box[idx] ^= 0xFF
	_, err := UnwrapKeyBox(box, priv, testAAD)
	require.Error(t, err, "expected auth error after tamper")
}

func TestKeyBoxBadVersion(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	if _, err := randReader.Read(sym); err != nil {
		require.NoError(t, err, "sym rand: %v", err)
	}
	box, _ := WrapKeyBox(sym, pub, testAAD)
	box[0] = 9
	_, err := UnwrapKeyBox(box, priv, testAAD)
	require.Error(t, err, "expected version error")
}

func TestKeyBoxAADMismatch(t *testing.T) {
	priv := mustPriv(t)
	pub, _ := curve25519.X25519(priv, curve25519.Basepoint)
	sym := make([]byte, 32)
	if _, err := randReader.Read(sym); err != nil {
		require.NoError(t, err, "sym rand: %v", err)
	}
	box, _ := WrapKeyBox(sym, pub, []byte("aad-one"))
	_, err := UnwrapKeyBox(box, priv, []byte("aad-two"))
	require.Error(t, err, "expected failure on AAD mismatch")
}
