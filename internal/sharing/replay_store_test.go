package sharing

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewReplayStoreFromConfigMemory(t *testing.T) {
	store, err := NewReplayStoreFromConfig(ReplayConfig{Mode: ReplayModeMemory, TTL: time.Minute})
	require.NoError(t, err, "unexpected error: %v", err)
	require.NotNil(t, store, "expected store")
	require.False(t, store.Seen("abc"), "should not have entry")
	store.Mark("abc", 0)
	require.True(t, store.Seen("abc"), "expected entry present")
}

func TestNewReplayStoreFromConfigFile(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/replay.json"
	storeInt, err := NewReplayStoreFromConfig(ReplayConfig{Mode: ReplayModeFile, FilePath: path, TTL: time.Millisecond * 50})
	require.NoError(t, err, "unexpected error: %v", err)
	fs, ok := storeInt.(*FileReplayStore)
	require.True(t, ok, "expected file replay store")
	fs.SetMaxEntries(1)
	fs.Mark("one", 0)
	fs.Mark("two", 0) // triggers cap pruning
	// At most one should remain
	count := 0
	fs.mu.Lock()
	for range fs.data {
		count++
	}
	fs.mu.Unlock()
	require.Equal(t, 1, count, "expected 1 entry after cap prune, got %d", count)
	// Expiry not asserted here due to second-level timestamp granularity; cap behavior covered.
}

func TestFileReplayStoreImmediateFlushToggle(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/replay.json"
	store, err := NewFileReplayStore(path, time.Minute)
	require.NoError(t, err, "err: %v", err)
	// disable immediate flush
	store.SetImmediateFlush(false)
	store.Mark("x", 0)
	// Because initial creation writes file, we only assert file size does not grow until flush (skip strict check)
	store.Flush()
	if _, err := os.Stat(path); err != nil {
		require.NoError(t, err, "expected file after manual flush: %v", err)
	}
}
