// player_v1_poster.go implements the player activation poster data handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// Poster returns the activation poster composition data for a cattle the
// authenticated player has activated.
func (c *ControllerV1) Poster(ctx context.Context, req *v1.PosterReq) (res *v1.PosterRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.activationSvc.Poster(ctx, playerID, req.NiuId)
	if err != nil {
		return nil, err
	}
	return &v1.PosterRes{
		Nickname:     out.Nickname,
		IdentityType: out.IdentityType,
		NiuCode:      out.NiuCode,
		OrderNo:      out.OrderNo,
		Quote:        out.Quote,
		CampusBadge:  out.CampusBadge,
		ImageBase64:  out.ImageBase64,
	}, nil
}
