package sharing

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReplayStore defines storage for processed bundle IDs to prevent replays.
// Implementations should be concurrency-safe.
type ReplayStore interface {
	Seen(id string) bool // returns true if ID currently recorded (not expired)
	Mark(id string, ttl time.Duration)
	SetTTL(ttl time.Duration) // adjust default TTL
}

// ReplayMode enumerates available store types.
type ReplayMode string

const (
	ReplayModeMemory ReplayMode = "memory"
	ReplayModeFile   ReplayMode = "file"
)

// ReplayConfig centralizes simple, dependency-free configuration.
type ReplayConfig struct {
	Mode          ReplayMode    // memory | file
	TTL           time.Duration // default 1h if zero
	FilePath      string        // required if Mode = file
	MaxEntries    int           // 0 = unlimited
	FlushInterval time.Duration // for file store periodic flush (0 disables)
}

func (c *ReplayConfig) normalize() {
	if c.TTL <= 0 {
		c.TTL = time.Hour
	}
	if c.Mode == "" {
		c.Mode = ReplayModeMemory
	}
}

// NewReplayStoreFromConfig builds a ReplayStore using only stdlib.
func NewReplayStoreFromConfig(cfg ReplayConfig) (ReplayStore, error) {
	cfg.normalize()
	switch cfg.Mode {
	case ReplayModeMemory:
		return NewInMemoryReplayStore(cfg.TTL), nil
	case ReplayModeFile:
		if cfg.FilePath == "" {
			return nil, fmt.Errorf("file path required for file mode")
		}
		store, err := NewFileReplayStore(cfg.FilePath, cfg.TTL)
		if err != nil {
			return nil, err
		}
		if cfg.MaxEntries > 0 {
			store.SetMaxEntries(cfg.MaxEntries)
		}
		if cfg.FlushInterval > 0 {
			store.EnablePeriodicFlush(cfg.FlushInterval)
		}
		return store, nil
	default:
		return nil, fmt.Errorf("unsupported replay mode: %s", cfg.Mode)
	}
}

// DefaultReplayStorePath returns an OS-appropriate location for the replay store file.
// macOS: ~/Library/Application Support/<appName>/replay_store.json
// Linux: ~/.local/share/<appName>/replay_store.json
// Windows: %APPDATA%/<appName>/replay_store.json
// Falls back to current working directory on error.
func DefaultReplayStorePath(appName string) string {
	home, _ := os.UserHomeDir()
	var base string
	switch {
	case home == "":
		if wd, err := os.Getwd(); err == nil {
			base = wd
		} else {
			base = "."
		}
	default:
		base = home
	}
	var dir string
	if isWindows() {
		if appData := os.Getenv("APPDATA"); appData != "" {
			base = appData
		}
		dir = filepath.Join(base, appName)
	} else if isMac() {
		dir = filepath.Join(base, "Library", "Application Support", appName)
	} else { // assume Linux/Unix
		dir = filepath.Join(base, ".local", "share", appName)
	}
	_ = os.MkdirAll(dir, 0o700)
	return filepath.Join(dir, "replay_store.json")
}

func isWindows() bool { return os.PathSeparator == '\\' }
func isMac() bool     { return os.Getenv("GOOS_OVERRIDE") == "darwin" }

// InMemoryReplayStore is a TTL-based in-memory implementation.
type InMemoryReplayStore struct {
	ttl   time.Duration
	items map[string]int64 // id -> expiry unix
	// simple mutex; low contention expected
	mu sync.Mutex
}

// NewInMemoryReplayStore creates a new store with provided ttl.
func NewInMemoryReplayStore(ttl time.Duration) *InMemoryReplayStore {
	return &InMemoryReplayStore{ttl: ttl, items: make(map[string]int64)}
}

func (s *InMemoryReplayStore) prune(now int64) {
	for k, exp := range s.items {
		if exp < now {
			delete(s.items, k)
		}
	}
}

// Seen returns true if ID present and not expired.
func (s *InMemoryReplayStore) Seen(id string) bool {
	now := time.Now().Unix()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(now)
	exp, ok := s.items[id]
	if !ok {
		return false
	}
	return exp >= now
}

// Mark records an ID with the provided ttl (if <=0 uses default).
func (s *InMemoryReplayStore) Mark(id string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = s.ttl
	}
	now := time.Now().Unix()
	s.mu.Lock()
	s.items[id] = now + int64(ttl.Seconds())
	s.mu.Unlock()
}

// SetTTL resets the default TTL.
func (s *InMemoryReplayStore) SetTTL(ttl time.Duration) { s.mu.Lock(); s.ttl = ttl; s.mu.Unlock() }

// FileReplayStore is a simple JSON-backed persistent replay store.
// Structure: { "id": expiryUnix, ... }
type FileReplayStore struct {
	path           string
	ttl            time.Duration
	mu             sync.Mutex
	data           map[string]int64
	maxEntries     int           // 0 == unlimited
	flushInterval  time.Duration // 0 == disable periodic flush
	stopCh         chan struct{}
	immediateFlush bool // default true
}

// NewFileReplayStore creates/loads a JSON file replay store.
func NewFileReplayStore(path string, ttl time.Duration) (*FileReplayStore, error) {
	abs := path
	if !filepath.IsAbs(abs) {
		if wd, err := os.Getwd(); err == nil {
			abs = filepath.Join(wd, path)
		}
	}
	store := &FileReplayStore{path: abs, ttl: ttl, data: make(map[string]int64), immediateFlush: true}
	if err := store.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	store.pruneLocked(time.Now().Unix())
	if err := store.flush(); err != nil { // ensure file exists
		return nil, err
	}
	return store, nil
}

func (s *FileReplayStore) load() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, &s.data)
}

func (s *FileReplayStore) flush() error {
	tmp := s.path + ".tmp"
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *FileReplayStore) pruneLocked(now int64) {
	for k, exp := range s.data {
		if exp < now {
			delete(s.data, k)
		}
	}
}

func (s *FileReplayStore) Seen(id string) bool {
	now := time.Now().Unix()
	s.mu.Lock()
	s.pruneLocked(now)
	exp, ok := s.data[id]
	s.mu.Unlock()
	if !ok {
		return false
	}
	return exp >= now
}

func (s *FileReplayStore) Mark(id string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = s.ttl
	}
	now := time.Now().Unix()
	s.mu.Lock()
	s.data[id] = now + int64(ttl.Seconds())
	s.pruneLocked(now)
	// size cap enforcement (simple oldest-prune strategy)
	if s.maxEntries > 0 && len(s.data) > s.maxEntries {
		// collect keys & expiry then prune oldest
		oldestKey := ""
		oldestExp := int64(1 << 62)
		for k, exp := range s.data {
			if exp < oldestExp {
				oldestExp = exp
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(s.data, oldestKey)
		}
	}
	if s.immediateFlush {
		_ = s.flush()
	} // best effort
	s.mu.Unlock()
}

func (s *FileReplayStore) SetTTL(ttl time.Duration) { s.mu.Lock(); s.ttl = ttl; s.mu.Unlock() }
func (s *FileReplayStore) SetImmediateFlush(v bool) { s.mu.Lock(); s.immediateFlush = v; s.mu.Unlock() }
func (s *FileReplayStore) Flush()                   { s.mu.Lock(); _ = s.flush(); s.mu.Unlock() }

// SetMaxEntries configures an upper bound on retained entries (0 = unlimited)
func (s *FileReplayStore) SetMaxEntries(n int) { s.mu.Lock(); s.maxEntries = n; s.mu.Unlock() }

// EnablePeriodicFlush starts a background goroutine writing the file at interval.
func (s *FileReplayStore) EnablePeriodicFlush(interval time.Duration) {
	if interval <= 0 {
		return
	}
	s.mu.Lock()
	if s.stopCh != nil {
		s.mu.Unlock()
		return
	} // already running
	s.flushInterval = interval
	stop := make(chan struct{})
	s.stopCh = stop
	s.mu.Unlock()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				_ = s.flush()
				s.mu.Unlock()
			case <-stop:
				return
			}
		}
	}()
}

// Shutdown stops periodic flush if enabled and performs a final flush.
func (s *FileReplayStore) Shutdown() {
	s.mu.Lock()
	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}
	_ = s.flush()
	s.mu.Unlock()
}
