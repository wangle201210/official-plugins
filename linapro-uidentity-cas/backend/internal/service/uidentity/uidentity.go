// Package uidentity implements tenant-scoped identity, CAS, OAuth, password,
// blacklist, audit, and statistics services for the linapro-uidentity-cas
// source plugin. It owns only plugin-prefixed tables and consumes host
// capabilities through explicit constructor injection.
package uidentity

import (
	"context"

	plugincontract "lina-core/pkg/plugin/capability/contract"
)

// Account status values.
const (
	AccountStatusNotActive = 0 // Account is created but not active.
	AccountStatusNormal    = 1 // Account can pass runtime access checks.
	AccountStatusLocked    = 2 // Account is locked.
)

// Application status values.
const (
	ApplicationStatusDisabled = 0 // Application is disabled.
	ApplicationStatusEnabled  = 1 // Application is enabled.
)

// Login type values recorded in CAS login logs.
const (
	LoginTypePassword = "pwd"  // Password login.
	LoginTypeSMS      = "sms"  // SMS login.
	LoginTypeCAS      = "cas"  // CAS ticket login.
	LoginTypeAuto     = "auto" // Automatic login.
)

// Service defines UIdentity resource CRUD, runtime authentication, password,
// OAuth, audit, and statistics behavior.
type Service interface {
	// ListResource returns a tenant-scoped paged list for one supported resource.
	// It applies database-side filtering, ordering, pagination, and batch
	// projection before returning API-ready records.
	ListResource(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error)
	// GetResource returns one tenant-visible resource record by ID or a
	// structured not-found error when absent or outside tenant scope.
	GetResource(ctx context.Context, resource string, id int64) (Record, error)
	// CreateResource creates one resource record from API field values and
	// returns the new ID. Audit fields and tenant ownership are filled from ctx.
	CreateResource(ctx context.Context, resource string, body map[string]any) (int64, error)
	// UpdateResource updates one resource record by ID from partial API field
	// values. It returns not-found when the row is outside tenant scope.
	UpdateResource(ctx context.Context, resource string, id int64, body map[string]any) error
	// DeleteResource soft-deletes or hard-deletes supported records after
	// tenant visibility checks. IDs are capped and validated before delete.
	DeleteResource(ctx context.Context, resource string, ids string) error
	// ResetAccountPassword resets one account password using active password
	// rules and records plugin-owned password metadata.
	ResetAccountPassword(ctx context.Context, accountID int64, newPassword string) error
	// CreatePasswordChallenge creates a short-lived self-service password reset
	// challenge for an account number and returns its account status.
	CreatePasswordChallenge(ctx context.Context, number string) (*PasswordChallengeOutput, error)
	// VerifyPasswordChallengePhone verifies challenge ownership by phone and SMS
	// code using plugin-owned SMS records.
	VerifyPasswordChallengePhone(ctx context.Context, challengeID string, phone string, code string) (string, error)
	// ResetPasswordByChallenge consumes a verified challenge and resets the
	// matched account password.
	ResetPasswordByChallenge(ctx context.Context, challengeID string, newPassword string) error
	// LoginByCASTicket validates a CAS ticket, enforces application access
	// rules, records a CAS log, and returns the resolved account projection.
	LoginByCASTicket(ctx context.Context, in CASLoginInput) (*CASLoginOutput, error)
	// IssueOAuthToken creates an OAuth token and log for an account and
	// application after runtime access checks.
	IssueOAuthToken(ctx context.Context, in OAuthIssueInput) (*OAuthIssueOutput, error)
	// Stats returns aggregate identity statistics using database-side grouping
	// and batch name projection.
	Stats(ctx context.Context) (*StatsOutput, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc    plugincontract.BizCtxService       // Business context bridge.
	configSvc    plugincontract.ConfigService       // Plugin-scoped static config reader.
	tenantFilter plugincontract.TenantFilterService // Tenant query filter bridge.
}

// New creates and returns a new UIdentity service instance.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	configSvc plugincontract.ConfigService,
	tenantFilter plugincontract.TenantFilterService,
) Service {
	return &serviceImpl{
		bizCtxSvc:    bizCtxSvc,
		configSvc:    configSvc,
		tenantFilter: tenantFilter,
	}
}

// Record is an API-ready field map using JSON field names.
type Record map[string]any

// ResourceListInput carries all supported generic list filters.
type ResourceListInput struct {
	Resource    string
	PageNum     int
	PageSize    int
	Keyword     string
	AccountId   int64
	AppId       int64
	GroupId     int64
	ContainerId int64
	UnitId      int64
	Status      *int
	PassLevels  []int64
	GroupIds    []int64
	OrderBy     string
	Order       string
}

// ResourceListOutput carries paged generic resource records.
type ResourceListOutput struct {
	List  []Record
	Total int
}

// PasswordChallengeOutput carries challenge creation result metadata.
type PasswordChallengeOutput struct {
	ChallengeID string
	Status      int
}

// CASLoginInput carries CAS login validation input.
type CASLoginInput struct {
	Ticket string
	UserID int64
	AppID  int64
}

// CASLoginOutput carries resolved CAS login metadata.
type CASLoginOutput struct {
	Number    string
	AccountID int64
	AppID     int64
}

// OAuthIssueInput carries OAuth token issue input.
type OAuthIssueInput struct {
	AccountID   int64
	AppID       int64
	RedirectURI string
	Scope       string
	TtlSeconds  int64
}

// OAuthIssueOutput carries issued OAuth token values.
type OAuthIssueOutput struct {
	Code      string
	Access    string
	Refresh   string
	ExpiredAt *int64
}

// StatItem carries one aggregate statistic bucket.
type StatItem struct {
	Name  string
	Total int64
}

// StatsOutput carries aggregate identity statistics.
type StatsOutput struct {
	AccountCount     int64
	AuthCount        int64
	AppCount         int64
	UserByContainer  []*StatItem
	AppByType        []*StatItem
	AuthByType       []*StatItem
	CasByAccountType []*StatItem
	PassLevel        []*StatItem
	LoginType        []*StatItem
	LoginApp         []*StatItem
}
