// Package jobs registers and executes UIdentity CAS plugin scheduled-job
// handlers through LinaPro's built-in task-management module. It does not own
// scheduler state or plugin-local job tables; the host scheduler decides
// when handlers run and this package only performs plugin business work.
package jobs

import (
	"context"
	"sync"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap"
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
	Register(ctx context.Context, registrar pluginhost.JobsRegistrar) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc    bizctxcap.Service
	configSvc    plugincap.ConfigService
	tenantFilter tenantcap.FilterService

	// running tracks per-job in-process execution so a slow run is never
	// overlapped by the next trigger on this node, mirroring the old per-job
	// Redis lock's skip-if-held behavior.
	running sync.Map
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates a UIdentity scheduled-job handler service with explicit host
// capability dependencies supplied by the source-plugin registrar.
func New(
	bizCtxSvc bizctxcap.Service,
	configSvc plugincap.ConfigService,
	tenantFilter tenantcap.FilterService,
) Service {
	return &serviceImpl{
		bizCtxSvc:    bizCtxSvc,
		configSvc:    configSvc,
		tenantFilter: tenantFilter,
	}
}
