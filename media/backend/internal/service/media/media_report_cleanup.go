// This file implements cleanup of closed media report projections.

package media

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
)

const mediaReportRetention = 30 * 24 * time.Hour

// CleanupClosedReports removes stream and session projections closed before the retention cutoff.
func (s *serviceImpl) CleanupClosedReports(ctx context.Context, now time.Time) (int64, error) {
	if err := validateMediaReportTablesReady(ctx); err != nil {
		return 0, err
	}
	cutoff := gtime.NewFromTime(now.Add(-mediaReportRetention))
	streamDeleted, err := dao.MediaReportStream.Ctx(ctx).
		WhereLT(dao.MediaReportStream.Columns().CloseTime, cutoff).
		Delete()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	streamCount, err := streamDeleted.RowsAffected()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	sessionDeleted, err := dao.MediaReportSession.Ctx(ctx).
		WhereLT(dao.MediaReportSession.Columns().CloseTime, cutoff).
		Delete()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	sessionCount, err := sessionDeleted.RowsAffected()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return streamCount + sessionCount, nil
}
