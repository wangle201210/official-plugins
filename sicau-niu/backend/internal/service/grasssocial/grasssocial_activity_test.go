package grasssocial

import (
	"context"
	"testing"
	"time"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

func TestActivitiesReturnsOnlyEventsAffectingCurrentPlayer(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()
	me := insertUserRow(t, ctx, "openid-activity-me")
	thief := insertUserRow(t, ctx, "openid-activity-thief")
	helper := insertUserRow(t, ctx, "openid-activity-helper")
	other := insertUserRow(t, ctx, "openid-activity-other")
	if _, err := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, thief).Data(do.User{Nickname: "偷草同学"}).Update(); err != nil {
		t.Fatalf("name thief failed: %v", err)
	}
	if _, err := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, helper).Data(do.User{Nickname: "赠草同学"}).Update(); err != nil {
		t.Fatalf("name helper failed: %v", err)
	}
	older := time.Now().Add(-time.Minute)
	newer := time.Now()
	if _, err := dao.Steal.Ctx(ctx).Data(do.Steal{ActorUserId: thief, TargetUserId: me, Amount: 2, StealDate: "2026-08-11", CreatedAt: &older}).Insert(); err != nil {
		t.Fatalf("insert steal failed: %v", err)
	}
	if _, err := dao.Gift.Ctx(ctx).Data(do.Gift{FromUserId: helper, ToUserId: me, Amount: 5, GiftDate: "2026-08-11", CreatedAt: &newer}).Insert(); err != nil {
		t.Fatalf("insert gift failed: %v", err)
	}
	if _, err := dao.Steal.Ctx(ctx).Data(do.Steal{ActorUserId: thief, TargetUserId: other, Amount: 9, StealDate: "2026-08-11"}).Insert(); err != nil {
		t.Fatalf("insert unrelated steal failed: %v", err)
	}

	items, err := svc.Activities(ctx, me)
	if err != nil {
		t.Fatalf("activities failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two own events, got %+v", items)
	}
	if items[0].Type != "help" || items[0].ActorName != "赠草同学" || items[0].Amount != 5 {
		t.Fatalf("unexpected newest gift event: %+v", items[0])
	}
	if items[1].Type != "steal" || items[1].ActorName != "偷草同学" || items[1].Amount != 2 {
		t.Fatalf("unexpected steal event: %+v", items[1])
	}
}
