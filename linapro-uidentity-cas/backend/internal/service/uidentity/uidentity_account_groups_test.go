// This file verifies legacy account payload groupIds synchronization behavior.

package uidentity

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

func TestAccountResourceSyncsGroupIDs(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.Account.Table(), dao.AccountDetail.Table(), dao.AccountChangeLog.Table(), dao.Group.Table(), dao.AccountGroup.Table())

	tenantID := int(time.Now().UnixNano()%1000000) + 920000
	actorID := 8821
	number := fmt.Sprintf("groups-%d", time.Now().UnixNano())
	service := &serviceImpl{
		tenantFilter: testTenantFilter{current: plugincontract.TenantFilterContext{
			TenantID: tenantID,
			UserID:   actorID,
		}},
	}

	cleanupAccountGroupTestRows(t, ctx, tenantID, number)
	t.Cleanup(func() {
		cleanupAccountGroupTestRows(t, ctx, tenantID, number)
	})

	groupA := insertAccountGroupTestGroup(t, ctx, tenantID, "group-a-"+number)
	groupB := insertAccountGroupTestGroup(t, ctx, tenantID, "group-b-"+number)
	groupC := insertAccountGroupTestGroup(t, ctx, tenantID, "group-c-"+number)

	accountID, err := service.CreateResource(ctx, "accounts", map[string]any{
		"number":   number,
		"name":     "Group User",
		"phone":    "15500008821",
		"groupIds": []int64{groupA, groupB, groupB, 0},
	})
	if err != nil {
		t.Fatalf("create account with groupIds: %v", err)
	}
	if got := accountGroupIDs(t, ctx, tenantID, accountID); !reflect.DeepEqual(got, []int64{groupA, groupB}) {
		t.Fatalf("expected created account groups [%d %d], got %v", groupA, groupB, got)
	}

	if err := service.UpdateResource(ctx, "accounts", accountID, map[string]any{"name": "Group User Updated", "groupIds": []int64{groupC}}); err != nil {
		t.Fatalf("update account with groupIds: %v", err)
	}
	if got := accountGroupIDs(t, ctx, tenantID, accountID); !reflect.DeepEqual(got, []int64{groupC}) {
		t.Fatalf("expected updated account groups [%d], got %v", groupC, got)
	}

	if err := service.UpdateResource(ctx, "accounts", accountID, map[string]any{"groupIds": []int64{groupA, 999999999}}); err == nil {
		t.Fatal("expected missing group to be rejected")
	}
	if got := accountGroupIDs(t, ctx, tenantID, accountID); !reflect.DeepEqual(got, []int64{groupC}) {
		t.Fatalf("expected invalid update to keep groups [%d], got %v", groupC, got)
	}

	if err := service.UpdateResource(ctx, "accounts", accountID, map[string]any{"groupIds": []int64{}}); err != nil {
		t.Fatalf("clear account groupIds: %v", err)
	}
	if got := accountGroupIDs(t, ctx, tenantID, accountID); len(got) != 0 {
		t.Fatalf("expected account groups to be cleared, got %v", got)
	}
}

func insertAccountGroupTestGroup(t *testing.T, ctx context.Context, tenantID int, name string) int64 {
	t.Helper()
	id, err := dao.Group.Ctx(ctx).Data(do.Group{
		TenantId: tenantID,
		Name:     name,
		Alias:    name,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert group %s: %v", name, err)
	}
	return id
}

func accountGroupIDs(t *testing.T, ctx context.Context, tenantID int, accountID int64) []int64 {
	t.Helper()
	rows, err := dao.AccountGroup.Ctx(ctx).
		Fields(dao.AccountGroup.Columns().GroupId).
		Where(do.AccountGroup{TenantId: tenantID, AccountId: accountID}).
		OrderAsc(dao.AccountGroup.Columns().GroupId).
		Array()
	if err != nil {
		t.Fatalf("query account groups: %v", err)
	}
	result := make([]int64, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Int64())
	}
	return result
}

func cleanupAccountGroupTestRows(t *testing.T, ctx context.Context, tenantID int, number string) {
	t.Helper()
	var accountIDs []int64
	if rows, err := dao.Account.Ctx(ctx).Unscoped().
		Fields(dao.Account.Columns().Id).
		Where(do.Account{TenantId: tenantID, Number: number}).
		Array(); err == nil {
		for _, id := range rows {
			if value := id.Int64(); value > 0 {
				accountIDs = append(accountIDs, value)
			}
		}
	} else {
		t.Fatalf("list account group cleanup accounts: %v", err)
	}
	if len(accountIDs) > 0 {
		if _, err := dao.AccountGroup.Ctx(ctx).Unscoped().WhereIn(dao.AccountGroup.Columns().AccountId, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup account groups: %v", err)
		}
		if _, err := dao.AccountDetail.Ctx(ctx).Unscoped().WhereIn(dao.AccountDetail.Columns().AccountId, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup account details: %v", err)
		}
		if _, err := dao.Account.Ctx(ctx).Unscoped().WhereIn(dao.Account.Columns().Id, accountIDs).Delete(); err != nil {
			t.Fatalf("cleanup accounts: %v", err)
		}
	}
	if _, err := dao.AccountChangeLog.Ctx(ctx).Unscoped().Where(do.AccountChangeLog{TenantId: tenantID}).Delete(); err != nil {
		t.Fatalf("cleanup account change logs: %v", err)
	}
	if _, err := dao.Group.Ctx(ctx).Unscoped().Where(dao.Group.Columns().TenantId, tenantID).Delete(); err != nil {
		t.Fatalf("cleanup groups: %v", err)
	}
}
