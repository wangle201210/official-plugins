// player_v1_mark_message_read.go implements the player mark-message-read handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// MarkMessageRead marks one of the authenticated player's own messages read.
func (c *ControllerV1) MarkMessageRead(ctx context.Context, req *v1.MarkMessageReadReq) (res *v1.MarkMessageReadRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.grassSocialSvc.MarkRead(ctx, playerID, req.Id); err != nil {
		return nil, err
	}
	return &v1.MarkMessageReadRes{}, nil
}
