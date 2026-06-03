// Package playerctx stores and reads the authenticated sicau-niu player ID on a
// request-scoped context. The player auth middleware writes the ID after token
// verification; player-facing controllers read it to constrain all reads and
// writes to the current player's own data. The context key is unexported so no
// other package can forge a player identity.
package playerctx

import "context"

// contextKeyType is a private key type that prevents key collisions and forging
// from other packages.
type contextKeyType struct{}

// playerIDKey is the single private context key carrying the player ID.
var playerIDKey = contextKeyType{}

// With returns a child context carrying playerID. It is called by the player
// auth middleware after the player token is verified.
func With(ctx context.Context, playerID int64) context.Context {
	return context.WithValue(ctx, playerIDKey, playerID)
}

// From returns the authenticated player ID stored on ctx. The ok result is
// false when no player identity is present, in which case playerID is 0.
func From(ctx context.Context) (playerID int64, ok bool) {
	value := ctx.Value(playerIDKey)
	if value == nil {
		return 0, false
	}
	id, valid := value.(int64)
	if !valid || id <= 0 {
		return 0, false
	}
	return id, true
}
