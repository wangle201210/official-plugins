// This file verifies node network probes retain source identity, null/zero values,
// subtree boundaries, direct instance resources, and a bounded query result.

package media

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// TestDecodeDashboardReportPayload verifies full objects retain numeric precision and invalid data fails explicitly.
func TestDecodeDashboardReportPayload(t *testing.T) {
	for _, raw := range []string{"", "  ", "null"} {
		got, err := decodeDashboardReportPayload(raw)
		if err != nil || got != nil {
			t.Fatalf("missing raw report must return nil: input=%q got=%#v err=%v", raw, got, err)
		}
	}
	for _, raw := range []string{`[]`, `"text"`, `1`, `{`, `{} {}`, `{} garbage`} {
		if got, err := decodeDashboardReportPayload(raw); err == nil {
			t.Fatalf("invalid or non-object report must fail: input=%q got=%#v", raw, got)
		}
	}
	got, err := decodeDashboardReportPayload(`{"counter":9007199254740993,"int64":"9007199254740993","extra":{"nul":"a\u0000b"},"zero":0,"flag":false}`)
	if err != nil {
		t.Fatalf("decode complete object: %v", err)
	}
	if got["counter"] != json.Number("9007199254740993") || got["int64"] != "9007199254740993" ||
		got["extra"].(map[string]any)["nul"] != "a\x00b" || got["zero"] != json.Number("0") || got["flag"] != false {
		t.Fatalf("payload field types or values changed: %#v", got)
	}
}

// TestDashboardQueriesReturnCompletePayloads exercises real SQL selection and source grouping for all four queries.
func TestDashboardQueriesReturnCompletePayloads(t *testing.T) {
	ctx := t.Context()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	insertDashboardReportNodes(t, ctx)
	const (
		machinePayload = `{"machine_id":"machine-original","instance_id":"inst-a","timestamp":"1791784485","cpu_usage":0,"disk_read_count":"9007199254740993","extra":{"vendor":"a\u0000b"}}`
		networkPayload = `{"node_id":"node-a","machine_id":"machine-original","rtt":18,"jitter":0,"packet_loss":0.01,"timestamp":"1791784485","extra":{"url":"https://probe.test","reachable":"true"}}`
		streamPayload  = `{"stream_id":"stream-a","machine_type":777,"video_codec":779,"audio_codec":780,"total_sessions_lifetime":"9007199254740993","extra":{"vendor":"stream-extra"}}`
		sessionPayload = `{"session_id":"session-a","client_type":"custom","current_fps":29.97654321,"extra":{"vendor":"session-extra"},"link_hops":[{"hop_id":"hop-a","hop_name":"完整跳点","hop_type":"gateway","node_id":"4","node_name":"节点四","latency":0}]}`
	)
	insertDashboardReports(t, ctx, []any{
		do.MediaReportInstance{InstanceId: "inst-a", NodeId: "node-a", Status: "running", ReportTime: int64(1791789999000), ReportPayload: machinePayload},
		do.MediaReportInstance{InstanceId: "inst-legacy", NodeId: "node-a", Status: "running", ReportTime: int64(1791784485000)},
		do.MediaReportInstance{InstanceId: "inst-other", NodeId: "node-b", Status: "running", ReportTime: int64(1791784485000), ReportPayload: `{"scope":"other"}`},
		do.MediaReportStream{StreamId: "stream-a", NodeId: "node-a", TenantId: "tenant-a", DeviceId: "device-a", InstanceId: "inst-a", ProtocolType: "HLS", Status: "running", ReportTime: int64(1791784485000), ReportPayload: streamPayload, ReportSubcommand: 22},
		do.MediaReportStream{StreamId: "stream-other", NodeId: "node-b", TenantId: "tenant-b", ProtocolType: "HLS", Status: "running", ReportTime: int64(1791784485000), ReportPayload: `{"scope":"other"}`},
		do.MediaReportSession{SessionId: "session-a", StreamId: "stream-a", TenantId: "tenant-a", NodeId: "node-a", InstanceId: "inst-a", ProtocolType: "HLS", ReportTime: int64(1791784485000), ReportPayload: sessionPayload, ReportSubcommand: 32},
		do.MediaReportSession{SessionId: "session-legacy", StreamId: "stream-a", TenantId: "tenant-a", NodeId: "node-a", ProtocolType: "HLS", ReportTime: int64(1791784485000)},
		do.MediaReportSession{SessionId: "session-other", StreamId: "stream-other", TenantId: "tenant-b", NodeId: "node-b", ProtocolType: "HLS", ReportTime: int64(1791784485000), ReportPayload: `{"scope":"other"}`},
		do.MediaReportSession{SessionId: "session-closed", StreamId: "stream-a", TenantId: "tenant-a", NodeId: "node-a", ProtocolType: "HLS", CloseTime: gtime.Now(), ReportTime: int64(1791784485000), ReportPayload: `{"scope":"closed"}`, ReportSubcommand: 31},
	})
	if _, err := dao.ReportNetworkLink.Ctx(ctx).Data([]do.ReportNetworkLink{
		{NodeId: "node-a", MachineId: "inst-a", SourceIp: "192.0.2.1", DestinationIp: "192.0.2.2", Rtt: 18, Jitter: 0, PacketLoss: 0.01, StatusCode: 200, ReportTime: int64(1791784485000), ReportPayload: networkPayload},
		{NodeId: "node-a", MachineId: "inst-legacy", SourceIp: "192.0.2.1", DestinationIp: "192.0.2.3", Rtt: 0, Jitter: 0, PacketLoss: 0, StatusCode: 0, ReportTime: int64(1791784485000)},
		{NodeId: "node-b", MachineId: "inst-other", SourceIp: "192.0.2.4", DestinationIp: "192.0.2.5", Rtt: 0, Jitter: 0, PacketLoss: 0, StatusCode: 0, ReportTime: int64(1791784485000), ReportPayload: `{"scope":"other"}`},
	}).Insert(); err != nil {
		t.Fatalf("insert network reports: %v", err)
	}
	assertPayload := func(got map[string]any, want string) {
		t.Helper()
		decoder := json.NewDecoder(strings.NewReader(want))
		decoder.UseNumber()
		var expected map[string]any
		if err := decoder.Decode(&expected); err != nil {
			t.Fatalf("decode expected raw fixture: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("raw fields were omitted or converted: got=%#v want=%#v", got, expected)
		}
	}
	svc := newTestMediaService(t)
	instances, err := svc.ListDashboardInstances(ctx, ListDashboardInstancesInput{NodeId: "node-a"})
	if err != nil || len(instances.List) != 2 {
		t.Fatalf("instance scope changed: result=%#v err=%v", instances, err)
	}
	assertPayload(instances.List[0].ReportPayload, machinePayload)
	var storedInstance dashboardInstanceEntity
	if err := dao.MediaReportInstance.Ctx(ctx).Where(do.MediaReportInstance{InstanceId: "inst-a"}).Scan(&storedInstance); err != nil {
		t.Fatalf("read stored timestamp fixture: %v", err)
	}
	if instances.List[0].ReportTime == nil || *instances.List[0].ReportTime != 1791784485000 ||
		storedInstance.UpdatedAt == nil || instances.List[0].UpdatedAt == nil ||
		*instances.List[0].UpdatedAt != storedInstance.UpdatedAt.TimestampMilli() {
		t.Fatalf("query must return raw sample time and stored update time: %#v", instances.List[0])
	}
	if instances.List[1].ReportPayload != nil {
		t.Fatal("legacy instance must retain a nil payload")
	}
	if instances.List[1].ReportTime != nil {
		t.Fatal("legacy instance must not use row report_time as raw sample time")
	}
	streams, err := svc.ListDashboardStreams(ctx, ListDashboardStreamsInput{TenantId: "tenant-a", DeviceId: "device-a"})
	if err != nil || len(streams.List) != 1 {
		t.Fatalf("stream scope changed: result=%#v err=%v", streams, err)
	}
	assertPayload(streams.List[0].ReportPayload, streamPayload)
	if streams.List[0].ReportSubcommand != 22 {
		t.Fatal("stream command was omitted")
	}
	sessions, err := svc.ListDashboardSessions(ctx, ListDashboardSessionsInput{StreamId: "stream-a", TenantId: "tenant-a"})
	if err != nil || len(sessions.Protocols) != 1 || len(sessions.Protocols[0].Sessions) != 2 {
		t.Fatalf("session active/scope filters changed: result=%#v err=%v", sessions, err)
	}
	assertPayload(sessions.Protocols[0].Sessions[0].ReportPayload, sessionPayload)
	if sessions.Protocols[0].Sessions[0].ReportSubcommand != 32 || sessions.Protocols[0].Sessions[1].ReportPayload != nil || sessions.Protocols[0].Sessions[1].ReportSubcommand != 0 {
		t.Fatal("session command or legacy-null semantics changed")
	}
	nodes, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: "node-a"})
	if err != nil || len(nodes.Node.NetworkLinks) != 2 {
		t.Fatalf("network grouping changed: result=%#v err=%v", nodes, err)
	}
	assertPayload(nodes.Node.NetworkLinks[0].ReportPayload, networkPayload)
	if nodes.Node.NetworkLinks[1].ReportPayload != nil {
		t.Fatal("legacy network link must retain a nil payload")
	}
	// The stream summary used for grouping still avoids loading unrelated raw content.
	impl := svc.(*serviceImpl)
	summary, err := impl.dashboardStreamEntity(ctx, "stream-a")
	if err != nil || summary.ReportPayload != "" {
		t.Fatalf("summary unnecessarily selected raw payload: summary=%#v err=%v", summary, err)
	}
}

// TestDashboardNetworkLinksAreScopedAndKeepRawMetrics exercises real database projections.
func TestDashboardNetworkLinksAreScopedAndKeepRawMetrics(t *testing.T) {
	ctx := t.Context()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	insertDashboardReportFixtures(t, ctx)
	var (
		reachable   = true
		unreachable = false
	)
	rows := []do.ReportNetworkLink{
		{NodeId: "node-a", MachineId: "inst-a", SourceIp: "192.0.2.1", DestinationIp: "192.0.2.50", Rtt: 18, Jitter: 3, PacketLoss: 0.02, StatusCode: 200, Reachable: &reachable, ReportTime: int64(1791784485000)},
		{NodeId: "node-b", MachineId: "inst-c", SourceIp: "192.0.2.2", DestinationIp: "192.0.2.50", Rtt: 5000, Jitter: 5001, PacketLoss: 1, StatusCode: 0, Reachable: &unreachable, ReportTime: int64(1791784486000)},
		{NodeId: "node-b", MachineId: "inst-d", SourceIp: "192.0.2.3", DestinationIp: "192.0.2.50", Rtt: 0, Jitter: 0, PacketLoss: 0, StatusCode: 0, ReportTime: int64(1791784487000)},
		{NodeId: "node-d", MachineId: "inst-e", SourceIp: "192.0.2.4", DestinationIp: "192.0.2.51", Rtt: 9, Jitter: 1, PacketLoss: 0, StatusCode: 200, Reachable: &reachable, ReportTime: int64(1791784488000)},
	}
	if _, err := dao.ReportNetworkLink.Ctx(ctx).Data(rows).Insert(); err != nil {
		t.Fatalf("insert link probes: %v", err)
	}
	svc := newTestMediaService(t)
	out, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: "node-b"})
	if err != nil {
		t.Fatalf("get node-b: %v", err)
	}
	root := out.Node
	if len(root.NetworkLinks) != 2 || root.NetworkLinksLimited {
		t.Fatalf("expected two untruncated probes for node-b, got %#v", root)
	}
	failed, unknown := root.NetworkLinks[0], root.NetworkLinks[1]
	if failed.MachineId != "inst-c" || failed.SourceIp != "192.0.2.2" || failed.DestinationIp != "192.0.2.50" ||
		failed.Rtt != 5000 || failed.Jitter != 5001 || failed.PacketLoss != 1 || failed.StatusCode != 0 ||
		failed.Reachable == nil || *failed.Reachable || failed.ReportTime != 1791784486000 {
		t.Fatalf("failed probe was lost or altered: %#v", failed)
	}
	if unknown.MachineId != "inst-d" || unknown.Reachable != nil || unknown.Rtt != 0 || unknown.Jitter != 0 || unknown.PacketLoss != 0 {
		t.Fatalf("unknown and zero metrics must remain distinct: %#v", unknown)
	}
	if len(root.ChildNodes) != 1 || len(root.ChildNodes[0].NetworkLinks) != 1 || root.ChildNodes[0].NetworkLinks[0].MachineId != "inst-e" {
		t.Fatalf("expected descendant probes only in their own node: %#v", root.ChildNodes)
	}
	if root.InstanceResources == nil || root.InstanceResources.InstanceCount != 1 || root.InstanceResources.CpuAllocated != 12 || root.ChildNodes[0].InstanceResources.InstanceCount != 0 {
		t.Fatalf("direct instance resources must exclude child totals: %#v", root)
	}
	all, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: "node-a"})
	if err != nil {
		t.Fatalf("get node-a: %v", err)
	}
	if len(all.Node.NetworkLinks) != 1 || all.Node.NetworkLinks[0].Reachable == nil || !*all.Node.NetworkLinks[0].Reachable || all.Node.NetworkLinks[0].PacketLoss != 0.02 {
		t.Fatalf("successful fractional probe must survive: %#v", all.Node.NetworkLinks)
	}
	if all.Node.InstanceResources.InstanceCount != 2 || all.Node.InstanceResources.CpuAllocated != 12 || all.Node.CpuAllocated != 30 {
		t.Fatalf("expected direct allocation 12 versus subtree allocation 30: %#v", all.Node)
	}
	if len(all.Node.ChildNodes[1].NetworkLinks) != 0 {
		t.Fatal("a node with no probe must not inherit other nodes' probes")
	}
}

// TestDashboardNetworkLinksLimitReportsTruncation verifies the public 10000-row boundary.
func TestDashboardNetworkLinksLimitReportsTruncation(t *testing.T) {
	ctx := t.Context()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	insertDashboardReportNodes(t, ctx)
	reachable := true
	rows := make([]do.ReportNetworkLink, dashboardReadLimit+1)
	for i := range rows {
		rows[i] = do.ReportNetworkLink{
			NodeId: "node-a", MachineId: "instance", SourceIp: "192.0.2.1",
			DestinationIp: fmt.Sprintf("destination-%05d", i),
			Rtt:           1, Jitter: 0, PacketLoss: 0, StatusCode: 200, Reachable: &reachable, ReportTime: int64(1791784485000),
		}
	}
	if _, err := dao.ReportNetworkLink.Ctx(ctx).Data(rows).Batch(500).Insert(); err != nil {
		t.Fatalf("insert link burst: %v", err)
	}
	svc := newTestMediaService(t)
	out, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: "node-a"})
	if err != nil {
		t.Fatalf("get limited links: %v", err)
	}
	if !out.Node.NetworkLinksLimited || len(out.Node.NetworkLinks) != dashboardReadLimit || out.Node.NetworkLinks[0].DestinationIp != "destination-00000" {
		t.Fatalf("expected deterministic limit and truncation marker, count=%d limited=%v", len(out.Node.NetworkLinks), out.Node.NetworkLinksLimited)
	}
	child, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: "node-b"})
	if err != nil {
		t.Fatalf("get unrelated subtree: %v", err)
	}
	if child.Node.NetworkLinksLimited || len(child.Node.NetworkLinks) != 0 {
		t.Fatal("probes outside the selected subtree must not consume its limit")
	}
	if _, err := dao.ReportNetworkLink.Ctx(ctx).Where(do.ReportNetworkLink{DestinationIp: "destination-10000"}).Delete(); err != nil {
		t.Fatalf("remove overflow probe: %v", err)
	}
	exact, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: "node-a"})
	if err != nil {
		t.Fatalf("get exact-boundary links: %v", err)
	}
	if exact.Node.NetworkLinksLimited || len(exact.Node.NetworkLinks) != dashboardReadLimit {
		t.Fatal("exactly 10000 probes must not claim truncation")
	}
}

// TestDashboardResourcesResolveReportedInstanceNodes reproduces net-flux's
// instance-ID node placeholders and verifies unique latest ownership without N+1 reads.
func TestDashboardResourcesResolveReportedInstanceNodes(t *testing.T) {
	ctx := t.Context()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	if _, err := dao.MediaReportNode.Ctx(ctx).Data([]do.MediaReportNode{
		{NodeId: "4", NodeName: "Node 4", ParentNodeId: "0", ReportTime: 1791784485000},
		{NodeId: "5", NodeName: "Node 5", ParentNodeId: "0", ReportTime: 1791784485000},
	}).Insert(); err != nil {
		t.Fatalf("insert business nodes: %v", err)
	}
	writeProbe := func(nodeID, machineID, destination string, timestamp int64) {
		t.Helper()
		if _, err := dao.ReportNetworkLink.Ctx(ctx).Data(do.ReportNetworkLink{
			NodeId: nodeID, MachineId: machineID, SourceIp: "192.0.2.1", DestinationIp: destination,
			Rtt: 18, Jitter: 3, PacketLoss: 0.02, StatusCode: 200, ReportTime: timestamp,
		}).Insert(); err != nil {
			t.Fatalf("insert probe: %v", err)
		}
	}
	svc := newTestMediaService(t)
	assertResources := func(nodeID string, count int, cores float64) {
		t.Helper()
		out, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: nodeID})
		if err != nil {
			t.Fatalf("get node %s: %v", nodeID, err)
		}
		resources := out.Node.InstanceResources
		if resources == nil || resources.InstanceCount != count || resources.CpuAllocated != cores {
			t.Fatalf("node %s expected count=%d cores=%v, got %#v", nodeID, count, cores, resources)
		}
	}
	// Network-before-machine is initially empty and resolves as soon as resources arrive.
	writeProbe("4", "pod-a", "192.0.2.10", 1791784485000)
	assertResources("4", 0, 0)
	if _, err := dao.MediaReportInstance.Ctx(ctx).Data([]do.MediaReportInstance{
		{InstanceId: "pod-a", NodeId: "pod-a", CpuAllocated: 8, CpuLoad: 2, ReportTime: 1791784485000},
		{InstanceId: "pod-b", NodeId: "", CpuAllocated: 4, CpuLoad: 1, ReportTime: 1791784485000},
		{InstanceId: "explicit", NodeId: "5", CpuAllocated: 2, CpuLoad: 0.5, ReportTime: 1791784485000},
		{InstanceId: "conflict", NodeId: "conflict", CpuAllocated: 64, CpuLoad: 32, ReportTime: 1791784485000},
		{InstanceId: "unmapped", NodeId: "unmapped", CpuAllocated: 64, CpuLoad: 32, ReportTime: 1791784485000},
	}).Insert(); err != nil {
		t.Fatalf("insert instance resources: %v", err)
	}
	assertResources("4", 1, 8)
	// More destinations cannot multiply pod-a; pod-b's machine-before-network order also resolves.
	writeProbe("4", "pod-a", "192.0.2.11", 1791784485000)
	writeProbe("4", "pod-b", "192.0.2.12", 1791784485001)
	writeProbe("4", "explicit", "192.0.2.13", 1791784485001)
	writeProbe("4", "conflict", "192.0.2.14", 1791784485001)
	writeProbe("5", "conflict", "192.0.2.15", 1791784485001)
	assertResources("4", 2, 12)
	assertResources("5", 1, 2)
	// Ownership must be selected globally before subtree filtering; pod-a has moved out of node 4.
	writeProbe("5", "pod-a", "192.0.2.16", 1791784485002)
	assertResources("4", 1, 4)
	assertResources("5", 2, 10)
	// A later resource report can retain the upstream placeholder without losing the derived node.
	columns := dao.MediaReportInstance.Columns()
	if _, err := dao.MediaReportInstance.Ctx(ctx).
		Data(do.MediaReportInstance{InstanceId: "pod-a", NodeId: "pod-a", CpuAllocated: 16, ReportTime: 1791784485003}).
		OnConflict(columns.InstanceId).
		OnDuplicate(columns.NodeId, columns.CpuAllocated, columns.ReportTime).Save(); err != nil {
		t.Fatalf("update placeholder resource report: %v", err)
	}
	assertResources("4", 1, 4)
	assertResources("5", 2, 18)
	// Resolve the formerly ambiguous machine only when a unique newer sample is available.
	writeProbe("4", "conflict", "192.0.2.17", 1791784485004)
	assertResources("4", 2, 68)
}
