package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) NiuDetail(ctx context.Context, req *v1.NiuDetailReq) (res *v1.NiuDetailRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.activationSvc.NiuDetail(ctx, playerID, req.Id)
	if err != nil {
		return nil, err
	}
	res = &v1.NiuDetailRes{Niu: toVisibleNiuItem(out.Niu), Quote: out.Quote}
	if out.Card != nil {
		res.Card = &v1.NiuDetailCard{Id: out.Card.ID, Category: out.Card.Category, Title: out.Card.Title, Content: out.Card.Content, ImagePath: out.Card.ImagePath, Owned: out.Card.Owned}
	}
	return res, nil
}
