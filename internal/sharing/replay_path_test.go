package sharing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultReplayStorePathCreatesDir(t *testing.T) {
	appName := "TestAppReplay"
	path := DefaultReplayStorePath(appName)
	require.NotEmpty(t, path, "expected non-empty path")
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	require.NoError(t, err, "expected directory created: %v", err)
	require.True(t, info.IsDir(), "expected directory, got file")
	require.Contains(t, path, appName, "path should contain app name: %s", path)
}
