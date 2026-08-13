// cron_register.go declares the managed sicau-niu scheduled-job entries.
package cron

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
)

// Register contributes the cloud-moving expiry job and optional locator refresh job.
func (s *serviceImpl) Register(ctx context.Context, registrar pluginhost.JobsRegistrar) error {
	if registrar == nil {
		return gerror.New("sicau-niu cron requires jobs registrar")
	}
	if err := registrar.AddWithMetadata(
		ctx,
		transportExpiryPattern,
		transportExpiryName,
		transportExpiryDisplayName,
		transportExpiryDescription,
		s.primaryOnly(registrar, s.expireInactiveTransportTeams),
	); err != nil {
		return err
	}
	if !s.ironRefreshEnabled {
		return nil
	}
	return registrar.AddWithMetadata(
		ctx,
		"@every "+s.ironRefreshInterval.String(),
		ironLocationRefreshName,
		ironLocationRefreshDisplayName,
		ironLocationRefreshDescription,
		s.primaryOnly(registrar, s.refreshIronLocations),
	)
}

func (s *serviceImpl) primaryOnly(registrar pluginhost.JobsRegistrar, handler pluginhost.JobHandler) pluginhost.JobHandler {
	return func(ctx context.Context) error {
		if !registrar.IsPrimaryNode() {
			return nil
		}
		return handler(ctx)
	}
}

func (s *serviceImpl) expireInactiveTransportTeams(ctx context.Context) error {
	_, err := s.transport.ExpireInactive(ctx, time.Time{})
	return err
}

func (s *serviceImpl) refreshIronLocations(ctx context.Context) error {
	_, err := s.ironRefresher.Refresh(ctx)
	return err
}
