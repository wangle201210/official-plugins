// player_new.go defines the player-facing controller and its constructor. The
// controller holds the identity, college, activation, grass, feeding,
// grass-social, ranking and honor services as fields injected at route assembly
// time; it never constructs services on the request path. Player endpoints read
// the authenticated player ID from the request context populated by the plugin
// player-auth middleware, ensuring each player only reads and writes their own
// data.

package player

import (
	"lina-plugin-sicau-niu/backend/api/player"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
	grasssocialsvc "lina-plugin-sicau-niu/backend/internal/service/grasssocial"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
	rankingsvc "lina-plugin-sicau-niu/backend/internal/service/ranking"
)

// ControllerV1 is the sicau-niu player-facing controller.
type ControllerV1 struct {
	identitySvc    identitysvc.Service    // identitySvc handles login, phone binding and profile.
	collegeSvc     collegesvc.Service     // collegeSvc provides the player college dropdown.
	activationSvc  activationsvc.Service  // activationSvc serves the visible map, activation, collection and poster.
	grassSvc       grasssvc.Service       // grassSvc serves daily check-in and the grass ledger account.
	feedingSvc     feedingsvc.Service     // feedingSvc serves feeding and the recent feeding trail.
	grassSocialSvc grasssocialsvc.Service // grassSocialSvc serves steal, gift and the player inbox.
	rankingSvc     rankingsvc.Service     // rankingSvc serves the three leaderboards.
	honorSvc       honorsvc.Service       // honorSvc serves the player honor unlock list.
}

// NewV1 creates the player controller with explicit service dependencies.
func NewV1(
	identitySvc identitysvc.Service,
	collegeSvc collegesvc.Service,
	activationSvc activationsvc.Service,
	grassSvc grasssvc.Service,
	feedingSvc feedingsvc.Service,
	grassSocialSvc grasssocialsvc.Service,
	rankingSvc rankingsvc.Service,
	honorSvc honorsvc.Service,
) player.IPlayerV1 {
	return &ControllerV1{
		identitySvc:    identitySvc,
		collegeSvc:     collegeSvc,
		activationSvc:  activationSvc,
		grassSvc:       grassSvc,
		feedingSvc:     feedingSvc,
		grassSocialSvc: grassSocialSvc,
		rankingSvc:     rankingSvc,
		honorSvc:       honorSvc,
	}
}
