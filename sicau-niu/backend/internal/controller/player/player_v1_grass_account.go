// player_v1_grass_account.go implements the player grass account handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// GrassAccount returns the authenticated player's grass balance and recent ledger.
func (c *ControllerV1) GrassAccount(ctx context.Context, req *v1.GrassAccountReq) (res *v1.GrassAccountRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.grassSvc.Account(ctx, playerID)
	if err != nil {
		return nil, err
	}
	recent := make([]*v1.GrassTxnItem, 0, len(out.Recent))
	for _, txn := range out.Recent {
		recent = append(recent, &v1.GrassTxnItem{
			TxnType:   txn.TxnType,
			Delta:     txn.Delta,
			CreatedAt: txn.CreatedAt,
		})
	}
	return &v1.GrassAccountRes{Balance: out.Balance, Recent: recent}, nil
}
