// admin_v1_list_players.go implements the operator read-only player list handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// ListPlayers returns one DB-side paged, read-only player list page.
func (c *ControllerV1) ListPlayers(ctx context.Context, req *v1.ListPlayersReq) (res *v1.ListPlayersRes, err error) {
	out, err := c.identitySvc.ListPlayers(ctx, &identitysvc.ListPlayersInput{
		Keyword:      req.Keyword,
		IdentityType: req.IdentityType,
		PageNum:      req.PageNum,
		PageSize:     req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.PlayerItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.PlayerItem{
			Id:             item.Id,
			Nickname:       item.Nickname,
			Phone:          item.Phone,
			IdentityType:   item.IdentityType,
			CollegeId:      item.CollegeId,
			Grade:          item.Grade,
			GraduationYear: item.GraduationYear,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return &v1.ListPlayersRes{List: items, Total: out.Total}, nil
}
