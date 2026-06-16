// Package cron schedules media plugin maintenance jobs.
package cron

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/pkg/logger"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

const mediaReportCleanupPattern = "0 0 3 * * 0"

var addSingletonCron = gcron.AddSingleton

// Service defines the media plugin cron contract.
type Service interface {
	// Start registers media maintenance jobs and returns without blocking; registration errors are returned to startup.
	Start(ctx context.Context) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	mediaSvc mediasvc.Service // mediaSvc owns cleanup business logic.
	mu       sync.Mutex       // mu protects start state across repeated startup hooks.
	started  bool             // started prevents duplicate registrations after a successful start.
}

// New creates a media cron service.
func New(mediaSvc mediasvc.Service) (Service, error) {
	if mediaSvc == nil {
		return nil, gerror.New("media cron service requires media service")
	}
	return &serviceImpl{mediaSvc: mediaSvc}, nil
}

// Start registers media maintenance jobs and returns without blocking.
func (s *serviceImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	if _, err := addSingletonCron(ctx, mediaReportCleanupPattern, func(ctx context.Context) {
		s.cleanupClosedReports(ctx)
	}, "media-report-cleanup"); err != nil {
		return err
	}
	s.started = true
	return nil
}

// cleanupClosedReports runs one retention cleanup pass.
func (s *serviceImpl) cleanupClosedReports(ctx context.Context) {
	if s == nil || s.mediaSvc == nil {
		return
	}
	cleaned, err := s.mediaSvc.CleanupClosedReports(ctx, time.Now())
	if err != nil {
		logger.Warningf(ctx, "media report cleanup failed: %v", err)
		return
	}
	if cleaned > 0 {
		logger.Infof(ctx, "media report cleanup removed %d closed rows", cleaned)
	}
}
