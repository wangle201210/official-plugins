// admin_v1_update_iron_reporting_cycle_test.go verifies that the operator
// controller passes the entered device code and fixed 10-second cycle to its
// injected updater and propagates updater errors.

package admin

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "lina-plugin-sicau-niu/backend/api/admin/v1"
)

type fakeReportingCycleUpdater struct {
	code  string
	cycle time.Duration
	err   error
}

func (f *fakeReportingCycleUpdater) SetReportingCycle(
	_ context.Context,
	code string,
	cycle time.Duration,
) error {
	f.code = code
	f.cycle = cycle
	return f.err
}

func TestUpdateIronReportingCycleUsesFixedTenSeconds(t *testing.T) {
	updater := &fakeReportingCycleUpdater{}
	controller := &ControllerV1{reportingCycleUpdater: updater}

	res, err := controller.UpdateIronReportingCycle(
		context.Background(),
		&v1.UpdateIronReportingCycleReq{Code: "50275156712"},
	)
	if err != nil {
		t.Fatalf("UpdateIronReportingCycle returned error: %v", err)
	}
	if res == nil {
		t.Fatal("expected a non-nil response")
	}
	if updater.code != "50275156712" {
		t.Fatalf("expected device code 50275156712, got %q", updater.code)
	}
	if updater.cycle != 10*time.Second {
		t.Fatalf("expected fixed 10-second cycle, got %s", updater.cycle)
	}
}

func TestUpdateIronReportingCyclePropagatesUpdaterError(t *testing.T) {
	want := gerror.New("setConfig failed")
	updater := &fakeReportingCycleUpdater{err: want}
	controller := &ControllerV1{reportingCycleUpdater: updater}

	res, err := controller.UpdateIronReportingCycle(
		context.Background(),
		&v1.UpdateIronReportingCycleReq{Code: "50275156712"},
	)
	if err != want {
		t.Fatalf("expected updater error, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil response on failure, got %+v", res)
	}
}
