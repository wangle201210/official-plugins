// This file verifies instance resource attribution through the public dashboard
// service against opt-in local PostgreSQL, using isolated and fully cleaned rows.

package media

import (
	"context"
	"crypto/rand"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	_ "lina-core/pkg/dbdriver"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// TestDashboardNetworkInstanceOwnershipPostgres covers real reporter identities,
// multiple destinations, explicit ownership, migration and ambiguous latest samples.
func TestDashboardNetworkInstanceOwnershipPostgres(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_POSTGRES") != "1" {
		t.Skip("set LINAPRO_TEST_POSTGRES=1 to run against a local PostgreSQL database")
	}
	// g.DB caches the default group's driver independently of gdb.SetConfig.
	// A dedicated test process prevents a PostgreSQL driver from surviving into
	// the package's SQLite tests, including when test execution order is shuffled.
	const childProcessEnv = "LINAPRO_MEDIA_NETWORK_POSTGRES_CHILD"
	if os.Getenv(childProcessEnv) != "1" {
		executable, err := os.Executable()
		if err != nil {
			t.Fatalf("locate test executable: %v", err)
		}
		cmd := exec.CommandContext(t.Context(), executable,
			"-test.run=^TestDashboardNetworkInstanceOwnershipPostgres$", "-test.count=1", "-test.timeout=45s")
		cmd.Env = append(os.Environ(), childProcessEnv+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated PostgreSQL ownership regression: %v\n%s", err, output)
		}
		t.Logf("isolated PostgreSQL ownership regression: %s", output)
		return
	}
	var (
		ctx            = t.Context()
		originalConfig = gdb.GetAllConfig()
		link           = os.Getenv("LINAPRO_TEST_POSTGRES_LINK")
		prefix         = "media-network-pg-" + rand.Text()[:10]
		node4          = prefix + "-4"
		explicitNode   = prefix + "-explicit"
		outsideNode    = prefix + "-outside"
		podA           = prefix + "-pod-a"
		emptyPod       = prefix + "-empty"
		explicitPod    = prefix + "-explicit-pod"
		migrantPod     = prefix + "-migrant"
		conflictingPod = prefix + "-conflicting"
		unknownPod     = prefix + "-unknown"
		nodeIDs        = []string{node4, explicitNode, outsideNode}
		instanceIDs    = []string{podA, emptyPod, explicitPod, migrantPod, conflictingPod, unknownPod}
	)
	if link == "" {
		link = "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable&search_path=public"
	}
	if err := gdb.SetConfig(gdb.Config{"default": {{Link: link}}}); err != nil {
		t.Fatalf("set PostgreSQL configuration: %v", err)
	}
	t.Cleanup(func() {
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore database configuration: %v", err)
		}
	})
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if _, err := dao.ReportNetworkLink.Ctx(cleanupCtx).
			WhereIn(dao.ReportNetworkLink.Columns().MachineId, instanceIDs).Delete(); err != nil {
			t.Errorf("clean test network links: %v", err)
		}
		if _, err := dao.MediaReportInstance.Ctx(cleanupCtx).
			WhereIn(dao.MediaReportInstance.Columns().InstanceId, instanceIDs).Delete(); err != nil {
			t.Errorf("clean test instances: %v", err)
		}
		if _, err := dao.MediaReportNode.Ctx(cleanupCtx).
			WhereIn(dao.MediaReportNode.Columns().NodeId, nodeIDs).Delete(); err != nil {
			t.Errorf("clean test nodes: %v", err)
		}
	})

	const sampledAt int64 = 1_790_000_000_000
	nodes := []do.MediaReportNode{
		{NodeId: node4, NodeName: "Business node 4", ParentNodeId: "0", Status: "healthy", ReportTime: sampledAt},
		{NodeId: explicitNode, NodeName: "Explicit owner", ParentNodeId: "0", Status: "healthy", ReportTime: sampledAt},
		{NodeId: outsideNode, NodeName: "Outside selected subtree", ParentNodeId: "0", Status: "healthy", ReportTime: sampledAt},
	}
	if _, err := dao.MediaReportNode.Ctx(ctx).Data(nodes).Insert(); err != nil {
		t.Fatalf("insert isolated dashboard nodes: %v", err)
	}
	instances := []do.MediaReportInstance{
		{InstanceId: podA, NodeId: podA, CpuAllocated: 8, CpuLoad: 2, MemoryAllocated: 1024, MemoryUsed: 512,
			DiskIoRead: 3, DiskIoWrite: 4, NetworkIn: 5, NetworkOut: 6, ReportTime: sampledAt},
		{InstanceId: emptyPod, NodeId: "", CpuAllocated: 2, CpuLoad: 0.5, MemoryAllocated: 256, MemoryUsed: 128,
			DiskIoRead: 1, DiskIoWrite: 2, NetworkIn: 3, NetworkOut: 4, ReportTime: sampledAt},
		{InstanceId: explicitPod, NodeId: explicitNode, CpuAllocated: 4, CpuLoad: 1, ReportTime: sampledAt},
		{InstanceId: migrantPod, NodeId: migrantPod, CpuAllocated: 16, CpuLoad: 4, ReportTime: sampledAt},
		{InstanceId: conflictingPod, NodeId: conflictingPod, CpuAllocated: 32, CpuLoad: 8, ReportTime: sampledAt},
		{InstanceId: unknownPod, NodeId: unknownPod, CpuAllocated: 64, CpuLoad: 16, ReportTime: sampledAt},
	}
	if _, err := dao.MediaReportInstance.Ctx(ctx).Data(instances).Insert(); err != nil {
		t.Fatalf("insert isolated instance resources: %v", err)
	}
	links := []do.ReportNetworkLink{
		{NodeId: node4, MachineId: podA, SourceIp: "192.0.2.1", DestinationIp: "192.0.2.2", ReportTime: sampledAt},
		{NodeId: node4, MachineId: podA, SourceIp: "192.0.2.1", DestinationIp: "192.0.2.3", ReportTime: sampledAt},
		{NodeId: node4, MachineId: emptyPod, SourceIp: "192.0.2.4", DestinationIp: "192.0.2.2", ReportTime: sampledAt},
		{NodeId: node4, MachineId: explicitPod, SourceIp: "192.0.2.5", DestinationIp: "192.0.2.2", ReportTime: sampledAt + 1},
		{NodeId: node4, MachineId: migrantPod, SourceIp: "192.0.2.6", DestinationIp: "192.0.2.2", ReportTime: sampledAt},
		{NodeId: outsideNode, MachineId: migrantPod, SourceIp: "192.0.2.6", DestinationIp: "192.0.2.2", ReportTime: sampledAt + 1},
		{NodeId: node4, MachineId: conflictingPod, SourceIp: "192.0.2.7", DestinationIp: "192.0.2.2", ReportTime: sampledAt + 2},
		{NodeId: outsideNode, MachineId: conflictingPod, SourceIp: "192.0.2.7", DestinationIp: "192.0.2.3", ReportTime: sampledAt + 2},
	}
	if _, err := dao.ReportNetworkLink.Ctx(ctx).Data(links).Insert(); err != nil {
		t.Fatalf("insert isolated network ownership samples: %v", err)
	}

	svc := newTestMediaService(t)
	for _, test := range []struct {
		nodeID string
		want   DashboardNodeResources
	}{
		{nodeID: node4, want: DashboardNodeResources{
			InstanceCount: 2, CpuAllocated: 10, CpuLoad: 2.5, MemoryAllocated: 1280, MemoryUsed: 640,
			DiskIoRead: 4, DiskIoWrite: 6, NetworkIn: 8, NetworkOut: 10,
		}},
		{nodeID: explicitNode, want: DashboardNodeResources{InstanceCount: 1, CpuAllocated: 4, CpuLoad: 1}},
		{nodeID: outsideNode, want: DashboardNodeResources{InstanceCount: 1, CpuAllocated: 16, CpuLoad: 4}},
	} {
		out, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: test.nodeID})
		if err != nil {
			t.Fatalf("query node %s resources: %v", test.nodeID, err)
		}
		if out.Node == nil || out.Node.NodeId != test.nodeID || out.Node.InstanceResources == nil {
			t.Fatalf("node %s is missing its resource projection: %#v", test.nodeID, out)
		}
		if got := *out.Node.InstanceResources; got != test.want {
			t.Fatalf("node %s attributed or counted resources incorrectly: got=%+v want=%+v", test.nodeID, got, test.want)
		}
		if out.Node.CpuAllocated != test.want.CpuAllocated || out.Node.CpuLoad != test.want.CpuLoad {
			t.Fatalf("node %s aggregate resources disagree with direct resources: %#v", test.nodeID, out.Node)
		}
	}
}
