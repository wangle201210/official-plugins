// This file tests water plugin runtime configuration loading.

package water

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/encoding/gjson"
)

// TestLoadRuntimeConfigReadsPluginConsumerCount verifies worker count comes from plugin config.
func TestLoadRuntimeConfigReadsPluginConsumerCount(t *testing.T) {
	ctx := context.Background()
	config, err := LoadRuntimeConfig(ctx, waterTestConfig{
		values: map[string]any{
			"water": map[string]any{
				"consumerCount": 3,
			},
		},
	})
	if err != nil {
		t.Fatalf("load runtime config: %v", err)
	}
	if config.ConsumerCount != 3 {
		t.Fatalf("expected plugin consumer count 3, got %d", config.ConsumerCount)
	}
}

// TestLoadRuntimeConfigBoundsConsumerCount verifies invalid and excessive counts are normalized.
func TestLoadRuntimeConfigBoundsConsumerCount(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		raw  int
		want int
	}{
		{name: "zero", raw: 0, want: defaultConsumerCount},
		{name: "negative", raw: -1, want: defaultConsumerCount},
		{name: "max", raw: maxConsumerCount, want: maxConsumerCount},
		{name: "excessive", raw: maxConsumerCount + 1, want: maxConsumerCount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := LoadRuntimeConfig(ctx, waterTestConfig{
				values: map[string]any{
					"water": map[string]any{
						"consumerCount": tt.raw,
					},
				},
			})
			if err != nil {
				t.Fatalf("load runtime config: %v", err)
			}
			if config.ConsumerCount != tt.want {
				t.Fatalf("expected consumer count %d, got %d", tt.want, config.ConsumerCount)
			}
		})
	}
}

type waterTestConfig struct {
	values map[string]any
}

// Get returns one raw plugin config value.
func (s waterTestConfig) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value := gjson.New(s.values).Get(key)
	if value == nil || value.IsNil() {
		if defaultValue == nil {
			return nil, nil
		}
		return gvar.New(defaultValue), nil
	}
	return value, nil
}

// Exists reports whether one plugin config key is present.
func (s waterTestConfig) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	return value != nil && !value.IsNil(), err
}

// Scan scans one plugin config section into target.
func (s waterTestConfig) Scan(ctx context.Context, key string, target any) error {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return err
	}
	return value.Scan(target)
}

// String reads one string plugin config value.
func (s waterTestConfig) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads one bool plugin config value.
func (s waterTestConfig) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads one int plugin config value.
func (s waterTestConfig) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration reads one duration plugin config value.
func (s waterTestConfig) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	duration, err := time.ParseDuration(value.String())
	if err != nil {
		return 0, err
	}
	return duration, nil
}
