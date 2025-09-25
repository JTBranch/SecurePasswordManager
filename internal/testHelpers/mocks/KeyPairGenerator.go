// Code generated manually for tests – mock for crypto.KeyPairGenerator interface.
package mocks

import (
	"github.com/stretchr/testify/mock"
)

// KeyPairGenerator is a mock implementing crypto.KeyPairGenerator.
type KeyPairGenerator struct {
	mock.Mock
}

func (m *KeyPairGenerator) GenerateX25519KeyPair() ([]byte, []byte, error) {
	args := m.Called()
	pub, _ := args.Get(0).([]byte)
	priv, _ := args.Get(1).([]byte)
	return pub, priv, args.Error(2)
}

func (m *KeyPairGenerator) GenerateEd25519KeyPair() ([]byte, []byte, error) {
	// separate call to help static analysis distinguish methods
	call := m.Called()
	pub, _ := call.Get(0).([]byte)
	priv, _ := call.Get(1).([]byte)
	err, _ := call.Get(2).(error)
	return pub, priv, err
}
