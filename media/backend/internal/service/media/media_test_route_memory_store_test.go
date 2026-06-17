// This file provides a route memory store test double for media service unit tests.

package media

import (
	"context"
	"fmt"
	"time"

	"lina-core/pkg/plugin/capability/cachecap"
)

// memoryRouteMemoryCache records route-memory host cache operations in memory for service tests.
type memoryRouteMemoryCache struct {
	items         map[string]*cachecap.CacheItem
	lastNamespace string
	lastKey       string
	lastTTL       time.Duration
}

// newMemoryRouteMemoryCache creates an empty route memory cache test double.
func newMemoryRouteMemoryCache() *memoryRouteMemoryCache {
	return &memoryRouteMemoryCache{items: make(map[string]*cachecap.CacheItem)}
}

// Get returns one cached route memory value.
func (s *memoryRouteMemoryCache) Get(_ context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
	s.lastNamespace = namespace
	s.lastKey = key
	item, ok := s.items[namespace+"\x00"+key]
	if !ok {
		return nil, false, nil
	}
	copied := *item
	copied.Key = key
	return &copied, true, nil
}

// Set records one route memory value.
func (s *memoryRouteMemoryCache) Set(_ context.Context, namespace string, key string, value string, ttl time.Duration) (*cachecap.CacheItem, error) {
	s.lastNamespace = namespace
	s.lastKey = key
	s.lastTTL = ttl
	item := &cachecap.CacheItem{Key: key, ValueKind: cachecap.CacheValueKindString, Value: value}
	s.items[namespace+"\x00"+key] = item
	return item, nil
}

// Delete removes one in-memory route memory value.
func (s *memoryRouteMemoryCache) Delete(_ context.Context, namespace string, key string) error {
	s.lastNamespace = namespace
	s.lastKey = key
	delete(s.items, namespace+"\x00"+key)
	return nil
}

// Incr is implemented to satisfy the host cache contract in media service tests.
func (s *memoryRouteMemoryCache) Incr(_ context.Context, namespace string, key string, delta int64, ttl time.Duration) (*cachecap.CacheItem, error) {
	s.lastNamespace = namespace
	s.lastKey = key
	s.lastTTL = ttl
	next := delta
	if item, ok := s.items[namespace+"\x00"+key]; ok {
		if item.ValueKind != cachecap.CacheValueKindInt {
			return nil, fmt.Errorf("cache item %s/%s is not an integer", namespace, key)
		}
		next = item.IntValue + delta
	}
	item := &cachecap.CacheItem{Key: key, ValueKind: cachecap.CacheValueKindInt, IntValue: next}
	s.items[namespace+"\x00"+key] = item
	return item, nil
}

// Expire is implemented to satisfy the host cache contract in media service tests.
func (s *memoryRouteMemoryCache) Expire(_ context.Context, namespace string, key string, ttl time.Duration) (bool, *time.Time, error) {
	s.lastNamespace = namespace
	s.lastKey = key
	s.lastTTL = ttl
	return true, nil, nil
}
