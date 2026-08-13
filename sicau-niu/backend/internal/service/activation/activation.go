// Package activation implements the sicau-niu C3 player gameplay capability:
// the visible-cattle map list (filtered by online-time visibility), GPS check-in
// activation that matches a nearby unactivated cattle, the per-day activation
// limit, on-activation card issuance, the player's personal
// card collection (图鉴) and the activation poster composition data. Every
// player-facing operation is constrained to the authenticated player's own data
// by the caller passing the player ID resolved by the player auth middleware;
// activation, the collection and the poster are isolated to the current player,
// while the visible-cattle list and shared-pool status are a public activity view.
// All store access uses the generated DAO/DO objects so GoFrame manages
// soft-delete and timestamp columns automatically. List and assembly paths run
// DB-side filtering plus bounded batch queries to avoid N+1; the check-in match
// path is bounded and the first-activator race is serialized with a per-cattle row
// lock inside a transaction.
package activation

import (
	"context"
	"time"

	"lina-plugin-sicau-niu/backend/internal/service/activation/internal/posterrender"
	activationphotosvc "lina-plugin-sicau-niu/backend/internal/service/activationphoto"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// Config carries the plain-value runtime configuration for the activation
// capability. It holds only scalar tuning values and never runtime service
// dependencies, which are injected as separate constructor parameters.
type Config struct {
	// LBSThresholdMeters is the maximum allowed distance in meters between the
	// player's reported location and the cattle GPS anchor for an activation to be
	// accepted.
	LBSThresholdMeters float64
	// CampusBadge is the campus anniversary badge text rendered onto the activation
	// poster; empty leaves the badge area blank.
	CampusBadge string
}

// Service defines the C3 player gameplay contract: visible-cattle map list, GPS
// check-in activation, personal card collection and activation poster data.
type Service interface {
	// VisibleNiu returns the cattle currently visible to playerID, filtered by
	// online time plus optional weekday/time window. Each item carries its GPS
	// anchor, shared-pool status and whether the player has already activated it.
	// The set is bounded and the per-player activation flags are batch-assembled in
	// one query to avoid N+1. It returns a query bizerr on store failure.
	VisibleNiu(ctx context.Context, playerID int64) (out []*VisibleNiuItem, err error)
	// Activate matches and activates the nearest currently visible inactive cattle
	// within the LBS threshold for playerID's reported GPS check-in location
	// (GCJ-02). The request does not require a cattle ID, but requires a
	// player-scoped request ID whose first successful response is replayed on
	// retries. The per-day success limit, daily attempt quota (failures count) and
	// movement-speed anti-cheat guard are rechecked under the player row lock; the
	// match path then locks and rechecks the matched cattle before flipping it to
	// active and issuing the main card. It returns the relevant validation bizerr
	// on rejection or a store bizerr on failure; rejection errors never carry
	// distance or bearing hints.
	Activate(ctx context.Context, playerID int64, in *ActivateInput) (out *ActivateOutput, err error)
	// Collection returns the bounded card catalog with playerID's ownership state,
	// optionally filtered by category. Locked entries retain only the cattle
	// reference and category needed for progress display; card-face content is
	// withheld. It uses bounded batch queries to avoid N+1 and returns a query
	// bizerr on store failure or a parameter bizerr on an invalid category filter.
	Collection(ctx context.Context, playerID int64, category string) (out []*CollectionItem, err error)
	// Poster returns the activation poster composition data for a cattle playerID
	// has activated: nickname, identity type, cattle code, arrival order, a random
	// enabled quote and the campus badge. It returns CodeActivationNotFound when the
	// player has not activated the cattle, or a query bizerr on store failure.
	Poster(ctx context.Context, playerID int64, niuID int64) (out *PosterOutput, err error)
	// NiuDetail returns one visible cattle with safe location projection, current
	// aggregate state, a quote and the player's owned main card.
	NiuDetail(ctx context.Context, playerID, niuID int64) (out *NiuDetailOutput, err error)
	// Revoke removes one erroneous activation from active gameplay while retaining
	// its photo evidence and repairing the cattle first-activator state.
	Revoke(ctx context.Context, activationID int64) error
}

// Interface compliance assertion for the default activation service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned activation, niu, card
// and quote tables. Its runtime dependencies are injected explicitly as separate
// parameters so dependency changes surface at compile time.
type serviceImpl struct {
	identitySvc    identitysvc.Service         // identitySvc supplies the poster nickname and identity type.
	posterRenderer posterrender.PosterRenderer // posterRenderer is the replaceable PNG output seam.
	rulesSvc       rulessvc.Service            // rulesSvc supplies operator-maintained runtime thresholds and badge text.
	photoSvc       activationphotosvc.Service  // photoSvc validates and consumes player-owned activation evidence.
	lbsThreshold   float64                     // lbsThreshold is the LBS activation distance threshold in meters.
	campusBadge    string                      // campusBadge is the poster campus anniversary badge text.
	now            func() time.Time            // now supplies the authoritative time after player locking.
}

// PosterRenderer is the activation-poster PNG output seam re-exported from this
// package so the route-assembly layer can construct the default renderer and
// inject it without reaching across the package-internal boundary.
type PosterRenderer = posterrender.PosterRenderer

// NewBasicPosterRenderer creates the default basic activation-poster renderer.
// The route-assembly layer constructs it once and injects it into New so the PNG
// output implementation can be replaced behind the stable seam.
func NewBasicPosterRenderer() PosterRenderer {
	return posterrender.New()
}

// New creates an activation service with explicit dependencies: the identity
// service used for poster player data, the poster renderer seam used for PNG
// output, the optional runtime-rule service used for operator-maintained
// thresholds and badge text, and the fallback plain-value activation
// configuration.
func New(identitySvc identitysvc.Service, posterRenderer PosterRenderer, rulesSvc rulessvc.Service, photoSvc activationphotosvc.Service, config Config) Service {
	return &serviceImpl{
		identitySvc:    identitySvc,
		posterRenderer: posterRenderer,
		rulesSvc:       rulesSvc,
		photoSvc:       photoSvc,
		lbsThreshold:   config.LBSThresholdMeters,
		campusBadge:    config.CampusBadge,
		now:            time.Now,
	}
}

func (s *serviceImpl) nowTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}
