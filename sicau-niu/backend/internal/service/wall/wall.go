// Package wall implements the sicau-niu C6 public memorial-wall capability: the
// first-activator wall, the campus-history highlights and the public activity
// stats. All three are public read-only activity views (no authentication, no
// player token) derived from the existing C1 player, C2 cattle/card/quote and C3
// activation tables; the capability owns no table of its own. Privacy is enforced
// at the data-assembly boundary: the service projects only nicknames, identity
// labels, cattle names/codes and aggregate counts, and never reads or returns
// phone numbers, openids or device fingerprints. Every list is bounded (the
// first-activator wall is capped at 120, the highlights samples are capped) and
// every aggregate count runs on the database side, so the assembly never loads an
// unbounded set into memory and never issues a per-row lookup. All store access
// uses the generated DAO/DO objects so GoFrame manages soft-delete and timestamp
// columns.
package wall

import "context"

// Bound constants for the public memorial wall. They cap every list the wall
// returns so a public endpoint can never trigger an unbounded scan.
const (
	// firstActivatorLimit caps the first-activator wall at the campaign's 120
	// commemorated first activators.
	firstActivatorLimit = 120
	// highlightCardLimit caps the campus-history card sample.
	highlightCardLimit = 12
	// highlightQuoteLimit caps the campus-history quote sample.
	highlightQuoteLimit = 12
)

// Service defines the C6 public memorial-wall contract: the first-activator wall,
// the campus-history highlights and the public activity stats.
type Service interface {
	// FirstActivators returns the first-activator wall: the players who were the
	// first to activate a cattle (is_first), ordered by activation time ascending
	// and capped at 120. Nicknames, identity labels and cattle names/codes are
	// batch-assembled to avoid N+1; no privacy field is read or returned. It
	// returns a query bizerr on store failure.
	FirstActivators(ctx context.Context) (out *FirstActivatorBoard, err error)
	// Highlights returns a bounded sample of campus-history cards and enabled
	// campus-history quotes. Cattle names for the sampled cards are batch-assembled
	// in one query to avoid N+1. It returns a query bizerr on store failure.
	Highlights(ctx context.Context) (out *Highlights, err error)
	// Stats returns the public activity statistics aggregated on the database
	// side: the activated cattle count, the total cattle count, the first-activator
	// count and the participating-player count. It returns a query bizerr on store
	// failure.
	Stats(ctx context.Context) (out *Stats, err error)
	// WallConfig returns the public memorial-wall configuration: the mini-program
	// back-link URL the H5 wall uses for its return-to-mini-program entry. The URL
	// is supplied by plugin configuration and is empty when unconfigured.
	WallConfig(ctx context.Context) (out *WallConfig, err error)
}

// Interface compliance assertion for the default wall service implementation.
var _ Service = (*serviceImpl)(nil)

// Config carries the plain-value public memorial-wall configuration: only the
// mini-program back-link URL exposed to the H5 wall.
type Config struct {
	// MiniappURL is the return-to-mini-program URL the H5 wall links to; empty when
	// unconfigured.
	MiniappURL string
}

// WallConfig is the public memorial-wall configuration result.
type WallConfig struct {
	// MiniappURL is the return-to-mini-program URL; empty when unconfigured.
	MiniappURL string
}

// serviceImpl implements Service against the plugin-owned activation, user,
// cattle, card and quote tables. Its only runtime input is the plain-value
// mini-program back-link URL injected at assembly time; every view reads the
// plugin's own tables through the generated DAO and the list bounds are fixed
// package constants.
type serviceImpl struct {
	miniappURL string // miniappURL is the H5 wall's return-to-mini-program link.
}

// New creates a public memorial-wall service with the plain-value configuration.
// The component reads the plugin's own tables through the generated DAO; the list
// bounds are fixed package constants so a public endpoint is always bounded.
func New(config Config) Service {
	return &serviceImpl{miniappURL: config.MiniappURL}
}

// WallConfig returns the public memorial-wall configuration.
func (s *serviceImpl) WallConfig(_ context.Context) (*WallConfig, error) {
	return &WallConfig{MiniappURL: s.miniappURL}, nil
}
