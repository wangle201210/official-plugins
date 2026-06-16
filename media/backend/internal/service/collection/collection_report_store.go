// This file persists normalized media collection reports into dashboard read models.

package collection

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/cachecap"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
	"lina-plugin-media/backend/internal/model/entity"
)

// reportWriter handles data-report command business writes.
type reportWriter interface {
	HandleMachineMetric(ctx context.Context, metric *gen.MachineMetric) error
	HandleNetworkMetric(ctx context.Context, metric *gen.NetworkMetric) error
	HandleStreamMetric(ctx context.Context, subcmd uint8, metric *gen.StreamMetric) error
	HandleSessionMetric(ctx context.Context, subcmd uint8, metric *gen.SessionMetric) error
}

const (
	reportCounterCacheNamespace  = "collection-report-counters"
	reportCounterStreamOwnerKey  = "stream-owner:"
	reportCounterSessionOwnerKey = "session-owner:"
	reportCounterLiveStreamsKey  = "instance-live-streams:"
	reportCounterSessionsKey     = "instance-sessions:"
)

// reportRuntime implements reportWriter using media plugin DAO projections and host shared cache.
type reportRuntime struct {
	cache cachecap.Service // cache stores lifecycle ownership and counters across pods.
}

// newReportRuntime creates the default collection report writer.
func newReportRuntime(cacheSvc cachecap.Service) reportWriter {
	return &reportRuntime{cache: cacheSvc}
}

// HandleMachineMetric writes one latest instance projection.
func (w *reportRuntime) HandleMachineMetric(ctx context.Context, metric *gen.MachineMetric) error {
	report, ok := normalizeMachineMetric(metric)
	if !ok {
		return nil
	}
	return upsertReportInstance(ctx, report)
}

// HandleNetworkMetric updates one node's network projection fields.
func (w *reportRuntime) HandleNetworkMetric(ctx context.Context, metric *gen.NetworkMetric) error {
	report, ok := normalizeNetworkMetric(metric)
	if !ok {
		return nil
	}
	return upsertReportNodeNetwork(ctx, report)
}

// HandleStreamMetric writes one latest stream projection and applies stream lifecycle counters.
func (w *reportRuntime) HandleStreamMetric(ctx context.Context, subcmd uint8, metric *gen.StreamMetric) error {
	report, ok := normalizeStreamMetric(metric)
	if !ok {
		return nil
	}
	isDelete := subcmd == uint8(gen.SCMDDataReport_STREAM_DELETE)
	if subcmd != uint8(gen.SCMDDataReport_STREAM_ADD) && subcmd != uint8(gen.SCMDDataReport_STREAM_DELETE) {
		return upsertReportStream(ctx, report)
	}
	if !isDelete {
		if err := upsertReportStream(ctx, report); err != nil {
			return err
		}
		if err := clearReportStreamCloseTime(ctx, report.streamID); err != nil {
			return err
		}
	}
	_, err := w.applyCounterEvent(ctx, counterEvent{
		kind:       counterEventKindStream,
		add:        !isDelete,
		resourceID: report.streamID,
		instanceID: report.instanceID,
		reportTime: report.reportTime,
	})
	if err != nil {
		return err
	}
	if isDelete {
		return markReportStreamClosed(ctx, report.streamID, report.reportTime)
	}
	return nil
}

// HandleSessionMetric writes one latest session projection and applies session lifecycle counters.
func (w *reportRuntime) HandleSessionMetric(ctx context.Context, subcmd uint8, metric *gen.SessionMetric) error {
	report, ok := normalizeSessionMetric(metric)
	if !ok {
		return nil
	}
	isDelete := subcmd == uint8(gen.SCMDDataReport_SESSION_DELETE)
	if subcmd != uint8(gen.SCMDDataReport_SESSION_ADD) && subcmd != uint8(gen.SCMDDataReport_SESSION_DELETE) {
		return upsertReportSession(ctx, report)
	}
	if !isDelete {
		if err := upsertReportSession(ctx, report); err != nil {
			return err
		}
		if err := clearReportSessionCloseTime(ctx, report.sessionID); err != nil {
			return err
		}
	}
	_, err := w.applyCounterEvent(ctx, counterEvent{
		kind:       counterEventKindSession,
		add:        !isDelete,
		resourceID: report.sessionID,
		instanceID: report.instanceID,
		reportTime: report.reportTime,
	})
	if err != nil {
		return err
	}
	if isDelete {
		return markReportSessionClosed(ctx, report.sessionID, report.reportTime)
	}
	return nil
}

// upsertReportInstance stores the latest instance projection by instance_id.
func upsertReportInstance(ctx context.Context, report instanceReport) error {
	cols := dao.MediaReportInstance.Columns()
	_, err := dao.MediaReportInstance.Ctx(ctx).
		Data(do.MediaReportInstance{
			InstanceId:      report.instanceID,
			InstanceName:    report.instanceName,
			NodeId:          report.nodeID,
			NodeName:        report.nodeName,
			Region:          report.region,
			NodeStatus:      report.nodeStatus,
			Status:          report.status,
			CpuAllocated:    report.cpuAllocated,
			CpuLoad:         report.cpuLoad,
			MemoryAllocated: report.memoryAllocated,
			MemoryUsed:      report.memoryUsed,
			DiskIoRead:      report.diskIoRead,
			DiskIoWrite:     report.diskIoWrite,
			NetworkIn:       report.networkIn,
			NetworkOut:      report.networkOut,
			StartTime:       report.startTime,
			Version:         report.version,
			ReportTime:      report.reportTime,
		}).
		OnConflict(cols.InstanceId).
		OnDuplicate(
			cols.InstanceName,
			cols.NodeId,
			cols.NodeName,
			cols.Region,
			cols.NodeStatus,
			cols.Status,
			cols.CpuAllocated,
			cols.CpuLoad,
			cols.MemoryAllocated,
			cols.MemoryUsed,
			cols.DiskIoRead,
			cols.DiskIoWrite,
			cols.NetworkIn,
			cols.NetworkOut,
			cols.StartTime,
			cols.Version,
			cols.ReportTime,
		).
		Save()
	return err
}

type counterEventKind uint8

const (
	counterEventKindStream counterEventKind = iota + 1
	counterEventKindSession
)

type counterEvent struct {
	kind       counterEventKind
	add        bool
	resourceID string
	instanceID string
	reportTime int64
}

type counterEventResult struct {
	counters []instanceCounterSnapshot
}

type instanceCounterSnapshot struct {
	instanceID  string
	liveStreams int32
	sessions    int32
	reportTime  int64
}

// applyCounterEvent updates shared stream/session ownership and returns affected instance counters.
func (w *reportRuntime) applyCounterEvent(ctx context.Context, event counterEvent) (counterEventResult, error) {
	event.resourceID = firstNonBlank(event.resourceID)
	event.instanceID = firstNonBlank(event.instanceID)
	if event.resourceID == "" || (event.add && event.instanceID == "") {
		return counterEventResult{}, nil
	}
	if w == nil || w.cache == nil {
		return counterEventResult{}, gerror.New("media collection report counters require host cache service")
	}
	switch event.kind {
	case counterEventKindStream, counterEventKindSession:
	default:
		return counterEventResult{}, nil
	}
	return w.applyCounterEventLocked(ctx, event)
}

// applyCounterEventLocked serializes one resource's lifecycle event through its projection row.
func (w *reportRuntime) applyCounterEventLocked(ctx context.Context, event counterEvent) (counterEventResult, error) {
	if event.kind == counterEventKindStream {
		return w.applyStreamCounterEventLocked(ctx, event)
	}
	return w.applySessionCounterEventLocked(ctx, event)
}

// applyStreamCounterEventLocked locks one stream row before changing shared counters.
func (w *reportRuntime) applyStreamCounterEventLocked(ctx context.Context, event counterEvent) (counterEventResult, error) {
	var result counterEventResult
	err := dao.MediaReportStream.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := lockReportStream(ctx, event.resourceID); err != nil {
			return err
		}
		applied, err := w.applySharedCounterEvent(ctx, event)
		if err != nil {
			return err
		}
		if err = updateReportInstanceCounters(ctx, applied.counters); err != nil {
			return err
		}
		result = applied
		return nil
	})
	return result, err
}

// applySessionCounterEventLocked locks one session row before changing shared counters.
func (w *reportRuntime) applySessionCounterEventLocked(ctx context.Context, event counterEvent) (counterEventResult, error) {
	var result counterEventResult
	err := dao.MediaReportSession.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := lockReportSession(ctx, event.resourceID); err != nil {
			return err
		}
		applied, err := w.applySharedCounterEvent(ctx, event)
		if err != nil {
			return err
		}
		if err = updateReportInstanceCounters(ctx, applied.counters); err != nil {
			return err
		}
		result = applied
		return nil
	})
	return result, err
}

// applySharedCounterEvent applies one idempotent lifecycle event to shared cache state.
func (w *reportRuntime) applySharedCounterEvent(ctx context.Context, event counterEvent) (counterEventResult, error) {
	ownerKey := event.ownerCacheKey()
	current, exists, err := w.currentCounterOwner(ctx, ownerKey)
	if err != nil {
		return counterEventResult{}, err
	}

	var deltas map[string]instanceCounterDelta
	if event.add {
		if exists && current == event.instanceID {
			counters, err := w.readCounterSnapshots(ctx, []string{event.instanceID}, event.reportTime)
			return counterEventResult{counters: counters}, err
		}
		if _, err = w.cache.Set(ctx, reportCounterCacheNamespace, ownerKey, event.instanceID, 0); err != nil {
			return counterEventResult{}, err
		}
		deltas = event.addDeltas(current, exists)
	} else {
		if !exists {
			if event.instanceID == "" {
				return counterEventResult{}, nil
			}
			counters, err := w.readCounterSnapshots(ctx, []string{event.instanceID}, event.reportTime)
			return counterEventResult{counters: counters}, err
		}
		if event.instanceID == "" {
			event.instanceID = current
		}
		if current != event.instanceID {
			counters, err := w.readCounterSnapshots(ctx, []string{event.instanceID}, event.reportTime)
			return counterEventResult{counters: counters}, err
		}
		if err = w.cache.Delete(ctx, reportCounterCacheNamespace, ownerKey); err != nil {
			return counterEventResult{}, err
		}
		deltas = event.deleteDeltas()
	}

	counters, err := w.applyCounterDeltas(ctx, deltas, event.reportTime)
	if err != nil {
		return counterEventResult{}, err
	}
	return counterEventResult{counters: counters}, nil
}

type instanceCounterDelta struct {
	liveStreams int64
	sessions    int64
}

// ownerCacheKey returns the shared cache key that stores one resource owner.
func (event counterEvent) ownerCacheKey() string {
	if event.kind == counterEventKindStream {
		return reportCounterStreamOwnerKey + event.resourceID
	}
	return reportCounterSessionOwnerKey + event.resourceID
}

// addDeltas returns all instance counter changes caused by an ADD event.
func (event counterEvent) addDeltas(current string, exists bool) map[string]instanceCounterDelta {
	deltas := map[string]instanceCounterDelta{
		event.instanceID: event.delta(1),
	}
	if exists && current != "" && current != event.instanceID {
		deltas[current] = event.delta(-1)
	}
	return deltas
}

// deleteDeltas returns all instance counter changes caused by a DELETE event.
func (event counterEvent) deleteDeltas() map[string]instanceCounterDelta {
	return map[string]instanceCounterDelta{
		event.instanceID: event.delta(-1),
	}
}

// delta returns a typed counter delta for this event kind.
func (event counterEvent) delta(value int64) instanceCounterDelta {
	if event.kind == counterEventKindStream {
		return instanceCounterDelta{liveStreams: value}
	}
	return instanceCounterDelta{sessions: value}
}

// currentCounterOwner reads the existing resource owner from shared cache.
func (w *reportRuntime) currentCounterOwner(ctx context.Context, ownerKey string) (string, bool, error) {
	item, ok, err := w.cache.Get(ctx, reportCounterCacheNamespace, ownerKey)
	if err != nil || !ok {
		return "", ok, err
	}
	return firstNonBlank(item.Value), true, nil
}

// applyCounterDeltas applies shared cache increments and returns the affected instance snapshots.
func (w *reportRuntime) applyCounterDeltas(
	ctx context.Context,
	deltas map[string]instanceCounterDelta,
	reportTime int64,
) ([]instanceCounterSnapshot, error) {
	instanceIDs := sortedCounterInstanceIDs(deltas)
	for _, instanceID := range instanceIDs {
		if err := lockReportInstance(ctx, instanceID); err != nil {
			return nil, err
		}
	}
	counters := make([]instanceCounterSnapshot, 0, len(instanceIDs))
	for _, instanceID := range instanceIDs {
		delta := deltas[instanceID]
		if delta.liveStreams != 0 {
			if err := w.incrementCounterValue(ctx, reportCounterLiveStreamsKey+instanceID, delta.liveStreams); err != nil {
				return nil, err
			}
		}
		if delta.sessions != 0 {
			if err := w.incrementCounterValue(ctx, reportCounterSessionsKey+instanceID, delta.sessions); err != nil {
				return nil, err
			}
		}
		counter, err := w.readCounterSnapshot(ctx, instanceID, reportTime)
		if err != nil {
			return nil, err
		}
		counters = append(counters, counter)
	}
	return counters, nil
}

// incrementCounterValue applies one delta and keeps the shared cache value non-negative.
func (w *reportRuntime) incrementCounterValue(ctx context.Context, key string, delta int64) error {
	item, err := w.cache.Incr(ctx, reportCounterCacheNamespace, key, delta, 0)
	if err != nil {
		return err
	}
	if item != nil && item.ValueKind == cachecap.CacheValueKindInt && item.IntValue < 0 {
		_, err = w.cache.Incr(ctx, reportCounterCacheNamespace, key, -item.IntValue, 0)
		return err
	}
	return nil
}

// readCounterSnapshots reads current shared cache counters for the requested instances.
func (w *reportRuntime) readCounterSnapshots(ctx context.Context, instanceIDs []string, reportTime int64) ([]instanceCounterSnapshot, error) {
	sort.Strings(instanceIDs)
	counters := make([]instanceCounterSnapshot, 0, len(instanceIDs))
	for _, instanceID := range instanceIDs {
		if err := lockReportInstance(ctx, instanceID); err != nil {
			return nil, err
		}
		counter, err := w.readCounterSnapshot(ctx, instanceID, reportTime)
		if err != nil {
			return nil, err
		}
		counters = append(counters, counter)
	}
	return counters, nil
}

// readCounterSnapshot reads one instance's shared counter state.
func (w *reportRuntime) readCounterSnapshot(ctx context.Context, instanceID string, reportTime int64) (instanceCounterSnapshot, error) {
	liveStreams, err := w.currentCounterValue(ctx, reportCounterLiveStreamsKey+instanceID)
	if err != nil {
		return instanceCounterSnapshot{}, err
	}
	sessions, err := w.currentCounterValue(ctx, reportCounterSessionsKey+instanceID)
	if err != nil {
		return instanceCounterSnapshot{}, err
	}
	return instanceCounterSnapshot{
		instanceID:  instanceID,
		liveStreams: normalizeCounterValue(liveStreams),
		sessions:    normalizeCounterValue(sessions),
		reportTime:  reportTime,
	}, nil
}

// sortedCounterInstanceIDs returns non-empty instance keys in a deadlock-safe order.
func sortedCounterInstanceIDs(deltas map[string]instanceCounterDelta) []string {
	instanceIDs := make([]string, 0, len(deltas))
	for instanceID := range deltas {
		if instanceID != "" {
			instanceIDs = append(instanceIDs, instanceID)
		}
	}
	sort.Strings(instanceIDs)
	return instanceIDs
}

// currentCounterValue reads one integer counter from shared cache.
func (w *reportRuntime) currentCounterValue(ctx context.Context, key string) (int64, error) {
	item, ok, err := w.cache.Get(ctx, reportCounterCacheNamespace, key)
	if err != nil || !ok {
		return 0, err
	}
	if item.ValueKind == cachecap.CacheValueKindInt {
		return item.IntValue, nil
	}
	return 0, nil
}

// normalizeCounterValue keeps counters non-negative before writing dashboard projections.
func normalizeCounterValue(value int64) int32 {
	if value <= 0 {
		return 0
	}
	if value > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(value)
}

// updateReportInstanceCounters stores server-side live stream and session counters.
func updateReportInstanceCounters(ctx context.Context, counters []instanceCounterSnapshot) error {
	if len(counters) == 0 {
		return nil
	}
	cols := dao.MediaReportInstance.Columns()
	for _, counter := range counters {
		if counter.instanceID == "" {
			continue
		}
		_, err := dao.MediaReportInstance.Ctx(ctx).
			Data(do.MediaReportInstance{
				InstanceId:  counter.instanceID,
				LiveStreams: counter.liveStreams,
				Sessions:    counter.sessions,
				ReportTime:  normalizeReportTime(counter.reportTime),
			}).
			OnConflict(cols.InstanceId).
			OnDuplicate(cols.LiveStreams, cols.Sessions, cols.ReportTime).
			Save()
		if err != nil {
			return err
		}
	}
	return nil
}

// upsertReportNodeNetwork stores the latest network projection by node_id.
func upsertReportNodeNetwork(ctx context.Context, report networkReport) error {
	if report.hasLatencyMerge {
		return upsertReportNodeNetworkWithLatency(ctx, report)
	}
	cols := dao.MediaReportNode.Columns()
	_, err := dao.MediaReportNode.Ctx(ctx).
		Data(do.MediaReportNode{
			NodeId:         report.nodeID,
			NodeName:       report.nodeID,
			ParentNodeId:   defaultReportParentNodeID,
			Status:         string(reportNodeStatusHealthy),
			NetworkOut:     report.networkOut,
			AvgDelay:       report.rtt,
			LastHeartbeat:  report.lastHeartbeat,
			NodeLatencyMap: report.nodeLatencyMap,
			ReportTime:     report.reportTime,
		}).
		OnConflict(cols.NodeId).
		OnDuplicate(
			cols.NetworkOut,
			cols.AvgDelay,
			cols.LastHeartbeat,
			cols.ReportTime,
		).
		Save()
	return err
}

// upsertReportNodeNetworkWithLatency locks one node row while merging latency map updates.
func upsertReportNodeNetworkWithLatency(ctx context.Context, report networkReport) error {
	cols := dao.MediaReportNode.Columns()
	return dao.MediaReportNode.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		_, err := dao.MediaReportNode.Ctx(ctx).
			Data(do.MediaReportNode{
				NodeId:         report.nodeID,
				NodeName:       report.nodeID,
				ParentNodeId:   defaultReportParentNodeID,
				Status:         string(reportNodeStatusHealthy),
				NodeLatencyMap: defaultReportJSONMap,
				ReportTime:     report.reportTime,
			}).
			OnConflict(cols.NodeId).
			OnDuplicate(cols.NodeId).
			Save()
		if err != nil {
			return err
		}

		latencyMap, err := mergeNodeLatencyMapLocked(ctx, report.nodeID, report.destinationID, report.rtt)
		if err != nil {
			return err
		}
		_, err = dao.MediaReportNode.Ctx(ctx).
			Data(do.MediaReportNode{
				NetworkOut:     report.networkOut,
				AvgDelay:       report.rtt,
				LastHeartbeat:  report.lastHeartbeat,
				NodeLatencyMap: latencyMap,
				ReportTime:     report.reportTime,
			}).
			Where(cols.NodeId, report.nodeID).
			Update()
		return err
	})
}

// upsertReportStream stores the latest stream projection by stream_id.
func upsertReportStream(ctx context.Context, report streamReport) error {
	cols := dao.MediaReportStream.Columns()
	_, err := dao.MediaReportStream.Ctx(ctx).
		Data(do.MediaReportStream{
			StreamId:              report.streamID,
			SourceType:            string(report.sourceType),
			TenantId:              report.tenantID,
			NodeId:                report.nodeID,
			NodeName:              report.nodeName,
			InstanceId:            report.instanceID,
			InstanceName:          report.instanceName,
			SourceUrl:             report.sourceURL,
			ProtocolType:          string(report.protocolType),
			StreamName:            report.streamName,
			Resolution:            report.resolution,
			Fps:                   report.fps,
			Bitrate:               report.bitrate,
			PacketLoss:            report.packetLoss,
			Status:                string(report.status),
			StartTime:             report.startTime,
			Duration:              report.duration,
			AvgDelay:              report.avgDelay,
			ProtocolCount:         report.protocolCount,
			TotalSessionsLifetime: report.totalSessionsLifetime,
			CurrentActiveSessions: report.currentSessions,
			WatermarkEnabled:      report.watermarkEnabled,
			ProtocolSummary:       report.protocolSummary,
			ReportTime:            report.reportTime,
		}).
		OnConflict(cols.StreamId).
		OnDuplicate(
			cols.SourceType,
			cols.TenantId,
			cols.NodeId,
			cols.NodeName,
			cols.InstanceId,
			cols.InstanceName,
			cols.SourceUrl,
			cols.ProtocolType,
			cols.StreamName,
			cols.Resolution,
			cols.Fps,
			cols.Bitrate,
			cols.PacketLoss,
			cols.Status,
			cols.StartTime,
			cols.Duration,
			cols.AvgDelay,
			cols.ProtocolCount,
			cols.TotalSessionsLifetime,
			cols.CurrentActiveSessions,
			cols.WatermarkEnabled,
			cols.ProtocolSummary,
			cols.ReportTime,
		).
		Save()
	return err
}

// upsertReportSession stores the latest session projection by session_id.
func upsertReportSession(ctx context.Context, report sessionReport) error {
	cols := dao.MediaReportSession.Columns()
	data := do.MediaReportSession{
		SessionId:         report.sessionID,
		StreamId:          report.streamID,
		StreamName:        report.streamName,
		TenantId:          report.tenantID,
		ClientId:          report.clientID,
		ClientType:        report.clientType,
		UserName:          report.userName,
		ProtocolType:      report.protocolType,
		StartTime:         report.startTime,
		PlayDuration:      report.playDuration,
		CurrentFps:        report.currentFPS,
		CurrentBitrate:    report.currentBitrate,
		CurrentResolution: report.currentResolution,
		NodeId:            report.nodeID,
		NodeName:          report.nodeName,
		InstanceId:        report.instanceID,
		InstanceName:      report.instanceName,
		LinkHops:          report.linkHops,
		TotalLinkLatency:  report.totalLinkLatency,
		ReportTime:        report.reportTime,
	}
	duplicateCols := []any{
		cols.StreamId,
		cols.StreamName,
		cols.TenantId,
		cols.ClientId,
		cols.ClientType,
		cols.UserName,
		cols.ProtocolType,
		cols.StartTime,
		cols.PlayDuration,
		cols.CurrentFps,
		cols.CurrentBitrate,
		cols.CurrentResolution,
		cols.NodeId,
		cols.NodeName,
		cols.InstanceId,
		cols.InstanceName,
		cols.LinkHops,
		cols.TotalLinkLatency,
		cols.ReportTime,
	}
	if report.clientIP != "" {
		data.ClientIp = report.clientIP
		duplicateCols = append(duplicateCols, cols.ClientIp)
	}
	_, err := dao.MediaReportSession.Ctx(ctx).
		Data(data).
		OnConflict(cols.SessionId).
		OnDuplicate(duplicateCols...).
		Save()
	return err
}

// clearReportStreamCloseTime reopens a stream projection after a lifecycle add event.
func clearReportStreamCloseTime(ctx context.Context, streamID string) error {
	if streamID == "" {
		return nil
	}
	cols := dao.MediaReportStream.Columns()
	_, err := dao.MediaReportStream.Ctx(ctx).
		Where(cols.StreamId, streamID).
		Data(cols.CloseTime, gdb.Raw("NULL")).
		Update()
	return err
}

// clearReportSessionCloseTime reopens a session projection after a lifecycle add event.
func clearReportSessionCloseTime(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	cols := dao.MediaReportSession.Columns()
	_, err := dao.MediaReportSession.Ctx(ctx).
		Where(cols.SessionId, sessionID).
		Data(cols.CloseTime, gdb.Raw("NULL")).
		Update()
	return err
}

// markReportStreamClosed records one stream close time after a lifecycle delete event.
func markReportStreamClosed(ctx context.Context, streamID string, reportTime int64) error {
	if streamID == "" {
		return nil
	}
	_, err := dao.MediaReportStream.Ctx(ctx).
		Data(do.MediaReportStream{
			CloseTime:  reportTimeToGTime(normalizeReportTime(reportTime)),
			ReportTime: normalizeReportTime(reportTime),
		}).
		Where(dao.MediaReportStream.Columns().StreamId, streamID).
		Update()
	return err
}

// markReportSessionClosed records one session close time after a lifecycle delete event.
func markReportSessionClosed(ctx context.Context, sessionID string, reportTime int64) error {
	if sessionID == "" {
		return nil
	}
	_, err := dao.MediaReportSession.Ctx(ctx).
		Data(do.MediaReportSession{
			CloseTime:  reportTimeToGTime(normalizeReportTime(reportTime)),
			ReportTime: normalizeReportTime(reportTime),
		}).
		Where(dao.MediaReportSession.Columns().SessionId, sessionID).
		Update()
	return err
}

// mergeNodeLatencyMap reads and merges one destination latency for a node.
func mergeNodeLatencyMapLocked(ctx context.Context, nodeID string, destinationID string, rtt int32) (string, error) {
	if destinationID == "" {
		return defaultReportJSONMap, nil
	}
	var record *entity.MediaReportNode
	err := dao.MediaReportNode.Ctx(ctx).
		Fields(dao.MediaReportNode.Columns().NodeLatencyMap).
		Where(dao.MediaReportNode.Columns().NodeId, nodeID).
		LockUpdate().
		Scan(&record)
	if err != nil {
		return "", err
	}
	latencyMap := map[string]int32{}
	if record != nil && record.NodeLatencyMap != "" {
		if err = json.Unmarshal([]byte(record.NodeLatencyMap), &latencyMap); err != nil {
			return "", gerror.Wrap(err, "decode media report node latency map failed")
		}
	}
	latencyMap[destinationID] = rtt
	return mustEncodeJSON(latencyMap, defaultReportJSONMap), nil
}

// lockReportStream locks the stream projection row that serializes one stream lifecycle resource.
func lockReportStream(ctx context.Context, streamID string) error {
	var record *entity.MediaReportStream
	return dao.MediaReportStream.Ctx(ctx).
		Fields(dao.MediaReportStream.Columns().StreamId).
		Where(dao.MediaReportStream.Columns().StreamId, streamID).
		LockUpdate().
		Scan(&record)
}

// lockReportSession locks the session projection row that serializes one session lifecycle resource.
func lockReportSession(ctx context.Context, sessionID string) error {
	var record *entity.MediaReportSession
	return dao.MediaReportSession.Ctx(ctx).
		Fields(dao.MediaReportSession.Columns().SessionId).
		Where(dao.MediaReportSession.Columns().SessionId, sessionID).
		LockUpdate().
		Scan(&record)
}

// lockReportInstance locks or creates an instance projection row before counter projection updates.
func lockReportInstance(ctx context.Context, instanceID string) error {
	if instanceID == "" {
		return nil
	}
	cols := dao.MediaReportInstance.Columns()
	_, err := dao.MediaReportInstance.Ctx(ctx).
		Data(do.MediaReportInstance{
			InstanceId: instanceID,
			ReportTime: normalizeReportTime(0),
		}).
		OnConflict(cols.InstanceId).
		OnDuplicate(cols.InstanceId).
		Save()
	if err != nil {
		return err
	}
	var record *entity.MediaReportInstance
	return dao.MediaReportInstance.Ctx(ctx).
		Fields(cols.InstanceId).
		Where(cols.InstanceId, instanceID).
		LockUpdate().
		Scan(&record)
}

var _ reportWriter = (*reportRuntime)(nil)
