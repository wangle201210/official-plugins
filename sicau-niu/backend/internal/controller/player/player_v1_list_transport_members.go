// player_v1_list_transport_members.go handles member-only team roster queries.
package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func (c *ControllerV1) ListTransportMembers(ctx context.Context, req *v1.ListTransportMembersReq) (res *v1.ListTransportMembersRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.ListMembers(ctx, playerID, req.Id, &irontransportsvc.PageInput{PageNum: req.PageNum, PageSize: req.PageSize})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.IronTransportMember, 0, len(out.List))
	for _, member := range out.List {
		list = append(list, toIronTransportMember(member))
	}
	return &v1.ListTransportMembersRes{List: list, Total: out.Total}, nil
}
