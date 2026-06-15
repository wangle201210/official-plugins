// This file implements dashboard read APIs backed by collection report projections.

package media

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
	entitymodel "lina-plugin-media/backend/internal/model/entity"
)

const (
	dashboardRootNodeID       = "0"
	dashboardProtocolActive   = "active"
	dashboardProtocolInactive = "inactive"
	dashboardReadLimit        = maxPageSize
)

// DashboardNodeOverviewInput defines dashboard node overview filters.
type DashboardNodeOverviewInput struct {
	RootNodeId   string // RootNodeId optionally selects one node subtree.
	IncludeEmpty bool   // IncludeEmpty returns a placeholder root when RootNodeId is missing.
}

// DashboardNodeOverviewOutput defines dashboard node overview result.
type DashboardNodeOverviewOutput struct {
	Node *DashboardNodeOverviewItem // Node contains the requested node tree.
}

// DashboardNodeOverviewItem defines one dashboard node overview tree item.
type DashboardNodeOverviewItem struct {
	NodeId          string
	NodeName        string
	Region          string
	Status          string
	ParentNodeId    string
	TotalNodes      int
	AliveNodes      int
	CpuAllocated    float64
	CpuLoad         float64
	MemoryAllocated float64
	MemoryUsed      float64
	DiskIoRead      float64
	DiskIoWrite     float64
	NetworkIn       float64
	NetworkOut      float64
	LiveStreams     int
	Sessions        int
	AvgDelay        int
	LastHeartbeat   *int64
	ReportTime      int64
	NodeLatencyMap  map[string]int
	ChildNodes      []*DashboardNodeOverviewItem
}

// ListDashboardInstancesInput defines dashboard instance list filters.
type ListDashboardInstancesInput struct {
	NodeId  string
	Status  string
	Keyword string
}

// ListDashboardInstancesOutput defines dashboard instances for the current bounded result set.
type ListDashboardInstancesOutput struct {
	NodeInfo *DashboardInstanceNodeInfo
	List     []*DashboardInstanceItem
}

// DashboardInstanceNodeInfo defines the node summary returned with instances.
type DashboardInstanceNodeInfo struct {
	NodeId   string
	NodeName string
	Region   string
	Status   string
}

// DashboardInstanceItem defines one dashboard instance row.
type DashboardInstanceItem struct {
	InstanceId      string
	InstanceName    string
	Status          string
	CpuAllocated    float64
	CpuLoad         float64
	MemoryAllocated float64
	MemoryUsed      float64
	DiskIoRead      float64
	DiskIoWrite     float64
	NetworkIn       float64
	NetworkOut      float64
	LiveStreams     int
	Sessions        int
	StartTime       *int64
	Version         string
}

// ListDashboardStreamsInput defines dashboard stream list filters.
type ListDashboardStreamsInput struct {
	SourceType string
	SourceId   string
	TenantId   string
	NodeId     string
	InstanceId string
	Status     string
	Keyword    string
}

// ListDashboardStreamsOutput defines dashboard streams for the current bounded result set.
type ListDashboardStreamsOutput struct {
	SourceType string
	SourceId   string
	List       []*DashboardStreamItem
}

// DashboardStreamItem defines one dashboard stream row.
type DashboardStreamItem struct {
	SourceUrl             string
	StreamId              string
	StreamName            string
	Resolution            string
	Fps                   float64
	Bitrate               int
	PacketLoss            float64
	Status                string
	StartTime             *int64
	Duration              int
	AvgDelay              int
	ProtocolCount         int
	TotalSessionsLifetime int64
	CurrentActiveSessions int
	WatermarkEnabled      bool
	ProtocolSummary       []*DashboardProtocolItem
}

// DashboardProtocolItem defines one protocol summary row.
type DashboardProtocolItem struct {
	ProtocolType    string `json:"protocol_type"`
	TotalSessions   int    `json:"total_sessions"`
	CurrentSessions int    `json:"current_sessions"`
}

// ListDashboardSessionsInput defines dashboard session list filters.
type ListDashboardSessionsInput struct {
	StreamId     string
	TenantId     string
	ProtocolType string
	NodeId       string
	InstanceId   string
	Keyword      string
}

// ListDashboardSessionsOutput defines dashboard sessions grouped by protocol for the current bounded result set.
type ListDashboardSessionsOutput struct {
	StreamInfo *DashboardSessionStreamInfo
	Protocols  []*DashboardSessionProtocol
}

// DashboardSessionStreamInfo defines the stream summary returned with sessions.
type DashboardSessionStreamInfo struct {
	StreamId   string
	StreamName string
}

// DashboardSessionProtocol defines one protocol group and its active sessions.
type DashboardSessionProtocol struct {
	ProtocolType string
	Status       string
	SessionCount int
	Sessions     []*DashboardSessionItem
}

// DashboardSessionItem defines one dashboard session row.
type DashboardSessionItem struct {
	SessionId         string
	ClientId          string
	ClientIp          string
	ClientType        string
	TenantId          string
	UserName          string
	ProtocolType      string
	StartTime         *int64
	PlayDuration      int
	CurrentFps        float64
	CurrentBitrate    int
	CurrentResolution string
	NodeId            string
	InstanceId        string
	LinkHops          []*DashboardLinkHopItem
	TotalLinkLatency  int
}

// DashboardLinkHopItem defines one session link hop.
type DashboardLinkHopItem struct {
	HopIndex  int    `json:"hop_index"`
	NodeId    string `json:"node_id"`
	LatencyMs int    `json:"latency_ms"`
}

type (
	dashboardNodeEntity     = entitymodel.MediaReportNode
	dashboardInstanceEntity = entitymodel.MediaReportInstance
	dashboardStreamEntity   = entitymodel.MediaReportStream
	dashboardSessionEntity  = entitymodel.MediaReportSession
)

// GetDashboardNodeOverview returns a bounded node tree from latest report projections.
func (s *serviceImpl) GetDashboardNodeOverview(ctx context.Context, in DashboardNodeOverviewInput) (*DashboardNodeOverviewOutput, error) {
	if err := validateMediaReportTablesReady(ctx); err != nil {
		return nil, err
	}

	columns := dao.MediaReportNode.Columns()
	items := make([]*dashboardNodeEntity, 0)
	err := dao.MediaReportNode.Ctx(ctx).
		OrderAsc(columns.ParentNodeId).
		OrderAsc(columns.NodeId).
		Limit(maxPageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}

	nodesByID := make(map[string]*DashboardNodeOverviewItem, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		nodesByID[item.NodeId] = buildDashboardNodeOverviewItem(item)
	}

	roots := make([]*DashboardNodeOverviewItem, 0)
	for _, item := range items {
		if item == nil {
			continue
		}
		node := nodesByID[item.NodeId]
		parentID := dashboardParentNodeID(item.ParentNodeId)
		parent, hasParent := nodesByID[parentID]
		if hasParent && parentID != item.NodeId {
			parent.ChildNodes = append(parent.ChildNodes, node)
			continue
		}
		roots = append(roots, node)
	}

	rootNodeID := strings.TrimSpace(in.RootNodeId)
	if rootNodeID != "" && rootNodeID != dashboardRootNodeID {
		if node, ok := nodesByID[rootNodeID]; ok {
			return &DashboardNodeOverviewOutput{Node: node}, nil
		}
		if in.IncludeEmpty {
			return &DashboardNodeOverviewOutput{Node: &DashboardNodeOverviewItem{
				NodeId:         rootNodeID,
				ParentNodeId:   dashboardRootNodeID,
				NodeLatencyMap: map[string]int{},
				ChildNodes:     []*DashboardNodeOverviewItem{},
			}}, nil
		}
		return &DashboardNodeOverviewOutput{Node: emptyDashboardNodeOverview(rootNodeID)}, nil
	}
	if len(roots) == 1 {
		return &DashboardNodeOverviewOutput{Node: roots[0]}, nil
	}
	root := emptyDashboardNodeOverview(dashboardRootNodeID)
	root.ChildNodes = roots
	root.TotalNodes = len(nodesByID)
	return &DashboardNodeOverviewOutput{Node: root}, nil
}

// ListDashboardInstances returns bounded instance projections and node summary.
func (s *serviceImpl) ListDashboardInstances(ctx context.Context, in ListDashboardInstancesInput) (*ListDashboardInstancesOutput, error) {
	if err := validateMediaReportTablesReady(ctx); err != nil {
		return nil, err
	}

	columns := dao.MediaReportInstance.Columns()
	model := dao.MediaReportInstance.Ctx(ctx)
	model = dashboardWhereEq(model, columns.NodeId, in.NodeId)
	model = dashboardWhereEq(model, columns.Status, in.Status)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.InstanceId+" LIKE ? OR "+columns.InstanceName+" LIKE ? OR "+columns.NodeName+" LIKE ?)",
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}

	items := make([]*dashboardInstanceEntity, 0)
	err := model.
		OrderDesc(columns.ReportTime).
		OrderAsc(columns.InstanceId).
		Limit(dashboardReadLimit).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}

	nodeInfo, err := s.dashboardNodeInfo(ctx, in.NodeId)
	if err != nil {
		return nil, err
	}
	if nodeInfo == nil && len(items) > 0 {
		nodeInfo = buildDashboardInstanceNodeInfoFromInstance(items[0])
	}

	list := make([]*DashboardInstanceItem, 0, len(items))
	for _, item := range items {
		list = append(list, buildDashboardInstanceItem(item))
	}
	return &ListDashboardInstancesOutput{NodeInfo: nodeInfo, List: list}, nil
}

// ListDashboardStreams returns bounded stream projections.
func (s *serviceImpl) ListDashboardStreams(ctx context.Context, in ListDashboardStreamsInput) (*ListDashboardStreamsOutput, error) {
	if err := validateMediaReportTablesReady(ctx); err != nil {
		return nil, err
	}

	columns := dao.MediaReportStream.Columns()
	model := dao.MediaReportStream.Ctx(ctx)
	model = dashboardWhereEq(model, columns.SourceType, in.SourceType)
	model = dashboardWhereEq(model, columns.SourceId, in.SourceId)
	model = dashboardWhereEq(model, columns.TenantId, in.TenantId)
	model = dashboardWhereEq(model, columns.NodeId, in.NodeId)
	model = dashboardWhereEq(model, columns.InstanceId, in.InstanceId)
	model = dashboardWhereEq(model, columns.Status, in.Status)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.StreamId+" LIKE ? OR "+columns.StreamName+" LIKE ? OR "+columns.SourceUrl+" LIKE ? OR "+columns.NodeName+" LIKE ? OR "+columns.InstanceName+" LIKE ?)",
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}

	items := make([]*dashboardStreamEntity, 0)
	err := model.
		OrderDesc(columns.ReportTime).
		OrderAsc(columns.StreamId).
		Limit(dashboardReadLimit).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}

	list := make([]*DashboardStreamItem, 0, len(items))
	for _, item := range items {
		list = append(list, buildDashboardStreamItem(item))
	}
	return &ListDashboardStreamsOutput{
		SourceType: strings.TrimSpace(in.SourceType),
		SourceId:   strings.TrimSpace(in.SourceId),
		List:       list,
	}, nil
}

// ListDashboardSessions returns bounded session projections grouped by protocol.
func (s *serviceImpl) ListDashboardSessions(ctx context.Context, in ListDashboardSessionsInput) (*ListDashboardSessionsOutput, error) {
	if err := validateMediaReportTablesReady(ctx); err != nil {
		return nil, err
	}
	in.StreamId = strings.TrimSpace(in.StreamId)
	if in.StreamId == "" {
		return nil, bizerr.NewCode(CodeMediaDashboardStreamRequired)
	}

	stream, err := s.dashboardStreamEntity(ctx, in.StreamId)
	if err != nil {
		return nil, err
	}

	counts, err := dashboardSessionProtocolCounts(ctx, in)
	if err != nil {
		return nil, err
	}

	columns := dao.MediaReportSession.Columns()
	sessions := make([]*dashboardSessionEntity, 0)
	err = dashboardSessionModel(ctx, in).
		OrderAsc(columns.ProtocolType).
		OrderDesc(columns.ReportTime).
		OrderAsc(columns.SessionId).
		Limit(dashboardReadLimit).
		Scan(&sessions)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}

	streamInfo := buildDashboardSessionStreamInfo(stream, sessions, in.StreamId)
	protocols := buildDashboardSessionProtocols(stream, counts, sessions, in.ProtocolType)
	return &ListDashboardSessionsOutput{StreamInfo: streamInfo, Protocols: protocols}, nil
}

// dashboardSessionModel builds the database-side session filter model.
func dashboardSessionModel(ctx context.Context, in ListDashboardSessionsInput) *gdb.Model {
	columns := dao.MediaReportSession.Columns()
	model := dao.MediaReportSession.Ctx(ctx)
	model = dashboardWhereEq(model, columns.StreamId, in.StreamId)
	model = dashboardWhereEq(model, columns.TenantId, in.TenantId)
	model = dashboardWhereEq(model, columns.ProtocolType, in.ProtocolType)
	model = dashboardWhereEq(model, columns.NodeId, in.NodeId)
	model = dashboardWhereEq(model, columns.InstanceId, in.InstanceId)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.SessionId+" LIKE ? OR "+columns.ClientId+" LIKE ? OR "+columns.ClientIp+" LIKE ? OR "+columns.UserName+" LIKE ? OR "+columns.ProtocolType+" LIKE ?)",
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}
	return model
}

// dashboardSessionProtocolCounts returns per-protocol counts for the current query scope.
func dashboardSessionProtocolCounts(ctx context.Context, in ListDashboardSessionsInput) (map[string]int, error) {
	columns := dao.MediaReportSession.Columns()
	type protocolCountRow struct {
		ProtocolType string `orm:"protocol_type"`
		SessionCount int    `orm:"session_count"`
	}

	rows := make([]*protocolCountRow, 0)
	err := dashboardSessionModel(ctx, in).
		Fields(columns.ProtocolType, "COUNT(*) AS session_count").
		Group(columns.ProtocolType).
		OrderAsc(columns.ProtocolType).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}

	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		counts[row.ProtocolType] = row.SessionCount
	}
	return counts, nil
}

// dashboardWhereEq applies one equality filter when value is non-empty.
func dashboardWhereEq(model *gdb.Model, column string, value string) *gdb.Model {
	if strings.TrimSpace(value) == "" {
		return model
	}
	return model.Where(column, strings.TrimSpace(value))
}

// dashboardNodeInfo reads one node summary from latest node projections.
func (s *serviceImpl) dashboardNodeInfo(ctx context.Context, nodeID string) (*DashboardInstanceNodeInfo, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil, nil
	}
	var node *dashboardNodeEntity
	err := dao.MediaReportNode.Ctx(ctx).
		Where(dao.MediaReportNode.Columns().NodeId, nodeID).
		Scan(&node)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	if node == nil {
		return nil, nil
	}
	return &DashboardInstanceNodeInfo{
		NodeId:   node.NodeId,
		NodeName: node.NodeName,
		Region:   node.Region,
		Status:   node.Status,
	}, nil
}

// dashboardStreamEntity reads one stream projection by stream ID.
func (s *serviceImpl) dashboardStreamEntity(ctx context.Context, streamID string) (*dashboardStreamEntity, error) {
	var stream *dashboardStreamEntity
	err := dao.MediaReportStream.Ctx(ctx).
		Where(dao.MediaReportStream.Columns().StreamId, streamID).
		Scan(&stream)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return stream, nil
}

// validateMediaReportTablesReady verifies dashboard report tables exist.
func validateMediaReportTablesReady(ctx context.Context) error {
	tableNames := []string{
		dao.MediaReportNode.Table(),
		dao.MediaReportInstance.Table(),
		dao.MediaReportStream.Table(),
		dao.MediaReportSession.Table(),
	}
	for _, tableName := range tableNames {
		fields, err := dao.MediaReportNode.DB().TableFields(ctx, tableName)
		if err != nil {
			return bizerr.WrapCode(err, CodeMediaTableCheckFailed)
		}
		if len(fields) == 0 {
			return bizerr.NewCode(CodeMediaTableNotInstalled)
		}
	}
	return nil
}

// buildDashboardNodeOverviewItem converts one node entity to dashboard output.
func buildDashboardNodeOverviewItem(item *dashboardNodeEntity) *DashboardNodeOverviewItem {
	if item == nil {
		return &DashboardNodeOverviewItem{NodeLatencyMap: map[string]int{}, ChildNodes: []*DashboardNodeOverviewItem{}}
	}
	return &DashboardNodeOverviewItem{
		NodeId:          item.NodeId,
		NodeName:        item.NodeName,
		Region:          item.Region,
		Status:          item.Status,
		ParentNodeId:    dashboardParentNodeID(item.ParentNodeId),
		TotalNodes:      item.TotalNodes,
		AliveNodes:      item.AliveNodes,
		CpuAllocated:    item.CpuAllocated,
		CpuLoad:         item.CpuLoad,
		MemoryAllocated: item.MemoryAllocated,
		MemoryUsed:      item.MemoryUsed,
		DiskIoRead:      item.DiskIoRead,
		DiskIoWrite:     item.DiskIoWrite,
		NetworkIn:       item.NetworkIn,
		NetworkOut:      item.NetworkOut,
		LiveStreams:     item.LiveStreams,
		Sessions:        item.Sessions,
		AvgDelay:        item.AvgDelay,
		LastHeartbeat:   formatTime(item.LastHeartbeat),
		ReportTime:      item.ReportTime,
		NodeLatencyMap:  decodeDashboardIntMap(item.NodeLatencyMap),
		ChildNodes:      []*DashboardNodeOverviewItem{},
	}
}

// dashboardParentNodeID normalizes empty parent IDs into the dashboard root value.
func dashboardParentNodeID(parentID string) string {
	parentID = strings.TrimSpace(parentID)
	if parentID == "" {
		return dashboardRootNodeID
	}
	return parentID
}

// emptyDashboardNodeOverview returns an object-shaped empty node response for the fixed frontend contract.
func emptyDashboardNodeOverview(nodeID string) *DashboardNodeOverviewItem {
	return &DashboardNodeOverviewItem{
		NodeId:         strings.TrimSpace(nodeID),
		ParentNodeId:   dashboardRootNodeID,
		NodeLatencyMap: map[string]int{},
		ChildNodes:     []*DashboardNodeOverviewItem{},
	}
}

// buildDashboardInstanceNodeInfoFromInstance falls back to instance-denormalized node fields.
func buildDashboardInstanceNodeInfoFromInstance(item *dashboardInstanceEntity) *DashboardInstanceNodeInfo {
	if item == nil {
		return nil
	}
	return &DashboardInstanceNodeInfo{
		NodeId:   item.NodeId,
		NodeName: item.NodeName,
		Region:   item.Region,
		Status:   item.NodeStatus,
	}
}

// buildDashboardInstanceItem converts one instance entity to dashboard output.
func buildDashboardInstanceItem(item *dashboardInstanceEntity) *DashboardInstanceItem {
	if item == nil {
		return &DashboardInstanceItem{}
	}
	return &DashboardInstanceItem{
		InstanceId:      item.InstanceId,
		InstanceName:    item.InstanceName,
		Status:          item.Status,
		CpuAllocated:    item.CpuAllocated,
		CpuLoad:         item.CpuLoad,
		MemoryAllocated: item.MemoryAllocated,
		MemoryUsed:      item.MemoryUsed,
		DiskIoRead:      item.DiskIoRead,
		DiskIoWrite:     item.DiskIoWrite,
		NetworkIn:       item.NetworkIn,
		NetworkOut:      item.NetworkOut,
		LiveStreams:     item.LiveStreams,
		Sessions:        item.Sessions,
		StartTime:       formatTime(item.StartTime),
		Version:         item.Version,
	}
}

// buildDashboardStreamItem converts one stream entity to dashboard output.
func buildDashboardStreamItem(item *dashboardStreamEntity) *DashboardStreamItem {
	if item == nil {
		return &DashboardStreamItem{ProtocolSummary: []*DashboardProtocolItem{}}
	}
	return &DashboardStreamItem{
		SourceUrl:             item.SourceUrl,
		StreamId:              item.StreamId,
		StreamName:            item.StreamName,
		Resolution:            item.Resolution,
		Fps:                   item.Fps,
		Bitrate:               item.Bitrate,
		PacketLoss:            item.PacketLoss,
		Status:                item.Status,
		StartTime:             formatTime(item.StartTime),
		Duration:              item.Duration,
		AvgDelay:              item.AvgDelay,
		ProtocolCount:         item.ProtocolCount,
		TotalSessionsLifetime: item.TotalSessionsLifetime,
		CurrentActiveSessions: item.CurrentActiveSessions,
		WatermarkEnabled:      item.WatermarkEnabled,
		ProtocolSummary:       decodeDashboardProtocolSummary(item.ProtocolSummary),
	}
}

// buildDashboardSessionStreamInfo returns stream summary using stream row or session fallback.
func buildDashboardSessionStreamInfo(
	stream *dashboardStreamEntity,
	sessions []*dashboardSessionEntity,
	streamID string,
) *DashboardSessionStreamInfo {
	if stream != nil {
		return &DashboardSessionStreamInfo{StreamId: stream.StreamId, StreamName: stream.StreamName}
	}
	for _, session := range sessions {
		if session != nil && strings.TrimSpace(session.StreamName) != "" {
			return &DashboardSessionStreamInfo{StreamId: streamID, StreamName: session.StreamName}
		}
	}
	return &DashboardSessionStreamInfo{StreamId: streamID}
}

// buildDashboardSessionProtocols merges stream protocol summary, aggregate counts, and bounded sessions.
func buildDashboardSessionProtocols(
	stream *dashboardStreamEntity,
	counts map[string]int,
	sessions []*dashboardSessionEntity,
	protocolFilter string,
) []*DashboardSessionProtocol {
	protocolFilter = strings.TrimSpace(protocolFilter)
	orderedProtocols := make([]string, 0)
	seen := make(map[string]struct{})
	for _, item := range dashboardStreamProtocolSummary(stream) {
		if item == nil {
			continue
		}
		protocol := strings.TrimSpace(item.ProtocolType)
		if protocol == "" || (protocolFilter != "" && protocol != protocolFilter) {
			continue
		}
		if _, ok := seen[protocol]; ok {
			continue
		}
		seen[protocol] = struct{}{}
		orderedProtocols = append(orderedProtocols, protocol)
	}

	countProtocols := make([]string, 0, len(counts))
	for protocol := range counts {
		if protocolFilter != "" && protocol != protocolFilter {
			continue
		}
		if _, ok := seen[protocol]; ok {
			continue
		}
		countProtocols = append(countProtocols, protocol)
	}
	sort.Strings(countProtocols)
	orderedProtocols = append(orderedProtocols, countProtocols...)

	sessionsByProtocol := make(map[string][]*DashboardSessionItem)
	for _, session := range sessions {
		if session == nil {
			continue
		}
		protocol := strings.TrimSpace(session.ProtocolType)
		if protocolFilter != "" && protocol != protocolFilter {
			continue
		}
		if _, ok := seen[protocol]; !ok {
			seen[protocol] = struct{}{}
			orderedProtocols = append(orderedProtocols, protocol)
		}
		sessionsByProtocol[protocol] = append(sessionsByProtocol[protocol], buildDashboardSessionItem(session))
	}

	protocols := make([]*DashboardSessionProtocol, 0, len(orderedProtocols))
	for _, protocol := range orderedProtocols {
		sessionCount := counts[protocol]
		status := dashboardProtocolInactive
		if sessionCount > 0 {
			status = dashboardProtocolActive
		}
		protocols = append(protocols, &DashboardSessionProtocol{
			ProtocolType: protocol,
			Status:       status,
			SessionCount: sessionCount,
			Sessions:     sessionsByProtocol[protocol],
		})
	}
	return protocols
}

// buildDashboardSessionItem converts one session entity to dashboard output.
func buildDashboardSessionItem(item *dashboardSessionEntity) *DashboardSessionItem {
	if item == nil {
		return &DashboardSessionItem{LinkHops: []*DashboardLinkHopItem{}}
	}
	return &DashboardSessionItem{
		SessionId:         item.SessionId,
		ClientId:          item.ClientId,
		ClientIp:          item.ClientIp,
		ClientType:        item.ClientType,
		TenantId:          item.TenantId,
		UserName:          item.UserName,
		ProtocolType:      item.ProtocolType,
		StartTime:         formatTime(item.StartTime),
		PlayDuration:      item.PlayDuration,
		CurrentFps:        item.CurrentFps,
		CurrentBitrate:    item.CurrentBitrate,
		CurrentResolution: item.CurrentResolution,
		NodeId:            item.NodeId,
		InstanceId:        item.InstanceId,
		LinkHops:          decodeDashboardLinkHops(item.LinkHops),
		TotalLinkLatency:  item.TotalLinkLatency,
	}
}

// dashboardStreamProtocolSummary returns parsed protocol summary for one stream row.
func dashboardStreamProtocolSummary(stream *dashboardStreamEntity) []*DashboardProtocolItem {
	if stream == nil {
		return []*DashboardProtocolItem{}
	}
	return decodeDashboardProtocolSummary(stream.ProtocolSummary)
}

// decodeDashboardIntMap parses a JSON object into an int map.
func decodeDashboardIntMap(raw string) map[string]int {
	result := make(map[string]int)
	if strings.TrimSpace(raw) == "" {
		return result
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return map[string]int{}
	}
	return result
}

// decodeDashboardProtocolSummary parses protocol summary JSON.
func decodeDashboardProtocolSummary(raw string) []*DashboardProtocolItem {
	if strings.TrimSpace(raw) == "" {
		return []*DashboardProtocolItem{}
	}
	items := make([]*DashboardProtocolItem, 0)
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []*DashboardProtocolItem{}
	}
	if items == nil {
		return []*DashboardProtocolItem{}
	}
	return items
}

// decodeDashboardLinkHops parses session link hop JSON.
func decodeDashboardLinkHops(raw string) []*DashboardLinkHopItem {
	if strings.TrimSpace(raw) == "" {
		return []*DashboardLinkHopItem{}
	}
	items := make([]*DashboardLinkHopItem, 0)
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []*DashboardLinkHopItem{}
	}
	if items == nil {
		return []*DashboardLinkHopItem{}
	}
	return items
}
