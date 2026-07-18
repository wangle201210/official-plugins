// This file implements tenant media-node topology aggregates from dashboard report projections.

package media

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// DashboardTenantTopologyInput defines the single-tenant topology request.
type DashboardTenantTopologyInput struct {
	TenantId string
}

// DashboardTenantTopologyOutput defines tenant totals and bounded node aggregates.
type DashboardTenantTopologyOutput struct {
	TenantId           string
	ConcurrentSessions int
	LiveStreamCount    int
	ProtocolCount      int
	Protocols          []*DashboardTenantTopologyProtocol
	NodeCount          int
	NodeDetailLimited  bool
	GeneratedAt        int64
	Nodes              []*DashboardTenantTopologyNode
}

// DashboardTenantTopologyProtocol defines active stream statistics for one protocol.
type DashboardTenantTopologyProtocol struct {
	ProtocolType    string
	LiveStreamCount int
}

// DashboardTenantTopologyNode defines one media-node aggregate under the tenant.
type DashboardTenantTopologyNode struct {
	NodeId             string
	NodeName           string
	ConcurrentSessions int
	LiveStreamCount    int
	ProtocolCount      int
	Protocols          []*DashboardTenantTopologyProtocol
}

type dashboardTenantProtocolRow struct {
	ProtocolType    string `orm:"protocol_type"`
	LiveStreamCount int    `orm:"live_stream_count"`
}

type dashboardTenantNodeStreamRow struct {
	NodeId          string `orm:"node_id"`
	NodeName        string `orm:"node_name"`
	ProtocolType    string `orm:"protocol_type"`
	LiveStreamCount int    `orm:"live_stream_count"`
}

type dashboardTenantNodeSessionRow struct {
	NodeId             string `orm:"node_id"`
	NodeName           string `orm:"node_name"`
	ConcurrentSessions int    `orm:"concurrent_sessions"`
}

type dashboardTenantNodeBuildState struct {
	node           *DashboardTenantTopologyNode
	protocolCounts map[string]int
}

type dashboardTietaTenantContextKey struct{}

type dashboardTietaTenantContext struct {
	tenantID string
}

// WithDashboardTietaTenantContext binds the authenticated Tieta tenant to a request context.
func WithDashboardTietaTenantContext(ctx context.Context, tenantID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, dashboardTietaTenantContextKey{}, dashboardTietaTenantContext{
		tenantID: strings.TrimSpace(tenantID),
	})
}

// GetDashboardTenantTopology returns tenant totals and bounded media-node aggregates.
func (s *serviceImpl) GetDashboardTenantTopology(
	ctx context.Context,
	in DashboardTenantTopologyInput,
) (*DashboardTenantTopologyOutput, error) {
	tenantID, err := normalizeDashboardTenantID(in.TenantId)
	if err != nil {
		return nil, err
	}
	if err := authorizeDashboardTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := validateMediaReportTablesReady(ctx); err != nil {
		return nil, err
	}
	protocolRows, err := dashboardTenantProtocolRows(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	concurrentSessions, err := dashboardTenantConcurrentSessions(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	streamRows, err := dashboardTenantNodeStreamRows(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	sessionRows, err := dashboardTenantNodeSessionRows(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	states := buildDashboardTenantNodeStates(streamRows, sessionRows)
	nodeIDs := dashboardTenantNodeIDs(states)
	detailLimited := len(streamRows) == dashboardReadLimit || len(sessionRows) == dashboardReadLimit
	if len(nodeIDs) > dashboardReadLimit {
		nodeIDs = nodeIDs[:dashboardReadLimit]
		detailLimited = true
	}
	nodeNames, err := dashboardTenantNodeNames(ctx, nodeIDs)
	if err != nil {
		return nil, err
	}

	protocols := buildDashboardTenantProtocols(protocolRows)
	nodes := buildDashboardTenantNodes(states, nodeIDs, nodeNames)
	return &DashboardTenantTopologyOutput{
		TenantId:           tenantID,
		ConcurrentSessions: concurrentSessions,
		LiveStreamCount:    dashboardTenantLiveStreamCount(protocolRows),
		ProtocolCount:      len(protocols),
		Protocols:          protocols,
		NodeCount:          len(nodes),
		NodeDetailLimited:  detailLimited,
		GeneratedAt:        time.Now().UnixMilli(),
		Nodes:              nodes,
	}, nil
}

// dashboardTenantProtocolRows returns exact tenant-level active stream counts by protocol.
func dashboardTenantProtocolRows(ctx context.Context, tenantID string) ([]*dashboardTenantProtocolRow, error) {
	columns := dao.MediaReportStream.Columns()
	rows := make([]*dashboardTenantProtocolRow, 0)
	err := dao.MediaReportStream.Ctx(ctx).
		Fields(columns.ProtocolType, "COUNT(*) AS live_stream_count").
		Where(do.MediaReportStream{TenantId: tenantID}).
		WhereNull(columns.CloseTime).
		Group(columns.ProtocolType).
		OrderAsc(columns.ProtocolType).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return rows, nil
}

// dashboardTenantConcurrentSessions returns the exact active session count for one tenant.
func dashboardTenantConcurrentSessions(ctx context.Context, tenantID string) (int, error) {
	columns := dao.MediaReportSession.Columns()
	count, err := dao.MediaReportSession.Ctx(ctx).
		Where(do.MediaReportSession{TenantId: tenantID}).
		WhereNull(columns.CloseTime).
		Count()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return count, nil
}

// dashboardTenantNodeStreamRows returns bounded active stream aggregates by node and protocol.
func dashboardTenantNodeStreamRows(ctx context.Context, tenantID string) ([]*dashboardTenantNodeStreamRow, error) {
	columns := dao.MediaReportStream.Columns()
	rows := make([]*dashboardTenantNodeStreamRow, 0)
	err := dao.MediaReportStream.Ctx(ctx).
		Fields(
			columns.NodeId,
			"MAX("+columns.NodeName+") AS node_name",
			columns.ProtocolType,
			"COUNT(*) AS live_stream_count",
		).
		Where(do.MediaReportStream{TenantId: tenantID}).
		WhereNull(columns.CloseTime).
		Group(columns.NodeId, columns.ProtocolType).
		OrderAsc(columns.NodeId).
		OrderAsc(columns.ProtocolType).
		Limit(dashboardReadLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return rows, nil
}

// dashboardTenantNodeSessionRows returns bounded active session aggregates by node.
func dashboardTenantNodeSessionRows(ctx context.Context, tenantID string) ([]*dashboardTenantNodeSessionRow, error) {
	columns := dao.MediaReportSession.Columns()
	rows := make([]*dashboardTenantNodeSessionRow, 0)
	err := dao.MediaReportSession.Ctx(ctx).
		Fields(
			columns.NodeId,
			"MAX("+columns.NodeName+") AS node_name",
			"COUNT(*) AS concurrent_sessions",
		).
		Where(do.MediaReportSession{TenantId: tenantID}).
		WhereNull(columns.CloseTime).
		Group(columns.NodeId).
		OrderAsc(columns.NodeId).
		Limit(dashboardReadLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return rows, nil
}

// dashboardTenantNodeNames reads current node names in one bounded query.
func dashboardTenantNodeNames(ctx context.Context, nodeIDs []string) (map[string]string, error) {
	result := make(map[string]string, len(nodeIDs))
	if len(nodeIDs) == 0 {
		return result, nil
	}

	columns := dao.MediaReportNode.Columns()
	rows := make([]*dashboardNodeEntity, 0, len(nodeIDs))
	err := dao.MediaReportNode.Ctx(ctx).
		Fields(columns.NodeId, columns.NodeName).
		WhereIn(columns.NodeId, nodeIDs).
		OrderAsc(columns.NodeId).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.NodeId) == "" {
			continue
		}
		result[row.NodeId] = row.NodeName
	}
	return result, nil
}

// buildDashboardTenantNodeStates merges pre-aggregated stream and session rows in memory.
func buildDashboardTenantNodeStates(
	streamRows []*dashboardTenantNodeStreamRow,
	sessionRows []*dashboardTenantNodeSessionRow,
) map[string]*dashboardTenantNodeBuildState {
	states := make(map[string]*dashboardTenantNodeBuildState)
	for _, row := range streamRows {
		if row == nil {
			continue
		}
		state := dashboardTenantEnsureNodeState(states, row.NodeId, row.NodeName)
		if state == nil {
			continue
		}
		state.node.LiveStreamCount += row.LiveStreamCount
		protocol := strings.TrimSpace(row.ProtocolType)
		if protocol != "" {
			state.protocolCounts[protocol] += row.LiveStreamCount
		}
	}
	for _, row := range sessionRows {
		if row == nil {
			continue
		}
		state := dashboardTenantEnsureNodeState(states, row.NodeId, row.NodeName)
		if state != nil {
			state.node.ConcurrentSessions += row.ConcurrentSessions
		}
	}
	return states
}

// dashboardTenantEnsureNodeState returns one node build state for a non-empty node ID.
func dashboardTenantEnsureNodeState(
	states map[string]*dashboardTenantNodeBuildState,
	nodeID string,
	nodeName string,
) *dashboardTenantNodeBuildState {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil
	}
	if state, ok := states[nodeID]; ok {
		if state.node.NodeName == "" {
			state.node.NodeName = strings.TrimSpace(nodeName)
		}
		return state
	}
	state := &dashboardTenantNodeBuildState{
		node: &DashboardTenantTopologyNode{
			NodeId:    nodeID,
			NodeName:  strings.TrimSpace(nodeName),
			Protocols: []*DashboardTenantTopologyProtocol{},
		},
		protocolCounts: make(map[string]int),
	}
	states[nodeID] = state
	return state
}

// dashboardTenantNodeIDs returns stable sorted node IDs.
func dashboardTenantNodeIDs(states map[string]*dashboardTenantNodeBuildState) []string {
	nodeIDs := make([]string, 0, len(states))
	for nodeID := range states {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	return nodeIDs
}

// buildDashboardTenantNodes finalizes node names and protocol lists in stable order.
func buildDashboardTenantNodes(
	states map[string]*dashboardTenantNodeBuildState,
	nodeIDs []string,
	nodeNames map[string]string,
) []*DashboardTenantTopologyNode {
	nodes := make([]*DashboardTenantTopologyNode, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		state := states[nodeID]
		if state == nil || state.node == nil {
			continue
		}
		if nodeName := strings.TrimSpace(nodeNames[nodeID]); nodeName != "" {
			state.node.NodeName = nodeName
		}
		state.node.Protocols = dashboardTenantProtocolsFromCounts(state.protocolCounts)
		state.node.ProtocolCount = len(state.node.Protocols)
		nodes = append(nodes, state.node)
	}
	return nodes
}

// buildDashboardTenantProtocols converts tenant-level aggregate rows to response items.
func buildDashboardTenantProtocols(rows []*dashboardTenantProtocolRow) []*DashboardTenantTopologyProtocol {
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		protocol := strings.TrimSpace(row.ProtocolType)
		if protocol != "" {
			counts[protocol] += row.LiveStreamCount
		}
	}
	return dashboardTenantProtocolsFromCounts(counts)
}

// dashboardTenantProtocolsFromCounts returns protocol items in stable order.
func dashboardTenantProtocolsFromCounts(counts map[string]int) []*DashboardTenantTopologyProtocol {
	protocols := make([]string, 0, len(counts))
	for protocol := range counts {
		protocols = append(protocols, protocol)
	}
	sort.Strings(protocols)

	items := make([]*DashboardTenantTopologyProtocol, 0, len(protocols))
	for _, protocol := range protocols {
		items = append(items, &DashboardTenantTopologyProtocol{
			ProtocolType:    protocol,
			LiveStreamCount: counts[protocol],
		})
	}
	return items
}

// dashboardTenantLiveStreamCount sums all active stream protocol groups, including empty protocols.
func dashboardTenantLiveStreamCount(rows []*dashboardTenantProtocolRow) int {
	total := 0
	for _, row := range rows {
		if row != nil {
			total += row.LiveStreamCount
		}
	}
	return total
}

// normalizeDashboardTenantID validates the tenant ID against the report projection width.
func normalizeDashboardTenantID(tenantID string) (string, error) {
	normalized := strings.TrimSpace(tenantID)
	if normalized == "" {
		return "", bizerr.NewCode(CodeMediaDashboardTenantRequired)
	}
	if utf8.RuneCountInString(normalized) > 64 {
		return "", bizerr.NewCode(CodeMediaDashboardTenantTooLong)
	}
	return normalized, nil
}

// authorizeDashboardTenant enforces the tenant carried by Tieta fallback authentication.
func authorizeDashboardTenant(ctx context.Context, requestedTenantID string) error {
	current, ok := ctx.Value(dashboardTietaTenantContextKey{}).(dashboardTietaTenantContext)
	if !ok {
		return nil
	}
	if current.tenantID == "" {
		return bizerr.NewCode(CodeMediaTietaTenantMissing)
	}
	if current.tenantID != requestedTenantID {
		return bizerr.NewCode(CodeMediaTietaTenantMismatch)
	}
	return nil
}
