package crypto

import (
	"encoding/pem"
	"fmt"
	"go-password-manager/internal/config/devicekeys"
)

type PemUtils struct{}

func (p *PemUtils) EncodeKeyToPEM(key []byte, keyType devicekeys.KeyType) ([]byte, error) {
	block := &pem.Block{
		Type:  string(keyType),
		Bytes: key,
	}
	return pem.EncodeToMemory(block), nil
}

func (p *PemUtils) DecodeKeyFromPEM(pemBytes []byte, keyType devicekeys.KeyType) ([]byte, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	if block.Type != string(keyType) {
		return nil, fmt.Errorf("unexpected PEM type: got %s, want %s", block.Type, keyType)
	}
	return block.Bytes, nil
}
