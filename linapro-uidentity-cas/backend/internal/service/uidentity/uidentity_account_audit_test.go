// This file verifies legacy account/account-detail change audit behavior.

package uidentity

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

func TestResourceAccountAuditLifecycle(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.Account.Table(), dao.AccountDetail.Table(), dao.AccountChangeLog.Table())

	tenantID := int(time.Now().UnixNano()%1000000) + 910000
	actorID := 8812
	number := fmt.Sprintf("audit-%d", time.Now().UnixNano())
	service := &serviceImpl{
		tenantFilter: testTenantFilter{current: plugincontract.TenantFilterContext{
			TenantID: tenantID,
			UserID:   actorID,
		}},
	}

	cleanupAccountAuditTestRows(t, ctx, tenantID)
	t.Cleanup(func() {
		cleanupAccountAuditTestRows(t, ctx, tenantID)
	})

	accountID, err := service.CreateResource(ctx, "accounts", map[string]any{
		"number":       number,
		"name":         "Audit User",
		"phone":        "15500008812",
		"passwordHash": "do-not-store-me",
	})
	if err != nil {
		t.Fatalf("create account resource: %v", err)
	}
	accountCreate := accountAuditLog(t, ctx, tenantID, accountID, accountAuditTableAccount, accountAuditActionCreate)
	if accountCreate == nil || strings.Contains(accountCreate.DataNew, "do-not-store-me") || strings.Contains(accountCreate.DataNew, "passwordHash") {
		t.Fatalf("account create audit should exist and redact password hash, got %#v", accountCreate)
	}
	if got := accountAuditLog(t, ctx, tenantID, accountID, accountAuditTableDetail, accountAuditActionCreate); got == nil {
		t.Fatal("expected account detail create audit from automatic detail row")
	}

	if err := service.UpdateResource(ctx, "accounts", accountID, map[string]any{"name": "Audit User Updated"}); err != nil {
		t.Fatalf("update account resource: %v", err)
	}
	if got := accountAuditLog(t, ctx, tenantID, accountID, accountAuditTableAccount, accountAuditActionUpdate); got == nil || got.DataOld == "" || got.DataNew == "" {
		t.Fatalf("expected account update audit with old and new data, got %#v", got)
	}

	if err := service.UpdateResource(ctx, "account-details", accountID, map[string]any{"email": "audit@example.com"}); err != nil {
		t.Fatalf("update account detail resource: %v", err)
	}
	if got := accountAuditLog(t, ctx, tenantID, accountID, accountAuditTableDetail, accountAuditActionUpdate); got == nil || !strings.Contains(got.DataNew, "audit@example.com") {
		t.Fatalf("expected account detail update audit, got %#v", got)
	}

	if err := service.DeleteResource(ctx, "account-details", fmt.Sprintf("%d", accountID)); err != nil {
		t.Fatalf("delete account detail resource: %v", err)
	}
	if got := accountAuditLog(t, ctx, tenantID, accountID, accountAuditTableDetail, accountAuditActionDelete); got == nil || got.DataOld == "" || got.DataNew != "" {
		t.Fatalf("expected account detail delete audit with only old data, got %#v", got)
	}

	if err := service.DeleteResource(ctx, "accounts", fmt.Sprintf("%d", accountID)); err != nil {
		t.Fatalf("delete account resource: %v", err)
	}
	if got := accountAuditLog(t, ctx, tenantID, accountID, accountAuditTableAccount, accountAuditActionDelete); got == nil || got.DataOld == "" || got.DataNew != "" {
		t.Fatalf("expected account delete audit with only old data, got %#v", got)
	}
}

func accountAuditLog(t *testing.T, ctx context.Context, tenantID int, accountID int64, tableName string, action string) *entity.AccountChangeLog {
	t.Helper()
	var logRow *entity.AccountChangeLog
	err := dao.AccountChangeLog.Ctx(ctx).
		Where(do.AccountChangeLog{TenantId: tenantID, AccountId: accountID, TableName: tableName, Action: action}).
		OrderDesc(dao.AccountChangeLog.Columns().Id).
		Scan(&logRow)
	if err != nil {
		t.Fatalf("query audit log %s/%s: %v", tableName, action, err)
	}
	return logRow
}

func cleanupAccountAuditTestRows(t *testing.T, ctx context.Context, tenantID int) {
	t.Helper()
	var accountIDs []int64
	if rows, err := dao.Account.Ctx(ctx).Unscoped().
		Fields(dao.Account.Columns().Id).
		Where(do.Account{TenantId: tenantID}).
		Array(); err == nil {
		for _, id := range rows {
			if value := id.Int64(); value > 0 {
				accountIDs = append(accountIDs, value)
			}
		}
	} else {
		t.Fatalf("list audit cleanup accounts: %v", err)
	}
	if len(accountIDs) > 0 {
		if _, err := dao.AccountDetail.Ctx(ctx).Unscoped().WhereIn(dao.AccountDetail.Columns().AccountId, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup audit account details: %v", err)
		}
		if _, err := dao.Account.Ctx(ctx).Unscoped().WhereIn(dao.Account.Columns().Id, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup audit accounts: %v", err)
		}
	}
	if _, err := dao.AccountChangeLog.Ctx(ctx).Unscoped().
		Where(do.AccountChangeLog{TenantId: tenantID}).
		Delete(); err != nil {
		t.Fatalf("cleanup audit logs: %v", err)
	}
}
