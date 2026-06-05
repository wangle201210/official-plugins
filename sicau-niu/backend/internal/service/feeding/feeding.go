// Package feeding implements the sicau-niu C4 feeding capability: the player
// feeds grass to an already-activated cattle, the iron-cow proximity bonus, the
// random school-history quote returned with each feeding, and the player's recent
// feeding trail. Feeding deducts the base amount from the player's ledger account
// (through the grass capability) and records the feeding with the original
// amount, the bonus coefficient and the resulting effect, all inside one
// transaction so the debit and the feeding record stay consistent. The bonus is
// decided by the great-circle distance between the cattle anchor and any iron
// cow's current position (obtained through the replaceable iron-location seam);
// within the configured threshold the coefficient is 1.5, otherwise 1.0. Every
// player-facing operation is isolated to the authenticated player. All store
// access uses the generated DAO/DO objects so GoFrame manages soft-delete and
// timestamps; the trail and quote reads are bounded to avoid N+1.
package feeding

import (
	"context"

	"lina-plugin-sicau-niu/backend/internal/service/feeding/internal/ironlocation"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// ironBonusCoefficientBasis is the persisted coefficient basis (in hundredths)
// applied when an iron-cow proximity bonus is in range: 150 means a 1.5x effect.
const ironBonusCoefficientBasis = 150

// baseCoefficientBasis is the persisted coefficient basis (in hundredths) applied
// when no iron-cow proximity bonus is in range: 100 means a 1.0x effect.
const baseCoefficientBasis = 100

// Config carries the plain-value runtime configuration for the feeding
// capability. It holds only scalar tuning values and never runtime dependencies.
type Config struct {
	// IronBonusThresholdMeters is the maximum great-circle distance in meters
	// between a cattle anchor and any iron cow's current position for the feeding
	// bonus to apply.
	IronBonusThresholdMeters float64
}

// Service defines the C4 feeding contract: feed an activated cattle and read the
// player's recent feeding trail.
type Service interface {
	// Feed feeds in.BaseAmount grass from playerID to the activated cattle
	// in.NiuId. It rejects a non-activated cattle, a non-positive amount and an
	// insufficient balance. Inside one transaction it debits the player's ledger,
	// computes the iron-bonus coefficient from the cattle anchor and current iron
	// positions, and records the feeding with the original amount, coefficient and
	// effect. It returns the cattle info, a random enabled quote and the bonus
	// breakdown with the resulting balance, or the relevant bizerr on failure.
	Feed(ctx context.Context, playerID int64, in *FeedInput) (out *FeedOutput, err error)
	// Trail returns playerID's most recent feedings (latest 10, newest first) with
	// the fed cattle code/name, effect and feed time. It is isolated to the current
	// player, batch-assembles cattle info to avoid N+1, and returns an empty trail
	// when the player has not fed any cattle. It returns a query bizerr on failure.
	Trail(ctx context.Context, playerID int64) (out []*TrailItem, err error)
}

// Interface compliance assertion for the default feeding service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned feeding, niu and quote
// tables. Its runtime dependencies are injected explicitly as separate parameters
// so dependency changes surface at compile time.
type serviceImpl struct {
	grassSvc           grasssvc.Service     // grassSvc debits the player ledger inside the feeding transaction.
	ironLocation       ironlocation.Gateway // ironLocation supplies current iron-cow positions for the bonus check.
	rulesSvc           rulessvc.Service     // rulesSvc supplies operator-maintained iron-bonus threshold when injected.
	ironBonusThreshold float64              // ironBonusThreshold is the proximity-bonus distance threshold in meters.
}

// IronLocationGateway is the iron-cow real-time location seam re-exported from
// this package so the route-assembly layer can construct the mock gateway and
// inject it without reaching across the package-internal boundary.
type IronLocationGateway = ironlocation.Gateway

// NewMockIronLocation creates the development iron-cow location gateway backed by
// the stored iron last_lat/last_lng. The route-assembly layer constructs it once
// and injects it so the real external-API implementation can replace it behind
// the stable seam.
func NewMockIronLocation() IronLocationGateway {
	return ironlocation.NewMock()
}

// New creates a feeding service with explicit dependencies: the grass ledger
// service used to debit the player inside the feeding transaction, the iron-cow
// location gateway used for the proximity bonus, the optional runtime-rule
// service used for operator-maintained threshold tuning, and the fallback
// plain-value feeding configuration.
func New(grassSvc grasssvc.Service, ironLocation IronLocationGateway, rulesSvc rulessvc.Service, config Config) Service {
	return &serviceImpl{
		grassSvc:           grassSvc,
		ironLocation:       ironLocation,
		rulesSvc:           rulesSvc,
		ironBonusThreshold: config.IronBonusThresholdMeters,
	}
}
