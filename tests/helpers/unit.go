package helpers

import (
	"go-password-manager/tests/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UnitTestCase provides a consistent structure for unit tests.
type UnitTestCase struct {
	Assert   *assert.Assertions
	Require  *require.Assertions
	TestData *testdata.TestDataManager
}

// WithUnitTestCase is a runner function to ensure consistent test setup.
func WithUnitTestCase(t *testing.T, testName string, testFunc func(tc *UnitTestCase)) {
	t.Run(testName, func(t *testing.T) {
		tc := &UnitTestCase{
			Assert:   assert.New(t),
			Require:  require.New(t),
			TestData: testdata.NewTestDataManager(),
		}
		testFunc(tc)
	})
}
