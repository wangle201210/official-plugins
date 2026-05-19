// This file keeps media-specific dual authentication inside the source plugin.
// It composes the host-published LinaPro middleware chain with the media-owned
// Tieta token parser, while leaving core and other plugins untouched.

package backend

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginhost"
	"lina-core/pkg/pluginservice/contract"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// media dual-auth constants define request fields and plugin-local context keys.
const (
	mediaAuthorizationHeader = "Authorization"
	mediaBearerPrefix        = "Bearer "
	mediaTokenField          = "token"
	mediaAuthMessageFailed   = "LinaPro 与 Tieta 鉴权均未通过"
)

// mediaAuthContextKey isolates media authentication state from host context keys.
type mediaAuthContextKey string

// mediaTietaCurrentContextKey stores the media-owned Tieta context snapshot.
type mediaTietaCurrentContextKey struct{}

// media authentication context keys record whether Tieta or the host gate passed.
const (
	mediaTietaAuthenticatedKey mediaAuthContextKey = "media.auth.tieta.authenticated"
	mediaHostGatePassedKey     mediaAuthContextKey = "media.auth.host.gate.passed"
)

// mediaTietaAuthenticator defines the Tieta operation used by the route middleware.
type mediaTietaAuthenticator interface {
	// AuthenticateTietaToken validates one Tieta token and returns the mapped user.
	AuthenticateTietaToken(ctx context.Context, token string) (*mediasvc.TietaUser, error)
}

// mediaBizCtxWithTietaOverlay prefers Tieta-injected plugin context while
// preserving the host business context for LinaPro-authenticated calls.
func mediaBizCtxWithTietaOverlay(base contract.BizCtxService) contract.BizCtxService {
	return &mediaBizCtxOverlay{base: base}
}

// mediaBizCtxOverlay reads Tieta fallback context before delegating to the host context service.
type mediaBizCtxOverlay struct {
	base contract.BizCtxService // base is the host-published business context service.
}

// Current returns the Tieta fallback context when present, otherwise the host context.
func (s *mediaBizCtxOverlay) Current(ctx context.Context) contract.CurrentContext {
	if current, ok := mediaInjectedCurrentContext(ctx); ok {
		return current
	}
	if s != nil && s.base != nil {
		return s.base.Current(ctx)
	}
	return contract.CurrentContext{}
}

// mediaInjectedCurrentContext reports whether the context carries a plugin-injected identity.
func mediaInjectedCurrentContext(ctx context.Context) (contract.CurrentContext, bool) {
	if ctx == nil {
		return contract.CurrentContext{}, false
	}
	current, ok := ctx.Value(mediaTietaCurrentContextKey{}).(contract.CurrentContext)
	return current, ok
}

// mediaDualAuthMiddleware first runs the host LinaPro auth chain and then falls
// back to media-owned Tieta token authentication if the host chain blocks the route.
func mediaDualAuthMiddleware(
	hostAuth pluginhost.RouteMiddleware,
	tieta mediaTietaAuthenticator,
) pluginhost.RouteMiddleware {
	return func(r *ghttp.Request) {
		if r == nil {
			return
		}
		if hostAuth != nil {
			mediaResetHostGateState(r)
			hostAuth(r)
			if mediaContextBool(r, mediaHostGatePassedKey) {
				return
			}
			mediaClearHostFailure(r)
		}

		if err := mediaAuthenticateTietaRequest(r, tieta); err != nil {
			mediaWriteAuthFailure(r, err)
			return
		}
		r.Middleware.Next()
	}
}

// mediaSkipWhenTietaAuthenticated bypasses one host middleware after Tieta fallback succeeds.
func mediaSkipWhenTietaAuthenticated(handler pluginhost.RouteMiddleware) pluginhost.RouteMiddleware {
	return func(r *ghttp.Request) {
		if r == nil {
			return
		}
		if mediaContextBool(r, mediaTietaAuthenticatedKey) {
			r.Middleware.Next()
			return
		}
		if handler == nil {
			r.Middleware.Next()
			return
		}
		handler(r)
	}
}

// mediaMarkHostGatePassed marks that host Auth, Tenancy, and Permission all reached the handler gate.
func mediaMarkHostGatePassed(r *ghttp.Request) {
	if r == nil {
		return
	}
	r.SetCtxVar(mediaHostGatePassedKey, true)
	r.Middleware.Next()
}

// mediaAuthenticateTietaRequest authenticates the current request with media-owned Tieta tokens.
func mediaAuthenticateTietaRequest(r *ghttp.Request, tieta mediaTietaAuthenticator) error {
	if r == nil || tieta == nil {
		return bizerr.NewCode(mediasvc.CodeMediaAuthFailed, bizerr.P("message", mediaAuthMessageFailed))
	}
	tokens := mediaTietaTokenCandidates(r)
	if len(tokens) == 0 {
		return bizerr.NewCode(mediasvc.CodeMediaTietaTokenRequired)
	}
	var lastErr error
	for _, token := range tokens {
		user, err := tieta.AuthenticateTietaToken(r.Context(), token)
		if err == nil {
			mediaBindTietaContext(r, user)
			return nil
		}
		lastErr = err
	}
	if lastErr == nil {
		return bizerr.NewCode(mediasvc.CodeMediaAuthFailed, bizerr.P("message", mediaAuthMessageFailed))
	}
	return lastErr
}

// mediaBindTietaContext injects a plugin-visible platform context for Tieta-authenticated media calls.
func mediaBindTietaContext(r *ghttp.Request, user *mediasvc.TietaUser) {
	if r == nil || user == nil {
		return
	}
	current := contract.CurrentContext{
		UserID:   int(user.Id),
		Username: mediaTietaUsername(user),
		TenantID: 0,
	}
	ctx := contract.WithCurrentContext(r.Context(), current)
	ctx = context.WithValue(ctx, mediaTietaCurrentContextKey{}, contract.CurrentFromContext(ctx))
	r.SetCtx(ctx)
	r.SetCtxVar(mediaTietaAuthenticatedKey, true)
}

// mediaTietaUsername resolves a stable display name for the plugin-visible context.
func mediaTietaUsername(user *mediasvc.TietaUser) string {
	if user == nil {
		return ""
	}
	if username := strings.TrimSpace(user.Username); username != "" {
		return username
	}
	if realName := strings.TrimSpace(user.RealName); realName != "" {
		return realName
	}
	return strings.TrimSpace(user.Mobile)
}

// mediaTietaTokenCandidates returns unique request tokens accepted by media Tieta authentication.
func mediaTietaTokenCandidates(r *ghttp.Request) []string {
	if r == nil {
		return nil
	}
	candidates := []string{
		mediaNormalizeToken(r.GetHeader(mediaAuthorizationHeader)),
		mediaNormalizeToken(r.Get(mediaTokenField).String()),
	}
	tokens := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		tokens = append(tokens, candidate)
	}
	return tokens
}

// mediaNormalizeToken trims an optional Bearer prefix from a request token.
func mediaNormalizeToken(token string) string {
	normalized := strings.TrimSpace(token)
	if len(normalized) >= len(mediaBearerPrefix) &&
		strings.EqualFold(normalized[:len(mediaBearerPrefix)], mediaBearerPrefix) {
		normalized = strings.TrimSpace(normalized[len(mediaBearerPrefix):])
	}
	return normalized
}

// mediaResetHostGateState clears per-request markers before one LinaPro auth attempt.
func mediaResetHostGateState(r *ghttp.Request) {
	r.SetCtxVar(mediaHostGatePassedKey, false)
	r.SetCtxVar(mediaTietaAuthenticatedKey, false)
}

// mediaClearHostFailure removes buffered host auth failure output before Tieta fallback.
func mediaClearHostFailure(r *ghttp.Request) {
	if r == nil || r.Response == nil || r.Response.BytesWritten() > 0 {
		return
	}
	r.SetError(nil)
	r.Response.ClearBuffer()
	r.Response.Status = 0
}

// mediaWriteAuthFailure binds a structured unauthorized error for unified response handling.
func mediaWriteAuthFailure(r *ghttp.Request, cause error) {
	if r == nil {
		return
	}
	message := mediaAuthMessageFailed
	if cause != nil {
		message = cause.Error()
	}
	err := bizerr.WrapCode(cause, mediasvc.CodeMediaAuthFailed, bizerr.P("message", message))
	if err == nil {
		err = bizerr.NewCode(mediasvc.CodeMediaAuthFailed, bizerr.P("message", message))
	}
	r.SetError(err)
	r.Response.Status = http.StatusUnauthorized
}

// mediaContextBool reads one plugin-local boolean context marker.
func mediaContextBool(r *ghttp.Request, key mediaAuthContextKey) bool {
	return r != nil && r.GetCtxVar(key).Bool()
}
