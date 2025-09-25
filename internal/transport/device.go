package transport

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DeviceDescriptor captures minimal identifying & crypto routing information about a device.
type DeviceDescriptor struct {
	DeviceID   string    `json:"device_id"`
	UserID     string    `json:"user_id"`
	DeviceName string    `json:"device_name"`
	Ed25519Pub []byte    `json:"ed25519_pub"`
	X25519Pub  []byte    `json:"x25519_pub"`
	LastSeenAt time.Time `json:"last_seen_at"`
	LastAddr   string    `json:"last_addr"`
}

// DeviceRegistry provides lookup & persistence for devices.
type DeviceRegistry interface {
	Upsert(d DeviceDescriptor) error
	Get(deviceID string) (DeviceDescriptor, bool)
	List() []DeviceDescriptor
}

// InMemoryRegistry is a thread-safe registry stored in memory.
type InMemoryRegistry struct {
	mu   sync.RWMutex
	data map[string]DeviceDescriptor
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{data: map[string]DeviceDescriptor{}}
}

func (r *InMemoryRegistry) Upsert(d DeviceDescriptor) error {
	r.mu.Lock()
	d.LastSeenAt = time.Now()
	r.data[d.DeviceID] = d
	r.mu.Unlock()
	return nil
}

func (r *InMemoryRegistry) Get(id string) (DeviceDescriptor, bool) {
	r.mu.RLock()
	d, ok := r.data[id]
	r.mu.RUnlock()
	return d, ok
}

func (r *InMemoryRegistry) List() []DeviceDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]DeviceDescriptor, 0, len(r.data))
	for _, d := range r.data {
		out = append(out, d)
	}
	return out
}

// FileRegistry persists registry entries to a JSON file (atomic write via temp rename).
type FileRegistry struct {
	path string
	mem  *InMemoryRegistry
	mu   sync.Mutex
}

func NewFileRegistry(path string) (*FileRegistry, error) {
	if path == "" {
		return nil, errors.New("file registry path required")
	}
	fr := &FileRegistry{path: path, mem: NewInMemoryRegistry()}
	if err := fr.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return fr, nil
}

func (r *FileRegistry) Upsert(d DeviceDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.mem.Upsert(d); err != nil {
		return err
	}
	return r.flush()
}
func (r *FileRegistry) Get(id string) (DeviceDescriptor, bool) { return r.mem.Get(id) }
func (r *FileRegistry) List() []DeviceDescriptor               { return r.mem.List() }

func (r *FileRegistry) load() error {
	b, err := os.ReadFile(r.path)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	var arr []DeviceDescriptor
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	for _, d := range arr {
		_ = r.mem.Upsert(d)
	}
	return nil
}

func (r *FileRegistry) flush() error {
	dir := filepath.Dir(r.path)
	_ = os.MkdirAll(dir, 0o700)
	arr := r.mem.List()
	b, err := json.MarshalIndent(arr, "", "  ")
	if err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

// end of file
