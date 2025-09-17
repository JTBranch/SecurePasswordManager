package sharing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultReplayStorePathCreatesDir(t *testing.T) {
	appName := "TestAppReplay"
	path := DefaultReplayStorePath(appName)
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory, got file")
	}
	if !strings.Contains(path, appName) {
		t.Fatalf("path should contain app name: %s", path)
	}
}
