// This file verifies UIdentity managed scheduled-job registration and pure
// legacy field mappings without requiring Oracle, LDAP, or host scheduler state.

package jobs

import (
	"context"
	"testing"
	"time"

	"lina-core/pkg/plugin/pluginhost"
)

type fakeCronRegistrar struct {
	names    []string
	handlers map[string]pluginhost.CronJobHandler
}

func (r *fakeCronRegistrar) Add(ctx context.Context, pattern string, name string, handler pluginhost.CronJobHandler) error {
	return r.AddWithMetadata(ctx, pattern, name, name, "", handler)
}

func (r *fakeCronRegistrar) AddWithMetadata(_ context.Context, _ string, name string, _ string, _ string, handler pluginhost.CronJobHandler) error {
	r.names = append(r.names, name)
	if r.handlers == nil {
		r.handlers = map[string]pluginhost.CronJobHandler{}
	}
	r.handlers[name] = handler
	return nil
}

func (r *fakeCronRegistrar) IsPrimaryNode() bool {
	return true
}

func (r *fakeCronRegistrar) Services() pluginhost.Services {
	return nil
}

func TestRegisterContributesOldJobRegistry(t *testing.T) {
	service := New(nil, nil, nil)
	registrar := &fakeCronRegistrar{}
	if err := service.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	want := []string{
		jobSyncDept,
		jobSyncJzg,
		jobSyncStudent,
		jobSyncStudentYJS,
		jobSyncStudentWJ,
		jobChangeContainer,
		jobNewContainerAccount,
		jobSyncMysql2LDAP,
		jobWannaT,
	}
	if len(registrar.names) != len(want) {
		t.Fatalf("expected %d jobs, got %d: %#v", len(want), len(registrar.names), registrar.names)
	}
	for i, name := range want {
		if registrar.names[i] != name {
			t.Fatalf("job %d mismatch, want %s got %s", i, name, registrar.names[i])
		}
		if registrar.handlers[name] == nil {
			t.Fatalf("job %s registered without handler", name)
		}
	}
}

func TestStatusValueMatchesOldStatusMap(t *testing.T) {
	cutoff := timeInPast(t)
	if got := statusValue("在读", false, cutoff); got != 1 {
		t.Fatalf("expected on status to be normal, got %d", got)
	}
	if got := statusValue("毕业", false, cutoff); got != 2 {
		t.Fatalf("expected graduation status to be locked, got %d", got)
	}
	if got := statusValue("待报到", false, cutoff); got != 0 {
		t.Fatalf("expected pending status to be not-active, got %d", got)
	}
	if got := statusValue("在读", true, cutoff); got != 0 {
		t.Fatalf("expected new on-status account after cutoff to be not-active, got %d", got)
	}
}

func TestOracleInputMappingNormalizesLegacyFields(t *testing.T) {
	service := &serviceImpl{}
	input := service.studentInput(&oracleStudentInfo{
		XH:    " 2026001 ",
		XM:    " 张 三 ",
		XB:    "男",
		SFZHM: "510000200101011234",
		CSRQ:  "2001-01-01",
		NJ:    2026,
		XYDM:  " 101 ",
		XYMC:  "信息学院",
		ZDQK:  "在读",
		SJH:   "phone:13800138000",
	})
	sanitizeAccountInputs([]*accountSyncInput{input})
	if input.number != "2026001" || input.name != "张三" || input.phone != "13800138000" {
		t.Fatalf("unexpected normalized account fields: %#v", input)
	}
	if input.detail.gender != 1 || input.detail.birthday != "2001-01-01" || input.detail.collegeCode != "101" {
		t.Fatalf("unexpected normalized detail fields: %#v", input.detail)
	}
}

func timeInPast(t *testing.T) time.Time {
	t.Helper()
	value, err := time.Parse(time.DateOnly, "2000-01-01")
	if err != nil {
		t.Fatalf("parse test cutoff: %v", err)
	}
	return value
}
