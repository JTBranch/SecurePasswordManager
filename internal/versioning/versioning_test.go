package versioning

import (
	"go-password-manager/tests/helpers"
	"testing"
	"time"
)

const SecretValue = "<secret_value>"

func setupVersionsAndReturnSecret(tc *helpers.UnitTestCase, versioning *SecretVersioning) string {
	// Simulate a versioning scenario

	secretKey := tc.TestData.GenerateUniqueSecretName("TestVersioning")
	versioning.AddVersion(secretKey, SecretValue, time.Now().Unix())

	version, err := versioning.GetLatestVersion(secretKey)

	tc.Assert.Nil(err)
	tc.Assert.Equal(SecretValue, version.Value)
	return secretKey
}

func TestVersioning(t *testing.T) {

	helpers.WithUnitTestCase(t, "Can Add a version and retrieve it", func(tc *helpers.UnitTestCase) {
		versioning := NewSecretVersioning()
		setupVersionsAndReturnSecret(tc, versioning)
	})

	helpers.WithUnitTestCase(t, "Can Update and List Versions", func(tc *helpers.UnitTestCase) {
		versioning := NewSecretVersioning()
		secretKey := setupVersionsAndReturnSecret(tc, versioning)

		// Update the version
		newSecretValue := "<new_secret_value>"
		versioning.AddVersion(secretKey, newSecretValue, time.Now().Unix())

		allVersions, err := versioning.ListVersions(secretKey)

		tc.Assert.Nil(err)
		tc.Assert.Equal(len(allVersions), 2)
		tc.Assert.Equal(SecretValue, allVersions[0].Value)
		tc.Assert.Equal(newSecretValue, allVersions[1].Value)
	})

	helpers.WithUnitTestCase(t, "Can Get Previous Version", func(tc *helpers.UnitTestCase) {
		versioning := NewSecretVersioning()
		secretKey := setupVersionsAndReturnSecret(tc, versioning)

		// Update the version
		newSecretValue := "<new_secret_value>"
		versioning.AddVersion(secretKey, newSecretValue, time.Now().Unix())

		oldVersion, err := versioning.GetVersion(secretKey, 0)

		tc.Assert.Nil(err)
		tc.Assert.Equal(SecretValue, oldVersion.Value)
	})
}
