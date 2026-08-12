// Package cron owns the scheduled jobs contributed by the sicau-niu plugin.
package cron

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

const (
	transportExpiryPattern     = "@every 1m"
	transportExpiryName        = "sicau-niu-cloud-moving-expiry"
	transportExpiryDisplayName = "Sicau Niu Cloud-Moving Expiry"
	transportExpiryDescription = "Invalidates cloud-moving teams after 72 hours without a successful create, join or contribution."

	ironLocationRefreshName        = "sicau-niu-iron-location-refresh"
	ironLocationRefreshDisplayName = "Sicau Niu Iron Location Refresh"
	ironLocationRefreshDescription = "Refreshes registered iron-cow locator coordinates from the IOT positioning platform."
)

// Config carries optional locator-refresh scheduling settings.
type Config struct {
	IronLocationEnabled  bool
	IronLocationInterval time.Duration
}

// Service registers and executes plugin-owned scheduled jobs.
type Service interface {
	Register(ctx context.Context, registrar pluginhost.JobsRegistrar) error
}

type serviceImpl struct {
	transport           irontransportsvc.Service
	ironRefresher       feedingsvc.IronLocationRefresher
	ironRefreshEnabled  bool
	ironRefreshInterval time.Duration
}

var _ Service = (*serviceImpl)(nil)

// New validates the job dependencies and creates the plugin cron service.
func New(transport irontransportsvc.Service, ironRefresher feedingsvc.IronLocationRefresher, config Config) (Service, error) {
	if transport == nil {
		return nil, gerror.New("sicau-niu cron requires cloud-moving service")
	}
	if config.IronLocationEnabled && (ironRefresher == nil || config.IronLocationInterval <= 0) {
		return nil, gerror.New("sicau-niu cron requires enabled iron-location refresh dependencies")
	}
	return &serviceImpl{
		transport: transport, ironRefresher: ironRefresher,
		ironRefreshEnabled:  config.IronLocationEnabled,
		ironRefreshInterval: config.IronLocationInterval,
	}, nil
}
