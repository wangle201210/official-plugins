// player_v1_messages.go implements the player inbox list handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	grasssocialsvc "lina-plugin-sicau-niu/backend/internal/service/grasssocial"
)

// Messages returns the authenticated player's paged in-app inbox.
func (c *ControllerV1) Messages(ctx context.Context, req *v1.MessagesReq) (res *v1.MessagesRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.grassSocialSvc.Messages(ctx, playerID, &grasssocialsvc.MessagesInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.MessageItem, 0, len(out.List))
	for _, msg := range out.List {
		list = append(list, &v1.MessageItem{
			Id:        msg.Id,
			MsgType:   msg.MsgType,
			Content:   msg.Content,
			IsRead:    msg.IsRead,
			CreatedAt: msg.CreatedAt,
		})
	}
	return &v1.MessagesRes{List: list, Total: out.Total}, nil
}
