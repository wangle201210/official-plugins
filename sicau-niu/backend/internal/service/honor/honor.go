// Package honor implements the sicau-niu C5 honor capability: operator-facing
// honor-definition CRUD and the read-only player honor unlock computation. Honor
// definitions carry a stable honor-type enum (badge/avatar_frame/certificate), a
// unique code, a stable unlock-rule enum (participation/feed_count/
// activation_count/category_complete/full_complete), an optional threshold for
// count rules, an optional card category for the category-complete rule, an
// image/template path and a sort order. Operator CRUD enforces honor-type and
// unlock-type enum validation, threshold/category shape rules and code uniqueness,
// with DB-side paged listing. The player honor view computes, for every honor
// definition, whether the requesting player has met its unlock rule. The
// computation is read-only (it never writes the user_honor grant table, which the
// C7 settlement owns) and isolated to the requesting player. To avoid a per-honor
// query it pre-aggregates the player's feeding count, activation count and
// per-category collected-card counts together with the per-category active-card
// totals in a fixed number of batched queries, then evaluates every honor rule in
// memory against those counts. All store access uses the generated DAO/DO objects
// so GoFrame manages soft-delete and timestamp columns automatically.
package honor

import (
	"context"

	"lina-plugin-sicau-niu/backend/internal/service/honor/internal/certrender"
)

// Service defines the C5 honor contract: operator honor-definition CRUD and the
// read-only player honor unlock list.
type Service interface {
	// List returns one DB-side paged, sort-then-ID-descending honor-definition page
	// for the operator console. Filtering (keyword/honor-type/unlock-type),
	// ordering and pagination run in the database. It returns a query bizerr on
	// store failure.
	List(ctx context.Context, in *ListInput) (out *ListOutput, err error)
	// Get returns one honor-definition detail by ID. It returns CodeHonorNotFound
	// for a missing ID and a query bizerr on store failure.
	Get(ctx context.Context, id int64) (out *HonorItem, err error)
	// Create inserts one honor definition after honor-type/unlock-type enum,
	// threshold/category shape and code-uniqueness validation. It returns the
	// relevant validation bizerr or the new ID on success.
	Create(ctx context.Context, in *MutateInput) (id int64, err error)
	// Update modifies one honor definition after existence, enum, shape and
	// code-uniqueness validation. It returns CodeHonorNotFound for a missing ID,
	// the relevant validation bizerr, or nil on success.
	Update(ctx context.Context, id int64, in *MutateInput) error
	// Delete soft-deletes one honor definition. It returns CodeHonorNotFound for a
	// missing ID and a write bizerr on store failure.
	Delete(ctx context.Context, id int64) error
	// PlayerHonors returns every honor definition with the requesting player's
	// read-only unlock status, computed from the player's feeding count, activation
	// count and card-collection completion against each honor's unlock rule. It is
	// isolated to playerID, never persists a grant, and pre-aggregates the required
	// counts in a fixed number of batched queries to avoid per-honor queries. It
	// returns a query bizerr on store failure.
	PlayerHonors(ctx context.Context, playerID int64) (out []*PlayerHonorItem, err error)
	// PlayerCertificate renders the personalized electronic certificate PNG for a
	// certificate honor the requesting player already holds. It rejects when the
	// honor is missing, is not a certificate type, or the player does not hold it,
	// so a player can only obtain their own granted certificate. It returns the
	// base64-encoded PNG together with the structured certificate fields.
	PlayerCertificate(ctx context.Context, playerID, honorID int64) (out *Certificate, err error)
}

// Interface compliance assertion for the default honor service implementation.
var _ Service = (*serviceImpl)(nil)

// CertRenderer is the electronic-certificate PNG output seam re-exported from this
// package so the route-assembly layer can construct the default renderer and inject
// it without importing the internal seam package.
type CertRenderer = certrender.CertRenderer

// NewBasicCertRenderer creates the default basic electronic-certificate renderer.
func NewBasicCertRenderer() CertRenderer {
	return certrender.New()
}

// Config carries the plain-value runtime configuration for the honor capability:
// only the campus anniversary badge text composed onto generated certificates.
type Config struct {
	// CampusBadge is the campus anniversary badge text rendered on certificates.
	CampusBadge string
}

// serviceImpl implements Service against the plugin-owned honor_def, feeding,
// activation, card and user_honor tables. It reads those tables through the
// generated DAO and reuses the card package's exported category-enum validator for
// the category-complete rule; its only runtime dependencies are the replaceable
// certificate renderer and the plain-value campus badge injected at assembly time.
type serviceImpl struct {
	certRenderer CertRenderer // certRenderer is the replaceable certificate PNG output seam.
	campusBadge  string       // campusBadge is the anniversary badge text rendered on certificates.
}

// New creates a honor service with the certificate renderer and campus badge
// injected at assembly time. The component reads the plugin's own honor_def,
// feeding, activation, card and user_honor tables through the generated DAO and
// reuses the card package's exported category-enum contract.
func New(certRenderer CertRenderer, config Config) Service {
	return &serviceImpl{
		certRenderer: certRenderer,
		campusBadge:  config.CampusBadge,
	}
}
