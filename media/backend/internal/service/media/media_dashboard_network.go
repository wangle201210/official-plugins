// This file enriches the existing node overview with bounded, unscored network
// probe projections. It queries only the returned subtree in one database call.

package media

import (
	"context"
	"fmt"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
)

// DashboardNodeResources contains only the current node's direct instance totals.
type DashboardNodeResources struct {
	InstanceCount   int
	CpuAllocated    float64
	CpuLoad         float64
	MemoryAllocated float64
	MemoryUsed      float64
	DiskIoRead      float64
	DiskIoWrite     float64
	NetworkIn       float64
	NetworkOut      float64
}

// DashboardNetworkLink carries one raw probe for a source instance and destination.
// Reachable is nil when the reporter did not supply a valid reachability value.
type DashboardNetworkLink struct {
	MachineId     string
	SourceIp      string
	DestinationIp string
	Rtt           int
	Jitter        int
	PacketLoss    float64
	StatusCode    int
	Reachable     *bool
	ReportTime    int64
}

// dashboardNetworkLinkRow adds the database grouping key to the service projection.
type dashboardNetworkLinkRow struct {
	NodeId string
	DashboardNetworkLink
}

// dashboardNetworkOwnershipSQL resolves the upstream instance-ID placeholder to
// a business node using the unique latest probe timestamp across all nodes.
// Grouping before the join prevents multiple probe destinations from multiplying
// resource totals. Conflicting latest nodes and unreported ownership stay unknown.
// Only generated table identifiers enter this SQL; request values never do.
func dashboardNetworkOwnershipSQL() string {
	return fmt.Sprintf(`
SELECT probe.machine_id, MIN(probe.node_id) AS node_id
FROM %[1]s probe
INNER JOIN (
    SELECT candidate.machine_id, MAX(candidate.report_time) AS report_time
    FROM %[1]s candidate
    INNER JOIN %[2]s instance ON instance.instance_id = candidate.machine_id
    WHERE candidate.machine_id <> ''
      AND (instance.node_id = '' OR instance.node_id = instance.instance_id)
    GROUP BY candidate.machine_id
) latest ON latest.machine_id = probe.machine_id AND latest.report_time = probe.report_time
GROUP BY probe.machine_id
HAVING MIN(probe.node_id) = MAX(probe.node_id)
   AND MIN(probe.node_id) <> ''
   AND MIN(probe.node_id) <> probe.machine_id`, dao.ReportNetworkLink.Table(), dao.MediaReportInstance.Table())
}

// dashboardNodeOverviewWithLinks loads all selected nodes' probes together, with a
// shared response limit. Infrastructure metrics retain the overview permission boundary.
func (s *serviceImpl) dashboardNodeOverviewWithLinks(ctx context.Context, root *DashboardNodeOverviewItem) (*DashboardNodeOverviewOutput, error) {
	var (
		nodes   = make(map[string]*DashboardNodeOverviewItem)
		nodeIDs = make([]string, 0)
		pending = []*DashboardNodeOverviewItem{root}
	)
	for len(pending) > 0 {
		node := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if node == nil {
			continue
		}
		if _, seen := nodes[node.NodeId]; seen {
			continue
		}
		nodes[node.NodeId] = node
		node.NetworkLinks = []*DashboardNetworkLink{}
		if dashboardCountsAsReportedNode(node) {
			nodeIDs = append(nodeIDs, node.NodeId)
		}
		pending = append(pending, node.ChildNodes...)
	}
	if len(nodeIDs) == 0 {
		return &DashboardNodeOverviewOutput{Node: root}, nil
	}
	var (
		columns = dao.ReportNetworkLink.Columns()
		rows    = make([]*dashboardNetworkLinkRow, 0)
	)
	err := dao.ReportNetworkLink.Ctx(ctx).
		Fields(columns.NodeId, columns.MachineId, columns.SourceIp, columns.DestinationIp,
			columns.Rtt, columns.Jitter, columns.PacketLoss, columns.StatusCode, columns.Reachable, columns.ReportTime).
		WhereIn(columns.NodeId, nodeIDs).
		OrderAsc(columns.NodeId).
		OrderAsc(columns.MachineId).
		OrderAsc(columns.SourceIp).
		OrderAsc(columns.DestinationIp).
		Limit(dashboardReadLimit + 1).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	root.NetworkLinksLimited = len(rows) > dashboardReadLimit
	if root.NetworkLinksLimited {
		rows = rows[:dashboardReadLimit]
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		if node := nodes[row.NodeId]; node != nil {
			node.NetworkLinks = append(node.NetworkLinks, &row.DashboardNetworkLink)
		}
	}
	return &DashboardNodeOverviewOutput{Node: root}, nil
}
