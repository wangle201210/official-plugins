// admin_v1_list_transport_teams.go handles operator cloud-moving team queries.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func (c *ControllerV1) ListTransportTeams(ctx context.Context, req *v1.ListTransportTeamsReq) (res *v1.ListTransportTeamsRes, err error) {
	out, err := c.ironTransportSvc.ListAdminTeams(ctx, &irontransportsvc.AdminTeamListInput{PageNum: req.PageNum, PageSize: req.PageSize, Name: req.Name, Status: req.Status})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.AdminTransportTeam, 0, len(out.List))
	for _, team := range out.List {
		list = append(list, toAdminTransportTeam(team))
	}
	return &v1.ListTransportTeamsRes{List: list, Total: out.Total}, nil
}
