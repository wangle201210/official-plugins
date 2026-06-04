// Package settlement implements the sicau-niu C7 operator settlement capability:
// the operations dashboard, the player roster export, the batch certificate
// issuance, the shared-device risk view and the settlement archive. The dashboard,
// export and risk view are read-only aggregates derived from the C1-C5 tables; the
// certificate issuance writes the C5 user_honor grant table idempotently; the
// archive owns the C7 settlement table where a frozen dashboard snapshot is
// persisted. Every aggregate runs on the database side and every list is bounded,
// so no view loads an unbounded set into memory and no per-row lookup is issued.
// All store access uses the generated DAO/DO objects so GoFrame manages soft-delete
// and timestamp columns.
package settlement

import "context"

// Bound constants for the settlement views. They cap the roster export, the
// archive list and the risk cluster list so an operator endpoint can never trigger
// an unbounded scan.
const (
	// exportMaxRows hard-caps the player roster export; beyond it the response is
	// truncated and flags it.
	exportMaxRows = 5000
	// archiveListLimit caps the settlement archive list.
	archiveListLimit = 200
	// riskClusterLimit caps the shared-device risk cluster list.
	riskClusterLimit = 200
)

// Service defines the C7 operator settlement contract.
type Service interface {
	// Dashboard returns the activity operations dashboard, every figure aggregated
	// on the database side. It returns a query bizerr on store failure.
	Dashboard(ctx context.Context) (out *Dashboard, err error)
	// ExportPlayers returns a bounded player roster projection for CSV export. The
	// per-player activation count and feeding total are batch-assembled to avoid
	// N+1; when more players exist than the cap the result is truncated and flagged.
	// It returns a query bizerr on store failure.
	ExportPlayers(ctx context.Context) (out *PlayerExport, err error)
	// IssueCertificates batch-issues one certificate honor to the players who
	// satisfy its unlock rule, writing the grant table idempotently in a
	// transaction. It returns a business error when the honor is missing, is not a
	// certificate, or uses an unsupported (collection) unlock rule.
	IssueCertificates(ctx context.Context, honorID int64) (out *IssueResult, err error)
	// RiskDeviceClusters returns the shared-device risk view: clusters of players
	// sharing one non-empty device fingerprint, with member nicknames, bounded. It
	// returns a query bizerr on store failure.
	RiskDeviceClusters(ctx context.Context) (out *RiskClusters, err error)
	// CreateArchive freezes the current dashboard into a persisted settlement
	// snapshot with the given title and returns its ID. It returns a business error
	// on serialization or store failure.
	CreateArchive(ctx context.Context, title string) (id int64, err error)
	// ListArchives returns the settlement archives ordered by archive time
	// descending, bounded. It returns a query bizerr on store failure.
	ListArchives(ctx context.Context) (out *ArchiveList, err error)
	// Activity returns the M5 dashboard activity metrics: the daily active-user
	// series for the last days days (bounded) and the next-day / 7-day retention
	// over elapsed cohorts. It returns a query bizerr on store failure.
	Activity(ctx context.Context, days int) (out *Activity, err error)
}

// Interface compliance assertion for the default settlement service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned tables. It holds no
// runtime dependency: every view reads the plugin's own tables through the
// generated DAO and the list bounds are fixed package constants.
type serviceImpl struct{}

// New creates an operator settlement service. The component reads and writes the
// plugin's own tables through the generated DAO and therefore takes no
// dependencies; the list bounds are fixed package constants so every operator
// endpoint is always bounded.
func New() Service {
	return &serviceImpl{}
}
