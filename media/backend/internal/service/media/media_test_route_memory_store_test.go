// This file provides a route memory store test double for media service unit tests.

package media

import (
	"context"
	"time"
)

// memoryRouteMemoryStore records route-memory operations in memory for service tests.
type memoryRouteMemoryStore struct {
	items   map[string]string
	lastKey string
	lastTTL time.Duration
}

// newMemoryRouteMemoryStore creates an empty route memory store test double.
func newMemoryRouteMemoryStore() *memoryRouteMemoryStore {
	return &memoryRouteMemoryStore{items: make(map[string]string)}
}

// Set records one route memory value.
func (s *memoryRouteMemoryStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	s.lastKey = key
	s.lastTTL = ttl
	s.items[key] = value
	return nil
}

// Get reads one in-memory route memory value.
func (s *memoryRouteMemoryStore) Get(ctx context.Context, key string) (string, error) {
	s.lastKey = key
	return s.items[key], nil
}

// Delete removes one in-memory route memory value.
func (s *memoryRouteMemoryStore) Delete(ctx context.Context, key string) error {
	s.lastKey = key
	delete(s.items, key)
	return nil
}
