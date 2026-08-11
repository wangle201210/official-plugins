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

import (
	"context"

	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

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
	// defaultFeedDailyThreshold and defaultStealDailyThreshold are the fallback
	// per-day anomaly thresholds when config is absent.
	defaultFeedDailyThreshold  = 100
	defaultStealDailyThreshold = 5
	// defaultAnomalyLimit caps the anomaly alert list when config is absent.
	defaultAnomalyLimit = 200
)

// Config carries the plain-value runtime configuration for the settlement
// capability: the per-day anomaly thresholds and the anomaly list cap. It holds
// only scalar tuning values and never runtime dependencies.
type Config struct {
	// FeedDailyThreshold flags a player whose single-day feeding count exceeds it.
	FeedDailyThreshold int
	// StealDailyThreshold flags a player whose single-day steal count exceeds it.
	StealDailyThreshold int
	// AnomalyLimit caps the anomaly alert list size.
	AnomalyLimit int
}

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
	// CertificateOptions returns the certificate honors that can be issued through
	// the operator settlement batch action. It only includes certificate honors
	// whose unlock rule is supported by IssueCertificates, projected in one
	// bounded query for the selector UI.
	CertificateOptions(ctx context.Context) (out []*CertificateOption, err error)
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
	// series for the requested number of days (bounded) and the next-day / 7-day retention
	// over elapsed cohorts. It returns a query bizerr on store failure.
	Activity(ctx context.Context, days int) (out *Activity, err error)
	// Anomalies returns the bounded risk anomaly alerts: players whose single-day
	// feeding or steal count exceeds the configured threshold. It returns a query
	// bizerr on store failure.
	Anomalies(ctx context.Context) (out *AnomalyAlerts, err error)
}

// Interface compliance assertion for the default settlement service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned tables. The optional
// rules service supplies operator-maintained anomaly thresholds and cap; scalar
// constructor values remain fallback. Every view reads the plugin's own tables
// through the generated DAO and the remaining list bounds are fixed constants.
type serviceImpl struct {
	rulesSvc            rulessvc.Service // rulesSvc supplies current anomaly thresholds and cap when injected.
	feedDailyThreshold  int              // feedDailyThreshold flags single-day feeding above it.
	stealDailyThreshold int              // stealDailyThreshold flags single-day steal above it.
	anomalyLimit        int              // anomalyLimit caps the anomaly alert list.
}

// New creates an operator settlement service with an optional runtime-rule
// service and fallback plain-value anomaly configuration. The component reads and
// writes the plugin's own tables through the generated DAO; non-positive fallback
// thresholds and caps degrade to package defaults so the anomaly view is bounded.
func New(rulesSvc rulessvc.Service, config Config) Service {
	impl := &serviceImpl{
		rulesSvc:            rulesSvc,
		feedDailyThreshold:  config.FeedDailyThreshold,
		stealDailyThreshold: config.StealDailyThreshold,
		anomalyLimit:        config.AnomalyLimit,
	}
	if impl.feedDailyThreshold <= 0 {
		impl.feedDailyThreshold = defaultFeedDailyThreshold
	}
	if impl.stealDailyThreshold <= 0 {
		impl.stealDailyThreshold = defaultStealDailyThreshold
	}
	if impl.anomalyLimit <= 0 {
		impl.anomalyLimit = defaultAnomalyLimit
	}
	return impl
}
