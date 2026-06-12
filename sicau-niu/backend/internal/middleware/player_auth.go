// Package middleware implements the sicau-niu plugin-owned HTTP middlewares. Its
// component contract is the player authentication middleware: it verifies the
// plugin self-signed player token from the Authorization header and injects the
// authenticated player ID into the request context so player-facing controllers
// can constrain all access to the current player's own data. This middleware is
// distinct from the host Auth+Tenancy+Permission chain, which governs the
// operator (admin) API surface.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/middleware/playerctx"
	tokensvc "lina-plugin-sicau-niu/backend/internal/service/token"
)

// authorizationHeader is the request header carrying the player bearer token.
const authorizationHeader = "Authorization"

// bearerPrefix is the case-insensitive scheme prefix expected before the token.
const bearerPrefix = "Bearer "

// PlayerAuth verifies the player session token and injects the authenticated
// player identity into the request context for downstream player-facing
// handlers.
type PlayerAuth struct {
	tokenSvc tokensvc.Service // tokenSvc verifies player session tokens.
}

// NewPlayerAuth creates the player authentication middleware with an explicit
// token verifier dependency.
func NewPlayerAuth(tokenSvc tokensvc.Service) *PlayerAuth {
	return &PlayerAuth{tokenSvc: tokenSvc}
}

// Handle is the ghttp middleware handler. It extracts the bearer token, verifies
// it, and on success injects the player ID into the request context before
// continuing the chain. On any failure it aborts with an authentication bizerr
// and does not invoke the next handler.
func (m *PlayerAuth) Handle(r *ghttp.Request) {
	token, err := extractBearerToken(r.Header.Get(authorizationHeader))
	if err != nil {
		m.abort(r, err)
		return
	}

	playerID, err := m.tokenSvc.Verify(r.Context(), token)
	if err != nil {
		m.abort(r, err)
		return
	}

	r.SetCtx(playerctx.With(r.Context(), playerID))
	r.Middleware.Next()
}

// abort records the authentication error and stops the request with a 401 so
// the host unified response middleware renders the structured business error.
// The status is assigned without writing a body: WriteStatus would buffer the
// plain "Unauthorized" text and the host response middleware skips rendering
// whenever the buffer is non-empty, which would drop the structured error.
func (m *PlayerAuth) abort(r *ghttp.Request, err error) {
	r.SetError(err)
	r.Response.Status = http.StatusUnauthorized
}

// extractBearerToken parses the bearer token from the Authorization header value.
func extractBearerToken(headerValue string) (string, error) {
	headerValue = strings.TrimSpace(headerValue)
	if headerValue == "" {
		return "", bizerr.NewCode(tokensvc.CodeTokenInvalid)
	}
	if len(headerValue) < len(bearerPrefix) ||
		!strings.EqualFold(headerValue[:len(bearerPrefix)], bearerPrefix) {
		return "", bizerr.NewCode(tokensvc.CodeTokenInvalid)
	}
	token := strings.TrimSpace(headerValue[len(bearerPrefix):])
	if token == "" {
		return "", bizerr.NewCode(tokensvc.CodeTokenInvalid)
	}
	return token, nil
}
