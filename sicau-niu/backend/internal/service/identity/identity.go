// Package identity implements the sicau-niu player identity capability: WeChat
// code login with first-login provisioning and plugin token issuance, phone
// authorization binding under a one-phone-one-account constraint with device
// fingerprint capture, identity profile read/update with enum and college
// validation, and an operator-facing paged player query. Player accounts are
// plugin-owned and isolated from host admin/tenant users; every player-facing
// operation is constrained to the authenticated player's own row by the caller
// passing the player ID resolved by the player auth middleware. All store access
// uses the generated DAO/DO objects so soft-delete and timestamps are managed by
// GoFrame.
package identity

import (
	"context"

	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	tokensvc "lina-plugin-sicau-niu/backend/internal/service/token"
	wechatsvc "lina-plugin-sicau-niu/backend/internal/service/wechat"
)

// Service defines the player identity contract.
type Service interface {
	// Login exchanges a WeChat login code for an authenticated player session. It
	// resolves the openid through the WeChat gateway, provisions a new player on
	// first login, issues a player token, and reports whether the account was
	// newly created. It returns a WeChat/token bizerr on failure.
	Login(ctx context.Context, in *LoginInput) (out *LoginOutput, err error)
	// BindPhone binds a WeChat-authorized phone number to playerID under the
	// one-phone-one-account constraint and records the device fingerprint. It
	// returns CodePhoneTaken when the phone already belongs to another player and
	// a decode/store bizerr on other failures.
	BindPhone(ctx context.Context, playerID int64, in *BindPhoneInput) error
	// GetProfile returns playerID's own identity profile with timestamps as Unix
	// milliseconds. It returns CodePlayerNotFound when the player row is missing.
	GetProfile(ctx context.Context, playerID int64) (out *ProfileOutput, err error)
	// UpdateProfile updates playerID's own identity profile after identity-type,
	// college-existence and grade/graduation-year validation. It returns a
	// parameter bizerr on invalid input and CodePlayerNotFound for a missing row.
	UpdateProfile(ctx context.Context, playerID int64, in *UpdateProfileInput) error
	// ListPlayers returns one DB-side paged, read-only player page for the
	// operator console with timestamps as Unix milliseconds. It returns a query
	// bizerr on store failure.
	ListPlayers(ctx context.Context, in *ListPlayersInput) (out *ListPlayersOutput, err error)
}

// Interface compliance assertion for the default identity service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service. Its runtime dependencies are injected
// explicitly as separate interface parameters so dependency changes surface at
// compile time.
type serviceImpl struct {
	gateway    wechatsvc.Gateway  // gateway resolves WeChat openid and phone numbers.
	tokenSvc   tokensvc.Service   // tokenSvc issues player session tokens.
	collegeSvc collegesvc.Service // collegeSvc validates selected college existence.
}

// New creates a player identity service with explicit dependencies: the WeChat
// gateway, the player token issuer, and the college service used for college
// existence validation.
func New(gateway wechatsvc.Gateway, tokenSvc tokensvc.Service, collegeSvc collegesvc.Service) Service {
	return &serviceImpl{
		gateway:    gateway,
		tokenSvc:   tokenSvc,
		collegeSvc: collegeSvc,
	}
}
