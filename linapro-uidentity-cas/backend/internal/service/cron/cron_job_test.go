// This file tests dependency-light legacy cron helper behavior.

package cron

import (
	"testing"
	"time"

	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

func TestJobCronNameAndEntryID(t *testing.T) {
	name := jobCronName(3, 7)
	if name != "uidentity-cas:job:3:7" {
		t.Fatalf("unexpected cron name: %s", name)
	}
	if got := jobEntryID(name); got <= 0 {
		t.Fatalf("expected positive entry ID, got %d", got)
	}
	if got := jobEntryIDForJob(&entity.SysJob{TenantId: 3, JobId: 7}); got != jobEntryID(name) {
		t.Fatalf("expected job entry ID to match name hash, got %d", got)
	}
}

func TestRunningEntryIDRoundTrip(t *testing.T) {
	entryID := int64(42)
	runningEntryID := toRunningEntryID(entryID)
	if !isRunningEntryID(runningEntryID) {
		t.Fatalf("expected running entry ID, got %d", runningEntryID)
	}
	if got := toScheduledEntryID(runningEntryID); got != entryID {
		t.Fatalf("expected scheduled entry ID %d, got %d", entryID, got)
	}
	if got := toRunningEntryID(runningEntryID); got != runningEntryID {
		t.Fatalf("expected running entry ID to be stable, got %d", got)
	}
	if got := toScheduledEntryID(entryID); got != entryID {
		t.Fatalf("expected scheduled entry ID to be stable, got %d", got)
	}
}

func TestIsStaleRunningJob(t *testing.T) {
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-legacyJobRunLease + time.Second)
	stale := now.Add(-legacyJobRunLease - time.Second)
	if isStaleRunningJob(&fresh, now) {
		t.Fatal("expected fresh running marker to be kept")
	}
	if !isStaleRunningJob(&stale, now) {
		t.Fatal("expected expired running marker to be stale")
	}
	if !isStaleRunningJob(nil, now) {
		t.Fatal("expected missing timestamp to be recoverable")
	}
}

func TestValidateJobDefinition(t *testing.T) {
	validHTTP := &entity.SysJob{
		JobId:          1,
		Status:         jobStatusEnabled,
		JobType:        jobTypeHTTP,
		CronExpression: "* * * * * *",
		InvokeTarget:   "http://127.0.0.1/health",
	}
	if err := validateJobDefinition(validHTTP); err != nil {
		t.Fatalf("expected valid HTTP job, got %v", err)
	}
	disabled := *validHTTP
	disabled.Status = 1
	if err := validateJobDefinition(&disabled); err == nil {
		t.Fatal("expected disabled job to be invalid")
	}
	unsupported := *validHTTP
	unsupported.JobType = 99
	if err := validateJobDefinition(&unsupported); err == nil {
		t.Fatal("expected unsupported job type to be invalid")
	}
}

func TestNormalizeExecTarget(t *testing.T) {
	if got := normalizeExecTarget(" WannaT "); got != "wannat" {
		t.Fatalf("unexpected normalized target: %s", got)
	}
}

func TestNormalizeCronExpression(t *testing.T) {
	if got, err := normalizeCronExpression("*/5 * * * *"); err != nil || got != "# */5 * * * *" {
		t.Fatalf("unexpected five-field normalization got=%q err=%v", got, err)
	}
	if got, err := normalizeCronExpression("0 */5 * * * *"); err != nil || got != "0 */5 * * * *" {
		t.Fatalf("unexpected six-field normalization got=%q err=%v", got, err)
	}
	if _, err := normalizeCronExpression("* *"); err == nil {
		t.Fatal("expected invalid cron expression to fail")
	}
}
