// settlement_dashboard.go implements the operations dashboard. Every figure is a
// single database-side aggregate (Count or Sum), so the dashboard never loads rows
// into memory and never scales with player or cattle count. The activated-cattle
// filter reuses the cattle status enum and the certificate count reuses the honor
// type enum so the dashboard and those contracts never drift.

package settlement

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// activeNiuStatus is the persisted cattle status string the activated-cattle count
// filters on, reusing the cattle package's stable enum constant.
const activeNiuStatus = string(cattlesvc.NiuStatusActive)

// Dashboard is the operations dashboard result. The json tags double as the
// persisted archive snapshot field names.
type Dashboard struct {
	PlayerCount             int64 `json:"playerCount"`
	ActivatedNiuCount       int64 `json:"activatedNiuCount"`
	TotalNiuCount           int64 `json:"totalNiuCount"`
	FirstActivatorCount     int64 `json:"firstActivatorCount"`
	FeedingCount            int64 `json:"feedingCount"`
	FeedTotalEffect         int64 `json:"feedTotalEffect"`
	StealCount              int64 `json:"stealCount"`
	GiftCount               int64 `json:"giftCount"`
	CheckinCount            int64 `json:"checkinCount"`
	CertificateGrantedCount int64 `json:"certificateGrantedCount"`
}

// Dashboard returns the operations dashboard, each figure counted or summed on the
// database side.
func (s *serviceImpl) Dashboard(ctx context.Context) (*Dashboard, error) {
	out := &Dashboard{}

	players, err := dao.User.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.PlayerCount = int64(players)

	activatedNiu, err := dao.Niu.Ctx(ctx).Where(dao.Niu.Columns().Status, activeNiuStatus).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.ActivatedNiuCount = int64(activatedNiu)

	totalNiu, err := dao.Niu.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.TotalNiuCount = int64(totalNiu)

	firstActivators, err := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().IsFirst, 1).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.FirstActivatorCount = int64(firstActivators)

	feedingCount, err := dao.Feeding.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.FeedingCount = int64(feedingCount)

	feedEffect, err := dao.Feeding.Ctx(ctx).Sum(dao.Feeding.Columns().EffectAmount)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.FeedTotalEffect = int64(feedEffect)

	stealCount, err := dao.Steal.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.StealCount = int64(stealCount)

	giftCount, err := dao.Gift.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.GiftCount = int64(giftCount)

	checkinCount, err := dao.Checkin.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out.CheckinCount = int64(checkinCount)

	certGranted, err := s.certificateGrantedCount(ctx)
	if err != nil {
		return nil, err
	}
	out.CertificateGrantedCount = certGranted

	return out, nil
}

// certificateGrantedCount counts the granted certificate honors: the certificate
// honor IDs are resolved once, then the grant rows for those IDs are counted on the
// database side. It returns 0 without a second query when no certificate honor
// exists.
func (s *serviceImpl) certificateGrantedCount(ctx context.Context) (int64, error) {
	certIDs, err := s.certificateHonorIDs(ctx)
	if err != nil {
		return 0, err
	}
	if len(certIDs) == 0 {
		return 0, nil
	}
	count, err := dao.UserHonor.Ctx(ctx).WhereIn(dao.UserHonor.Columns().HonorId, certIDs).Count()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	return int64(count), nil
}
