package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type Storage struct {
	mu      sync.RWMutex
	data    map[string][]byte
	filePath string
}

func NewStorage(filePath string) (*Storage, error) {
	s := &Storage{
		data:    make(map[string][]byte),
		filePath: filePath,
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) Save(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.save()
}

func (s *Storage) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return s.save()
}

func (s *Storage) save() error {
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(s.data)
}

func (s *Storage) load() error {
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(&s.data)
}