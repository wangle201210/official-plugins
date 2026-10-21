// This file enriches the existing node overview with bounded, unscored network
// probe projections. It queries only the returned subtree in one database call.

package media

import (
	"context"

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
