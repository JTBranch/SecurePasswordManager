package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockConfigProvider is a mock implementation of the ConfigProvider interface for testing.
type mockConfigProvider struct {
	keyUUID string
}

// GetKeyUUID returns the mock key UUID.
func (m *mockConfigProvider) GetKeyUUID() string {
	return m.keyUUID
}

// TestNewCryptoServiceWithMockConfig tests the creation of a CryptoService
// using a mock configuration provider. This verifies that the service can be
// instantiated without a dependency on the actual runtimeconfig package.
func TestNewCryptoServiceWithMockConfig(t *testing.T) {
	// Arrange: Create a mock config provider
	mockProvider := &mockConfigProvider{
		keyUUID: "test-uuid-12345",
	}

	// Act: Create the CryptoService with the mock provider
	cryptoService, err := NewCryptoService(mockProvider)

	// Assert: Check that the service was created successfully
	assert.NoError(t, err, "NewCryptoService should not return an error with a valid mock provider")
	assert.NotNil(t, cryptoService, "NewCryptoService should return a non-nil service instance")
}
