package secretkeymetadata_test

import (
	"go-password-manager/internal/config/secretkeymetadata"
	"go-password-manager/internal/testHelpers/mocks"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSomeFunction(t *testing.T) {
	assert.True(t, true)
}

func TestNewSecretKeyFileServiceCreatesFileIfMissing(t *testing.T) {
	tmpFile := "test_secret_key_metadata.json"
	_ = os.Remove(tmpFile)
	provider := &mocks.ConfigProvider{}
	provider.On("GetSecretKeyMetadataFilePath").Return(tmpFile)
	service, err := secretkeymetadata.NewSecretKeyMetadataFileService(provider)
	require.NoError(t, err)
	assert.NotNil(t, service)
	metadata := service.GetSecretKeyMetadata()
	assert.Nil(t, metadata.CurrentKey)
	assert.Empty(t, metadata.Archived)

	// Clean up
	_ = os.Remove(tmpFile)
}

func TestUpdateCurrentKeyArchivesOldKey(t *testing.T) {
	tmpFile := "test_secret_key_metadata.json"
	_ = os.Remove(tmpFile)
	provider := &mocks.ConfigProvider{}
	provider.On("GetSecretKeyMetadataFilePath").Return(tmpFile)
	service, err := secretkeymetadata.NewSecretKeyMetadataFileService(provider)
	require.NoError(t, err)

	firstKey := secretkeymetadata.SecretsEncryptionKeyMetadata{
		UUID:        "uuid-1",
		DateCreated: time.Now(),
	}
	service.UpdateCurrentKey(firstKey)
	assert.Equal(t, firstKey.UUID, service.GetSecretKeyMetadata().CurrentKey.UUID)
	assert.Empty(t, service.GetSecretKeyMetadata().Archived)

	secondKey := secretkeymetadata.SecretsEncryptionKeyMetadata{
		UUID:        "uuid-2",
		DateCreated: time.Now(),
	}
	service.UpdateCurrentKey(secondKey)
	assert.Equal(t, secondKey.UUID, service.GetSecretKeyMetadata().CurrentKey.UUID)
	assert.Len(t, service.GetSecretKeyMetadata().Archived, 1)
	assert.Equal(t, firstKey.UUID, service.GetSecretKeyMetadata().Archived[0].UUID)

	// Clean up
	_ = os.Remove(tmpFile)
}
