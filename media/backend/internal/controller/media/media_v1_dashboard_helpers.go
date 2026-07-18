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

func dashboardTopologyToDTO(item *mediasvc.DashboardTopologyOutput) *v1.GetDashboardTopologyRes {
	if item == nil {
		return &v1.GetDashboardTopologyRes{Devices: []*v1.DashboardTopologyDevice{}}
	}
	return &v1.GetDashboardTopologyRes{
		NodeId:               item.NodeId,
		NodeName:             item.NodeName,
		DeviceId:             item.DeviceId,
		StreamCount:          item.StreamCount,
		SessionCount:         item.SessionCount,
		SessionDetailLimited: item.SessionDetailLimited,
		GeneratedAt:          item.GeneratedAt,
		Devices:              dashboardTopologyDevicesToDTO(item.Devices),
	}
}

func dashboardTopologyDevicesToDTO(items []*mediasvc.DashboardTopologyDevice) []*v1.DashboardTopologyDevice {
	list := make([]*v1.DashboardTopologyDevice, 0, len(items))
	for _, item := range items {
		if item == nil {
			list = append(list, &v1.DashboardTopologyDevice{Streams: []*v1.DashboardTopologyStream{}})
			continue
		}
		list = append(list, &v1.DashboardTopologyDevice{
			DeviceId:     item.DeviceId,
			StreamCount:  item.StreamCount,
			SessionCount: item.SessionCount,
			Streams:      dashboardTopologyStreamsToDTO(item.Streams),
		})
	}
	return list
}

func dashboardTopologyStreamsToDTO(items []*mediasvc.DashboardTopologyStream) []*v1.DashboardTopologyStream {
	list := make([]*v1.DashboardTopologyStream, 0, len(items))
	for _, item := range items {
		if item == nil {
			list = append(list, &v1.DashboardTopologyStream{Protocols: []*v1.DashboardTopologyProtocol{}})
			continue
		}
		list = append(list, &v1.DashboardTopologyStream{
			StreamId:     item.StreamId,
			StreamName:   item.StreamName,
			Status:       item.Status,
			BasePlatform: dashboardTopologyBasePlatformToDTO(item.BasePlatform),
			Gateway:      dashboardTopologyGatewayToDTO(item.Gateway),
			Protocols:    dashboardTopologyProtocolsToDTO(item.Protocols),
		})
	}
	return list
}

func dashboardTopologyBasePlatformToDTO(item *mediasvc.DashboardTopologyBasePlatform) *v1.DashboardTopologyBasePlatform {
	if item == nil {
		return &v1.DashboardTopologyBasePlatform{}
	}
	return &v1.DashboardTopologyBasePlatform{
		Bitrate:        item.Bitrate,
		Resolution:     item.Resolution,
		StreamProtocol: item.StreamProtocol,
		DeviceCode:     item.DeviceCode,
		SourceUrl:      item.SourceUrl,
	}
}

func dashboardTopologyGatewayToDTO(item *mediasvc.DashboardTopologyGateway) *v1.DashboardTopologyGateway {
	if item == nil {
		return &v1.DashboardTopologyGateway{}
	}
	return &v1.DashboardTopologyGateway{
		StreamProtocol: item.StreamProtocol,
		Resolution:     item.Resolution,
		Bitrate:        item.Bitrate,
		Fps:            item.Fps,
		AvgDelay:       item.AvgDelay,
		NodeId:         item.NodeId,
		NodeName:       item.NodeName,
		InstanceId:     item.InstanceId,
		InstanceName:   item.InstanceName,
	}
}

func dashboardTopologyProtocolsToDTO(items []*mediasvc.DashboardTopologyProtocol) []*v1.DashboardTopologyProtocol {
	list := make([]*v1.DashboardTopologyProtocol, 0, len(items))
	for _, item := range items {
		if item == nil {
			list = append(list, &v1.DashboardTopologyProtocol{Tenants: []*v1.DashboardTopologyTenant{}})
			continue
		}
		list = append(list, &v1.DashboardTopologyProtocol{
			ProtocolType: item.ProtocolType,
			ReuseCount:   item.ReuseCount,
			Tenants:      dashboardTopologyTenantsToDTO(item.Tenants),
		})
	}
	return list
}

func dashboardTopologyTenantsToDTO(items []*mediasvc.DashboardTopologyTenant) []*v1.DashboardTopologyTenant {
	list := make([]*v1.DashboardTopologyTenant, 0, len(items))
	for _, item := range items {
		if item == nil {
			list = append(list, &v1.DashboardTopologyTenant{Users: []*v1.DashboardTopologyUser{}})
			continue
		}
		list = append(list, &v1.DashboardTopologyTenant{
			TenantId:           item.TenantId,
			ConcurrentSessions: item.ConcurrentSessions,
			Users:              dashboardTopologyUsersToDTO(item.Users),
		})
	}
	return list
}

func dashboardTopologyUsersToDTO(items []*mediasvc.DashboardTopologyUser) []*v1.DashboardTopologyUser {
	list := make([]*v1.DashboardTopologyUser, 0, len(items))
	for _, item := range items {
		if item == nil {
			list = append(list, &v1.DashboardTopologyUser{})
			continue
		}
		list = append(list, &v1.DashboardTopologyUser{
			SessionId:    item.SessionId,
			UserName:     item.UserName,
			ClientId:     item.ClientId,
			ClientIp:     item.ClientIp,
			ClientType:   item.ClientType,
			ProtocolType: item.ProtocolType,
			TenantId:     item.TenantId,
			NodeId:       item.NodeId,
			InstanceId:   item.InstanceId,
			StartTime:    item.StartTime,
			PlayDuration: item.PlayDuration,
		})
	}
	return list
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
