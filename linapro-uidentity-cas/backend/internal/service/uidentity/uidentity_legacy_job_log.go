// This file adapts host scheduler execution logs to the old uidentity/admin
// job-log CRUD contract while keeping scheduling and log ownership in GF/host
// scheduler storage.

package uidentity

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/tenantcap"
)

const (
	legacyJobLogTable = "sys_job_log"
)

type legacyJobLogRow struct {
	Id             int64      `orm:"id"`
	TenantId       int        `orm:"tenant_id"`
	JobId          int64      `orm:"job_id"`
	JobSnapshot    string     `orm:"job_snapshot"`
	NodeId         string     `orm:"node_id"`
	Trigger        string     `orm:"trigger"`
	ParamsSnapshot string     `orm:"params_snapshot"`
	StartAt        *time.Time `orm:"start_at"`
	EndAt          *time.Time `orm:"end_at"`
	DurationMs     int64      `orm:"duration_ms"`
	Status         string     `orm:"status"`
	ErrMsg         string     `orm:"err_msg"`
	ResultJson     string     `orm:"result_json"`
	CreatedAt      *time.Time `orm:"created_at"`
}

// ListLegacyJobLogs returns host scheduler logs in the legacy job-log envelope.
func (s *serviceImpl) ListLegacyJobLogs(ctx context.Context, in LegacyJobLogListInput) (*ResourceListOutput, error) {
	model := s.legacyJobLogScopedModel(ctx)
	if in.JobID > 0 {
		model = model.Where("job_id", in.JobID)
	}
	if status := strings.TrimSpace(in.Status); status != "" {
		model = model.Where("status", status)
	}
	if trigger := strings.TrimSpace(in.Trigger); trigger != "" {
		model = model.Where("trigger", trigger)
	}
	if begin := strings.TrimSpace(in.BeginTime); begin != "" {
		model = model.WhereGTE("start_at", begin)
	}
	if end := strings.TrimSpace(in.EndTime); end != "" {
		model = model.WhereLTE("start_at", end)
	}
	if jobName := strings.TrimSpace(in.JobName); jobName != "" {
		model = model.WhereLike("job_snapshot", "%"+jobName+"%")
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	var rows []*legacyJobLogRow
	if err := model.Order(legacyJobLogOrder(in.OrderBy, in.Order)).
		Page(legacyJobLogPage(in.PageNum), legacyJobLogPageSize(in.PageSize)).
		Scan(&rows); err != nil {
		return nil, err
	}

	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		records = append(records, legacyJobLogRecord(row))
	}
	return &ResourceListOutput{List: records, Total: total}, nil
}

// GetLegacyJobLog returns one host scheduler log in legacy field names.
func (s *serviceImpl) GetLegacyJobLog(ctx context.Context, id int64) (Record, error) {
	var row *legacyJobLogRow
	if err := s.legacyJobLogScopedModel(ctx).Where("id", id).Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return legacyJobLogRecord(row), nil
}

// CreateLegacyJobLog writes one compatibility scheduler-log row.
func (s *serviceImpl) CreateLegacyJobLog(ctx context.Context, body map[string]any) (int64, error) {
	data := s.legacyJobLogMutationData(ctx, body)
	id, err := g.DB().Model(legacyJobLogTable).Ctx(ctx).Data(data).InsertAndGetId()
	return id, err
}

// UpdateLegacyJobLog updates one compatibility scheduler-log row.
func (s *serviceImpl) UpdateLegacyJobLog(ctx context.Context, id int64, body map[string]any) error {
	data := s.legacyJobLogMutationData(ctx, body)
	delete(data, "created_at")
	affected, err := s.legacyJobLogScopedModel(ctx).Where("id", id).Data(data).UpdateAndGetAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	return nil
}

// DeleteLegacyJobLogs deletes compatibility scheduler-log rows by IDs.
func (s *serviceImpl) DeleteLegacyJobLogs(ctx context.Context, ids string) error {
	normalized := legacyJobLogIDs(ids)
	if len(normalized) == 0 {
		return bizerr.NewCode(CodeDeleteIDsRequired)
	}
	model := s.legacyJobLogScopedModel(ctx).WhereIn("id", normalized)
	affected, err := model.Count()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	_, err = s.legacyJobLogScopedModel(ctx).WhereIn("id", normalized).Delete()
	return err
}

func (s *serviceImpl) legacyJobLogScopedModel(ctx context.Context) *gdb.Model {
	model := g.DB().Model(legacyJobLogTable).Ctx(ctx)
	if s.tenantFilter == nil {
		return model
	}
	current := s.tenantFilter.Context(ctx)
	if current.PlatformBypass {
		return model
	}
	return model.Where(tenantcap.TenantFilterColumn, current.TenantID)
}

func (s *serviceImpl) legacyJobLogMutationData(ctx context.Context, body map[string]any) map[string]any {
	now := time.Now()
	current := tenantcap.TenantFilterContext{}
	if s.tenantFilter != nil {
		current = s.tenantFilter.Context(ctx)
	}
	startAt := timeFromAny(body["startAt"], body["start_at"])
	endAt := timeFromAny(body["endAt"], body["end_at"])
	result := map[string]any{
		"tenant_id":       current.TenantID,
		"job_id":          gconv.Int64(firstMapValue(body, "jobId", "job_id")),
		"job_snapshot":    legacyJobLogSnapshot(body),
		"node_id":         strings.TrimSpace(gconv.String(firstMapValue(body, "nodeId", "node_id"))),
		"trigger":         fallbackString(firstMapValue(body, "trigger"), "manual"),
		"params_snapshot": strings.TrimSpace(gconv.String(firstMapValue(body, "paramsSnapshot", "params_snapshot"))),
		"start_at":        startAt,
		"end_at":          endAt,
		"duration_ms":     legacyDurationMS(startAt, endAt),
		"status":          legacyJobLogStatus(body),
		"err_msg":         strings.TrimSpace(gconv.String(firstMapValue(body, "errMsg", "err_msg"))),
		"result_json":     legacyJobLogResult(body),
		"created_at":      now,
	}
	return result
}

func legacyJobLogRecord(row *legacyJobLogRow) Record {
	if row == nil {
		return Record{}
	}
	jobName := legacyJobLogName(row)
	createNum, updateNum, deleteNum, errNum := legacyJobLogCounters(row.ResultJson)
	return Record{
		"id":              row.Id,
		"jobId":           row.JobId,
		"job_id":          row.JobId,
		"jobName":         jobName,
		"job_name":        jobName,
		"startAt":         legacyLocalClockTimePtr(row.StartAt),
		"start_at":        legacyLocalClockTimePtr(row.StartAt),
		"endAt":           legacyLocalClockTimePtr(row.EndAt),
		"end_at":          legacyLocalClockTimePtr(row.EndAt),
		"createNum":       createNum,
		"create_num":      createNum,
		"updateNum":       updateNum,
		"update_num":      updateNum,
		"deleteNum":       deleteNum,
		"delete_num":      deleteNum,
		"errNum":          errNum,
		"err_num":         errNum,
		"status":          row.Status,
		"trigger":         row.Trigger,
		"nodeId":          row.NodeId,
		"node_id":         row.NodeId,
		"durationMs":      row.DurationMs,
		"duration_ms":     row.DurationMs,
		"errMsg":          row.ErrMsg,
		"err_msg":         row.ErrMsg,
		"resultJson":      row.ResultJson,
		"result_json":     row.ResultJson,
		"jobSnapshot":     row.JobSnapshot,
		"job_snapshot":    row.JobSnapshot,
		"paramsSnapshot":  row.ParamsSnapshot,
		"params_snapshot": row.ParamsSnapshot,
		"createdAt":       legacyLocalClockTimePtr(row.CreatedAt),
		"createTime":      legacyLocalClockTimePtr(row.CreatedAt),
		"created_at":      legacyLocalClockTimePtr(row.CreatedAt),
	}
}

func legacyJobLogName(row *legacyJobLogRow) string {
	if row == nil {
		return ""
	}
	var snapshot map[string]any
	if err := json.Unmarshal([]byte(row.JobSnapshot), &snapshot); err != nil {
		return ""
	}
	for _, key := range []string{"name", "jobName", "displayName", "code"} {
		if value := strings.TrimSpace(gconv.String(snapshot[key])); value != "" {
			return value
		}
	}
	return ""
}

func legacyJobLogCounters(resultJSON string) (int64, int64, int64, int64) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(resultJSON), &payload); err != nil {
		return 0, 0, 0, 0
	}
	return gconv.Int64(firstMapValue(payload, "createNum", "create_num", "created", "inserted")),
		gconv.Int64(firstMapValue(payload, "updateNum", "update_num", "updated")),
		gconv.Int64(firstMapValue(payload, "deleteNum", "delete_num", "deleted")),
		gconv.Int64(firstMapValue(payload, "errNum", "err_num", "errors", "failed"))
}

func legacyJobLogOrder(orderBy string, order string) string {
	field := map[string]string{
		"id":         "id",
		"jobId":      "job_id",
		"job_id":     "job_id",
		"jobName":    "job_snapshot",
		"job_name":   "job_snapshot",
		"startAt":    "start_at",
		"start_at":   "start_at",
		"endAt":      "end_at",
		"end_at":     "end_at",
		"createNum":  "result_json",
		"create_num": "result_json",
		"updateNum":  "result_json",
		"update_num": "result_json",
		"deleteNum":  "result_json",
		"delete_num": "result_json",
		"errNum":     "result_json",
		"err_num":    "result_json",
		"createdAt":  "created_at",
		"createTime": "created_at",
		"created_at": "created_at",
		"status":     "status",
	}
	column := field[strings.TrimSpace(orderBy)]
	if column == "" {
		column = "start_at"
	}
	direction := "DESC"
	if strings.EqualFold(strings.TrimSpace(order), "asc") {
		direction = "ASC"
	}
	return column + " " + direction
}

func legacyJobLogSnapshot(body map[string]any) string {
	if value := strings.TrimSpace(gconv.String(firstMapValue(body, "jobSnapshot", "job_snapshot"))); value != "" {
		return value
	}
	payload := map[string]any{
		"id":      gconv.Int64(firstMapValue(body, "jobId", "job_id")),
		"jobId":   gconv.Int64(firstMapValue(body, "jobId", "job_id")),
		"name":    strings.TrimSpace(gconv.String(firstMapValue(body, "jobName", "job_name"))),
		"jobName": strings.TrimSpace(gconv.String(firstMapValue(body, "jobName", "job_name"))),
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func legacyJobLogResult(body map[string]any) string {
	if value := strings.TrimSpace(gconv.String(firstMapValue(body, "resultJson", "result_json"))); value != "" {
		return value
	}
	payload := map[string]any{
		"createNum": gconv.Int64(firstMapValue(body, "createNum", "create_num")),
		"updateNum": gconv.Int64(firstMapValue(body, "updateNum", "update_num")),
		"deleteNum": gconv.Int64(firstMapValue(body, "deleteNum", "delete_num")),
		"errNum":    gconv.Int64(firstMapValue(body, "errNum", "err_num")),
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func legacyJobLogStatus(body map[string]any) string {
	if value := strings.TrimSpace(gconv.String(firstMapValue(body, "status"))); value != "" {
		return value
	}
	if gconv.Int64(firstMapValue(body, "errNum", "err_num")) > 0 {
		return "failed"
	}
	return "success"
}

func legacyDurationMS(startAt *time.Time, endAt *time.Time) int64 {
	if startAt == nil || endAt == nil {
		return 0
	}
	return endAt.Sub(*startAt).Milliseconds()
}

func timeFromAny(values ...any) *time.Time {
	for _, value := range values {
		if value == nil {
			continue
		}
		switch typed := value.(type) {
		case time.Time:
			return &typed
		case *time.Time:
			return typed
		default:
			if parsed := gconv.Time(value); !parsed.IsZero() {
				return &parsed
			}
		}
	}
	return nil
}

func firstMapValue(body map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := body[key]; ok {
			return value
		}
	}
	return nil
}

func fallbackString(value any, fallback string) string {
	if result := strings.TrimSpace(gconv.String(value)); result != "" {
		return result
	}
	return fallback
}

func legacyJobLogPage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func legacyJobLogPageSize(size int) int {
	if size <= 0 {
		return 10
	}
	if size > 200 {
		return 200
	}
	return size
}

func legacyJobLogIDs(ids string) []int64 {
	parts := strings.Split(strings.TrimSpace(ids), ",")
	out := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		id := gconv.Int64(strings.TrimSpace(part))
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
