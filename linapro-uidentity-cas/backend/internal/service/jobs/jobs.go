// Package jobs registers and executes UIdentity CAS plugin scheduled-job
// handlers through LinaPro's built-in task-management module. It does not own
// cron scheduling state or plugin-local job tables; the host scheduler decides
// when handlers run and this package only performs plugin business work.
package jobs

import (
	"context"

	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/pluginhost"
)

// Service defines UIdentity managed scheduled-job handler registration and
// execution. Register contributes all old uidentity/admin app/jobs entries to
// LinaPro task management; individual Exec methods return plugin business
// errors for missing external Oracle or LDAP configuration.
type Service interface {
	// Register contributes all old uidentity/admin job registry entries to the
	// host task-management registry. It requires a host registrar and returns
	// validation or registration errors before any task is persisted by the host.
	Register(ctx context.Context, registrar pluginhost.CronRegistrar) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc    plugincontract.BizCtxService
	configSvc    plugincontract.ConfigService
	tenantFilter plugincontract.TenantFilterService
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates a UIdentity scheduled-job handler service with explicit host
// capability dependencies supplied by the source-plugin registrar.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	configSvc plugincontract.ConfigService,
	tenantFilter plugincontract.TenantFilterService,
) Service {
	return &serviceImpl{
		bizCtxSvc:    bizCtxSvc,
		configSvc:    configSvc,
		tenantFilter: tenantFilter,
	}
}
