// This file verifies ChangeContainer database side effects with a live test
// database when one is available.

package jobs

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

const jobsTestDBLink = "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"

func TestExecuteChangeContainerJobUpdatesOnlyLDAPSuccesses(t *testing.T) {
	ctx := context.Background()
	configureJobsTestDB(t, ctx, dao.Account.Table(), dao.AccountDetail.Table(), dao.Container.Table())

	tenantID := 919001
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	sourceContainerID := insertJobsTestContainer(t, ctx, tenantID, "src-"+token)
	targetContainerID := insertJobsTestContainer(t, ctx, tenantID, legacyContainerAlumni)
	okAccountID := insertJobsTestAccount(t, ctx, tenantID, "ok-"+token, sourceContainerID)
	failAccountID := insertJobsTestAccount(t, ctx, tenantID, "fail-"+token, sourceContainerID)
	insertJobsTestDetail(t, ctx, tenantID, okAccountID, time.Now().Year())
	insertJobsTestDetail(t, ctx, tenantID, failAccountID, time.Now().Year())
	t.Cleanup(func() {
		cleanupJobsTestRows(t, ctx, tenantID, []int64{okAccountID, failAccountID}, []int64{sourceContainerID, targetContainerID})
	})

	stats, err := executeChangeContainerJob(ctx, tenantID, time.Now().Year(), func(_ context.Context, account *entity.Account, _ *entity.Container) error {
		if account.Id == failAccountID {
			return fmt.Errorf("ldap move failed")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("executeChangeContainerJob returned error: %v", err)
	}
	if stats.updateNum != 1 || stats.errNum != 1 {
		t.Fatalf("expected one update and one failure, got %#v", stats)
	}
	if got := jobsTestAccountContainerID(t, ctx, tenantID, okAccountID); got != targetContainerID {
		t.Fatalf("expected LDAP-success account container=%d, got %d", targetContainerID, got)
	}
	if got := jobsTestAccountContainerID(t, ctx, tenantID, failAccountID); got != sourceContainerID {
		t.Fatalf("expected LDAP-failed account to keep source container=%d, got %d", sourceContainerID, got)
	}
}

func configureJobsTestDB(t *testing.T, ctx context.Context, tables ...string) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: jobsTestDBLink}},
	}); err != nil {
		t.Fatalf("configure jobs test database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close jobs test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore jobs test database config failed: %v", err)
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

func insertJobsTestContainer(t *testing.T, ctx context.Context, tenantID int, name string) int64 {
	t.Helper()
	id, err := dao.Container.Ctx(ctx).Data(do.Container{
		TenantId: tenantID,
		Name:     name,
		Alias:    name,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert container %s: %v", name, err)
	}
	return id
}

func insertJobsTestAccount(t *testing.T, ctx context.Context, tenantID int, number string, containerID int64) int64 {
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

func insertJobsTestDetail(t *testing.T, ctx context.Context, tenantID int, accountID int64, graduationYear int) {
	t.Helper()
	if _, err := dao.AccountDetail.Ctx(ctx).Data(do.AccountDetail{
		TenantId:    tenantID,
		AccountId:   accountID,
		GraduatedAt: fmt.Sprintf("%d", graduationYear),
	}).Insert(); err != nil {
		t.Fatalf("insert account detail %d: %v", accountID, err)
	}
}

func jobsTestAccountContainerID(t *testing.T, ctx context.Context, tenantID int, accountID int64) int64 {
	t.Helper()
	var account *entity.Account
	if err := dao.Account.Ctx(ctx).Where(do.Account{TenantId: tenantID, Id: accountID}).Scan(&account); err != nil {
		t.Fatalf("query account %d: %v", accountID, err)
	}
	if account == nil {
		t.Fatalf("account %d not found", accountID)
	}
	return account.ContainerId
}

func cleanupJobsTestRows(t *testing.T, ctx context.Context, tenantID int, accountIDs []int64, containerIDs []int64) {
	t.Helper()
	if len(accountIDs) > 0 {
		if _, err := dao.AccountDetail.Ctx(ctx).Unscoped().WhereIn(dao.AccountDetail.Columns().AccountId, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup account details: %v", err)
		}
		if _, err := dao.Account.Ctx(ctx).Unscoped().Where(do.Account{TenantId: tenantID}).WhereIn(dao.Account.Columns().Id, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup accounts: %v", err)
		}
	}
	if len(containerIDs) > 0 {
		if _, err := dao.Container.Ctx(ctx).Unscoped().Where(do.Container{TenantId: tenantID}).WhereIn(dao.Container.Columns().Id, containerIDs).Delete(); err != nil {
			t.Fatalf("cleanup containers: %v", err)
		}
	}
}
