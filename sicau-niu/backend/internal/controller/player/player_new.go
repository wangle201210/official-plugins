// player_new.go defines the player-facing controller and its constructor. The
// controller holds the identity and college services as fields injected at route
// assembly time; it never constructs services on the request path. Player
// endpoints read the authenticated player ID from the request context populated
// by the plugin player-auth middleware, ensuring each player only reads and
// writes their own data.

package player

import (
	"lina-plugin-sicau-niu/backend/api/player"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// ControllerV1 is the sicau-niu player-facing controller.
type ControllerV1 struct {
	identitySvc identitysvc.Service // identitySvc handles login, phone binding and profile.
	collegeSvc  collegesvc.Service  // collegeSvc provides the player college dropdown.
}

// NewV1 creates the player controller with explicit service dependencies.
func NewV1(identitySvc identitysvc.Service, collegeSvc collegesvc.Service) player.IPlayerV1 {
	return &ControllerV1{
		identitySvc: identitySvc,
		collegeSvc:  collegeSvc,
	}
}
