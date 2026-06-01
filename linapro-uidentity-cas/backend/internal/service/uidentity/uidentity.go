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
	LoginTypePassword = "pwd"     // Password login.
	LoginTypeSMS      = "sms"     // SMS login.
	LoginTypeCAS      = "cas"     // CAS ticket login.
	LoginTypeAuto     = "auto"    // Automatic login.
	LoginTypeWechat   = "wechat"  // Wechat login.
	LoginTypeUnionID  = "unionID" // Union ID login.
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
	// CheckAccountImport validates one legacy account import workbook without
	// writing data and returns the number of importable rows.
	CheckAccountImport(ctx context.Context, in AccountImportInput) (*AccountImportCheckOutput, error)
	// ImportAccounts imports or updates plugin account rows from one legacy
	// workbook, matching existing accounts by tenant-scoped account number.
	ImportAccounts(ctx context.Context, in AccountImportInput) (*AccountImportOutput, error)
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
	// LoginByPassword validates application, account password, account status,
	// blacklist rules, and delegated accounts before issuing CAS TGT/ST values.
	LoginByPassword(ctx context.Context, in PasswordLoginInput) (*RuntimeLoginOutput, error)
	// LoginByPhone validates application, phone-bound account, and SMS code
	// before issuing CAS TGT/ST values for legacy phone login clients.
	LoginByPhone(ctx context.Context, in PhoneLoginInput) (*RuntimeLoginOutput, error)
	// LoginByUnionID resolves a Wechat union ID to an account and issues CAS
	// TGT/ST values when the account can access the target application.
	LoginByUnionID(ctx context.Context, in UnionIDLoginInput) (*RuntimeLoginOutput, error)
	// IssueServiceTicketFromTGT issues a new one-time ST from an existing TGT
	// and optional selected delegated account.
	IssueServiceTicketFromTGT(ctx context.Context, in ServiceTicketInput) (*ServiceTicketOutput, error)
	// ValidateServiceTicket consumes one ST, enforces selected-account access,
	// records the selected account in the login log, and returns projections.
	ValidateServiceTicket(ctx context.Context, in ServiceValidateInput) (*ServiceValidateOutput, error)
	// DeleteTicket deletes one TGT/ST/access/challenge token by runtime value.
	DeleteTicket(ctx context.Context, ticket string) error
	// IssueOAuthToken creates an OAuth token and log for an account and
	// application after runtime access checks.
	IssueOAuthToken(ctx context.Context, in OAuthIssueInput) (*OAuthIssueOutput, error)
	// IssueRuntimeToken validates a client secret and account password before
	// issuing a runtime access token compatible with the old token API.
	IssueRuntimeToken(ctx context.Context, in RuntimeTokenInput) (*RuntimeTokenOutput, error)
	// GetUserInfoByRuntimeToken validates a runtime access token and returns
	// account and application projections without consuming the token.
	GetUserInfoByRuntimeToken(ctx context.Context, accessToken string) (*RuntimeTokenInfoOutput, error)
	// StartActivation verifies base account information and creates an
	// activation challenge backed by plugin token storage.
	StartActivation(ctx context.Context, in ActivationStartInput) (*ActivationOutput, error)
	// RecordActivationFace stores a face proof marker for an activation
	// challenge without requiring external face-service integration.
	RecordActivationFace(ctx context.Context, in ActivationFaceInput) (*ActivationStepOutput, error)
	// SetActivationPassword validates and stores a new password for the account
	// attached to an activation challenge.
	SetActivationPassword(ctx context.Context, in ActivationPasswordInput) (*ActivationStepOutput, error)
	// SetActivationPhone validates an activation SMS code, binds phone, and
	// activates the account attached to the challenge.
	SetActivationPhone(ctx context.Context, in ActivationPhoneInput) (*ActivationStepOutput, error)
	// SetActivationWechat binds a Wechat union ID and activates the account
	// attached to the challenge.
	SetActivationWechat(ctx context.Context, in ActivationWechatInput) (*ActivationStepOutput, error)
	// ActivationState returns the current activation challenge stage and
	// account status.
	ActivationState(ctx context.Context, challengeID string) (*ActivationStateOutput, error)
	// LookupUnionID resolves a Wechat union ID or creates a bind challenge when
	// the union ID is not bound to any account.
	LookupUnionID(ctx context.Context, unionID string) (*UnionIDLookupOutput, error)
	// BindUnionID consumes a bind challenge and attaches the union ID to an
	// account verified by phone/SMS or number/password.
	BindUnionID(ctx context.Context, in UnionIDBindInput) (*UnionIDBindOutput, error)
	// ChangeRuntimePassword updates an account password through runtime
	// self-service policy checks.
	ChangeRuntimePassword(ctx context.Context, number string, newPassword string) error
	// ChangeRuntimePhone verifies an SMS bind code and updates account phone.
	ChangeRuntimePhone(ctx context.Context, in ChangePhoneInput) error
	// ChangeRuntimeEmail updates account detail email for one account.
	ChangeRuntimeEmail(ctx context.Context, number string, email string) error
	// ChangeRuntimeQQ updates account detail QQ for one account.
	ChangeRuntimeQQ(ctx context.Context, number string, qq string) error
	// UnbindRuntimeWechat clears the Wechat union ID for one account.
	UnbindRuntimeWechat(ctx context.Context, number string) error
	// GetRuntimeUserInfo returns the account projection used by legacy user
	// self-service endpoints.
	GetRuntimeUserInfo(ctx context.Context, number string) (*RuntimeAccount, error)
	// ListRuntimeUserLoginLogs returns a bounded paged login-log list for one
	// runtime account.
	ListRuntimeUserLoginLogs(ctx context.Context, in UserLogListInput) (*ResourceListOutput, error)
	// ListRuntimeApplications returns enabled applications not blocked for one
	// runtime account using set-based blacklist reads.
	ListRuntimeApplications(ctx context.Context, in UserApplicationListInput) (*RuntimeApplicationListOutput, error)
	// ListRuntimeAppRoles returns delegated account-app roles granted by one
	// runtime account.
	ListRuntimeAppRoles(ctx context.Context, in UserAppRoleListInput) (*ResourceListOutput, error)
	// CreateRuntimeAppRole creates one delegated account-app role for a runtime
	// account after resolving both accounts in tenant scope.
	CreateRuntimeAppRole(ctx context.Context, in UserAppRoleCreateInput) (int64, error)
	// UpdateRuntimeAppRole updates delegated role expiration when owned by the
	// runtime granting account.
	UpdateRuntimeAppRole(ctx context.Context, in UserAppRoleUpdateInput) error
	// SendSMSCode records one bounded plugin-local SMS verification code for
	// CAS login, activation, phone binding, or password reset.
	SendSMSCode(ctx context.Context, in SMSSendInput) (*SMSSendOutput, error)
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

// AccountImportInput carries account import workbook options.
type AccountImportInput struct {
	Filepath string
	Limit    int
}

// AccountImportCheckOutput carries import validation metadata.
type AccountImportCheckOutput struct {
	Rows int
}

// AccountImportOutput carries account import results.
type AccountImportOutput struct {
	Success      int
	FailedNumber []string
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

// RuntimeApplication is the service-level application projection for legacy
// CAS and token clients.
type RuntimeApplication struct {
	ID          int64
	Name        string
	Alias       string
	ClientID    string
	AccessModel string
	CallbackURL string
}

// RuntimeAccountDetail is the service-level account detail projection.
type RuntimeAccountDetail struct {
	Birthday string
	Email    string
	Gender   int
	QQ       string
	Wechat   string
	Idcard   string
	Avatar   string
	Face     string
}

// RuntimeAccount is the service-level account projection for runtime clients.
type RuntimeAccount struct {
	ID            int64
	Number        string
	Name          string
	Phone         string
	Status        int
	PassLevel     int
	ContainerID   int64
	ContainerName string
	UnitID        int64
	UnitName      string
	ExpireAt      *int64
	Groups        []string
	Detail        *RuntimeAccountDetail
}

// PasswordLoginInput carries legacy CAS password-login input.
type PasswordLoginInput struct {
	ClientID string
	Number   string
	Password string
}

// PhoneLoginInput carries legacy CAS phone-login input.
type PhoneLoginInput struct {
	ClientID string
	Phone    string
	Code     string
}

// UnionIDLoginInput carries legacy CAS union-ID-login input.
type UnionIDLoginInput struct {
	ClientID string
	UnionID  string
}

// RuntimeLoginOutput carries issued CAS ticket values and accessible accounts.
type RuntimeLoginOutput struct {
	CallbackURL string
	TGT         string
	ST          string
	User        *RuntimeAccount
	Users       []*RuntimeAccount
	App         *RuntimeApplication
}

// ServiceTicketInput carries TGT-to-ST issue input.
type ServiceTicketInput struct {
	ClientID  string
	TGT       string
	AccountID int64
}

// ServiceTicketOutput carries one newly issued service ticket.
type ServiceTicketOutput struct {
	ST          string
	CallbackURL string
}

// ServiceValidateInput carries one service ticket validation request.
type ServiceValidateInput struct {
	Ticket string
	UserID int64
}

// ServiceValidateOutput carries consumed ticket validation result.
type ServiceValidateOutput struct {
	Ticket  string
	User    *RuntimeAccount
	App     *RuntimeApplication
	Success bool
}

// RuntimeTokenInput carries legacy token issue input.
type RuntimeTokenInput struct {
	ClientID string
	Secret   string
	Number   string
	Password string
}

// RuntimeTokenOutput carries issued runtime access token data.
type RuntimeTokenOutput struct {
	AccessToken string
	ExpiredAt   *int64
}

// RuntimeTokenInfoOutput carries runtime token user-info result.
type RuntimeTokenInfoOutput struct {
	User  *RuntimeAccount
	Users []*RuntimeAccount
	App   *RuntimeApplication
}

// ActivationStartInput carries basic activation verification data.
type ActivationStartInput struct {
	Number string
	Name   string
	Idcard string
}

// ActivationFaceInput carries face proof update data.
type ActivationFaceInput struct {
	ChallengeID string
	FaceURL     string
}

// ActivationPasswordInput carries activation password setup data.
type ActivationPasswordInput struct {
	ChallengeID string
	Password    string
}

// ActivationPhoneInput carries activation phone binding data.
type ActivationPhoneInput struct {
	ChallengeID string
	Phone       string
	Code        string
}

// ActivationWechatInput carries activation Wechat binding data.
type ActivationWechatInput struct {
	ChallengeID string
	UnionID     string
}

// ActivationOutput carries activation challenge metadata.
type ActivationOutput struct {
	ChallengeID string
	NeedFace    bool
	Status      int
}

// ActivationStepOutput carries mutation result for one activation step.
type ActivationStepOutput struct {
	ChallengeID string
	Success     bool
}

// ActivationStateOutput carries activation state read result.
type ActivationStateOutput struct {
	ChallengeID string
	Success     bool
	Status      int
	Stage       string
}

// UnionIDLookupOutput carries union-ID lookup or bind challenge metadata.
type UnionIDLookupOutput struct {
	Number      string
	ChallengeID string
	CallbackURL string
}

// UnionIDBindInput carries union-ID binding data.
type UnionIDBindInput struct {
	ChallengeID string
	BindType    int
	Phone       string
	Code        string
	Number      string
	Password    string
}

// UnionIDBindOutput carries bound account metadata.
type UnionIDBindOutput struct {
	Number string
}

// ChangePhoneInput carries runtime phone update data.
type ChangePhoneInput struct {
	Number string
	Phone  string
	Code   string
}

// UserLogListInput carries runtime login-log list filters.
type UserLogListInput struct {
	Number   string
	PageNum  int
	PageSize int
}

// UserAppRoleListInput carries runtime delegated-role list filters.
type UserAppRoleListInput struct {
	Number   string
	PageNum  int
	PageSize int
}

// UserApplicationListInput carries runtime application list filters.
type UserApplicationListInput struct {
	Number   string
	PageNum  int
	PageSize int
}

// RuntimeApplicationListOutput carries paged runtime application results.
type RuntimeApplicationListOutput struct {
	List  []*RuntimeApplication
	Total int
}

// UserAppRoleCreateInput carries delegated role creation data.
type UserAppRoleCreateInput struct {
	Number          string
	EmpoweredNumber string
	AppID           int64
	ExpireAt        *int64
}

// UserAppRoleUpdateInput carries delegated role update data.
type UserAppRoleUpdateInput struct {
	Number   string
	ID       int64
	ExpireAt *int64
}

// SMSSendInput carries one SMS verification-code send request.
type SMSSendInput struct {
	Type  string
	Phone string
}

// SMSSendOutput carries the plugin SMS record ID.
type SMSSendOutput struct {
	ID int64
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
