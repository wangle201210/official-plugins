// This file verifies legacy activation and union-ID binding database behavior
// that cannot be proven by dependency-light projection tests alone.

package uidentity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const uidentityTestDBLink = "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"

type testTenantFilter struct {
	current plugincontract.TenantFilterContext
}

func (f testTenantFilter) Context(context.Context) plugincontract.TenantFilterContext {
	return f.current
}

func (f testTenantFilter) Apply(_ context.Context, model *gdb.Model, qualifier string) *gdb.Model {
	if model == nil || f.current.PlatformBypass {
		return model
	}
	column := plugincontract.TenantFilterColumn
	if qualifier != "" {
		column = qualifier + "." + column
	}
	return model.Where(column, f.current.TenantID)
}

func TestRebindUnionIDToAccountMigratesExistingBinding(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.Account.Table(), dao.AccountDetail.Table(), dao.AccountActiveLog.Table())

	tenantID := 909001
	actorID := 7721
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	unionID := "union-" + token
	oldNumber := "old-" + token
	newNumber := "new-" + token

	cleanupUIdentityTestRows(t, ctx, tenantID, oldNumber, newNumber, unionID)
	t.Cleanup(func() {
		cleanupUIdentityTestRows(t, ctx, tenantID, oldNumber, newNumber, unionID)
	})

	oldAccountID := insertUIdentityTestAccount(t, ctx, tenantID, oldNumber, "15500000001")
	newAccountID := insertUIdentityTestAccount(t, ctx, tenantID, newNumber, "15500000002")
	insertUIdentityTestDetail(t, ctx, tenantID, oldAccountID, unionID)
	insertUIdentityTestDetail(t, ctx, tenantID, newAccountID, "")

	service := &serviceImpl{
		tenantFilter: testTenantFilter{current: plugincontract.TenantFilterContext{
			TenantID: tenantID,
			UserID:   actorID,
		}},
	}
	account := &entity.Account{Id: newAccountID, TenantId: tenantID, Number: newNumber, Phone: "15500000002"}
	if err := service.rebindUnionIDToAccount(ctx, account, unionID); err != nil {
		t.Fatalf("rebindUnionIDToAccount returned error: %v", err)
	}

	var oldDetail *entity.AccountDetail
	if err := dao.AccountDetail.Ctx(ctx).Unscoped().Where(do.AccountDetail{TenantId: tenantID, AccountId: oldAccountID}).Scan(&oldDetail); err != nil {
		t.Fatalf("query old account detail: %v", err)
	}
	if oldDetail == nil || oldDetail.Wechat != "" {
		t.Fatalf("expected old account union ID to be cleared, got %#v", oldDetail)
	}

	var newDetail *entity.AccountDetail
	if err := dao.AccountDetail.Ctx(ctx).Unscoped().Where(do.AccountDetail{TenantId: tenantID, AccountId: newAccountID}).Scan(&newDetail); err != nil {
		t.Fatalf("query new account detail: %v", err)
	}
	if newDetail == nil || newDetail.Wechat != unionID {
		t.Fatalf("expected new account to own union ID %q, got %#v", unionID, newDetail)
	}

	var logRow *entity.AccountActiveLog
	if err := dao.AccountActiveLog.Ctx(ctx).Where(do.AccountActiveLog{TenantId: tenantID, Number: newNumber, Wechat: unionID}).Scan(&logRow); err != nil {
		t.Fatalf("query account active log: %v", err)
	}
	if logRow == nil || logRow.Type != int(accountActiveLogTypeUnionBind) || logRow.Phone != account.Phone {
		t.Fatalf("expected union-bind active log, got %#v", logRow)
	}
}

func configureUIdentityTestDB(t *testing.T, ctx context.Context, tables ...string) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: uidentityTestDBLink}},
	}); err != nil {
		t.Fatalf("configure uidentity test database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close uidentity test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore uidentity test database config failed: %v", err)
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

func insertUIdentityTestAccount(t *testing.T, ctx context.Context, tenantID int, number string, phone string) int64 {
	t.Helper()
	id, err := dao.Account.Ctx(ctx).Data(do.Account{
		TenantId: tenantID,
		Number:   number,
		Name:     number,
		Phone:    phone,
		Status:   AccountStatusNormal,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert account %s: %v", number, err)
	}
	return id
}

func insertUIdentityTestDetail(t *testing.T, ctx context.Context, tenantID int, accountID int64, unionID string) {
	t.Helper()
	if _, err := dao.AccountDetail.Ctx(ctx).Data(do.AccountDetail{
		TenantId:  tenantID,
		AccountId: accountID,
		Wechat:    unionID,
	}).Insert(); err != nil {
		t.Fatalf("insert account detail %d: %v", accountID, err)
	}
}

func cleanupUIdentityTestRows(t *testing.T, ctx context.Context, tenantID int, oldNumber string, newNumber string, unionID string) {
	t.Helper()
	var accountIDs []int64
	if rows, err := dao.Account.Ctx(ctx).Unscoped().
		Fields(dao.Account.Columns().Id).
		Where(do.Account{TenantId: tenantID}).
		WhereIn(dao.Account.Columns().Number, []string{oldNumber, newNumber}).
		Array(); err == nil {
		for _, id := range rows {
			if value := id.Int64(); value > 0 {
				accountIDs = append(accountIDs, value)
			}
		}
	} else {
		t.Fatalf("list cleanup accounts: %v", err)
	}

	if len(accountIDs) > 0 {
		if _, err := dao.AccountDetail.Ctx(ctx).Unscoped().WhereIn(dao.AccountDetail.Columns().AccountId, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup account details: %v", err)
		}
		if _, err := dao.Account.Ctx(ctx).Unscoped().WhereIn(dao.Account.Columns().Id, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup accounts: %v", err)
		}
	}
	if _, err := dao.AccountActiveLog.Ctx(ctx).Unscoped().
		Where(do.AccountActiveLog{TenantId: tenantID}).
		Where(dao.AccountActiveLog.Columns().Wechat, unionID).
		Delete(); err != nil {
		t.Fatalf("cleanup account active logs: %v", err)
	}
}
