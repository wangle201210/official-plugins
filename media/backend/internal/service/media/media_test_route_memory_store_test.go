// This file provides a route memory store test double for media service unit tests.

package media

import (
	"context"
	"time"

	"lina-core/pkg/pluginservice/contract"
)

// memoryRouteMemoryCache records route-memory host cache operations in memory for service tests.
type memoryRouteMemoryCache struct {
	items         map[string]string
	lastNamespace string
	lastKey       string
	lastTTL       time.Duration
}

// newMemoryRouteMemoryCache creates an empty route memory cache test double.
func newMemoryRouteMemoryCache() *memoryRouteMemoryCache {
	return &memoryRouteMemoryCache{items: make(map[string]string)}
}

// Get returns one cached route memory value.
func (s *memoryRouteMemoryCache) Get(_ context.Context, namespace string, key string) (*contract.CacheItem, bool, error) {
	s.lastNamespace = namespace
	s.lastKey = key
	value, ok := s.items[namespace+"\x00"+key]
	if !ok {
		return nil, false, nil
	}
	return &contract.CacheItem{Key: key, ValueKind: contract.CacheValueKindString, Value: value}, true, nil
}

// Set records one route memory value.
func (s *memoryRouteMemoryCache) Set(_ context.Context, namespace string, key string, value string, ttl time.Duration) (*contract.CacheItem, error) {
	s.lastNamespace = namespace
	s.lastKey = key
	s.lastTTL = ttl
	s.items[namespace+"\x00"+key] = value
	return &contract.CacheItem{Key: key, ValueKind: contract.CacheValueKindString, Value: value}, nil
}

// Delete removes one in-memory route memory value.
func (s *memoryRouteMemoryCache) Delete(_ context.Context, namespace string, key string) error {
	s.lastNamespace = namespace
	s.lastKey = key
	delete(s.items, namespace+"\x00"+key)
	return nil
}
