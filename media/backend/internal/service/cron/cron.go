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

// Service defines the media plugin cron contract.
type Service interface {
	// Start registers media maintenance jobs and returns without blocking.
	Start(ctx context.Context)
}

// serviceImpl implements Service.
type serviceImpl struct {
	mediaSvc mediasvc.Service // mediaSvc owns cleanup business logic.
	once     sync.Once        // once prevents duplicate registrations across repeated startup hooks.
}

// New creates a media cron service.
func New(mediaSvc mediasvc.Service) (Service, error) {
	if mediaSvc == nil {
		return nil, gerror.New("media cron service requires media service")
	}
	return &serviceImpl{mediaSvc: mediaSvc}, nil
}

// Start registers media maintenance jobs and returns without blocking.
func (s *serviceImpl) Start(ctx context.Context) {
	s.once.Do(func() {
		if _, err := gcron.AddSingleton(ctx, mediaReportCleanupPattern, func(ctx context.Context) {
			s.cleanupClosedReports(ctx)
		}, "media-report-cleanup"); err != nil {
			logger.Warningf(ctx, "failed to start media report cleanup cron: %v", err)
		}
	})
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
