// This file tests dependency-light legacy cron helper behavior.

package cron

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const cronTestDBLink = "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"

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
	if got := normalizeExecTarget(" WannaT "); got != legacyExecTargetWannaT {
		t.Fatalf("unexpected normalized target: %s", got)
	}
	if got := normalizeExecTarget(" ChangeContainer "); got != legacyExecTargetChangeContainer {
		t.Fatalf("unexpected normalized change-container target: %s", got)
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

func TestUpdateGraduatingAccountContainer(t *testing.T) {
	ctx := context.Background()
	configureCronTestDB(t, ctx, dao.Account.Table(), dao.AccountDetail.Table(), dao.Container.Table())

	tenantID := 900000 + int(time.Now().UnixNano()%1000000)
	year := 2099
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	numberGraduating := "grad-" + token
	numberOther := "other-" + token

	cleanupCronContainerRows(t, ctx, tenantID, numberGraduating, numberOther)
	t.Cleanup(func() {
		cleanupCronContainerRows(t, ctx, tenantID, numberGraduating, numberOther)
	})

	containerID, err := dao.Container.Ctx(ctx).Data(do.Container{
		TenantId: tenantID,
		Name:     "xy",
		Alias:    "xy",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert xy container: %v", err)
	}
	graduatingAccountID := insertCronAccount(t, ctx, tenantID, numberGraduating, 1)
	otherAccountID := insertCronAccount(t, ctx, tenantID, numberOther, 1)
	insertCronAccountDetail(t, ctx, tenantID, graduatingAccountID, year)
	insertCronAccountDetail(t, ctx, tenantID, otherAccountID, year+1)

	stats, err := executeChangeContainerJob(ctx, tenantID, year)
	if err != nil {
		t.Fatalf("execute change-container job: %v", err)
	}
	if stats.updateNum != 1 {
		t.Fatalf("expected one changed account, got %d", stats.updateNum)
	}

	graduatingContainerID := accountContainerID(t, ctx, tenantID, graduatingAccountID)
	if graduatingContainerID != containerID {
		t.Fatalf("expected graduating account container %d, got %d", containerID, graduatingContainerID)
	}
	otherContainerID := accountContainerID(t, ctx, tenantID, otherAccountID)
	if otherContainerID != 1 {
		t.Fatalf("expected non-graduating account to keep container 1, got %d", otherContainerID)
	}
}

func configureCronTestDB(t *testing.T, ctx context.Context, tables ...string) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: cronTestDBLink}},
	}); err != nil {
		t.Fatalf("configure cron test database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close cron test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore cron test database config failed: %v", err)
		}
	})
	if _, err := db.Exec(ctx, "SELECT 1"); err != nil {
		t.Skipf("database is unavailable: %v", err)
	}
	for _, table := range tables {
		value, err := db.GetValue(ctx, fmt.Sprintf("SELECT to_regclass('public.%s')", table))
		if err != nil {
			t.Skipf("database table check failed: %v", err)
		}
		if value == nil || value.IsNil() || value.String() == "" {
			t.Skipf("database table %s is unavailable", table)
		}
	}
}

func insertCronAccount(t *testing.T, ctx context.Context, tenantID int, number string, containerID int64) int64 {
	t.Helper()
	id, err := dao.Account.Ctx(ctx).Data(do.Account{
		TenantId:    tenantID,
		Number:      number,
		Name:        number,
		Phone:       number,
		ContainerId: containerID,
		Status:      1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert account %s: %v", number, err)
	}
	return id
}

func insertCronAccountDetail(t *testing.T, ctx context.Context, tenantID int, accountID int64, graduationYear int) {
	t.Helper()
	if _, err := dao.AccountDetail.Ctx(ctx).Data(do.AccountDetail{
		TenantId:    tenantID,
		AccountId:   accountID,
		GraduatedAt: fmt.Sprintf("%d", graduationYear),
	}).Insert(); err != nil {
		t.Fatalf("insert account detail %d: %v", accountID, err)
	}
}

func accountContainerID(t *testing.T, ctx context.Context, tenantID int, accountID int64) int64 {
	t.Helper()
	var account *entity.Account
	if err := dao.Account.Ctx(ctx).Unscoped().
		Fields(dao.Account.Columns().Id, dao.Account.Columns().ContainerId).
		Where(do.Account{TenantId: tenantID, Id: accountID}).
		Scan(&account); err != nil {
		t.Fatalf("query account %d: %v", accountID, err)
	}
	if account == nil {
		t.Fatalf("account %d not found", accountID)
	}
	return account.ContainerId
}

func cleanupCronContainerRows(t *testing.T, ctx context.Context, tenantID int, numbers ...string) {
	t.Helper()
	var accountIDs []int64
	if rows, err := dao.Account.Ctx(ctx).Unscoped().
		Fields(dao.Account.Columns().Id).
		Where(do.Account{TenantId: tenantID}).
		WhereIn(dao.Account.Columns().Number, numbers).
		Array(); err == nil {
		for _, id := range rows {
			if value := id.Int64(); value > 0 {
				accountIDs = append(accountIDs, value)
			}
		}
	} else {
		t.Fatalf("list cron cleanup accounts: %v", err)
	}
	if len(accountIDs) > 0 {
		if _, err := dao.AccountDetail.Ctx(ctx).Unscoped().WhereIn(dao.AccountDetail.Columns().AccountId, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup cron account details: %v", err)
		}
		if _, err := dao.Account.Ctx(ctx).Unscoped().WhereIn(dao.Account.Columns().Id, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup cron accounts: %v", err)
		}
	}
	if _, err := dao.Container.Ctx(ctx).Unscoped().
		Where(do.Container{TenantId: tenantID, Name: "xy"}).
		Delete(); err != nil {
		t.Fatalf("cleanup cron containers: %v", err)
	}
}
