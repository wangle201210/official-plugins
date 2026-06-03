// settlement_v1_risk.go implements the shared-device risk view handler and its
// service-to-DTO projections.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
)

// RiskDeviceClusters returns the shared-device risk view.
func (c *ControllerV1) RiskDeviceClusters(ctx context.Context, req *v1.RiskDeviceClustersReq) (res *v1.RiskDeviceClustersRes, err error) {
	clusters, err := c.settlementSvc.RiskDeviceClusters(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.RiskDeviceClustersRes{List: toDeviceClusters(clusters.List)}, nil
}

// toDeviceClusters projects the risk clusters to their response DTOs.
func toDeviceClusters(rows []*settlementsvc.RiskCluster) []*v1.DeviceCluster {
	clusters := make([]*v1.DeviceCluster, 0, len(rows))
	for _, row := range rows {
		clusters = append(clusters, &v1.DeviceCluster{
			Fingerprint: row.Fingerprint,
			Count:       row.Count,
			Members:     toDeviceClusterMembers(row.Members),
		})
	}
	return clusters
}

// toDeviceClusterMembers projects the cluster members to their response DTOs.
func toDeviceClusterMembers(rows []*settlementsvc.RiskMember) []*v1.DeviceClusterMember {
	members := make([]*v1.DeviceClusterMember, 0, len(rows))
	for _, row := range rows {
		members = append(members, &v1.DeviceClusterMember{
			UserId:   row.UserId,
			Nickname: row.Nickname,
		})
	}
	return members
}
