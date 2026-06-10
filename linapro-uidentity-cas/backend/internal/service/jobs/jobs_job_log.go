// This file restores the old afterSync job_log summary: each data-mutating
// scheduled job records one plugin-owned job_log row with start/end time and
// create/update/delete/error counters, mirroring uidentity/admin's per-run log.

package jobs

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
)

// jobLogTable is the plugin-owned execution-summary table created by the
// install manifest. It is platform-global and not tenant-partitioned.
const jobLogTable = "plugin_linapro_uidentity_cas_job_log"

// recordJobLog writes one execution summary for a scheduled job run. It is
// best-effort: a logging failure never fails the job itself.
func (s *serviceImpl) recordJobLog(ctx context.Context, jobName string, startAt time.Time, stats jobRunStats, runErr error) {
	errNum := stats.errNum
	if runErr != nil {
		errNum++
	}
	_, err := g.DB().Model(jobLogTable).Ctx(ctx).Data(g.Map{
		"job_name":   jobName,
		"start_at":   startAt,
		"end_at":     time.Now(),
		"create_num": stats.createNum,
		"update_num": stats.updateNum,
		"delete_num": stats.deleteNum,
		"err_num":    errNum,
		"create_by":  0,
		"update_by":  0,
	}).Insert()
	if err != nil {
		logger.Warningf(ctx, "uidentity job log write failed job=%s err=%v", jobName, err)
	}
}
