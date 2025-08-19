package storage_test

import (
	"encoding/json"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/storage"
	"go-password-manager/tests/helpers"
	"os"
	"testing"
)

const (
	TestFileStoragePath    = "test.json"
	TestFileStorageVersion = "1.0.0"
	TestFileStorageUser    = "user"
)

func TestSecretStorage(t *testing.T) {

	helpers.WithUnitTestCase(t, "Writes and Reads secrets when they exist", func(tc *helpers.UnitTestCase) {
		t.Cleanup(func() {
			_ = os.RemoveAll(TestFileStoragePath)
		})

		storage := storage.NewFileStorage(TestFileStoragePath, TestFileStorageVersion, TestFileStorageUser)

		secretsData := domain.SecretsFile{
			AppVersion: TestFileStorageVersion,
			AppUser:    TestFileStorageUser,
			Secrets:    []domain.Secret{},
		}

		// Simulate saving a secret
		err := storage.WriteSecrets(secretsData)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		savedData, err := os.ReadFile(TestFileStoragePath)
		if err != nil {
			t.Fatalf("Failed to read saved data: %v", err)
		}

		var savedDataObj domain.SecretsFile
		err = json.Unmarshal(savedData, &savedDataObj)
		if err != nil {
			t.Fatalf("Failed to unmarshal saved data: %v", err)
		}

		tc.Assert.NotNil(savedDataObj.LastUpdated)

		savedDataObj.LastUpdated = ""

		tc.Assert.Equal(savedDataObj, secretsData)

		// Test reading the secret
		_, error := storage.ReadSecrets()
		if error != nil {
			t.Errorf("should not get an error on file load: %v", error)
		}
	})

	helpers.WithUnitTestCase(t, "Handles Errors with Read Secrets", func(tc *helpers.UnitTestCase) {
		t.Cleanup(func() {
			_ = os.RemoveAll(TestFileStoragePath)
		})

		os.WriteFile(TestFileStoragePath, []byte("invalid json"), 0644)

		storage := storage.NewFileStorage(TestFileStoragePath, TestFileStorageVersion, TestFileStorageUser)

		// Simulate a read error
		_, err := storage.ReadSecrets()
		tc.Assert.Error(err)
	})

	helpers.WithUnitTestCase(t, "Handles Errors with Write Secrets", func(tc *helpers.UnitTestCase) {
		storage := storage.NewFileStorage("fakePlace/FakePath", TestFileStorageVersion, TestFileStorageUser)

		// Simulate a write error
		err := storage.WriteSecrets(domain.SecretsFile{})
		tc.Assert.Error(err)
	})

}
