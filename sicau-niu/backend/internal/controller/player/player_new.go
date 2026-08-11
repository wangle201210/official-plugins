// player_new.go defines the player-facing controller and its constructor. The
// controller holds the identity, college, activation, grass, feeding,
// grass-social, ranking and honor services as fields injected at route assembly
// time; it never constructs services on the request path. Player endpoints read
// the authenticated player ID from the request context populated by the plugin
// player-auth middleware, ensuring each player only reads and writes their own
// data.

package player

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-plugin-sicau-niu/backend/api/player"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
	activationphotosvc "lina-plugin-sicau-niu/backend/internal/service/activationphoto"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
	grasssocialsvc "lina-plugin-sicau-niu/backend/internal/service/grasssocial"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
	miniappconfigsvc "lina-plugin-sicau-niu/backend/internal/service/miniappconfig"
	rankingsvc "lina-plugin-sicau-niu/backend/internal/service/ranking"
)

// ControllerV1 is the sicau-niu player-facing controller.
type ControllerV1 struct {
	identitySvc      identitysvc.Service        // identitySvc handles login, phone binding and profile.
	collegeSvc       collegesvc.Service         // collegeSvc provides the player college dropdown.
	activationSvc    activationsvc.Service      // activationSvc serves the visible map, activation, collection and poster.
	grassSvc         grasssvc.Service           // grassSvc serves daily check-in and the grass ledger account.
	feedingSvc       feedingsvc.Service         // feedingSvc serves feeding and the recent feeding trail.
	grassSocialSvc   grasssocialsvc.Service     // grassSocialSvc serves steal, gift and the player inbox.
	rankingSvc       rankingsvc.Service         // rankingSvc serves the three leaderboards.
	honorSvc         honorsvc.Service           // honorSvc serves the player honor unlock list.
	miniappConfigSvc miniappconfigsvc.Service   // miniappConfigSvc serves public runtime map and activity configuration.
	photoSvc         activationphotosvc.Service // photoSvc serves private activation photo upload and reads.
	ironTransportSvc irontransportsvc.Service   // ironTransportSvc serves persistent team transport gameplay.
}

// NewV1 creates the player controller with explicit service dependencies and
// rejects missing dependencies during route initialization.
func NewV1(
	identitySvc identitysvc.Service,
	collegeSvc collegesvc.Service,
	activationSvc activationsvc.Service,
	grassSvc grasssvc.Service,
	feedingSvc feedingsvc.Service,
	grassSocialSvc grasssocialsvc.Service,
	rankingSvc rankingsvc.Service,
	honorSvc honorsvc.Service,
	miniappConfigSvc miniappconfigsvc.Service,
	photoSvc activationphotosvc.Service,
	ironTransportSvc irontransportsvc.Service,
) (player.IPlayerV1, error) {
	if identitySvc == nil || collegeSvc == nil || activationSvc == nil || grassSvc == nil ||
		feedingSvc == nil || grassSocialSvc == nil || rankingSvc == nil || honorSvc == nil ||
		miniappConfigSvc == nil || photoSvc == nil || ironTransportSvc == nil {
		return nil, gerror.New("sicau-niu player controller requires all service dependencies")
	}
	return &ControllerV1{
		identitySvc:      identitySvc,
		collegeSvc:       collegeSvc,
		activationSvc:    activationSvc,
		grassSvc:         grassSvc,
		feedingSvc:       feedingSvc,
		grassSocialSvc:   grassSocialSvc,
		rankingSvc:       rankingSvc,
		honorSvc:         honorSvc,
		miniappConfigSvc: miniappConfigSvc,
		photoSvc:         photoSvc,
		ironTransportSvc: ironTransportSvc,
	}, nil
}
