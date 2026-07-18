// This file implements the gateway monitoring topology projection from dashboard report tables.

package media

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
)

// DashboardTopologyInput defines the single-device topology request.
type DashboardTopologyInput struct {
	DeviceId string
}

// DashboardTopologyOutput defines a bounded gateway monitoring topology result.
type DashboardTopologyOutput struct {
	NodeId               string
	NodeName             string
	DeviceId             string
	StreamCount          int
	SessionCount         int
	SessionDetailLimited bool
	GeneratedAt          int64
	Devices              []*DashboardTopologyDevice
}

// DashboardTopologyDevice defines one device subtree.
type DashboardTopologyDevice struct {
	DeviceId     string
	StreamCount  int
	SessionCount int
	Streams      []*DashboardTopologyStream
}

// DashboardTopologyStream defines one stream subtree.
type DashboardTopologyStream struct {
	StreamId     string
	StreamName   string
	Status       string
	BasePlatform *DashboardTopologyBasePlatform
	Gateway      *DashboardTopologyGateway
	Protocols    []*DashboardTopologyProtocol
}

// DashboardTopologyBasePlatform defines the upstream platform node.
type DashboardTopologyBasePlatform struct {
	Bitrate        int
	Resolution     string
	StreamProtocol string
	DeviceCode     string
	SourceUrl      string
}

// DashboardTopologyGateway defines the video gateway node.
type DashboardTopologyGateway struct {
	StreamProtocol string
	Resolution     string
	Bitrate        int
	Fps            float64
	AvgDelay       int
	NodeId         string
	NodeName       string
	InstanceId     string
	InstanceName   string
}

// DashboardTopologyProtocol defines one protocol node under a gateway stream.
type DashboardTopologyProtocol struct {
	ProtocolType string
	ReuseCount   int
	Tenants      []*DashboardTopologyTenant
}

// DashboardTopologyTenant defines one tenant node under a protocol.
type DashboardTopologyTenant struct {
	TenantId           string
	ConcurrentSessions int
	Users              []*DashboardTopologyUser
}

// DashboardTopologyUser defines one active user session node.
type DashboardTopologyUser struct {
	SessionId    string
	UserName     string
	ClientId     string
	ClientIp     string
	ClientType   int
	ProtocolType string
	TenantId     string
	NodeId       string
	InstanceId   string
	StartTime    *int64
	PlayDuration int
}

type dashboardTopologyCountRow struct {
	StreamId     string `orm:"stream_id"`
	ProtocolType string `orm:"protocol_type"`
	TenantId     string `orm:"tenant_id"`
	SessionCount int    `orm:"session_count"`
}

type dashboardTopologyStreamBuildState struct {
	stream      *DashboardTopologyStream
	protocols   map[string]*DashboardTopologyProtocol
	protocolSeq []string
}

type dashboardTopologyDeviceBuildState struct {
	device *DashboardTopologyDevice
}

// GetDashboardTopology returns a bounded topology projection for the gateway monitoring page.
func (s *serviceImpl) GetDashboardTopology(ctx context.Context, in DashboardTopologyInput) (*DashboardTopologyOutput, error) {
	if err := validateMediaReportTablesReady(ctx); err != nil {
		return nil, err
	}

	deviceID := strings.TrimSpace(in.DeviceId)
	if deviceID == "" {
		return nil, bizerr.NewCode(CodeMediaDeviceIDRequired)
	}

	streams, err := dashboardTopologyStreams(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	streamIDs := dashboardStreamEntityIDs(streams)
	nodeInfo, err := s.dashboardTopologyNodeInfo(ctx, streams)
	if err != nil {
		return nil, err
	}
	output := &DashboardTopologyOutput{
		DeviceId:    deviceID,
		GeneratedAt: time.Now().UnixMilli(),
	}
	if nodeInfo != nil {
		output.NodeId = nodeInfo.NodeId
		output.NodeName = nodeInfo.NodeName
	}
	if len(streamIDs) == 0 {
		output.Devices = []*DashboardTopologyDevice{}
		return output, nil
	}

	countRows, err := dashboardTopologySessionCounts(ctx, streamIDs, deviceID)
	if err != nil {
		return nil, err
	}
	sessions, err := dashboardTopologySessions(ctx, streamIDs, deviceID)
	if err != nil {
		return nil, err
	}

	output.SessionDetailLimited = len(sessions) == dashboardReadLimit
	output.Devices = buildDashboardTopologyDevices(streams, countRows, sessions)
	output.StreamCount = len(streams)
	output.SessionCount = dashboardTopologyTotalSessionCount(countRows)
	return output, nil
}

// dashboardTopologyStreams reads bounded active streams matching topology filters.
func dashboardTopologyStreams(
	ctx context.Context,
	deviceID string,
) ([]*dashboardStreamEntity, error) {
	columns := dao.MediaReportStream.Columns()
	items := make([]*dashboardStreamEntity, 0)
	err := dao.MediaReportStream.Ctx(ctx).
		Fields(
			columns.StreamId,
			columns.StreamName,
			columns.Status,
			columns.Bitrate,
			columns.Resolution,
			columns.ProtocolType,
			columns.DeviceId,
			columns.SourceUrl,
			columns.Fps,
			columns.AvgDelay,
			columns.NodeId,
			columns.NodeName,
			columns.InstanceId,
			columns.InstanceName,
			columns.ProtocolSummary,
		).
		Where(columns.DeviceId, deviceID).
		WhereNull(columns.CloseTime).
		OrderAsc(columns.DeviceId).
		OrderDesc(columns.ReportTime).
		OrderAsc(columns.StreamId).
		Limit(dashboardReadLimit).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return items, nil
}

// dashboardTopologySessions reads bounded active session details for the selected streams.
func dashboardTopologySessions(
	ctx context.Context,
	streamIDs []string,
	deviceID string,
) ([]*dashboardSessionEntity, error) {
	columns := dao.MediaReportSession.Columns()
	sessions := make([]*dashboardSessionEntity, 0)
	err := dashboardTopologySessionModel(ctx, streamIDs, deviceID).
		Fields(
			columns.StreamId,
			columns.ProtocolType,
			columns.TenantId,
			columns.SessionId,
			columns.UserName,
			columns.ClientId,
			columns.ClientIp,
			columns.ClientType,
			columns.NodeId,
			columns.InstanceId,
			columns.StartTime,
			columns.CloseTime,
			columns.PlayDuration,
		).
		OrderAsc(columns.StreamId).
		OrderAsc(columns.ProtocolType).
		OrderAsc(columns.TenantId).
		OrderDesc(columns.ReportTime).
		OrderAsc(columns.SessionId).
		Limit(dashboardReadLimit).
		Scan(&sessions)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return sessions, nil
}

// dashboardTopologySessionCounts returns accurate active session counts by stream, protocol and tenant.
func dashboardTopologySessionCounts(
	ctx context.Context,
	streamIDs []string,
	deviceID string,
) ([]*dashboardTopologyCountRow, error) {
	columns := dao.MediaReportSession.Columns()
	rows := make([]*dashboardTopologyCountRow, 0)
	err := dashboardTopologySessionModel(ctx, streamIDs, deviceID).
		Fields(columns.StreamId, columns.ProtocolType, columns.TenantId, "COUNT(*) AS session_count").
		Group(columns.StreamId, columns.ProtocolType, columns.TenantId).
		OrderAsc(columns.StreamId).
		OrderAsc(columns.ProtocolType).
		OrderAsc(columns.TenantId).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDashboardQueryFailed)
	}
	return rows, nil
}

// dashboardTopologySessionModel builds the shared database-side active session filter.
func dashboardTopologySessionModel(
	ctx context.Context,
	streamIDs []string,
	deviceID string,
) *gdb.Model {
	columns := dao.MediaReportSession.Columns()
	return dao.MediaReportSession.Ctx(ctx).
		WhereIn(columns.StreamId, streamIDs).
		Where(columns.DeviceId, deviceID).
		WhereNull(columns.CloseTime)
}

// dashboardTopologyNodeInfo resolves the selected device node label from stream projections.
func (s *serviceImpl) dashboardTopologyNodeInfo(
	ctx context.Context,
	streams []*dashboardStreamEntity,
) (*DashboardInstanceNodeInfo, error) {
	for _, stream := range streams {
		if stream == nil || strings.TrimSpace(stream.NodeId) == "" {
			continue
		}
		if strings.TrimSpace(stream.NodeName) != "" {
			return &DashboardInstanceNodeInfo{
				NodeId:   stream.NodeId,
				NodeName: stream.NodeName,
			}, nil
		}
		nodeInfo, err := s.dashboardNodeInfo(ctx, stream.NodeId)
		if err != nil || nodeInfo != nil {
			return nodeInfo, err
		}
		return &DashboardInstanceNodeInfo{NodeId: stream.NodeId}, nil
	}
	return nil, nil
}

// buildDashboardTopologyDevices assembles the bounded topology tree in memory.
func buildDashboardTopologyDevices(
	streams []*dashboardStreamEntity,
	countRows []*dashboardTopologyCountRow,
	sessions []*dashboardSessionEntity,
) []*DashboardTopologyDevice {
	countsByStream := dashboardTopologyCountsByStream(countRows)
	usersByKey := dashboardTopologyUsersByKey(sessions)
	devices := make(map[string]*dashboardTopologyDeviceBuildState)
	deviceSeq := make([]string, 0)

	for _, stream := range streams {
		if stream == nil {
			continue
		}
		deviceID := strings.TrimSpace(stream.DeviceId)
		if deviceID == "" {
			continue
		}
		deviceState, ok := devices[deviceID]
		if !ok {
			deviceState = &dashboardTopologyDeviceBuildState{device: &DashboardTopologyDevice{
				DeviceId: deviceID,
				Streams:  []*DashboardTopologyStream{},
			}}
			devices[deviceID] = deviceState
			deviceSeq = append(deviceSeq, deviceID)
		}

		streamState := buildDashboardTopologyStreamState(stream, countsByStream[stream.StreamId], usersByKey)
		deviceState.device.Streams = append(deviceState.device.Streams, streamState.stream)
		deviceState.device.StreamCount++
		deviceState.device.SessionCount += dashboardTopologyStreamSessionCount(streamState.stream)
	}

	sort.Strings(deviceSeq)
	result := make([]*DashboardTopologyDevice, 0, len(deviceSeq))
	for _, deviceID := range deviceSeq {
		device := devices[deviceID].device
		if len(device.Streams) == 0 {
			continue
		}
		result = append(result, device)
	}
	return result
}

// buildDashboardTopologyStreamState converts one stream and preloaded aggregates to a stream subtree.
func buildDashboardTopologyStreamState(
	stream *dashboardStreamEntity,
	countRows []*dashboardTopologyCountRow,
	usersByKey map[string][]*DashboardTopologyUser,
) *dashboardTopologyStreamBuildState {
	state := &dashboardTopologyStreamBuildState{
		stream: &DashboardTopologyStream{
			StreamId:     stream.StreamId,
			StreamName:   stream.StreamName,
			Status:       stream.Status,
			BasePlatform: buildDashboardTopologyBasePlatform(stream),
			Gateway:      buildDashboardTopologyGateway(stream),
			Protocols:    []*DashboardTopologyProtocol{},
		},
		protocols: map[string]*DashboardTopologyProtocol{},
	}
	for _, summary := range decodeDashboardProtocolSummary(stream.ProtocolSummary) {
		if summary == nil {
			continue
		}
		protocol := strings.TrimSpace(summary.ProtocolType)
		if protocol == "" {
			continue
		}
		dashboardTopologyEnsureProtocol(state, protocol)
	}
	if strings.TrimSpace(stream.ProtocolType) != "" {
		dashboardTopologyEnsureProtocol(state, stream.ProtocolType)
	}
	for _, row := range countRows {
		if row == nil {
			continue
		}
		protocol := strings.TrimSpace(row.ProtocolType)
		if protocol == "" {
			continue
		}
		protocolNode := dashboardTopologyEnsureProtocol(state, protocol)
		protocolNode.ReuseCount += row.SessionCount
		tenantNode := dashboardTopologyEnsureTenant(protocolNode, row.TenantId)
		tenantNode.ConcurrentSessions += row.SessionCount
		tenantNode.Users = append(tenantNode.Users, usersByKey[dashboardTopologyKey(stream.StreamId, protocol, row.TenantId)]...)
	}
	state.stream.Protocols = dashboardTopologyOrderedProtocols(state)
	return state
}

// buildDashboardTopologyBasePlatform converts stream projection fields to the upstream platform node.
func buildDashboardTopologyBasePlatform(stream *dashboardStreamEntity) *DashboardTopologyBasePlatform {
	return &DashboardTopologyBasePlatform{
		Bitrate:        stream.Bitrate,
		Resolution:     stream.Resolution,
		StreamProtocol: stream.ProtocolType,
		DeviceCode:     stream.DeviceId,
		SourceUrl:      stream.SourceUrl,
	}
}

// buildDashboardTopologyGateway converts stream projection fields to the video gateway node.
func buildDashboardTopologyGateway(stream *dashboardStreamEntity) *DashboardTopologyGateway {
	return &DashboardTopologyGateway{
		StreamProtocol: stream.ProtocolType,
		Resolution:     stream.Resolution,
		Bitrate:        stream.Bitrate,
		Fps:            stream.Fps,
		AvgDelay:       stream.AvgDelay,
		NodeId:         stream.NodeId,
		NodeName:       stream.NodeName,
		InstanceId:     stream.InstanceId,
		InstanceName:   stream.InstanceName,
	}
}

// dashboardTopologyEnsureProtocol returns a protocol node, creating it when needed.
func dashboardTopologyEnsureProtocol(state *dashboardTopologyStreamBuildState, protocol string) *DashboardTopologyProtocol {
	if item, ok := state.protocols[protocol]; ok {
		return item
	}
	item := &DashboardTopologyProtocol{
		ProtocolType: protocol,
		Tenants:      []*DashboardTopologyTenant{},
	}
	state.protocols[protocol] = item
	state.protocolSeq = append(state.protocolSeq, protocol)
	return item
}

// dashboardTopologyEnsureTenant returns a tenant node, creating it when needed.
func dashboardTopologyEnsureTenant(protocol *DashboardTopologyProtocol, tenantID string) *DashboardTopologyTenant {
	for _, item := range protocol.Tenants {
		if item != nil && item.TenantId == tenantID {
			return item
		}
	}
	item := &DashboardTopologyTenant{
		TenantId: tenantID,
		Users:    []*DashboardTopologyUser{},
	}
	protocol.Tenants = append(protocol.Tenants, item)
	return item
}

// dashboardTopologyOrderedProtocols returns protocol nodes in stream-summary order plus sorted extras.
func dashboardTopologyOrderedProtocols(state *dashboardTopologyStreamBuildState) []*DashboardTopologyProtocol {
	seen := make(map[string]struct{}, len(state.protocolSeq))
	result := make([]*DashboardTopologyProtocol, 0, len(state.protocols))
	for _, protocol := range state.protocolSeq {
		if _, ok := seen[protocol]; ok {
			continue
		}
		seen[protocol] = struct{}{}
		result = append(result, state.protocols[protocol])
	}

	extras := make([]string, 0, len(state.protocols))
	for protocol := range state.protocols {
		if _, ok := seen[protocol]; ok {
			continue
		}
		extras = append(extras, protocol)
	}
	sort.Strings(extras)
	for _, protocol := range extras {
		result = append(result, state.protocols[protocol])
	}
	return result
}

// dashboardTopologyCountsByStream groups aggregate rows by stream ID.
func dashboardTopologyCountsByStream(rows []*dashboardTopologyCountRow) map[string][]*dashboardTopologyCountRow {
	result := make(map[string][]*dashboardTopologyCountRow)
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.StreamId) == "" {
			continue
		}
		result[row.StreamId] = append(result[row.StreamId], row)
	}
	return result
}

// dashboardTopologyUsersByKey groups bounded user session detail rows by stream, protocol and tenant.
func dashboardTopologyUsersByKey(sessions []*dashboardSessionEntity) map[string][]*DashboardTopologyUser {
	result := make(map[string][]*DashboardTopologyUser)
	for _, session := range sessions {
		if session == nil {
			continue
		}
		key := dashboardTopologyKey(session.StreamId, session.ProtocolType, session.TenantId)
		result[key] = append(result[key], buildDashboardTopologyUser(session))
	}
	return result
}

// buildDashboardTopologyUser converts one active session to a user node.
func buildDashboardTopologyUser(session *dashboardSessionEntity) *DashboardTopologyUser {
	return &DashboardTopologyUser{
		SessionId:    session.SessionId,
		UserName:     session.UserName,
		ClientId:     session.ClientId,
		ClientIp:     session.ClientIp,
		ClientType:   session.ClientType,
		ProtocolType: session.ProtocolType,
		TenantId:     session.TenantId,
		NodeId:       session.NodeId,
		InstanceId:   session.InstanceId,
		StartTime:    formatTime(session.StartTime),
		PlayDuration: dashboardElapsedSeconds(session.StartTime, session.CloseTime, session.PlayDuration),
	}
}

// dashboardTopologyKey joins stream, protocol and tenant dimensions for in-memory grouping.
func dashboardTopologyKey(streamID string, protocolType string, tenantID string) string {
	return streamID + "\x00" + protocolType + "\x00" + tenantID
}

// dashboardTopologyTotalSessionCount sums aggregate count rows.
func dashboardTopologyTotalSessionCount(rows []*dashboardTopologyCountRow) int {
	total := 0
	for _, row := range rows {
		if row == nil {
			continue
		}
		total += row.SessionCount
	}
	return total
}

// dashboardTopologyStreamSessionCount sums protocol reuse counts under one stream.
func dashboardTopologyStreamSessionCount(stream *DashboardTopologyStream) int {
	if stream == nil {
		return 0
	}
	total := 0
	for _, protocol := range stream.Protocols {
		if protocol == nil {
			continue
		}
		total += protocol.ReuseCount
	}
	return total
}
