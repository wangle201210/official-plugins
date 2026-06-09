// Package activation implements the sicau-niu C3 player gameplay capability:
// the visible-cattle map list (filtered by online-time visibility), LBS activation with
// shared-pool first-activator concurrency, the per-day activation limit and the
// no-duplicate-per-cattle rule, on-activation card issuance, the player's personal
// card collection (图鉴) and the activation poster composition data. Every
// player-facing operation is constrained to the authenticated player's own data
// by the caller passing the player ID resolved by the player auth middleware;
// activation, the collection and the poster are isolated to the current player,
// while the visible-cattle list and shared-pool status are a public activity view.
// All store access uses the generated DAO/DO objects so GoFrame manages
// soft-delete and timestamp columns automatically. List and assembly paths run
// DB-side filtering plus bounded batch queries to avoid N+1; the first-activator
// race is serialized with a per-cattle row lock inside a transaction.
package activation

import (
	"context"

	"lina-plugin-sicau-niu/backend/internal/service/activation/internal/posterrender"
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

// Service defines the C3 player gameplay contract: visible-cattle map list, LBS
// activation, personal card collection and activation poster data.
type Service interface {
	// VisibleNiu returns the cattle currently visible to playerID, filtered by
	// online time plus optional weekday/time window. Each item carries its GPS
	// anchor, shared-pool status and whether the player has already activated it.
	// The set is bounded and the per-player activation flags are batch-assembled in
	// one query to avoid N+1. It returns a query bizerr on store failure.
	VisibleNiu(ctx context.Context, playerID int64) (out []*VisibleNiuItem, err error)
	// Activate activates the target cattle for playerID after online visibility,
	// LBS distance, daily limit and no-duplicate validation. The first-activator
	// race is serialized by a per-cattle row lock inside a transaction; the first
	// activator flips the cattle to active and arrival order starts at 1. The
	// cattle main card is returned on success (nil when the cattle has no main
	// card). It returns the relevant validation bizerr on rejection or a store
	// bizerr on failure.
	Activate(ctx context.Context, playerID int64, in *ActivateInput) (out *ActivateOutput, err error)
	// Collection returns playerID's personal card collection: the main cards of the
	// cattle the player has activated, optionally filtered by category, ordered by
	// activation recency. It is isolated to the current player and batch-assembles
	// cards from the activated cattle to avoid N+1. It returns a query bizerr on
	// store failure or a parameter bizerr on an invalid category filter.
	Collection(ctx context.Context, playerID int64, category string) (out []*CollectionItem, err error)
	// Poster returns the activation poster composition data for a cattle playerID
	// has activated: nickname, identity type, cattle code, arrival order, a random
	// enabled quote and the campus badge. It returns CodeActivationNotFound when the
	// player has not activated the cattle, or a query bizerr on store failure.
	Poster(ctx context.Context, playerID int64, niuID int64) (out *PosterOutput, err error)
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
	lbsThreshold   float64                     // lbsThreshold is the LBS activation distance threshold in meters.
	campusBadge    string                      // campusBadge is the poster campus anniversary badge text.
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
func New(identitySvc identitysvc.Service, posterRenderer PosterRenderer, rulesSvc rulessvc.Service, config Config) Service {
	return &serviceImpl{
		identitySvc:    identitySvc,
		posterRenderer: posterRenderer,
		rulesSvc:       rulesSvc,
		lbsThreshold:   config.LBSThresholdMeters,
		campusBadge:    config.CampusBadge,
	}
}
