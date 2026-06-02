// This file verifies legacy Oracle account synchronization decisions that can
// be checked without a live Oracle or plugin database.

package jobs

import (
	"context"
	"testing"
)

func TestPrepareAccountSyncInputsSkipsLegacyNonNumericUnitCodes(t *testing.T) {
	service := &serviceImpl{}
	inputs := []*accountSyncInput{
		service.studentInput(&oracleStudentInfo{XH: "bzks-1", XM: "本科生", XYDM: "abc"}),
		studentYJSInput(&oracleStudentYJS{Xh: "yjs-1", XsXm: "研究生", XsYxDm: "10A"}),
		staffJZGInput(&oracleStaffJZG{GH: "jzg-1", XM: "教职工", DWH: "X42"}),
		studentWJInput(&oracleStudentWJ{XH: "wj-1", XM: "网教生"}),
	}

	valid, stats := prepareAccountSyncInputs(inputs)
	if stats.errNum != 3 {
		t.Fatalf("expected three invalid unit-code rows, got %d", stats.errNum)
	}
	if len(valid) != 1 {
		t.Fatalf("expected only WJ fixed-unit input to remain, got %#v", valid)
	}
	if valid[0].number != "wj-1" || valid[0].unitCode != "424" {
		t.Fatalf("unexpected surviving input: %#v", valid[0])
	}
}

func TestUpsertDetailFromJobSkipsMissingExistingDetail(t *testing.T) {
	service := &serviceImpl{}
	stats := jobRunStats{}

	changed, err := service.upsertDetailFromJob(context.Background(), &syncContext{tenantID: 1}, 1001, nil, accountDetailSyncInput{
		email:  "new@example.test",
		wechat: "open-id",
	}, &stats)
	if err != nil {
		t.Fatalf("upsertDetailFromJob returned error: %v", err)
	}
	if changed {
		t.Fatal("expected missing existing detail to be skipped")
	}
	if stats.updateAccountDetailCount != 0 {
		t.Fatalf("expected no detail update count, got %d", stats.updateAccountDetailCount)
	}
}
