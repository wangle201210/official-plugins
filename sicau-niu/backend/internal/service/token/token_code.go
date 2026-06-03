// token_code.go defines the player session token business error codes.

package token

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTokenSecretMissing reports that the token signing secret is not configured.
	CodeTokenSecretMissing = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_TOKEN_SECRET_MISSING",
		"Player token secret is not configured",
		gcode.CodeInvalidConfiguration,
	)
	// CodeTokenTTLInvalid reports that the token TTL is not a positive duration.
	CodeTokenTTLInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_TOKEN_TTL_INVALID",
		"Player token TTL must be a positive duration",
		gcode.CodeInvalidConfiguration,
	)
	// CodeTokenSignFailed reports that issuing a player token failed.
	CodeTokenSignFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_TOKEN_SIGN_FAILED",
		"Failed to issue player token",
		gcode.CodeInternalError,
	)
	// CodeTokenInvalid reports that the player token is malformed or tampered.
	CodeTokenInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_TOKEN_INVALID",
		"Player token is invalid",
		gcode.CodeNotAuthorized,
	)
	// CodeTokenExpired reports that the player token has expired.
	CodeTokenExpired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_TOKEN_EXPIRED",
		"Player token has expired",
		gcode.CodeNotAuthorized,
	)
)
