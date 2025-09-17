package integration

import (
	"encoding/json"
	"fmt"
	"go-password-manager/internal/config/devicekeys"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/sharing"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/reporting"
	"go-password-manager/tests/testdata"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func SetupSuite(reporter *reporting.TestWrapper) *helpers.IntegrationTestSuite {
	suite := helpers.NewIntegrationTestSuite(reporter)
	suite.SetupTestEnvironment()
	defer suite.Cleanup()
	return suite
}

var TestSenderMetadata = sharing.SenderMetadata{
	DeviceName: "Test-Device",
	UserID:     "test@example.com",
	PublicKey:  []byte("test-public-key"),
}

const failedToCreateSecretMsg = "Failed to create secret"

func TestCanSignAndVerify(t *testing.T) {
	reporting.WithReporting(t, "TestCanSignAndVerify", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		testObj := map[string]string{"hello": "world"}
		testData, err := json.Marshal(testObj)
		require.NoError(t, err, "Failed to marshal test object")

		signingKey, err := suite.DeviceKeyManager.GetSigningDeviceKey()
		require.NoError(t, err, "Failed to get signing device key")

		testSig, err := suite.CryptoService.SignEd25519(testData, signingKey.PrivateKey)
		require.NoError(t, err, "Failed to sign test object")

		valid, err := suite.CryptoService.VerifyEd25519(testData, testSig, signingKey.PublicKey)
		require.NoError(t, err, "Signature verification should not error for test object")
		require.True(t, valid, "Signature should be valid for test object")
	})
}

func TestCanGenerateAndShareSecretBundle(t *testing.T) {
	reporting.WithReporting(t, "TestCanGenerateAndShareSecretBundle", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		testDataManager := testdata.NewTestDataManager()

		// Test 1: Create a secret using test data
		testSecret := testdata.TestSecrets.Simple

		err := testDataManager.CreateTestSecret(suite.SecretsService, testSecret.Name)
		require.NoError(reporter.T(), err, failedToCreateSecretMsg)

		// Test 1: Create a secret using test data
		testSecret = testdata.TestSecrets.Complex

		err = testDataManager.CreateTestSecret(suite.SecretsService, testSecret.Name)
		require.NoError(reporter.T(), err, failedToCreateSecretMsg)

		file, err := suite.SecretsService.LoadAllSecrets()
		require.NoError(reporter.T(), err, "Failed to load secrets")

		pub, priv, err := suite.CryptoService.GenerateX25519KeyPair()
		require.NoError(t, err)

		pemPub, err := suite.PemUtils.EncodeKeyToPEM(pub, devicekeys.KeyTypeX25519Public)
		require.NoError(t, err)

		// do the export
		exportBundle, err := suite.SharingService.ExportSecrets(file.Secrets, pemPub, 60)

		require.NoError(reporter.T(), err, "Failed to export secrets")
		require.NotNil(reporter.T(), exportBundle, "Expected non-nil export bundle")
		require.NotEmpty(t, exportBundle.Payload.ID, "Export bundle ID should not be empty")
		require.NotEmpty(t, exportBundle.Payload.Name, "Export bundle Name should not be empty")
		require.NotNil(t, exportBundle.Payload.EncryptedSecrets, "EncryptedSecrets should not be nil")
		require.NotNil(t, exportBundle.Payload.SecretsNonce, "SecretsNonce should not be nil")
		require.NotNil(t, exportBundle.Payload.SymmetricKeyBox, "SymmetricKeyBox should not be nil")
		require.NotNil(t, exportBundle.Signature, "Signature should not be nil")
		require.True(t, exportBundle.Payload.Timestamp > 0, "Timestamp should be set")
		require.True(t, exportBundle.Payload.ExpiresAt > exportBundle.Payload.Timestamp, "ExpiresAt should be after Timestamp")
		require.NotEmpty(t, exportBundle.Payload.SenderInfo.DeviceName, "SenderInfo.DeviceName should not be empty")
		require.NotEmpty(t, exportBundle.Payload.SenderInfo.UserID, "SenderInfo.UserID should not be empty")
		require.NotNil(t, exportBundle.Payload.SenderInfo.PublicKey, "SenderInfo.PublicKey should not be nil")
		require.NotNil(t, exportBundle.Payload.SenderInfo.SigningPublicKey, "SenderInfo.SigningPublicKey should not be nil")
		require.Equal(t, 32, len(exportBundle.Payload.SenderInfo.SigningPublicKey), "SigningPublicKey should be 32 bytes (Ed25519 public key)")
		require.NotEqual(t, file.Secrets, exportBundle.Payload.EncryptedSecrets, "EncryptedSecrets should not match plaintext")

		sig, err := json.Marshal(exportBundle.Payload)
		require.NoError(t, err, "Failed to marshal export bundle")

		valid, err := suite.CryptoService.VerifyEd25519(
			sig, exportBundle.Signature, exportBundle.Payload.SenderInfo.SigningPublicKey)
		require.NoError(t, err, "Signature verification should not error")
		require.True(t, valid, "Signature should be valid")
		// // delete the secrets

		err = suite.SecretsService.DeleteSecret(testdata.TestSecrets.Simple.Name)
		require.NoError(reporter.T(), err, "Failed to delete secret")
		err = suite.SecretsService.DeleteSecret(testdata.TestSecrets.Complex.Name)
		require.NoError(reporter.T(), err, "Failed to delete secret")

		// now import them

		pemPriv, err := suite.PemUtils.EncodeKeyToPEM(priv, devicekeys.KeyTypeX25519Private)
		require.NoError(t, err)

		importResult, err := suite.SharingService.ImportSecrets(exportBundle, pemPriv, pemPub)
		require.NoError(reporter.T(), err, "Failed to import secrets")

		require.Nil(reporter.T(), importResult.Error)
		require.True(reporter.T(), importResult.Success)
		require.Equal(reporter.T(), 2, importResult.ImportedSecretsCount, "Imported secret count should match exported count")

		// verify Log

		// logResults, err := suite.SharingService.GetLog()
		// require.NoError(reporter.T(), err, "Failed to get log")
		// require.NotEmpty(reporter.T(), logResults, "Expected non-empty log results")
		// require.Equal(reporter.T(), "import", logResults[0].Action, "Expected latest log action to be 'import'")

		// verify can decrypt the imported secrets

		secret1, err := suite.SecretsService.GetSecret(testdata.TestSecrets.Simple.Name)
		require.NoError(reporter.T(), err, "Failed to get secret1")
		require.NotNil(reporter.T(), secret1, "Expected secret1 to be found")
		secretVal, err := suite.SecretsService.GetSecretValue(secret1)
		require.NoError(reporter.T(), err, "Failed to get secret1 value")
		require.Equal(reporter.T(), testdata.TestSecrets.Simple.Value, secretVal, "Secret1 value should match")

		secret2, err := suite.SecretsService.GetSecret(testdata.TestSecrets.Complex.Name)
		require.NoError(reporter.T(), err, "Failed to get secret2")
		require.NotNil(reporter.T(), secret2, "Expected secret2 to be found")
		secretVal, err = suite.SecretsService.GetSecretValue(secret2)
		require.NoError(reporter.T(), err, "Failed to get secret2 value")
		require.Equal(reporter.T(), testdata.TestSecrets.Complex.Value, secretVal, "Secret2 value should match")
	})
}

func TestImportExpiredBundle(t *testing.T) {
	reporting.WithReporting(t, "TestImportExpiredBundle", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		testDataManager := testdata.NewTestDataManager()
		// Create and export a secret with a very short expiry
		testSecret := testdata.TestSecrets.Simple
		err := testDataManager.CreateTestSecret(suite.SecretsService, testSecret.Name)
		require.NoError(reporter.T(), err, failedToCreateSecretMsg)

		file, err := suite.SecretsService.LoadAllSecrets()
		require.NoError(reporter.T(), err, "Failed to load secrets")

		// Generate valid PEM-encoded X25519 key pairs
		pubPEM, privPEM, err := suite.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(reporter.T(), err, "Failed to generate PEM keys")

		// Export with 1 second expiry
		exportBundle, err := suite.SharingService.ExportSecrets(file.Secrets, pubPEM, 1)
		require.NoError(reporter.T(), err, "Failed to export secrets")
		require.NotNil(reporter.T(), exportBundle, "Expected non-nil export bundle")

		// Wait for expiry
		time.Sleep(2 * time.Second)

		// Attempt to import after expiry
		importResult, err := suite.SharingService.ImportSecrets(exportBundle, privPEM, pubPEM)
		require.Nil(reporter.T(), err, "No error should be returned as value; error is in result struct")
		require.NotNil(reporter.T(), importResult, "ImportResult should not be nil")
		require.NotNil(reporter.T(), importResult.Error, "Expected error in ImportResult when importing expired bundle")
		require.False(reporter.T(), importResult.Success, "Import should not succeed for expired bundle")
	})
}

func TestImportWithBadKeys(t *testing.T) {
	reporting.WithReporting(t, "TestImportWithBadKeys", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		testDataManager := testdata.NewTestDataManager()

		testSecret := testdata.TestSecrets.Simple
		err := testDataManager.CreateTestSecret(suite.SecretsService, testSecret.Name)
		require.NoError(reporter.T(), err, failedToCreateSecretMsg)

		file, err := suite.SecretsService.LoadAllSecrets()
		require.NoError(reporter.T(), err, "Failed to load secrets")

		// Generate valid PEM-encoded X25519 key pairs
		pubPEM, _, err := suite.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(reporter.T(), err, "Failed to generate PEM keys")
		// Generate a different (mismatched) private key for import
		_, badPrivPEM, err := suite.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(reporter.T(), err, "Failed to generate bad PEM private key")

		exportBundle, err := suite.SharingService.ExportSecrets(file.Secrets, pubPEM, 60)
		require.NoError(reporter.T(), err, "Failed to export secrets")
		require.NotNil(reporter.T(), exportBundle, "Expected non-nil export bundle")

		// Attempt to import with a mismatched private key (should fail at decryption, not PEM decode)
		importResult, err := suite.SharingService.ImportSecrets(exportBundle, badPrivPEM, pubPEM)
		require.Nil(reporter.T(), err, "No error should be returned as value; error is in result struct")
		require.NotNil(reporter.T(), importResult, "ImportResult should not be nil")
		require.NotNil(reporter.T(), importResult.Error, "Expected error in ImportResult when importing with bad private key")
		require.False(reporter.T(), importResult.Success, "Import should not succeed with bad private key")
	})
}

func TestExportWithEmptySecrets(t *testing.T) {
	reporting.WithReporting(t, "TestExportWithEmptySecrets", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		emptySecrets := []domain.Secret{}
		exportBundle, err := suite.SharingService.ExportSecrets(emptySecrets, []byte("recipient-public-key"), 60)
		require.Error(reporter.T(), err, "Expected error when exporting with empty secrets")
		require.Nil(reporter.T(), exportBundle, "Expected nil export bundle when secrets are empty")
	})
}

func TestImportCorruptedBundleData(t *testing.T) {
	reporting.WithReporting(t, "TestImportCorruptedBundleData", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		testDataManager := testdata.NewTestDataManager()
		secret := testdata.TestSecrets.Simple
		require.NoError(t, testDataManager.CreateTestSecret(suite.SecretsService, secret.Name))
		file, err := suite.SecretsService.LoadAllSecrets()
		require.NoError(t, err)
		pubPEM, privPEM, err := suite.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(t, err)
		exportBundle, err := suite.SharingService.ExportSecrets(file.Secrets, pubPEM, 300)
		require.NoError(t, err)
		// Corrupt a byte inside SymmetricKeyBox (avoid version byte); safe index offset 5 if length permits
		if len(exportBundle.Payload.SymmetricKeyBox) > 6 {
			exportBundle.Payload.SymmetricKeyBox[5] ^= 0xFF
		}
		res, ierr := suite.SharingService.ImportSecrets(exportBundle, privPEM, pubPEM)
		require.NoError(t, ierr)
		require.NotNil(t, res)
		require.False(t, res.Success)
		require.Error(t, res.Error)
	})
}

func TestPartialImportSomeSecretsExist(t *testing.T) {
	reporting.WithReporting(t, "TestPartialImportSomeSecretsExist", func(reporter *reporting.TestWrapper) {
		// suite := SetupSuite(reporter)
		// testDataManager := testdata.NewTestDataManager()
		// TODO: Implement test for partial import when some secrets already exist
	})
}

func TestExportImportLargeSecretValues(t *testing.T) {
	reporting.WithReporting(t, "TestExportImportLargeSecretValues", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		testDataManager := testdata.NewTestDataManager()
		// Generate N secrets with moderately large values
		N := 200
		large := make([]byte, 1024) // 1KB each
		for i := range large {
			large[i] = byte(i % 251)
		}
		// Use ASCII-safe content to avoid UTF-8 replacement during JSON marshal
		pattern := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		wantLen := 1024
		b := make([]byte, wantLen)
		for i := 0; i < wantLen; i++ {
			b[i] = pattern[i%len(pattern)]
		}
		origValue := string(b)
		for i := 0; i < N; i++ {
			name := fmt.Sprintf("perf-secret-%03d", i)
			err := testDataManager.CreateTestSecretWithValue(suite.SecretsService, name, origValue)
			require.NoError(t, err)
		}
		file, err := suite.SecretsService.LoadAllSecrets()
		require.NoError(t, err)
		pubPEM, privPEM, err := suite.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(t, err)
		start := time.Now()
		exportBundle, err := suite.SharingService.ExportSecrets(file.Secrets, pubPEM, 600)
		require.NoError(t, err)
		require.NotNil(t, exportBundle)
		durExport := time.Since(start)
		// prune one secret to validate import restores it
		require.NoError(t, suite.SecretsService.DeleteSecret("perf-secret-000"))
		impStart := time.Now()
		res, err := suite.SharingService.ImportSecrets(exportBundle, privPEM, pubPEM)
		require.NoError(t, err)
		require.True(t, res.Success)
		durImport := time.Since(impStart)
		reporter.LogInfo("large secrets performance", map[string]interface{}{
			"count":       N,
			"export_ms":   durExport.Milliseconds(),
			"import_ms":   durImport.Milliseconds(),
			"total_bytes": N * len(large),
		})
		// Spot check one restored secret
		sec, err := suite.SecretsService.GetSecret("perf-secret-000")
		require.NoError(t, err)
		val, err := suite.SecretsService.GetSecretValue(sec)
		require.NoError(t, err)
		require.Equal(t, origValue, val)
	})
}

func TestExportUnsupportedSecretTypes(t *testing.T) {
	reporting.WithReporting(t, "TestExportUnsupportedSecretTypes", func(reporter *reporting.TestWrapper) {
		// suite := SetupSuite(reporter)
		// testDataManager := testdata.NewTestDataManager()
		// TODO: Implement test for unsupported secret types
	})
}

func TestLogIntegrityForSharingActions(t *testing.T) {
	reporting.WithReporting(t, "TestLogIntegrityForSharingActions", func(reporter *reporting.TestWrapper) {
		// suite := SetupSuite(reporter)
		// testDataManager := testdata.NewTestDataManager()
		// TODO: Implement test for log integrity (all actions, including failures)
	})
}

func TestMultipleImportsOfSameBundle(t *testing.T) {
	reporting.WithReporting(t, "TestMultipleImportsOfSameBundle", func(reporter *reporting.TestWrapper) {
		suite := SetupSuite(reporter)
		testDataManager := testdata.NewTestDataManager()
		secret := testdata.TestSecrets.Simple
		require.NoError(t, testDataManager.CreateTestSecret(suite.SecretsService, secret.Name))
		file, err := suite.SecretsService.LoadAllSecrets()
		require.NoError(t, err)
		pubPEM, privPEM, err := suite.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(t, err)
		exportBundle, err := suite.SharingService.ExportSecrets(file.Secrets, pubPEM, 300)
		require.NoError(t, err)
		// Delete secret locally to prove re-import works once
		require.NoError(t, suite.SecretsService.DeleteSecret(secret.Name))
		// First import should succeed
		res1, err1 := suite.SharingService.ImportSecrets(exportBundle, privPEM, pubPEM)
		require.NoError(t, err1)
		require.True(t, res1.Success)
		// Second import should be blocked (replay)
		res2, err2 := suite.SharingService.ImportSecrets(exportBundle, privPEM, pubPEM)
		require.NoError(t, err2)
		require.False(t, res2.Success)
		require.Error(t, res2.Error)
		require.Contains(t, res2.Error.Error(), "replay")
	})
}
