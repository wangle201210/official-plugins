// Package cron owns plugin-local GoFrame scheduled jobs for the UIdentity CAS
// source plugin. It reconstructs legacy sys_job runtime entries from plugin
// tables, keeps scheduling state inside a shared gcron instance, and leaves
// external directory or system integrations behind explicit executor failures.
package cron

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gcron"

	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	// Job status values copied from the legacy sys_job table.
	jobStatusEnabled = 2
	// jobEntryRunningOffset marks one in-flight legacy job while preserving the scheduled entry ID.
	jobEntryRunningOffset = 1_000_000_000_000
	// Job type values copied from the legacy sys_job table.
	jobTypeHTTP = 1
	jobTypeExec = 2
	// Job concurrency values copied from the legacy sys_job table.
	jobConcurrentAllow = 1
	// Legacy HTTP jobs retried three times in the old implementation.
	jobHTTPRetryCount = 3
)

const (
	// defaultHTTPTimeout bounds one legacy HTTP job attempt.
	defaultHTTPTimeout = 30 * time.Second
	// legacyHTTPRetryBaseDelay preserves the old incremental retry delay shape.
	legacyHTTPRetryBaseDelay = 5 * time.Second
	// legacyJobRunLease preserves the old distributed lock lease used by Exec jobs.
	legacyJobRunLease = 50 * time.Minute
)

// legacy exec target names copied from the old jobs registry.
const (
	legacyExecTargetWannaT              = "wannat"
	legacyExecTargetContainerAccount    = "containeraccount"
	legacyExecTargetNewContainerAccount = "newcontaineraccount"
	legacyExecTargetChangeContainer     = "changecontainer"
)

// Service defines plugin-owned legacy job scheduling behavior.
type Service interface {
	// Start reloads all enabled legacy sys_job rows into the shared GoFrame cron.
	// It resets stale runtime entry IDs before scheduling, keeps plugin tables as
	// the authoritative state, and returns database or cron registration errors.
	Start(ctx context.Context, pluginState plugincontract.PluginStateService, isPrimaryNode func() bool) error
	// StartJob schedules one tenant-owned job row and persists its runtime entry
	// ID. The actor ID is optional audit metadata for request-triggered starts; a
	// zero value is used by startup reload. A disabled, malformed, or unsupported
	// job returns a structured error.
	StartJob(ctx context.Context, job *entity.SysJob, actorID int64) (int64, error)
	// RemoveJob removes one tenant-owned job from the local cron instance and
	// clears its persisted runtime entry ID. The actor ID is optional audit
	// metadata for request-triggered stops; the operation is idempotent.
	RemoveJob(ctx context.Context, tenantID int, jobID int64, actorID int64) error
	// Close stops all local cron entries owned by this plugin scheduler. It does
	// not delete plugin business rows and is safe to call more than once.
	Close(ctx context.Context) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	pluginID      string
	cron          *gcron.Cron
	httpClient    *http.Client
	pluginState   plugincontract.PluginStateService
	isPrimaryNode func() bool
	mu            sync.RWMutex
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates a plugin-owned legacy job scheduler backed by one GoFrame cron.
func New(pluginID string) Service {
	return &serviceImpl{
		pluginID:   pluginID,
		cron:       gcron.New(),
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
	}
}
