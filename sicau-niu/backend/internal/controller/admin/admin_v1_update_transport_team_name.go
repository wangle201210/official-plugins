// admin_v1_update_transport_team_name.go handles operator team-name updates.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

func (c *ControllerV1) UpdateTransportTeamName(ctx context.Context, req *v1.UpdateTransportTeamNameReq) (res *v1.UpdateTransportTeamNameRes, err error) {
	if err := c.ironTransportSvc.RenameTeam(ctx, req.Id, req.Name); err != nil {
		return nil, err
	}
	return &v1.UpdateTransportTeamNameRes{}, nil
}
