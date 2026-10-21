// This file verifies node network probes retain source identity, null/zero values,
// subtree boundaries, direct instance resources, and a bounded query result.

package media

import (
	"fmt"
	"testing"

	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

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
