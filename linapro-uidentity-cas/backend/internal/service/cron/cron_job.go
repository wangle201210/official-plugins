// This file implements legacy sys_job scheduling, execution, and logging on top
// of GoFrame gcron. The scheduler only touches UIdentity plugin-owned tables and
// rechecks persisted entry state before every execution.

package cron

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

// Start reloads enabled sys_job rows into the local GoFrame cron instance.
func (s *serviceImpl) Start(ctx context.Context, pluginState plugincontract.PluginStateService, isPrimaryNode func() bool) error {
	if s == nil {
		return bizerr.NewCode(CodeJobInvalid)
	}
	s.mu.Lock()
	s.pluginState = pluginState
	s.isPrimaryNode = isPrimaryNode
	s.mu.Unlock()
	if !s.shouldScheduleOnCurrentNode() {
		logger.Infof(ctx, "uidentity legacy job scheduler skipped on non-primary node")
		return nil
	}
	return s.reload(ctx)
}

// StartJob schedules one legacy job and writes its runtime entry ID.
func (s *serviceImpl) StartJob(ctx context.Context, job *entity.SysJob, actorID int64) (int64, error) {
	if s == nil || job == nil {
		return 0, bizerr.NewCode(CodeJobInvalid)
	}
	if !s.shouldScheduleOnCurrentNode() {
		return 0, bizerr.NewCode(CodeJobScheduleFailed)
	}
	if err := validateJobDefinition(job); err != nil {
		return 0, err
	}
	entryID, entryName, err := s.addJob(ctx, job)
	if err != nil {
		return 0, err
	}
	if isRunningEntryID(job.EntryId) {
		released, releaseErr := releaseStaleRunningJob(ctx, job)
		if releaseErr != nil {
			s.removeEntry(entryName)
			return 0, releaseErr
		}
		if !released {
			return entryID, nil
		}
	}
	if err = s.persistEntryID(ctx, job.TenantId, job.JobId, entryID, actorID); err != nil {
		s.removeEntry(entryName)
		return 0, err
	}
	return entryID, nil
}

// RemoveJob clears one legacy job from the local cron and DB runtime state.
func (s *serviceImpl) RemoveJob(ctx context.Context, tenantID int, jobID int64, actorID int64) error {
	if s == nil || jobID <= 0 {
		return bizerr.NewCode(CodeJobInvalid)
	}
	s.removeEntry(jobCronName(tenantID, jobID))
	return s.persistEntryID(ctx, tenantID, jobID, 0, actorID)
}

// Close stops all local cron entries owned by this scheduler.
func (s *serviceImpl) Close(ctx context.Context) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range s.cron.Entries() {
		s.cron.Remove(entry.Name)
	}
	logger.Infof(ctx, "uidentity legacy job scheduler closed")
	return nil
}

// reload reconstructs all enabled sys_job rows into the local cron instance.
func (s *serviceImpl) reload(ctx context.Context) error {
	s.mu.Lock()
	for _, entry := range s.cron.Entries() {
		s.cron.Remove(entry.Name)
	}
	s.mu.Unlock()

	cols := dao.SysJob.Columns()
	if err := releaseStaleRunningJobs(ctx); err != nil {
		return err
	}

	var jobs []*entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(cols.Status, jobStatusEnabled).
		OrderAsc(cols.TenantId).
		OrderAsc(cols.JobId).
		Scan(&jobs); err != nil {
		return err
	}
	for _, job := range jobs {
		entryID, entryName, err := s.addJob(ctx, job)
		if err != nil {
			logger.Warningf(ctx, "uidentity legacy job schedule skipped tenant=%d jobId=%d err=%v", job.TenantId, job.JobId, err)
			continue
		}
		if isRunningEntryID(job.EntryId) {
			continue
		}
		if err = s.persistEntryID(ctx, job.TenantId, job.JobId, entryID, 0); err != nil {
			s.removeEntry(entryName)
			return err
		}
	}
	logger.Infof(ctx, "uidentity legacy job scheduler loaded jobs=%d", len(jobs))
	return nil
}

// addJob registers one validated legacy job into the local GoFrame cron.
func (s *serviceImpl) addJob(ctx context.Context, job *entity.SysJob) (int64, string, error) {
	if err := validateJobDefinition(job); err != nil {
		return 0, "", err
	}
	pattern, err := normalizeCronExpression(job.CronExpression)
	if err != nil {
		return 0, "", err
	}
	entryName := jobCronName(job.TenantId, job.JobId)
	entryID := jobEntryID(entryName)
	s.removeEntry(entryName)
	handler := func(jobCtx context.Context) {
		s.runJob(jobCtx, job.TenantId, job.JobId)
	}
	if job.Concurrent == jobConcurrentAllow {
		_, err = s.cron.Add(ctx, pattern, handler, entryName)
	} else {
		_, err = s.cron.AddSingleton(ctx, pattern, handler, entryName)
	}
	if err != nil {
		return 0, "", bizerr.WrapCode(err, CodeJobScheduleFailed)
	}
	return entryID, entryName, nil
}

// runJob reloads job state from the database before each execution.
func (s *serviceImpl) runJob(ctx context.Context, tenantID int, jobID int64) {
	runCtx := plugincontract.WithCurrentContext(ctx, plugincontract.CurrentContext{TenantID: tenantID})
	if !s.shouldExecute(runCtx, tenantID) {
		return
	}
	job, err := loadRunnableJob(runCtx, tenantID, jobID)
	if err != nil {
		logger.Warningf(runCtx, "uidentity legacy job load failed tenant=%d jobId=%d err=%v", tenantID, jobID, err)
		return
	}
	if job == nil {
		return
	}
	releaseEntry, acquired, err := claimJobExecution(runCtx, job)
	if err != nil {
		logger.Warningf(runCtx, "uidentity legacy job claim failed tenant=%d jobId=%d err=%v", tenantID, jobID, err)
		return
	}
	if !acquired {
		return
	}
	defer releaseEntry()

	startAt := time.Now()
	err = s.executeJob(runCtx, job)
	endAt := time.Now()
	if logErr := insertJobLog(runCtx, job, startAt, endAt, err); logErr != nil {
		logger.Warningf(runCtx, "uidentity legacy job log failed tenant=%d jobId=%d err=%v", tenantID, jobID, logErr)
	}
	if err != nil {
		logger.Warningf(runCtx, "uidentity legacy job execution failed tenant=%d jobId=%d target=%s err=%v", tenantID, jobID, job.InvokeTarget, err)
		return
	}
	logger.Infof(runCtx, "uidentity legacy job execution finished tenant=%d jobId=%d", tenantID, jobID)
}

// executeJob dispatches one runnable legacy job by its stored job_type.
func (s *serviceImpl) executeJob(ctx context.Context, job *entity.SysJob) error {
	switch job.JobType {
	case jobTypeHTTP:
		return s.executeHTTPJob(ctx, job)
	case jobTypeExec:
		return executeExecJob(ctx, job)
	default:
		return bizerr.NewCode(CodeJobInvalid)
	}
}

// executeHTTPJob preserves the legacy HTTP GET job behavior with bounded retries.
func (s *serviceImpl) executeHTTPJob(ctx context.Context, job *entity.SysJob) error {
	target := strings.TrimSpace(job.InvokeTarget)
	var lastErr error
	for attempt := 1; attempt <= jobHTTPRetryCount; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return bizerr.WrapCode(err, CodeJobInvalid)
		}
		resp, err := s.httpClient.Do(req)
		if resp != nil {
			closeErr := closeHTTPResponse(resp)
			if err == nil {
				err = closeErr
			}
		}
		if err == nil && resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return nil
		}
		if err == nil {
			lastErr = gerror.Newf("legacy http job returned status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		if attempt < jobHTTPRetryCount {
			if !sleepWithContext(ctx, time.Duration(attempt)*legacyHTTPRetryBaseDelay) {
				return ctx.Err()
			}
		}
	}
	return lastErr
}

// executeExecJob dispatches plugin-local exec jobs and rejects external targets.
func executeExecJob(ctx context.Context, job *entity.SysJob) error {
	switch normalizeExecTarget(job.InvokeTarget) {
	case "wannat":
		for i := 0; i < 3; i++ {
			logger.Infof(ctx, "uidentity legacy WannaT job tick=%d", i+1)
			if !sleepWithContext(ctx, time.Second) {
				return ctx.Err()
			}
		}
		return nil
	case "containeraccount", "newcontaineraccount":
		return executeContainerAccountJob(ctx, job.TenantId)
	default:
		return bizerr.NewCode(CodeJobExecutorUnsupported)
	}
}

// executeContainerAccountJob refreshes container account counts with one set-based update.
func executeContainerAccountJob(ctx context.Context, tenantID int) error {
	accountCols := dao.Account.Columns()
	containerCols := dao.Container.Columns()
	tableAccount := dao.Account.Table()
	tableContainer := dao.Container.Table()
	accountCountSubquery := fmt.Sprintf(
		`(SELECT COUNT(*) FROM %s WHERE %s.%s = %d AND %s.%s IS NULL AND %s.%s = %s.%s)`,
		tableAccount,
		tableAccount,
		accountCols.TenantId,
		tenantID,
		tableAccount,
		accountCols.DeletedAt,
		tableAccount,
		accountCols.ContainerId,
		tableContainer,
		containerCols.Id,
	)
	_, err := dao.Container.Ctx(ctx).
		Where(containerCols.TenantId, tenantID).
		Data(do.Container{AccountCount: gdb.Raw(accountCountSubquery)}).
		Update()
	return err
}

// loadRunnableJob returns the job only when it is still enabled and scheduled.
func loadRunnableJob(ctx context.Context, tenantID int, jobID int64) (*entity.SysJob, error) {
	var job *entity.SysJob
	cols := dao.SysJob.Columns()
	err := dao.SysJob.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.JobId, jobID).
		Scan(&job)
	if err != nil || job == nil {
		return job, err
	}
	if job.Status != jobStatusEnabled || job.EntryId <= 0 {
		return nil, nil
	}
	if isRunningEntryID(job.EntryId) {
		released, releaseErr := releaseStaleRunningJob(ctx, job)
		if releaseErr != nil {
			return nil, releaseErr
		}
		if !released {
			return nil, nil
		}
	}
	if isRunningEntryID(job.EntryId) {
		return nil, nil
	}
	return job, nil
}

// claimJobExecution atomically marks one scheduled job as running across nodes.
func claimJobExecution(ctx context.Context, job *entity.SysJob) (release func(), acquired bool, err error) {
	if job == nil || job.EntryId <= 0 {
		return func() {}, false, nil
	}
	cols := dao.SysJob.Columns()
	runningEntryID := toRunningEntryID(job.EntryId)
	result, err := dao.SysJob.Ctx(ctx).
		Where(cols.TenantId, job.TenantId).
		Where(cols.JobId, job.JobId).
		Where(cols.Status, jobStatusEnabled).
		Where(cols.EntryId, job.EntryId).
		Data(do.SysJob{EntryId: runningEntryID}).
		Update()
	if err != nil {
		return func() {}, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return func() {}, false, err
	}
	if affected == 0 {
		return func() {}, false, nil
	}
	return func() {
		if _, releaseErr := dao.SysJob.Ctx(context.WithoutCancel(ctx)).
			Where(cols.TenantId, job.TenantId).
			Where(cols.JobId, job.JobId).
			Where(cols.EntryId, runningEntryID).
			Data(do.SysJob{EntryId: job.EntryId}).
			Update(); releaseErr != nil {
			logger.Warningf(ctx, "uidentity legacy job release failed tenant=%d jobId=%d err=%v", job.TenantId, job.JobId, releaseErr)
		}
	}, true, nil
}

// releaseStaleRunningJobs restores orphaned running markers left by crashed nodes.
func releaseStaleRunningJobs(ctx context.Context) error {
	cols := dao.SysJob.Columns()
	var staleJobs []*entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(cols.Status, jobStatusEnabled).
		WhereGT(cols.EntryId, jobEntryRunningOffset).
		WhereLT(cols.UpdatedAt, time.Now().Add(-legacyJobRunLease)).
		Scan(&staleJobs); err != nil {
		return err
	}
	for _, job := range staleJobs {
		if _, err := releaseStaleRunningJob(ctx, job); err != nil {
			return err
		}
	}
	return nil
}

// releaseStaleRunningJob restores one expired running marker when the lease elapsed.
func releaseStaleRunningJob(ctx context.Context, job *entity.SysJob) (bool, error) {
	if job == nil || !isRunningEntryID(job.EntryId) || !isStaleRunningJob(job.UpdatedAt, time.Now()) {
		return false, nil
	}
	cols := dao.SysJob.Columns()
	result, err := dao.SysJob.Ctx(ctx).
		Where(cols.TenantId, job.TenantId).
		Where(cols.JobId, job.JobId).
		Where(cols.Status, jobStatusEnabled).
		Where(cols.EntryId, job.EntryId).
		Data(do.SysJob{EntryId: toScheduledEntryID(job.EntryId)}).
		Update()
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, nil
	}
	job.EntryId = toScheduledEntryID(job.EntryId)
	return true, nil
}

// insertJobLog stores the legacy job run result in the plugin job log table.
func insertJobLog(ctx context.Context, job *entity.SysJob, startAt time.Time, endAt time.Time, runErr error) error {
	errNum := int64(0)
	if runErr != nil {
		errNum = 1
	}
	_, err := dao.JobLog.Ctx(ctx).Data(do.JobLog{
		TenantId: job.TenantId,
		JobId:    job.JobId,
		JobName:  job.JobName,
		StartAt:  &startAt,
		EndAt:    &endAt,
		ErrNum:   errNum,
	}).Insert()
	return err
}

// shouldExecute checks node role and plugin enablement before running a job.
func (s *serviceImpl) shouldExecute(ctx context.Context, tenantID int) bool {
	if !s.shouldScheduleOnCurrentNode() {
		return false
	}
	s.mu.RLock()
	pluginState := s.pluginState
	pluginID := s.pluginID
	s.mu.RUnlock()
	if pluginState == nil || strings.TrimSpace(pluginID) == "" {
		return true
	}
	tenantCtx := plugincontract.WithCurrentContext(ctx, plugincontract.CurrentContext{TenantID: tenantID})
	return pluginState.IsEnabledAuthoritative(tenantCtx, pluginID)
}

// shouldScheduleOnCurrentNode checks the host-provided primary-node predicate.
func (s *serviceImpl) shouldScheduleOnCurrentNode() bool {
	s.mu.RLock()
	isPrimaryNode := s.isPrimaryNode
	s.mu.RUnlock()
	return isPrimaryNode == nil || isPrimaryNode()
}

// persistEntryID updates persisted runtime scheduling state for one job.
func (s *serviceImpl) persistEntryID(ctx context.Context, tenantID int, jobID int64, entryID int64, actorID int64) error {
	cols := dao.SysJob.Columns()
	data := do.SysJob{EntryId: entryID}
	if actorID > 0 {
		data.UpdatedBy = actorID
	}
	_, err := dao.SysJob.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.JobId, jobID).
		Data(data).
		Update()
	return err
}

// removeEntry removes one local cron entry by name.
func (s *serviceImpl) removeEntry(entryName string) {
	if s == nil || strings.TrimSpace(entryName) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cron.Remove(entryName)
}

// validateJobDefinition rejects malformed or disabled legacy job definitions.
func validateJobDefinition(job *entity.SysJob) error {
	if job == nil || job.JobId <= 0 || job.Status != jobStatusEnabled || strings.TrimSpace(job.CronExpression) == "" {
		return bizerr.NewCode(CodeJobInvalid)
	}
	switch job.JobType {
	case jobTypeHTTP:
		if strings.TrimSpace(job.InvokeTarget) == "" {
			return bizerr.NewCode(CodeJobInvalid)
		}
	case jobTypeExec:
		if strings.TrimSpace(job.InvokeTarget) == "" {
			return bizerr.NewCode(CodeJobInvalid)
		}
	default:
		return bizerr.NewCode(CodeJobInvalid)
	}
	return nil
}

// normalizeCronExpression converts legacy five-field rows into GoFrame syntax.
func normalizeCronExpression(expr string) (string, error) {
	trimmed := strings.TrimSpace(expr)
	if strings.HasPrefix(trimmed, "@") {
		return trimmed, nil
	}
	fields := strings.Fields(trimmed)
	switch len(fields) {
	case 5:
		return "# " + strings.Join(fields, " "), nil
	case 6:
		return strings.Join(fields, " "), nil
	default:
		return "", bizerr.NewCode(CodeJobInvalid)
	}
}

// closeHTTPResponse drains and closes a legacy HTTP job response body.
func closeHTTPResponse(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return nil
	}
	_, readErr := io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	closeErr := resp.Body.Close()
	if readErr != nil {
		return readErr
	}
	return closeErr
}

// sleepWithContext preserves legacy retry delay while respecting cancellation.
func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// normalizeExecTarget normalizes legacy exec target names for dispatch.
func normalizeExecTarget(target string) string {
	return strings.ToLower(strings.TrimSpace(target))
}

// jobCronName returns the stable local GoFrame cron entry name for a sys_job row.
func jobCronName(tenantID int, jobID int64) string {
	return fmt.Sprintf("uidentity-cas:job:%d:%d", tenantID, jobID)
}

// jobEntryID derives a positive legacy-compatible numeric entry ID from a cron name.
func jobEntryID(entryName string) int64 {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(strings.TrimSpace(entryName))); err != nil {
		return 1
	}
	value := int64(hash.Sum32())
	if value == 0 {
		return 1
	}
	return value
}

// jobEntryIDForJob returns the persisted entry ID value for one sys_job row.
func jobEntryIDForJob(job *entity.SysJob) int64 {
	if job == nil {
		return 0
	}
	return jobEntryID(jobCronName(job.TenantId, job.JobId))
}

// isRunningEntryID reports whether entryID is the persisted cross-node running marker.
func isRunningEntryID(entryID int64) bool {
	return entryID > jobEntryRunningOffset
}

// toRunningEntryID converts one scheduled entry ID to its running marker.
func toRunningEntryID(entryID int64) int64 {
	if entryID <= 0 || isRunningEntryID(entryID) {
		return entryID
	}
	return entryID + jobEntryRunningOffset
}

// toScheduledEntryID restores the scheduled entry ID from its running marker.
func toScheduledEntryID(entryID int64) int64 {
	if !isRunningEntryID(entryID) {
		return entryID
	}
	return entryID - jobEntryRunningOffset
}

// isStaleRunningJob reports whether a running marker is old enough to recover.
func isStaleRunningJob(updatedAt *time.Time, now time.Time) bool {
	if updatedAt == nil {
		return true
	}
	return updatedAt.Before(now.Add(-legacyJobRunLease))
}
