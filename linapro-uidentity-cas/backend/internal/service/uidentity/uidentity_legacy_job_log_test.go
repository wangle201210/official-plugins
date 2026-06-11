// This file verifies legacy job-log projection helpers without requiring a
// scheduler database.

package uidentity

import (
	"testing"
	"time"
)

func TestLegacyJobLogRecordProjectsOldFields(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	created := end.Add(time.Second)
	record := legacyJobLogRecord(&legacyJobLogRow{
		Id:          7,
		JobId:       9,
		JobSnapshot: `{"jobName":"SyncStudent"}`,
		StartAt:     &start,
		EndAt:       &end,
		DurationMs:  2000,
		Status:      "success",
		ResultJson:  `{"createNum":3,"updateNum":4,"deleteNum":1,"errNum":2}`,
		CreatedAt:   &created,
	})
	if record["jobName"] != "SyncStudent" || record["job_name"] != "SyncStudent" {
		t.Fatalf("job name projection missing: %#v", record)
	}
	if record["createNum"] != int64(3) || record["updateNum"] != int64(4) ||
		record["deleteNum"] != int64(1) || record["errNum"] != int64(2) {
		t.Fatalf("counter projection missing: %#v", record)
	}
	wantCreated := legacyLocalClockTime(created)
	gotCreated, ok := record["createdAt"].(*time.Time)
	gotCreateTime, okTime := record["createTime"].(*time.Time)
	if !ok || !okTime || gotCreated == nil || gotCreateTime == nil ||
		!gotCreated.Equal(wantCreated) || !gotCreateTime.Equal(wantCreated) {
		t.Fatalf("create time projection missing: %#v", record)
	}
}

func TestLegacyJobLogIDsDeduplicates(t *testing.T) {
	t.Parallel()

	got := legacyJobLogIDs("7, 8,7,0,nope")
	if len(got) != 2 || got[0] != 7 || got[1] != 8 {
		t.Fatalf("legacyJobLogIDs() = %#v", got)
	}
}
