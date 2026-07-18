// This file verifies media collection report normalization logic.

package collection

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/gen"

	"lina-core/pkg/plugin/capability/cachecap"
)

// errPositiveTTLSetCacheStop stops the owner write test before DB-backed counter updates.
var errPositiveTTLSetCacheStop = errors.New("positive ttl set cache stop")

// positiveTTLSetCache verifies owner cache Set calls without reaching database-dependent flow.
type positiveTTLSetCache struct {
	*memoryCollectionCache
}

// Set verifies the owner cache write uses a positive TTL before stopping DB-backed flow.
func (c positiveTTLSetCache) Set(
	ctx context.Context,
	namespace string,
	key string,
	value string,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	if ttl <= 0 {
		return nil, errors.New("cache expiration seconds must be greater than 0")
	}
	if _, err := c.memoryCollectionCache.Set(ctx, namespace, key, value, ttl); err != nil {
		return nil, err
	}
	return nil, errPositiveTTLSetCacheStop
}

// TestNormalizeMachineMetricBuildsInstanceReport verifies machine metrics become instance projections.
func TestNormalizeMachineMetricBuildsInstanceReport(t *testing.T) {
	report, ok := normalizeMachineMetric(&gen.MachineMetric{
		MachineId:      " legacy-instance-a ",
		InstanceId:     " instance-a ",
		InstanceName:   "media-server-a",
		NodeId:         " node-a ",
		NodeName:       "边缘节点-A",
		Region:         "华东",
		NodeStatus:     "healthy",
		Status:         "running",
		CpuUsage:       42.5,
		CpuCount:       8,
		MemUsed:        2 * bytesPerMegabyte,
		MemTotal:       4 * bytesPerMegabyte,
		DiskReadBytes:  3 * bytesPerKilobyte,
		DiskWriteBytes: 5 * bytesPerKilobyte,
		NetworkIn:      2 * bitsPerByte * bytesPerKilobyte,
		NetworkOut:     5 * bitsPerByte * bytesPerKilobyte,
		StartTime:      1_780_000_001,
		Version:        "v1.2.3",
		Timestamp:      1_780_000_000,
	})
	if !ok {
		t.Fatal("expected machine metric to normalize")
	}
	if report.instanceID != "instance-a" || report.instanceName != "media-server-a" {
		t.Fatalf("unexpected instance identity: %#v", report)
	}
	if report.nodeID != "node-a" || report.nodeName != "边缘节点-A" {
		t.Fatalf("unexpected node ownership fields: %#v", report)
	}
	if report.region != "华东" || report.nodeStatus != "healthy" || report.status != "running" {
		t.Fatalf("unexpected status fields: %#v", report)
	}
	if report.cpuAllocated != 8 || report.cpuLoad != 42.5 {
		t.Fatalf("unexpected cpu values: %#v", report)
	}
	if report.memoryAllocated != 4 || report.memoryUsed != 2 {
		t.Fatalf("unexpected memory values: %#v", report)
	}
	if report.diskIoRead != 3 || report.diskIoWrite != 5 {
		t.Fatalf("unexpected disk values: %#v", report)
	}
	if report.networkIn != 2 || report.networkOut != 5 {
		t.Fatalf("unexpected instance runtime values: %#v", report)
	}
	if report.startTime == nil || report.startTime.Timestamp() != 1_780_000_001 {
		t.Fatalf("expected explicit start time, got %#v", report.startTime)
	}
	if report.reportTime != 1_780_000_000 {
		t.Fatalf("expected report timestamp preserved, got %d", report.reportTime)
	}
}

// TestNormalizeMachineMetricFallsBackToMachineID verifies old machine_id carries instance identity.
func TestNormalizeMachineMetricFallsBackToMachineID(t *testing.T) {
	report, ok := normalizeMachineMetric(&gen.MachineMetric{MachineId: " legacy-instance-a "})
	if !ok {
		t.Fatal("expected machine_id fallback to normalize")
	}
	if report.instanceID != "legacy-instance-a" || report.instanceName != "legacy-instance-a" {
		t.Fatalf("unexpected fallback identity: %#v", report)
	}
}

// TestNormalizeMachineMetricRejectsMissingInstanceID verifies missing business keys are ignored.
func TestNormalizeMachineMetricRejectsMissingInstanceID(t *testing.T) {
	if _, ok := normalizeMachineMetric(&gen.MachineMetric{}); ok {
		t.Fatal("expected metric without instance identity to be ignored")
	}
}

// TestCounterEventDeltas verifies lifecycle events produce typed instance deltas.
func TestCounterEventDeltas(t *testing.T) {
	streamMove := counterEvent{kind: counterEventKindStream, add: true, instanceID: "instance-b"}.
		addDeltas("instance-a", true)
	if streamMove["instance-a"].liveStreams != -1 || streamMove["instance-b"].liveStreams != 1 {
		t.Fatalf("unexpected stream move deltas: %#v", streamMove)
	}

	sessionDelete := counterEvent{kind: counterEventKindSession, add: false, instanceID: "instance-a"}.
		deleteDeltas()
	if sessionDelete["instance-a"].sessions != -1 {
		t.Fatalf("unexpected session delete deltas: %#v", sessionDelete)
	}
	if normalizeCounterValue(-2) != 0 {
		t.Fatal("negative counter values should be clamped before projection writes")
	}
}

// TestIncrementCounterValueKeepsSharedCacheNonNegative verifies shared cache counters never stay negative.
func TestIncrementCounterValueKeepsSharedCacheNonNegative(t *testing.T) {
	ctx := context.Background()
	cache := newMemoryCollectionCache()
	runtime := &reportRuntime{cache: cache}

	if err := runtime.incrementCounterValue(ctx, reportCounterSessionsKey+"instance-a", -1); err != nil {
		t.Fatalf("decrement missing session counter: %v", err)
	}
	got, err := runtime.currentCounterValue(ctx, reportCounterSessionsKey+"instance-a")
	if err != nil {
		t.Fatalf("read session counter: %v", err)
	}
	if got != 0 {
		t.Fatalf("expected shared cache counter to be clamped to 0, got %d", got)
	}
}

// TestSharedCounterOwnerUsesPositiveCacheTTL verifies owner cache writes satisfy the host cache contract.
func TestSharedCounterOwnerUsesPositiveCacheTTL(t *testing.T) {
	ctx := context.Background()
	runtime := &reportRuntime{cache: positiveTTLSetCache{memoryCollectionCache: newMemoryCollectionCache()}}

	_, err := runtime.applySharedCounterEvent(ctx, counterEvent{
		kind:       counterEventKindStream,
		add:        true,
		resourceID: "stream-a",
		instanceID: "instance-a",
		reportTime: 1_780_000_000,
	})
	if !errors.Is(err, errPositiveTTLSetCacheStop) {
		t.Fatalf("expected owner Set to use positive TTL and stop before DB access, got %v", err)
	}
}

// TestNormalizeNetworkMetricBuildsLatencyUpdate verifies network metrics update node latency data.
func TestNormalizeNetworkMetricBuildsLatencyUpdate(t *testing.T) {
	report, ok := normalizeNetworkMetric(&gen.NetworkMetric{
		MachineId:     "node-a",
		SourceIp:      "10.0.0.1",
		DestinationIp: "10.0.0.2",
		Rtt:           28,
		Extra:         map[string]string{"throughput_bps": "40960"},
		Timestamp:     1_780_000_000_123,
	})
	if !ok {
		t.Fatal("expected network metric to normalize")
	}
	if report.nodeID != "node-a" || report.destinationID != "10.0.0.2" {
		t.Fatalf("unexpected network keys: %#v", report)
	}
	if report.networkOut != 5 {
		t.Fatalf("expected throughput converted to 5 KB/S, got %f", report.networkOut)
	}
	if report.lastHeartbeat == nil {
		t.Fatal("expected heartbeat time")
	}
	if delta := report.lastHeartbeat.Time.Sub(time.UnixMilli(1_780_000_000_123)); delta != 0 {
		t.Fatalf("expected millisecond timestamp heartbeat, delta=%s", delta)
	}
	var latency map[string]int32
	if err := json.Unmarshal([]byte(report.nodeLatencyMap), &latency); err != nil {
		t.Fatalf("decode latency map: %v", err)
	}
	if latency["10.0.0.2"] != 28 {
		t.Fatalf("expected destination latency 28, got %#v", latency)
	}
}

// TestNormalizeStreamMetricBuildsStreamReport verifies stream metrics become stream projections.
func TestNormalizeStreamMetricBuildsStreamReport(t *testing.T) {
	report, ok := normalizeStreamMetric(&gen.StreamMetric{
		MachineId:             "legacy-instance-a",
		NodeId:                "node-a",
		StreamId:              "stream-a",
		Status:                gen.StreamStatus_SS_RUNNING,
		Protocol:              gen.StreamProtocol_SP_RTMP,
		Bitrate:               4096,
		Width:                 1920,
		Height:                1080,
		StreamPath:            "rtmp://example/live/stream-a",
		StreamAlias:           "camera-a",
		DeviceId:              " device-a ",
		Extra:                 map[string]string{"device_id": "extra-device-a"},
		TenantId:              "tenant-a",
		NodeName:              "node-a-name",
		InstanceId:            "instance-a",
		InstanceName:          "media-server-a",
		Fps:                   25,
		PacketLoss:            0.01,
		Duration:              120,
		AvgDelay:              35,
		TotalSessionsLifetime: 8,
		CurrentActiveSessions: 3,
		WatermarkEnabled:      true,
		Timestamp:             1_780_000_000,
	})
	if !ok {
		t.Fatal("expected stream metric to normalize")
	}
	if report.streamID != "stream-a" || report.nodeID != "node-a" || report.instanceID != "instance-a" {
		t.Fatalf("unexpected stream keys: %#v", report)
	}
	if report.sourceType != reportSourceTypeInstance || report.tenantID != "tenant-a" || report.deviceID != "device-a" ||
		report.instanceName != "media-server-a" {
		t.Fatalf("unexpected stream ownership: %#v", report)
	}
	if report.status != reportStreamStatusRunning {
		t.Fatalf("expected running status, got %q", report.status)
	}
	if report.protocolType != reportStreamProtocolRTMP {
		t.Fatalf("expected stream protocol RTMP, got %q", report.protocolType)
	}
	if report.resolution != "1920x1080" {
		t.Fatalf("expected resolution, got %q", report.resolution)
	}
	if report.sourceURL != "rtmp://example/live/stream-a" || report.streamName != "camera-a" {
		t.Fatalf("unexpected stream names: %#v", report)
	}
	if report.fps != 25 || report.packetLoss != 0.01 || report.duration != 120 || report.avgDelay != 35 {
		t.Fatalf("unexpected stream performance fields: %#v", report)
	}
	if report.protocolCount != 1 || report.currentSessions != 3 || report.totalSessionsLifetime != 8 {
		t.Fatalf("expected one active protocol, got %#v", report)
	}
	if !report.watermarkEnabled {
		t.Fatal("expected watermark flag from extra")
	}
	var summary []protocolSummaryItem
	if err := json.Unmarshal([]byte(report.protocolSummary), &summary); err != nil {
		t.Fatalf("decode protocol summary: %v", err)
	}
	if len(summary) != 1 || summary[0].ProtocolType != "RTMP" ||
		summary[0].TotalSessions != 8 || summary[0].CurrentSessions != 3 {
		t.Fatalf("unexpected protocol summary: %#v", summary)
	}
}

// TestNormalizeStreamMetricDerivesDuration verifies stream runtime can be derived server-side.
func TestNormalizeStreamMetricDerivesDuration(t *testing.T) {
	report, ok := normalizeStreamMetric(&gen.StreamMetric{
		StreamId:  "stream-a",
		StartTime: 1_780_000_000,
		Timestamp: 1_780_000_185,
	})
	if !ok {
		t.Fatal("expected stream metric to normalize")
	}
	if report.duration != 185 {
		t.Fatalf("expected derived stream duration 185, got %#v", report)
	}
}

// TestNormalizeSessionMetricBuildsSessionReport verifies session metrics become session projections.
func TestNormalizeSessionMetricBuildsSessionReport(t *testing.T) {
	report, ok := normalizeSessionMetric(&gen.SessionMetric{
		SessionId:      "session-a",
		StreamId:       "stream-a",
		StreamName:     "camera-a",
		TenantId:       "tenant-a",
		DeviceId:       " device-a ",
		Extra:          map[string]string{"deviceId": "extra-device-a"},
		ClientId:       "client-a",
		ClientIp:       "192.0.2.10",
		ClientType:     "mobile",
		UserName:       "alice",
		Protocol:       gen.StreamProtocol_SP_HLS,
		StartTime:      1_780_000_001,
		PlayDuration:   60,
		CurrentFps:     25,
		CurrentBitrate: 2048,
		CurrentWidth:   1280,
		CurrentHeight:  720,
		MachineId:      "legacy-instance-a",
		NodeId:         "node-a",
		NodeName:       "edge-a",
		InstanceId:     "instance-a",
		InstanceName:   "media-server-a",
		LinkHops: []*gen.LinkHop{{
			HopId:    "hop-a",
			HopName:  "edge-a",
			HopType:  "node",
			NodeId:   "node-a",
			NodeName: "edge-a",
			Latency:  12,
		}},
		TotalLinkLatency: 12,
		Timestamp:        1_780_000_100,
	})
	if !ok {
		t.Fatal("expected session metric to normalize")
	}
	if report.sessionID != "session-a" || report.tenantID != "tenant-a" || report.deviceID != "device-a" ||
		report.protocolType != "HLS" {
		t.Fatalf("unexpected session fields: %#v", report)
	}
	if report.clientType != sessionClientTypeMobile {
		t.Fatalf("expected mobile client type enum, got %#v", report.clientType)
	}
	if report.currentResolution != "1280x720" || report.totalLinkLatency != 12 {
		t.Fatalf("unexpected session quality fields: %#v", report)
	}
	var hops []linkHopReportItem
	if err := json.Unmarshal([]byte(report.linkHops), &hops); err != nil {
		t.Fatalf("decode link hops: %v", err)
	}
	if len(hops) != 1 || hops[0].HopIndex != 1 || hops[0].NodeID != "node-a" || hops[0].LatencyMs != 12 {
		t.Fatalf("unexpected link hops: %#v", hops)
	}
}

// TestNormalizeMetricsIgnoreExtraDeviceID verifies device IDs only come from typed protocol fields.
func TestNormalizeMetricsIgnoreExtraDeviceID(t *testing.T) {
	stream, ok := normalizeStreamMetric(&gen.StreamMetric{
		StreamId: "stream-a",
		Extra:    map[string]string{"device_id": "extra-device-a", "deviceId": "extra-device-b"},
	})
	if !ok {
		t.Fatal("expected stream metric to normalize")
	}
	if stream.deviceID != "" {
		t.Fatalf("expected stream to ignore extra device id, got %#v", stream)
	}

	session, ok := normalizeSessionMetric(&gen.SessionMetric{
		SessionId: "session-a",
		Extra:     map[string]string{"device_id": "extra-device-a", "deviceId": "extra-device-b"},
	})
	if !ok {
		t.Fatal("expected session metric to normalize")
	}
	if session.deviceID != "" {
		t.Fatalf("expected session to ignore extra device id, got %#v", session)
	}
}

// TestNormalizeSessionMetricDerivesDurationAndLatency verifies missing computed fields are derived.
func TestNormalizeSessionMetricDerivesDurationAndLatency(t *testing.T) {
	report, ok := normalizeSessionMetric(&gen.SessionMetric{
		SessionId: "session-a",
		StartTime: 1_780_000_000,
		Timestamp: 1_780_000_042,
		LinkHops: []*gen.LinkHop{
			{NodeId: "origin-node", Latency: 10},
			{NodeId: "edge-node", Latency: 18},
		},
	})
	if !ok {
		t.Fatal("expected session metric to normalize")
	}
	if report.playDuration != 42 || report.totalLinkLatency != 28 {
		t.Fatalf("expected derived session fields, got %#v", report)
	}
	var hops []linkHopReportItem
	if err := json.Unmarshal([]byte(report.linkHops), &hops); err != nil {
		t.Fatalf("decode link hops: %v", err)
	}
	if len(hops) != 2 || hops[0].HopIndex != 1 || hops[1].HopIndex != 2 ||
		hops[0].LatencyMs != 10 || hops[1].LatencyMs != 18 {
		t.Fatalf("unexpected dashboard hop shape: %#v", hops)
	}
}

// TestNormalizeSessionMetricRejectsMissingSessionID verifies session business key is required.
func TestNormalizeSessionMetricRejectsMissingSessionID(t *testing.T) {
	if _, ok := normalizeSessionMetric(&gen.SessionMetric{StreamId: "stream-a"}); ok {
		t.Fatal("expected metric without session id to be ignored")
	}
}

// TestNormalizeStreamMetricRejectsMissingStreamID verifies stream business key is required.
func TestNormalizeStreamMetricRejectsMissingStreamID(t *testing.T) {
	if _, ok := normalizeStreamMetric(&gen.StreamMetric{NodeId: "node-a"}); ok {
		t.Fatal("expected metric without stream id to be ignored")
	}
}
