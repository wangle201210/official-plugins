// token_unconfigured.go provides the deferred-failure token service used when
// source plugin route registration must continue without a configured
// token.secret. It never signs or verifies tokens and only returns the existing
// missing-secret business error from request-time token operations.

package token

import (
	"context"

	"lina-core/pkg/bizerr"
)

// Interface compliance assertion for the deferred-configuration token service.
var _ Service = (*unconfiguredService)(nil)

// unconfiguredService defers the missing-secret failure from startup to the
// specific player-token operation that requires signing or verification.
type unconfiguredService struct{}

// Sign reports that token.secret is missing and never issues a token.
func (*unconfiguredService) Sign(_ context.Context, _ int64) (string, error) {
	return "", bizerr.NewCode(CodeTokenSecretMissing)
}

// Verify reports that token.secret is missing and never authenticates a token.
func (*unconfiguredService) Verify(_ context.Context, _ string) (int64, error) {
	return 0, bizerr.NewCode(CodeTokenSecretMissing)
}
