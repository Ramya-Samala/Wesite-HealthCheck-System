package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

// All returns a sorted copy of all checks — sorted by endpoint alphabetically
func (s *Store) All() []HealthCheck {
	s.lock.RLock()
	defer s.lock.RUnlock()

	out := make([]HealthCheck, 0, len(s.checks))
	for _, c := range s.checks {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Endpoint < out[j].Endpoint
	})
	return out
}

func (s *Store) Find(id string) (HealthCheck, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	c, ok := s.checks[id]
	if !ok {
		return HealthCheck{}, false
	}
	return *c, true
}

// HasEndpoint checks if a URL is already being monitored
func (s *Store) HasEndpoint(ep string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, c := range s.checks {
		if c.Endpoint == ep {
			return true
		}
	}
	return false
}

func (s *Store) Put(hc *HealthCheck) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.checks[hc.ID] = hc
	return s.save()
}

func (s *Store) Update(hc *HealthCheck) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.checks[hc.ID]; !ok {
		return fmt.Errorf("check %s does not exist", hc.ID)
	}
	s.checks[hc.ID] = hc
	return s.save()
}

func (s *Store) Remove(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.checks[id]; !ok {
		return fmt.Errorf("check %s does not exist", id)
	}
	delete(s.checks, id)
	return s.save()
}
