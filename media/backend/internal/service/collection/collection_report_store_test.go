// This file verifies collection report persistence against a local PostgreSQL database.

package collection

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/gogf/gf/v2/database/gdb"

	_ "lina-core/pkg/dbdriver"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/entity"
)

// TestReportRuntimePersistsMetrics verifies metric reports are written into dashboard projections.
func TestReportRuntimePersistsMetrics(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_POSTGRES") != "1" {
		t.Skip("set LINAPRO_TEST_POSTGRES=1 to run against a local PostgreSQL database")
	}

	ctx := context.Background()
	setupCollectionReportPostgres(t, ctx)

	const (
		nodeID     = "collection-test-node"
		instanceID = "collection-test-instance"
		streamID   = "collection-test-stream"
		sessionID  = "collection-test-session"
		tenantID   = "collection-test-tenant"
		reportTime = int64(1_780_000_000)
	)
	cleanupCollectionReportRows(t, ctx, nodeID, instanceID, streamID, sessionID)
	t.Cleanup(func() {
		cleanupCollectionReportRows(t, ctx, nodeID, instanceID, streamID, sessionID)
	})

	cache := newMemoryCollectionCache()
	writer := newReportRuntime(cache)
	if err := writer.HandleMachineMetric(ctx, &gen.MachineMetric{
		InstanceId:     instanceID,
		InstanceName:   "media-server-a",
		MachineId:      instanceID,
		NodeId:         nodeID,
		NodeName:       "edge-a",
		Region:         "华东",
		NodeStatus:     "healthy",
		Status:         "running",
		CpuCount:       8,
		CpuUsage:       31.5,
		MemUsed:        3 * bytesPerGigabyte,
		MemTotal:       12 * bytesPerGigabyte,
		DiskReadBytes:  2 * bytesPerMegabyte,
		DiskWriteBytes: 4 * bytesPerMegabyte,
		NetworkIn:      2 * bitsPerMegabit,
		NetworkOut:     6 * bitsPerMegabit,
		StartTime:      reportTime,
		Version:        "v1.2.3",
		Timestamp:      reportTime,
	}); err != nil {
		t.Fatalf("handle machine metric: %v", err)
	}
	if err := writer.HandleMachineMetric(ctx, &gen.MachineMetric{
		InstanceId:     instanceID,
		InstanceName:   "media-server-a",
		MachineId:      instanceID,
		NodeId:         nodeID,
		NodeName:       "edge-a",
		Region:         "华东",
		NodeStatus:     "healthy",
		Status:         "running",
		CpuCount:       8,
		CpuUsage:       32.5,
		MemUsed:        3 * bytesPerGigabyte,
		MemTotal:       12 * bytesPerGigabyte,
		DiskReadBytes:  2 * bytesPerMegabyte,
		DiskWriteBytes: 4 * bytesPerMegabyte,
		NetworkIn:      2 * bitsPerMegabit,
		NetworkOut:     6 * bitsPerMegabit,
		StartTime:      reportTime,
		Version:        "v1.2.3",
		Timestamp:      reportTime,
	}); err != nil {
		t.Fatalf("replay machine metric: %v", err)
	}
	if err := writer.HandleNetworkMetric(ctx, &gen.NetworkMetric{
		MachineId:     nodeID,
		DestinationIp: "10.0.0.2",
		Rtt:           33,
		Throughput:    3 * bitsPerMegabit,
		Timestamp:     reportTime,
	}); err != nil {
		t.Fatalf("handle network metric: %v", err)
	}
	if err := writer.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_ADD), &gen.StreamMetric{
		MachineId:             instanceID,
		NodeId:                nodeID,
		StreamId:              streamID,
		TenantId:              tenantID,
		NodeName:              "edge-a",
		InstanceId:            instanceID,
		InstanceName:          "media-server-a",
		Status:                gen.StreamStatus_SS_RUNNING,
		Protocol:              gen.StreamProtocol_SP_HLS,
		Bitrate:               2048,
		Width:                 1280,
		Height:                720,
		Fps:                   25,
		PacketLoss:            0.01,
		Duration:              180,
		AvgDelay:              41,
		CurrentActiveSessions: 5,
		TotalSessionsLifetime: 10,
		WatermarkEnabled:      true,
		StreamPath:            "https://example.test/hls/stream.m3u8",
		Timestamp:             reportTime,
	}); err != nil {
		t.Fatalf("handle stream metric: %v", err)
	}
	if err := writer.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_ADD), &gen.SessionMetric{
		SessionId:        sessionID,
		StreamId:         streamID,
		StreamName:       "camera-a",
		TenantId:         tenantID,
		ClientId:         "client-a",
		ClientIp:         "192.0.2.10",
		ClientType:       "web",
		UserName:         "alice",
		Protocol:         gen.StreamProtocol_SP_HLS,
		StartTime:        reportTime,
		PlayDuration:     60,
		CurrentFps:       25,
		CurrentBitrate:   2048,
		CurrentWidth:     1280,
		CurrentHeight:    720,
		MachineId:        instanceID,
		NodeId:           nodeID,
		NodeName:         "edge-a",
		InstanceId:       instanceID,
		InstanceName:     "media-server-a",
		LinkHops:         []*gen.LinkHop{{HopId: "hop-a", NodeId: nodeID, NodeName: "edge-a", Latency: 12}},
		TotalLinkLatency: 12,
		Timestamp:        reportTime,
	}); err != nil {
		t.Fatalf("handle session metric: %v", err)
	}

	var node *entity.MediaReportNode
	if err := dao.MediaReportNode.Ctx(ctx).Where(dao.MediaReportNode.Columns().NodeId, nodeID).Scan(&node); err != nil {
		t.Fatalf("query node report: %v", err)
	}
	if node == nil {
		t.Fatal("expected node report row")
	}
	if node.NetworkOut != 3 || node.AvgDelay != 33 {
		t.Fatalf("expected network projection update, got networkOut=%f avgDelay=%d", node.NetworkOut, node.AvgDelay)
	}
	if node.NodeLatencyMap == "" {
		t.Fatal("expected node latency map")
	}

	var instance *entity.MediaReportInstance
	if err := dao.MediaReportInstance.Ctx(ctx).Where(dao.MediaReportInstance.Columns().InstanceId, instanceID).Scan(&instance); err != nil {
		t.Fatalf("query instance report: %v", err)
	}
	if instance == nil {
		t.Fatal("expected instance report row")
	}
	if instance.NodeId != nodeID || instance.CpuLoad != 32.5 || instance.CpuAllocated != 8 {
		t.Fatalf("unexpected replayed instance resource projection: %#v", instance)
	}
	if instance.Status != "running" || instance.NetworkOut != 6 || instance.LiveStreams != 1 || instance.Sessions != 1 {
		t.Fatalf("unexpected instance projection: %#v", instance)
	}

	var stream *entity.MediaReportStream
	if err := dao.MediaReportStream.Ctx(ctx).Where(dao.MediaReportStream.Columns().StreamId, streamID).Scan(&stream); err != nil {
		t.Fatalf("query stream report: %v", err)
	}
	if stream == nil {
		t.Fatal("expected stream report row")
	}
	if stream.TenantId != tenantID || stream.NodeId != nodeID || stream.InstanceId != instanceID ||
		stream.Resolution != "1280x720" || stream.Status != string(reportStreamStatusRunning) {
		t.Fatalf("unexpected stream projection: %#v", stream)
	}
	if stream.ProtocolCount != 1 || stream.CurrentActiveSessions != 5 || stream.TotalSessionsLifetime != 10 ||
		!stream.WatermarkEnabled || stream.ProtocolSummary == "" {
		t.Fatalf("unexpected stream protocol projection: %#v", stream)
	}

	var session *entity.MediaReportSession
	if err := dao.MediaReportSession.Ctx(ctx).Where(dao.MediaReportSession.Columns().SessionId, sessionID).Scan(&session); err != nil {
		t.Fatalf("query session report: %v", err)
	}
	if session == nil {
		t.Fatal("expected session report row")
	}
	if session.StreamId != streamID || session.TenantId != tenantID || session.InstanceId != instanceID ||
		session.ProtocolType != "HLS" || session.TotalLinkLatency != 12 {
		t.Fatalf("unexpected session projection: %#v", session)
	}
	if session.LinkHops == "" {
		t.Fatal("expected session link hops")
	}

	if err := writer.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_ADD), &gen.StreamMetric{
		MachineId:  instanceID,
		NodeId:     nodeID,
		StreamId:   streamID,
		InstanceId: instanceID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("handle duplicate stream add: %v", err)
	}
	if err := writer.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_DELETE), &gen.SessionMetric{
		SessionId:  sessionID,
		StreamId:   streamID,
		MachineId:  instanceID,
		NodeId:     nodeID,
		InstanceId: instanceID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("handle session delete: %v", err)
	}
	var counted *entity.MediaReportInstance
	if err := dao.MediaReportInstance.Ctx(ctx).Where(dao.MediaReportInstance.Columns().InstanceId, instanceID).Scan(&counted); err != nil {
		t.Fatalf("query counted instance report: %v", err)
	}
	if counted.LiveStreams != 1 || counted.Sessions != 0 {
		t.Fatalf("expected idempotent counters live_streams=1 sessions=0, got %#v", counted)
	}
	if err := writer.HandleMachineMetric(ctx, &gen.MachineMetric{
		InstanceId:   instanceID,
		InstanceName: "media-server-a",
		MachineId:    instanceID,
		NodeId:       nodeID,
		Status:       "running",
		CpuCount:     8,
		CpuUsage:     33.5,
		Timestamp:    reportTime + 1,
	}); err != nil {
		t.Fatalf("handle machine metric after counters: %v", err)
	}
	if err := dao.MediaReportInstance.Ctx(ctx).Where(dao.MediaReportInstance.Columns().InstanceId, instanceID).Scan(&counted); err != nil {
		t.Fatalf("query counted instance after machine metric: %v", err)
	}
	if counted.CpuLoad != 33.5 || counted.LiveStreams != 1 || counted.Sessions != 0 {
		t.Fatalf("expected machine metric not to overwrite counters, got %#v", counted)
	}
}

// TestReportRuntimeSharesLifecycleCounters verifies separate runtimes use shared cache state.
func TestReportRuntimeSharesLifecycleCounters(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_POSTGRES") != "1" {
		t.Skip("set LINAPRO_TEST_POSTGRES=1 to run against a local PostgreSQL database")
	}

	ctx := context.Background()
	setupCollectionReportPostgres(t, ctx)

	const (
		nodeID       = "collection-shared-cache-node"
		instanceID   = "collection-shared-cache-instance"
		streamID     = "collection-shared-cache-stream"
		sessionID    = "collection-shared-cache-session"
		nextInstance = "collection-shared-cache-instance-next"
		reportTime   = int64(1_780_000_200)
	)
	cleanupCollectionReportRows(t, ctx, nodeID, instanceID, streamID, sessionID)
	cleanupCollectionReportRows(t, ctx, nodeID, nextInstance, "", "")
	t.Cleanup(func() {
		cleanupCollectionReportRows(t, ctx, nodeID, instanceID, streamID, sessionID)
		cleanupCollectionReportRows(t, ctx, nodeID, nextInstance, "", "")
	})

	cache := newMemoryCollectionCache()
	firstPod := newReportRuntime(cache)
	secondPod := newReportRuntime(cache)
	if err := firstPod.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_ADD), &gen.StreamMetric{
		StreamId:   streamID,
		MachineId:  instanceID,
		InstanceId: instanceID,
		NodeId:     nodeID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("first pod stream add: %v", err)
	}
	if err := secondPod.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_ADD), &gen.StreamMetric{
		StreamId:   streamID,
		MachineId:  instanceID,
		InstanceId: instanceID,
		NodeId:     nodeID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("second pod duplicate stream add: %v", err)
	}
	if err := secondPod.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_ADD), &gen.SessionMetric{
		SessionId:  sessionID,
		StreamId:   streamID,
		MachineId:  instanceID,
		InstanceId: instanceID,
		NodeId:     nodeID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("second pod session add: %v", err)
	}

	assertReportInstanceCounters(t, ctx, instanceID, 1, 1)

	if err := firstPod.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_DELETE), &gen.SessionMetric{
		SessionId:  sessionID,
		StreamId:   streamID,
		MachineId:  instanceID,
		InstanceId: instanceID,
		NodeId:     nodeID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("first pod session delete: %v", err)
	}
	if err := secondPod.HandleSessionMetric(ctx, uint8(gen.SCMDDataReport_SESSION_DELETE), &gen.SessionMetric{
		SessionId:  sessionID,
		StreamId:   streamID,
		MachineId:  instanceID,
		InstanceId: instanceID,
		NodeId:     nodeID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("second pod duplicate session delete: %v", err)
	}

	assertReportInstanceCounters(t, ctx, instanceID, 1, 0)

	if err := secondPod.HandleStreamMetric(ctx, uint8(gen.SCMDDataReport_STREAM_ADD), &gen.StreamMetric{
		StreamId:   streamID,
		MachineId:  nextInstance,
		InstanceId: nextInstance,
		NodeId:     nodeID,
		Timestamp:  reportTime,
	}); err != nil {
		t.Fatalf("second pod stream move: %v", err)
	}

	assertReportInstanceCounters(t, ctx, instanceID, 0, 0)
	assertReportInstanceCounters(t, ctx, nextInstance, 1, 0)
}

// TestReportRuntimeMergesNetworkLatencyConcurrently verifies latency map updates survive concurrent reports.
func TestReportRuntimeMergesNetworkLatencyConcurrently(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_POSTGRES") != "1" {
		t.Skip("set LINAPRO_TEST_POSTGRES=1 to run against a local PostgreSQL database")
	}

	ctx := context.Background()
	setupCollectionReportPostgres(t, ctx)

	const nodeID = "collection-test-node-concurrent"
	cleanupCollectionReportRows(t, ctx, nodeID, "", "", "")
	t.Cleanup(func() {
		cleanupCollectionReportRows(t, ctx, nodeID, "", "", "")
	})

	writer := newReportRuntime(newMemoryCollectionCache())
	destinations := map[string]int32{
		"10.0.0.1": 11,
		"10.0.0.2": 22,
		"10.0.0.3": 33,
		"10.0.0.4": 44,
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(destinations))
	for destination, rtt := range destinations {
		wg.Add(1)
		go func(destination string, rtt int32) {
			defer wg.Done()
			errCh <- writer.HandleNetworkMetric(ctx, &gen.NetworkMetric{
				MachineId:     nodeID,
				DestinationIp: destination,
				Rtt:           rtt,
				Throughput:    bitsPerMegabit,
				Timestamp:     1_780_000_101,
			})
		}(destination, rtt)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("handle concurrent network metric: %v", err)
		}
	}

	var node *entity.MediaReportNode
	if err := dao.MediaReportNode.Ctx(ctx).Where(dao.MediaReportNode.Columns().NodeId, nodeID).Scan(&node); err != nil {
		t.Fatalf("query node report: %v", err)
	}
	if node == nil {
		t.Fatal("expected node report row")
	}
	latencyMap := map[string]int32{}
	if err := json.Unmarshal([]byte(node.NodeLatencyMap), &latencyMap); err != nil {
		t.Fatalf("decode latency map: %v", err)
	}
	for destination, rtt := range destinations {
		if latencyMap[destination] != rtt {
			t.Fatalf("expected latency %s=%d, got map=%v", destination, rtt, latencyMap)
		}
	}
}

// setupCollectionReportPostgres configures tests to use the local PostgreSQL database.
func setupCollectionReportPostgres(t *testing.T, _ context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	link := os.Getenv("LINAPRO_TEST_POSTGRES_LINK")
	if link == "" {
		link = "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable&search_path=public"
	}
	if err := gdb.SetConfig(gdb.Config{
		"default": {
			{Link: link},
		},
	}); err != nil {
		t.Fatalf("set postgres config: %v", err)
	}
	t.Cleanup(func() {
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Fatalf("restore db config: %v", err)
		}
	})
}

// cleanupCollectionReportRows removes rows written by collection report persistence tests.
func cleanupCollectionReportRows(t *testing.T, ctx context.Context, nodeID string, instanceID string, streamID string, sessionID string) {
	t.Helper()

	if sessionID != "" {
		if _, err := dao.MediaReportSession.Ctx(ctx).Where(dao.MediaReportSession.Columns().SessionId, sessionID).Delete(); err != nil {
			t.Fatalf("cleanup session report: %v", err)
		}
	}
	if streamID != "" {
		if _, err := dao.MediaReportStream.Ctx(ctx).Where(dao.MediaReportStream.Columns().StreamId, streamID).Delete(); err != nil {
			t.Fatalf("cleanup stream report: %v", err)
		}
	}
	if instanceID != "" {
		if _, err := dao.MediaReportInstance.Ctx(ctx).Where(dao.MediaReportInstance.Columns().InstanceId, instanceID).Delete(); err != nil {
			t.Fatalf("cleanup instance report: %v", err)
		}
	}
	if _, err := dao.MediaReportNodeSnapshot.Ctx(ctx).Where(dao.MediaReportNodeSnapshot.Columns().NodeId, nodeID).Delete(); err != nil {
		t.Fatalf("cleanup node snapshot report: %v", err)
	}
	if _, err := dao.MediaReportNode.Ctx(ctx).Where(dao.MediaReportNode.Columns().NodeId, nodeID).Delete(); err != nil {
		t.Fatalf("cleanup node report: %v", err)
	}
}

// assertReportInstanceCounters verifies one instance projection's live counters.
func assertReportInstanceCounters(
	t *testing.T,
	ctx context.Context,
	instanceID string,
	liveStreams int,
	sessions int,
) {
	t.Helper()

	var instance *entity.MediaReportInstance
	if err := dao.MediaReportInstance.Ctx(ctx).Where(dao.MediaReportInstance.Columns().InstanceId, instanceID).Scan(&instance); err != nil {
		t.Fatalf("query instance counters: %v", err)
	}
	if instance == nil {
		t.Fatalf("expected instance %s report row", instanceID)
	}
	if instance.LiveStreams != liveStreams || instance.Sessions != sessions {
		t.Fatalf(
			"expected instance %s counters live_streams=%d sessions=%d, got live_streams=%d sessions=%d",
			instanceID,
			liveStreams,
			sessions,
			instance.LiveStreams,
			instance.Sessions,
		)
	}
}
