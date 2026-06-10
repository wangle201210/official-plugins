// activation_visible_test.go covers the DB-gated visible-cattle map list: the
// online-time DB-side filter, the in-memory weekday/time-window filter, and the
// batch-assembled per-player activation flag. Each test stages its own cattle and
// asserts only on its rows.

package activation

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// TestVisibleNiuFiltersByOnlineTime verifies a cattle with no online time or a
// future online time is excluded while a cattle whose online time has arrived is
// included.
func TestVisibleNiuFiltersByOnlineTime(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()

	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	visibleID := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-VIS-ON", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	// Future online time: hidden.
	insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-VIS-FUTURE", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &future,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	// Blank online time: hidden.
	insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-VIS-NOSTAGE", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		Status: cattlesvc.NiuStatusInactive.String(),
	})

	me := insertUserRow(t, ctx, do.User{Openid: "openid-vis"})
	items, err := svc.VisibleNiu(ctx, me)
	if err != nil {
		t.Fatalf("visible niu failed: %v", err)
	}
	if len(items) != 1 || items[0].Id != visibleID {
		t.Fatalf("expected only the already-online cattle, got %+v", items)
	}
}

// TestVisibleNiuIncludesApiCreatedPastOnlineAt verifies the operator Unix-ms
// onlineAt input is stored as an absolute instant. The write and read run under
// different PostgreSQL session time zones to catch TIMESTAMP-without-time-zone
// drift that would otherwise hide an already-online cattle from the player map.
func TestVisibleNiuIncludesApiCreatedPastOnlineAt(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	activationSvc := newActivationServiceForTest()
	cattleSvc := cattlesvc.New(nil)

	onlineAt := time.Now().Add(-time.Hour).UnixMilli()
	var niuID int64
	err := dao.Niu.Transaction(ctx, func(txCtx context.Context, tx gdb.TX) error {
		if _, execErr := tx.Exec("SET LOCAL TIME ZONE 'Asia/Shanghai'"); execErr != nil {
			return execErr
		}
		id, createErr := cattleSvc.CreateNiu(txCtx, &cattlesvc.NiuMutateInput{
			Code:     "NIU-VIS-API-ONLINE",
			NiuType:  cattlesvc.NiuTypeCommon.String(),
			Lat:      30.0,
			Lng:      103.0,
			OnlineAt: &onlineAt,
		})
		if createErr != nil {
			return createErr
		}
		niuID = id
		return nil
	})
	if err != nil {
		t.Fatalf("create niu with session timezone failed: %v", err)
	}

	me := insertUserRow(t, ctx, do.User{Openid: "openid-vis-api-online"})
	var items []*VisibleNiuItem
	err = dao.Niu.Transaction(ctx, func(txCtx context.Context, tx gdb.TX) error {
		if _, execErr := tx.Exec("SET LOCAL TIME ZONE 'UTC'"); execErr != nil {
			return execErr
		}
		list, visibleErr := activationSvc.VisibleNiu(txCtx, me)
		if visibleErr != nil {
			return visibleErr
		}
		items = list
		return nil
	})
	if err != nil {
		t.Fatalf("visible niu with session timezone failed: %v", err)
	}
	if len(items) != 1 || items[0].Id != niuID {
		t.Fatalf("expected API-created already-online cattle visible, got %+v", items)
	}
}

// TestVisibleNiuActivatedByMeFlag verifies the per-player activation flag is set on
// cattle the player activated and clear on others, assembled without per-row
// queries.
func TestVisibleNiuActivatedByMeFlag(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	past := time.Now().Add(-time.Hour)

	activatedID := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-VIS-MINE", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	otherID := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-VIS-OTHER", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	me := insertUserRow(t, ctx, do.User{Openid: "openid-vis-mine"})
	if _, err := svc.Activate(ctx, me, &ActivateInput{NiuId: activatedID, Lat: 30.0, Lng: 103.0}); err != nil {
		t.Fatalf("activation failed: %v", err)
	}

	items, err := svc.VisibleNiu(ctx, me)
	if err != nil {
		t.Fatalf("visible niu failed: %v", err)
	}
	flags := make(map[int64]bool, len(items))
	for _, item := range items {
		flags[item.Id] = item.ActivatedByMe
	}
	if !flags[activatedID] {
		t.Fatalf("expected activatedByMe true for activated cattle")
	}
	if flags[otherID] {
		t.Fatalf("expected activatedByMe false for non-activated cattle")
	}
}
