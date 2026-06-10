// This file tests snapshot processing orchestration.

package water

import (
	"context"
	"testing"

	mediastrategy "lina-plugin-media/backend/provider/strategy"
)

// TestProcessSnapshotRendersWithoutInnerEnabled verifies media_strategy.enable
// is the only switch required for snapshot watermark rendering.
func TestProcessSnapshotRendersWithoutInnerEnabled(t *testing.T) {
	service := &serviceImpl{
		strategyResolver: fakeStrategyResolver{
			strategy: `snapshot_watermark:
  text: LinaPro Water
  fontSize: 18
  color: "#ffffff"
  align: bottomRight
  opacity: 0.8`,
		},
	}

	out, err := service.processSnapshot(context.Background(), SubmitSnapInput{
		Tenant:   "tenant-water-unit",
		DeviceId: "device-water-unit",
		Image:    encodePNGDataURL(testPNGBytes(t)),
	})
	if err != nil {
		t.Fatalf("process snapshot without inner enabled failed: %v", err)
	}
	if out.Status != TaskStatusSuccess {
		t.Fatalf("expected watermark rendering success, got status=%s message=%q", out.Status, out.Message)
	}
	if out.StrategyId != 99 || out.Source != StrategySourceDevice {
		t.Fatalf("expected fake device strategy metadata, got id=%d source=%s", out.StrategyId, out.Source)
	}
}

type fakeStrategyResolver struct {
	strategy string
}

func (f fakeStrategyResolver) ResolveStrategy(_ context.Context, _ mediastrategy.ResolveStrategyInput) (*mediastrategy.ResolveStrategyOutput, error) {
	return &mediastrategy.ResolveStrategyOutput{
		Matched:      true,
		Source:       string(StrategySourceDevice),
		SourceLabel:  strategySourceLabel(StrategySourceDevice),
		StrategyId:   99,
		StrategyName: "unit strategy",
		Strategy:     f.strategy,
	}, nil
}
