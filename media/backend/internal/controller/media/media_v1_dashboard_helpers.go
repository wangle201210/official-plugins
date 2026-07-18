// This file maps dashboard service outputs to fixed frontend API DTOs.

package media

import (
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

func dashboardNodeOverviewToDTO(item *mediasvc.DashboardNodeOverviewItem) *v1.DashboardNodeOverviewItem {
	if item == nil {
		return &v1.DashboardNodeOverviewItem{
			NodeLatencyMap: map[string]int{},
			ChildNodes:     []*v1.DashboardNodeOverviewItem{},
		}
	}
	children := make([]*v1.DashboardNodeOverviewItem, 0, len(item.ChildNodes))
	for _, child := range item.ChildNodes {
		children = append(children, dashboardNodeOverviewToDTO(child))
	}
	return &v1.DashboardNodeOverviewItem{
		NodeId:          item.NodeId,
		NodeName:        item.NodeName,
		Region:          item.Region,
		Status:          item.Status,
		ParentNodeId:    item.ParentNodeId,
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
		LastHeartbeat:   item.LastHeartbeat,
		ReportTime:      item.ReportTime,
		NodeLatencyMap:  cloneDashboardIntMap(item.NodeLatencyMap),
		ChildNodes:      children,
	}
}

func dashboardInstanceNodeInfoToDTO(item *mediasvc.DashboardInstanceNodeInfo) *v1.DashboardInstanceNodeInfo {
	if item == nil {
		return &v1.DashboardInstanceNodeInfo{}
	}
	return &v1.DashboardInstanceNodeInfo{
		NodeId:   item.NodeId,
		NodeName: item.NodeName,
		Region:   item.Region,
		Status:   item.Status,
	}
}

func dashboardInstanceListToDTO(items []*mediasvc.DashboardInstanceItem) []*v1.DashboardInstanceItem {
	list := make([]*v1.DashboardInstanceItem, 0, len(items))
	for _, item := range items {
		list = append(list, dashboardInstanceToDTO(item))
	}
	return list
}

func dashboardInstanceToDTO(item *mediasvc.DashboardInstanceItem) *v1.DashboardInstanceItem {
	if item == nil {
		return &v1.DashboardInstanceItem{}
	}
	return &v1.DashboardInstanceItem{
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
		StartTime:       item.StartTime,
		Version:         item.Version,
	}
}

func dashboardStreamListToDTO(items []*mediasvc.DashboardStreamItem) []*v1.DashboardStreamItem {
	list := make([]*v1.DashboardStreamItem, 0, len(items))
	for _, item := range items {
		list = append(list, dashboardStreamToDTO(item))
	}
	return list
}

func dashboardStreamToDTO(item *mediasvc.DashboardStreamItem) *v1.DashboardStreamItem {
	if item == nil {
		return &v1.DashboardStreamItem{ProtocolSummary: []*v1.DashboardProtocolItem{}}
	}
	return &v1.DashboardStreamItem{
		SourceUrl:             item.SourceUrl,
		StreamId:              item.StreamId,
		StreamName:            item.StreamName,
		DeviceId:              item.DeviceId,
		Resolution:            item.Resolution,
		Fps:                   item.Fps,
		Bitrate:               item.Bitrate,
		PacketLoss:            item.PacketLoss,
		Status:                item.Status,
		StartTime:             item.StartTime,
		Duration:              item.Duration,
		AvgDelay:              item.AvgDelay,
		ProtocolCount:         item.ProtocolCount,
		TotalSessionsLifetime: item.TotalSessionsLifetime,
		CurrentActiveSessions: item.CurrentActiveSessions,
		WatermarkEnabled:      item.WatermarkEnabled,
		ProtocolSummary:       dashboardProtocolSummaryToDTO(item.ProtocolSummary),
	}
}

func dashboardProtocolSummaryToDTO(items []*mediasvc.DashboardProtocolItem) []*v1.DashboardProtocolItem {
	list := make([]*v1.DashboardProtocolItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			list = append(list, &v1.DashboardProtocolItem{})
			continue
		}
		list = append(list, &v1.DashboardProtocolItem{
			ProtocolType:    item.ProtocolType,
			TotalSessions:   item.TotalSessions,
			CurrentSessions: item.CurrentSessions,
		})
	}
	return list
}

func dashboardSessionStreamInfoToDTO(item *mediasvc.DashboardSessionStreamInfo) *v1.DashboardSessionStreamInfo {
	if item == nil {
		return nil
	}
	return &v1.DashboardSessionStreamInfo{
		StreamId:   item.StreamId,
		StreamName: item.StreamName,
		DeviceId:   item.DeviceId,
	}
}

func dashboardSessionProtocolListToDTO(items []*mediasvc.DashboardSessionProtocol) []*v1.DashboardSessionProtocol {
	list := make([]*v1.DashboardSessionProtocol, 0, len(items))
	for _, item := range items {
		list = append(list, dashboardSessionProtocolToDTO(item))
	}
	return list
}

func dashboardSessionProtocolToDTO(item *mediasvc.DashboardSessionProtocol) *v1.DashboardSessionProtocol {
	if item == nil {
		return &v1.DashboardSessionProtocol{ActiveSessionList: []*v1.DashboardSessionItem{}}
	}
	return &v1.DashboardSessionProtocol{
		ProtocolType:      item.ProtocolType,
		Status:            item.Status,
		SessionCount:      item.SessionCount,
		ActiveSessionList: dashboardSessionListToDTO(item.Sessions),
	}
}

func dashboardSessionListToDTO(items []*mediasvc.DashboardSessionItem) []*v1.DashboardSessionItem {
	list := make([]*v1.DashboardSessionItem, 0, len(items))
	for _, item := range items {
		list = append(list, dashboardSessionToDTO(item))
	}
	return list
}

func dashboardSessionToDTO(item *mediasvc.DashboardSessionItem) *v1.DashboardSessionItem {
	if item == nil {
		return &v1.DashboardSessionItem{LinkHops: []*v1.DashboardLinkHopItem{}}
	}
	return &v1.DashboardSessionItem{
		SessionId:         item.SessionId,
		DeviceId:          item.DeviceId,
		ClientId:          item.ClientId,
		ClientIp:          item.ClientIp,
		ClientType:        item.ClientType,
		TenantId:          item.TenantId,
		UserName:          item.UserName,
		ProtocolType:      item.ProtocolType,
		StartTime:         item.StartTime,
		PlayDuration:      item.PlayDuration,
		CurrentFps:        item.CurrentFps,
		CurrentBitrate:    item.CurrentBitrate,
		CurrentResolution: item.CurrentResolution,
		NodeId:            item.NodeId,
		InstanceId:        item.InstanceId,
		LinkHops:          dashboardLinkHopsToDTO(item.LinkHops),
		TotalLinkLatency:  item.TotalLinkLatency,
	}
}

func dashboardLinkHopsToDTO(items []*mediasvc.DashboardLinkHopItem) []*v1.DashboardLinkHopItem {
	list := make([]*v1.DashboardLinkHopItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			list = append(list, &v1.DashboardLinkHopItem{})
			continue
		}
		list = append(list, &v1.DashboardLinkHopItem{
			HopIndex:  item.HopIndex,
			NodeId:    item.NodeId,
			LatencyMs: item.LatencyMs,
		})
	}
	return list
}

func cloneDashboardIntMap(input map[string]int) map[string]int {
	output := make(map[string]int, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
