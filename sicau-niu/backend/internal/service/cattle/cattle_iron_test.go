// cattle_iron_test.go verifies iron-cow CRUD and code uniqueness against a gated
// PostgreSQL database. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is set.

package cattle

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

// TestCreateIronRejectsDuplicateCode verifies a second iron-cow reusing an active
// code is rejected with CodeIronCodeExists while the first remains.
func TestCreateIronRejectsDuplicateCode(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	const code = "IRON-DEV-001"
	if _, err := svc.CreateIron(ctx, &IronMutateInput{Code: code, Name: "Gate Ox"}); err != nil {
		t.Fatalf("first iron create failed: %v", err)
	}

	_, err := svc.CreateIron(ctx, &IronMutateInput{Code: code, Name: "Yard Ox"})
	assertBizCode(t, err, CodeIronCodeExists.RuntimeCode())

	count, err := dao.Iron.Ctx(ctx).Where(dao.Iron.Columns().Code, code).Count()
	if err != nil {
		t.Fatalf("count iron-cows failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one iron-cow after duplicate create, got %d", count)
	}
}

// TestCreateIronRequiresCodeAndName verifies a blank code and a blank name are
// each rejected with their specific bizerr.
func TestCreateIronRequiresCodeAndName(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()

	_, err := svc.CreateIron(ctx, &IronMutateInput{Code: "  ", Name: "Gate Ox"})
	assertBizCode(t, err, CodeIronCodeRequired.RuntimeCode())

	_, err = svc.CreateIron(ctx, &IronMutateInput{Code: "IRON-NO-NAME", Name: "  "})
	assertBizCode(t, err, CodeIronNameRequired.RuntimeCode())
}

// TestUpdateIronRejectsConflictingCode verifies updating one iron-cow to another
// active iron-cow's code is rejected with CodeIronCodeExists, while keeping its
// own code succeeds.
func TestUpdateIronRejectsConflictingCode(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	firstID, err := svc.CreateIron(ctx, &IronMutateInput{Code: "IRON-A", Name: "Ox A"})
	if err != nil {
		t.Fatalf("create iron A failed: %v", err)
	}
	if _, err = svc.CreateIron(ctx, &IronMutateInput{Code: "IRON-B", Name: "Ox B"}); err != nil {
		t.Fatalf("create iron B failed: %v", err)
	}

	// Updating iron A to iron B's code conflicts.
	err = svc.UpdateIron(ctx, firstID, &IronMutateInput{Code: "IRON-B", Name: "Ox A"})
	assertBizCode(t, err, CodeIronCodeExists.RuntimeCode())

	// Keeping iron A's own code succeeds (the row is excluded from the check).
	if err = svc.UpdateIron(ctx, firstID, &IronMutateInput{Code: "IRON-A", Name: "Ox A Renamed"}); err != nil {
		t.Fatalf("self-code update failed: %v", err)
	}
}

// TestDeleteIronMissingReturnsNotFound verifies deleting an unknown iron-cow
// returns CodeIronNotFound.
func TestDeleteIronMissingReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	err := svc.DeleteIron(ctx, 525252)
	assertBizCode(t, err, CodeIronNotFound.RuntimeCode())
}

// TestListIronReturnsLocatedAtAsAbsoluteMillis verifies located_at is stored and
// returned as a real instant, not shifted by the PostgreSQL session time zone.
func TestListIronReturnsLocatedAtAsAbsoluteMillis(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	locatedAt := time.Date(2026, time.June, 11, 16, 15, 23, 193000000, time.FixedZone("Asia/Shanghai", 8*60*60))
	err := dao.Iron.Transaction(ctx, func(txCtx context.Context, tx gdb.TX) error {
		if _, execErr := tx.Exec("SET LOCAL TIME ZONE 'UTC'"); execErr != nil {
			return execErr
		}
		_, insertErr := dao.Iron.Ctx(txCtx).Data(do.Iron{
			Code:      "IRON-TZ",
			Name:      "Time Zone Iron",
			LastLat:   29.982093672013686,
			LastLng:   102.99276695413675,
			LocatedAt: &locatedAt,
		}).Insert()
		return insertErr
	})
	if err != nil {
		t.Fatalf("insert iron with UTC session failed: %v", err)
	}

	svc := newCattleServiceForTest()
	var out *ListIronOutput
	err = dao.Iron.Transaction(ctx, func(txCtx context.Context, tx gdb.TX) error {
		if _, execErr := tx.Exec("SET LOCAL TIME ZONE 'Asia/Shanghai'"); execErr != nil {
			return execErr
		}
		list, listErr := svc.ListIron(txCtx, &ListIronInput{Keyword: "IRON-TZ", PageNum: 1, PageSize: 10})
		if listErr != nil {
			return listErr
		}
		out = list
		return nil
	})
	if err != nil {
		t.Fatalf("list iron with Asia/Shanghai session failed: %v", err)
	}
	if out == nil || len(out.List) != 1 {
		t.Fatalf("expected one iron row, got %+v", out)
	}
	if out.List[0].LocatedAt == nil || *out.List[0].LocatedAt != locatedAt.UnixMilli() {
		t.Fatalf("expected locatedAt %d, got %+v", locatedAt.UnixMilli(), out.List[0].LocatedAt)
	}
}
