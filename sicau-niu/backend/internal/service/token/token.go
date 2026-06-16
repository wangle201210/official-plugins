// Package token implements the sicau-niu player session token: a stateless,
// plugin self-signed token of the form `header.payload.signature`. The payload
// carries the player ID and an expiry instant; the signature is an HMAC-SHA256
// over `header.payload` using the plugin-configured secret. The component owns
// no host session state, so it is safe across nodes for the platform_only
// deployment. It depends only on the Go standard library (no third-party JWT
// dependency) per the plugin token constraint.
package token

import (
	"context"
	"time"

	"lina-core/pkg/bizerr"
)

// Service defines the player session token contract.
type Service interface {
	// Sign issues a signed player session token for playerID. The token embeds
	// the player ID and an expiry computed from the configured TTL. It returns a
	// bizerr when playerID is non-positive, signing fails, or token.secret is not
	// configured.
	Sign(ctx context.Context, playerID int64) (token string, err error)
	// Verify validates token signature and expiry and returns the embedded
	// player ID. It returns CodeTokenInvalid for malformed or tampered tokens and
	// CodeTokenExpired for expired tokens; when token.secret is not configured it
	// returns CodeTokenSecretMissing. On any error playerID is 0.
	Verify(ctx context.Context, token string) (playerID int64, err error)
}

// Interface compliance assertion for the default token service implementation.
var _ Service = (*serviceImpl)(nil)

// Config carries the pure-value token configuration. It intentionally holds no
// runtime interface dependency so it can be constructed directly from plugin
// config values.
type Config struct {
	// Secret is the HMAC signing secret. It must be non-empty.
	Secret string
	// TTL is the token validity duration applied at signing time.
	TTL time.Duration
}

// serviceImpl implements Service using HMAC-SHA256 over a compact JSON payload.
type serviceImpl struct {
	secret []byte        // secret is the HMAC signing key.
	ttl    time.Duration // ttl is the token validity duration.
	now    func() time.Time
}

// New creates a player token service from pure-value configuration. It returns
// a bizerr when the secret is empty or the TTL is non-positive. Callers that
// need to defer a missing-secret failure to request time should use
// NewUnconfigured.
func New(cfg Config) (Service, error) {
	if cfg.Secret == "" {
		return nil, bizerr.NewCode(CodeTokenSecretMissing)
	}
	if cfg.TTL <= 0 {
		return nil, bizerr.NewCode(CodeTokenTTLInvalid)
	}
	return &serviceImpl{
		secret: []byte(cfg.Secret),
		ttl:    cfg.TTL,
		now:    time.Now,
	}, nil
}

// NewUnconfigured returns a token service used when token.secret is intentionally
// absent. It lets the source plugin register routes while preserving the same
// business error for player-token operations.
func NewUnconfigured() Service {
	return &unconfiguredService{}
}
