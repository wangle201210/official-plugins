// Package player implements the sicau-niu player-facing HTTP controllers. These
// endpoints serve the WeChat mini-program: public login plus token-protected
// phone binding, profile read/update and the college dropdown. Every protected
// handler resolves the authenticated player ID from the request context that the
// plugin player-auth middleware populates, so a player can only read and write
// their own data.
package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/internal/middleware/playerctx"
	tokensvc "lina-plugin-sicau-niu/backend/internal/service/token"

	"lina-core/pkg/bizerr"
)

// currentPlayerID returns the authenticated player ID from ctx or an
// authentication bizerr when the player identity is absent. The player-auth
// middleware guards these routes, so a missing identity indicates a routing or
// middleware misconfiguration rather than a normal request.
func currentPlayerID(ctx context.Context) (int64, error) {
	playerID, ok := playerctx.From(ctx)
	if !ok {
		return 0, bizerr.NewCode(tokensvc.CodeTokenInvalid)
	}
	return playerID, nil
}
