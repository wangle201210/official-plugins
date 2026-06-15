// This file verifies media collection server startup behavior.

package collection

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/pkg/network"
	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability/cachecap"
)

// TestStartRequiresCacheWhenEnabled verifies enabled collection server needs shared host cache.
func TestStartRequiresCacheWhenEnabled(t *testing.T) {
	err := New().Start(context.Background(), staticConfigService{
		values: map[string]any{
			configKeyCollectionServerEnabled: true,
			configKeyCollectionServerAddr:    "127.0.0.1:0",
		},
	}, nil)
	if err == nil {
		t.Fatal("expected enabled collection server to require host cache service")
	}
}

// TestStartSkipsDisabledConfig verifies disabled configuration does not open a server.
func TestStartSkipsDisabledConfig(t *testing.T) {
	restore := replaceTCPServerFactory(t, func(_ string, _ network.EventHandler) tcpServerRunner {
		t.Fatal("disabled collection server must not create TCP server")
		return nil
	})
	defer restore()

	if err := New().Start(context.Background(), staticConfigService{
		values: map[string]any{
			configKeyCollectionServerEnabled: false,
		},
	}, nil); err != nil {
		t.Fatalf("start disabled collection server: %v", err)
	}
}

// TestStartIsIdempotent verifies multiple Start calls only launch one server.
func TestStartIsIdempotent(t *testing.T) {
	var (
		mu          sync.Mutex
		createCount int
	)
	restore := replaceTCPServerFactory(t, func(_ string, _ network.EventHandler) tcpServerRunner {
		mu.Lock()
		createCount++
		mu.Unlock()
		return fakeRunner{run: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}}
	})
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := New()
	configSvc := staticConfigService{
		values: map[string]any{
			configKeyCollectionServerEnabled: true,
			configKeyCollectionServerAddr:    "127.0.0.1:0",
		},
	}
	if err := svc.Start(ctx, configSvc, newMemoryCollectionCache()); err != nil {
		t.Fatalf("first start collection server: %v", err)
	}
	if err := svc.Start(ctx, configSvc, newMemoryCollectionCache()); err != nil {
		t.Fatalf("second start collection server: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if createCount != 1 {
		t.Fatalf("expected one server creation, got %d", createCount)
	}
}

// replaceTCPServerFactory swaps the TCP server constructor for one test case.
func replaceTCPServerFactory(
	t *testing.T,
	factory func(addr string, handler network.EventHandler) tcpServerRunner,
) func() {
	t.Helper()

	original := newTCPServer
	newTCPServer = factory
	return func() {
		newTCPServer = original
	}
}

// fakeRunner records one fake TCP server run.
type fakeRunner struct {
	run func(ctx context.Context) error
}

// Run executes the fake runner callback.
func (r fakeRunner) Run(ctx context.Context) error {
	return r.run(ctx)
}

// staticConfigService is a minimal plugincap.ConfigService test double.
type staticConfigService struct {
	values map[string]any
}

// Get returns the raw test config value.
func (s staticConfigService) Get(_ context.Context, key string) (*gvar.Var, error) {
	value, ok := s.values[key]
	if !ok {
		return nil, nil
	}
	return gvar.New(value), nil
}

// Exists reports whether the test config key exists.
func (s staticConfigService) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key)
	return value != nil, err
}

// Scan is a no-op because tests provide flattened keys.
func (s staticConfigService) Scan(_ context.Context, _ string, _ any) error {
	return nil
}

// String reads one string config value.
func (s staticConfigService) String(_ context.Context, key string, defaultValue string) (string, error) {
	value, ok := s.values[key]
	if !ok {
		return defaultValue, nil
	}
	if text, ok := value.(string); ok {
		return text, nil
	}
	return defaultValue, nil
}

// Bool reads one bool config value.
func (s staticConfigService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	value, ok := s.values[key]
	if !ok {
		return defaultValue, nil
	}
	if flag, ok := value.(bool); ok {
		return flag, nil
	}
	return defaultValue, nil
}

// Int reads one int config value.
func (s staticConfigService) Int(_ context.Context, key string, defaultValue int) (int, error) {
	value, ok := s.values[key]
	if !ok {
		return defaultValue, nil
	}
	if number, ok := value.(int); ok {
		return number, nil
	}
	return defaultValue, nil
}

// Duration reads one duration config value.
func (s staticConfigService) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// memoryCollectionCache is a shared cachecap.Service test double.
type memoryCollectionCache struct {
	mu     sync.Mutex
	values map[string]*cachecap.CacheItem
}

// newMemoryCollectionCache creates an in-memory shared cache test double.
func newMemoryCollectionCache() *memoryCollectionCache {
	return &memoryCollectionCache{values: make(map[string]*cachecap.CacheItem)}
}

// Get returns one stored cache item.
func (c *memoryCollectionCache) Get(_ context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.values[c.cacheKey(namespace, key)]
	if !ok {
		return nil, false, nil
	}
	copied := *item
	return &copied, true, nil
}

// Set stores one string cache item.
func (c *memoryCollectionCache) Set(
	_ context.Context,
	namespace string,
	key string,
	value string,
	_ time.Duration,
) (*cachecap.CacheItem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := &cachecap.CacheItem{Key: key, ValueKind: cachecap.CacheValueKindString, Value: value}
	c.values[c.cacheKey(namespace, key)] = item
	return item, nil
}

// Delete removes one cache item.
func (c *memoryCollectionCache) Delete(_ context.Context, namespace string, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.values, c.cacheKey(namespace, key))
	return nil
}

// Incr increments one integer cache item.
func (c *memoryCollectionCache) Incr(
	_ context.Context,
	namespace string,
	key string,
	delta int64,
	_ time.Duration,
) (*cachecap.CacheItem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheKey := c.cacheKey(namespace, key)
	var current int64
	if item, ok := c.values[cacheKey]; ok {
		switch item.ValueKind {
		case cachecap.CacheValueKindInt:
			current = item.IntValue
		case cachecap.CacheValueKindString:
			parsed, err := strconv.ParseInt(strings.TrimSpace(item.Value), 10, 64)
			if err != nil {
				return nil, err
			}
			current = parsed
		default:
			return nil, fmt.Errorf("unsupported cache item kind %d", item.ValueKind)
		}
	}
	current += delta
	item := &cachecap.CacheItem{
		Key:       key,
		ValueKind: cachecap.CacheValueKindInt,
		Value:     strconv.FormatInt(current, 10),
		IntValue:  current,
	}
	c.values[cacheKey] = item
	return item, nil
}

// Expire reports whether one cache item exists.
func (c *memoryCollectionCache) Expire(
	_ context.Context,
	namespace string,
	key string,
	_ time.Duration,
) (bool, *time.Time, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.values[c.cacheKey(namespace, key)]
	return ok, nil, nil
}

// cacheKey joins namespace and logical key for the in-memory map.
func (c *memoryCollectionCache) cacheKey(namespace string, key string) string {
	return namespace + "\x00" + key
}
