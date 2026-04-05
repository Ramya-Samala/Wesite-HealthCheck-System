package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const savePath = "healthchecks.json"

// Store keeps checks in memory and syncs to disk on every change
type Store struct {
	lock   sync.RWMutex
	checks map[string]*HealthCheck
}

func NewStore() *Store {
	return &Store{checks: make(map[string]*HealthCheck)}
}

// Load reads saved checks from disk into memory
func (s *Store) Load() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	data, err := os.ReadFile(savePath)
	if err != nil {
		if os.IsNotExist(err) {
			// first run, nothing saved yet — that's fine
			return nil
		}
		return fmt.Errorf("reading %s: %w", savePath, err)
	}

	var items []HealthCheck
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("parsing %s: %w", savePath, err)
	}

	for idx := range items {
		c := items[idx]
		s.checks[c.ID] = &c
	}
	return nil
}

// save writes everything to disk — caller must hold the write lock
func (s *Store) save() error {
	list := make([]HealthCheck, 0, len(s.checks))
	for _, c := range s.checks {
		list = append(list, *c)
	}
	raw, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(savePath, raw, 0644)
}
