// This file implements media plugin authentication middleware. The middleware
// keeps LinaPro JWT/session/permission validation as the first path and falls
// back to Tieta token validation when LinaPro auth cannot authorize the route.

package backend

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginhost"
	plugincontract "lina-core/pkg/pluginservice/contract"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

const (
	// mediaPermissionMetaTag is the canonical g.Meta permission tag.
	mediaPermissionMetaTag = "permission"
	// mediaPermissionMetaTagAlias preserves compatibility with the legacy permission tag alias.
	mediaPermissionMetaTagAlias = "perms"
	// mediaAuthorizationHeader is the HTTP authorization header used by both LinaPro and Tieta tokens.
	mediaAuthorizationHeader = "Authorization"
	// mediaBearerPrefix is the standard Bearer token prefix.
	mediaBearerPrefix = "Bearer "
	// mediaTietaTokenParam is the request field accepted by HotGo-compatible Tieta endpoints.
	mediaTietaTokenParam = "token"
)

// mediaTietaAuthenticator validates Tieta tokens without coupling the
// middleware to the full media service contract.
type mediaTietaAuthenticator interface {
	// AuthenticateTietaToken validates one Tieta token and returns its user identity.
	AuthenticateTietaToken(ctx context.Context, token string) (*mediasvc.TietaUser, error)
}

// mediaDualAuthMiddleware returns a plugin route middleware that authenticates
// through LinaPro first, then through Tieta when LinaPro does not authorize.
func mediaDualAuthMiddleware(
	authSvc plugincontract.AuthService,
	tietaAuthenticator mediaTietaAuthenticator,
) pluginhost.RouteMiddleware {
	return func(r *ghttp.Request) {
		if r == nil {
			return
		}
		result := authenticateMediaRequest(r, authSvc, tietaAuthenticator)
		if result.allowed {
			r.Middleware.Next()
			return
		}
		r.SetError(result.err)
		r.Response.Status = result.status
	}
}

// mediaAuthResult captures middleware decision state.
type mediaAuthResult struct {
	allowed bool  // allowed reports whether the request may proceed.
	status  int   // status is the HTTP status to emit when rejected.
	err     error // err is the business error attached to the request.
}

// authenticateMediaRequest executes the LinaPro-then-Tieta auth flow.
func authenticateMediaRequest(
	r *ghttp.Request,
	authSvc plugincontract.AuthService,
	tietaAuthenticator mediaTietaAuthenticator,
) mediaAuthResult {
	if r == nil {
		return mediaAuthResult{status: http.StatusUnauthorized, err: bizerr.NewCode(mediasvc.CodeMediaAuthFailed)}
	}
	if current, err := authenticateMediaWithLinaPro(r, authSvc); err == nil {
		r.SetCtx(plugincontract.WithCurrentContext(r.Context(), current))
		return mediaAuthResult{allowed: true}
	}
	if current, err := authenticateMediaWithTieta(r, tietaAuthenticator); err == nil {
		r.SetCtx(plugincontract.WithCurrentContext(r.Context(), current))
		return mediaAuthResult{allowed: true}
	}
	return mediaAuthResult{status: http.StatusUnauthorized, err: bizerr.NewCode(mediasvc.CodeMediaAuthFailed)}
}

// authenticateMediaWithLinaPro validates LinaPro token and route permissions.
func authenticateMediaWithLinaPro(
	r *ghttp.Request,
	authSvc plugincontract.AuthService,
) (plugincontract.CurrentContext, error) {
	if authSvc == nil {
		return plugincontract.CurrentContext{}, bizerr.NewCode(mediasvc.CodeMediaAuthFailed)
	}
	token := mediaAuthorizationToken(r)
	identity, err := authSvc.AuthenticateBearer(r.Context(), token)
	if err != nil {
		return plugincontract.CurrentContext{}, err
	}
	requiredPermissions := mediaDeclaredPermissions(r)
	if !identity.HasPermissions(requiredPermissions) {
		return plugincontract.CurrentContext{}, bizerr.NewCode(
			mediasvc.CodeMediaPermissionDenied,
			bizerr.P("permissions", strings.Join(requiredPermissions, ", ")),
		)
	}
	return identity.CurrentContext(), nil
}

// authenticateMediaWithTieta validates Tieta token from Authorization or token parameter.
func authenticateMediaWithTieta(
	r *ghttp.Request,
	tietaAuthenticator mediaTietaAuthenticator,
) (plugincontract.CurrentContext, error) {
	if tietaAuthenticator == nil {
		return plugincontract.CurrentContext{}, bizerr.NewCode(mediasvc.CodeMediaAuthFailed)
	}
	var lastErr error
	for _, token := range mediaTietaTokenCandidates(r) {
		user, err := tietaAuthenticator.AuthenticateTietaToken(r.Context(), token)
		if err != nil {
			lastErr = err
			continue
		}
		if user == nil || user.Id <= 0 {
			lastErr = bizerr.NewCode(mediasvc.CodeMediaTietaTokenInvalid, bizerr.P("message", "用户信息为空"))
			continue
		}
		return plugincontract.CurrentContext{
			UserID:         int(user.Id),
			Username:       user.Username,
			TenantID:       0,
			PlatformBypass: true,
		}, nil
	}
	if lastErr != nil {
		return plugincontract.CurrentContext{}, lastErr
	}
	return plugincontract.CurrentContext{}, bizerr.NewCode(mediasvc.CodeMediaTietaTokenRequired)
}

// mediaAuthorizationToken extracts the Authorization header for LinaPro validation.
func mediaAuthorizationToken(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.GetHeader(mediaAuthorizationHeader))
}

// mediaTietaTokenCandidates returns unique Tieta token candidates in fallback order.
func mediaTietaTokenCandidates(r *ghttp.Request) []string {
	if r == nil {
		return nil
	}
	candidates := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	appendCandidate := func(token string) {
		normalized := normalizeMediaBearerToken(token)
		if normalized == "" {
			return
		}
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		candidates = append(candidates, normalized)
	}
	appendCandidate(r.GetHeader(mediaAuthorizationHeader))
	if bodyToken := r.Get(mediaTietaTokenParam); bodyToken != nil {
		appendCandidate(bodyToken.String())
	}
	return candidates
}

// mediaDeclaredPermissions reads and normalizes permission metadata from the matched handler.
func mediaDeclaredPermissions(r *ghttp.Request) []string {
	if r == nil {
		return nil
	}
	handler := r.GetServeHandler()
	if handler == nil {
		return nil
	}
	if permissions := normalizeMediaPermissionList(handler.GetMetaTag(mediaPermissionMetaTag)); len(permissions) > 0 {
		return permissions
	}
	return normalizeMediaPermissionList(handler.GetMetaTag(mediaPermissionMetaTagAlias))
}

// normalizeMediaPermissionList trims, deduplicates, and preserves permission order.
func normalizeMediaPermissionList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var (
		parts  = strings.Split(raw, ",")
		result = make([]string, 0, len(parts))
		seen   = make(map[string]struct{}, len(parts))
	)
	for _, part := range parts {
		permission := strings.TrimSpace(part)
		if permission == "" {
			continue
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		result = append(result, permission)
	}
	return result
}

// normalizeMediaBearerToken strips an optional Bearer prefix from a token value.
func normalizeMediaBearerToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(trimmed), strings.ToLower(mediaBearerPrefix)) {
		return strings.TrimSpace(trimmed[len(mediaBearerPrefix):])
	}
	return trimmed
}
