// Package collection implements the media plugin TCP data collection server.
package collection

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/plugincap"
)

// Service defines the media collection server contract.
type Service interface {
	// Start loads plugin configuration and starts the collection server when enabled.
	// A disabled configuration is a no-op. Calling Start more than once is idempotent;
	// after the server has been started, later calls return without opening another
	// listener. Startup errors are returned when configuration is invalid or the
	// listen address cannot be bound. Enabled servers require the host-published
	// shared cache service because lifecycle counters must be visible across pods.
	Start(ctx context.Context, configSvc plugincap.ConfigService, cacheSvc cachecap.Service) error
}

// Interface compliance assertion for the default collection service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	mu      sync.Mutex // mu protects started.
	started bool       // started records whether the TCP server has already been launched.
}

// New creates and returns a media collection service.
func New() Service {
	return &serviceImpl{}
}

// Start loads configuration and starts the collection server when enabled.
func (s *serviceImpl) Start(ctx context.Context, configSvc plugincap.ConfigService, cacheSvc cachecap.Service) error {
	if configSvc == nil {
		return gerror.New("media collection server requires host plugin config service")
	}
	cfg, err := LoadConfig(ctx, configSvc)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		logger.Infof(ctx, "media collection server disabled")
		return nil
	}
	if cacheSvc == nil {
		return gerror.New("media collection server requires host cache service")
	}
	return s.startConfigured(ctx, cfg, cacheSvc)
}

// startConfigured starts the TCP server once for an enabled configuration.
func (s *serviceImpl) startConfigured(ctx context.Context, cfg *Config, cacheSvc cachecap.Service) error {
	if cfg == nil {
		return gerror.New("media collection server config cannot be nil")
	}
	if cacheSvc == nil {
		return gerror.New("media collection server requires host cache service")
	}
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		logger.Infof(ctx, "media collection server already started addr=%s", cfg.Addr)
		return nil
	}
	s.started = true
	s.mu.Unlock()

	if err := runServer(ctx, cfg, cacheSvc); err != nil {
		s.mu.Lock()
		s.started = false
		s.mu.Unlock()
		return err
	}
	logger.Infof(ctx, "media collection server started addr=%s", cfg.Addr)
	return nil
}
