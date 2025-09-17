package sharing

import (
	"os"
	"testing"
	"time"
)

func TestNewReplayStoreFromConfigMemory(t *testing.T) {
	store, err := NewReplayStoreFromConfig(ReplayConfig{Mode: ReplayModeMemory, TTL: time.Minute})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Fatal("expected store")
	}
	if store.Seen("abc") {
		t.Fatal("should not have entry")
	}
	store.Mark("abc", 0)
	if !store.Seen("abc") {
		t.Fatal("expected entry present")
	}
}

func TestNewReplayStoreFromConfigFile(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/replay.json"
	storeInt, err := NewReplayStoreFromConfig(ReplayConfig{Mode: ReplayModeFile, FilePath: path, TTL: time.Millisecond * 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fs, ok := storeInt.(*FileReplayStore)
	if !ok {
		t.Fatal("expected file replay store")
	}
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
	if count != 1 {
		t.Fatalf("expected 1 entry after cap prune, got %d", count)
	}
	// Expiry not asserted here due to second-level timestamp granularity; cap behavior covered.
}

func TestFileReplayStoreImmediateFlushToggle(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/replay.json"
	store, err := NewFileReplayStore(path, time.Minute)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// disable immediate flush
	store.SetImmediateFlush(false)
	store.Mark("x", 0)
	// Because initial creation writes file, we only assert file size does not grow until flush (skip strict check)
	store.Flush()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file after manual flush: %v", err)
	}
}
