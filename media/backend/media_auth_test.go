// This file verifies the media plugin dual-channel authentication middleware.

package backend

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/pluginservice/contract"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// mediaAuthTestReq defines one media route with declarative permission metadata.
type mediaAuthTestReq struct {
	g.Meta `path:"/media/auth-test" method:"get" permission:"media:management:query"`
}

// mediaAuthTestRes returns the identity injected by the media auth middleware.
type mediaAuthTestRes struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
}

// mediaAuthBodyTokenReq defines one HotGo-compatible route that reads token from request body.
type mediaAuthBodyTokenReq struct {
	g.Meta `path:"/route/auth-test" method:"post"`
	Token  string `json:"token"`
}

// mediaAuthBodyTokenRes returns the identity injected by body-token fallback auth.
type mediaAuthBodyTokenRes struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
}

// mediaAuthTestHandler handles auth middleware tests.
type mediaAuthTestHandler struct{}

// MediaAuthTest returns the current plugin-visible identity.
func (h *mediaAuthTestHandler) MediaAuthTest(ctx context.Context, req *mediaAuthTestReq) (*mediaAuthTestRes, error) {
	_ = req
	current := contract.CurrentFromContext(ctx)
	return &mediaAuthTestRes{UserID: current.UserID, Username: current.Username}, nil
}

// MediaAuthBodyToken returns the current plugin-visible identity for body-token tests.
func (h *mediaAuthTestHandler) MediaAuthBodyToken(
	ctx context.Context,
	req *mediaAuthBodyTokenReq,
) (*mediaAuthBodyTokenRes, error) {
	_ = req
	current := contract.CurrentFromContext(ctx)
	return &mediaAuthBodyTokenRes{UserID: current.UserID, Username: current.Username}, nil
}

// TestMediaDualAuthAllowsLinaProToken verifies LinaPro token and permissions are tried first.
func TestMediaDualAuthAllowsLinaProToken(t *testing.T) {
	authSvc := &fakeMediaAuthService{
		identities: map[string]*contract.AuthenticatedIdentity{
			"lina-ok": {
				UserID:      7,
				Username:    "lina-user",
				Permissions: []string{"media:management:query"},
			},
		},
	}
	tietaSvc := &fakeMediaTietaAuthService{users: map[string]*mediasvc.TietaUser{
		"lina-ok": {Id: 13, Username: "tieta-user"},
	}}
	baseURL, shutdown := startMediaAuthTestServer(t, authSvc, tietaSvc)
	defer shutdown()

	response := doMediaAuthRequest(t, http.MethodGet, baseURL+"/media/auth-test", "Bearer lina-ok", "")
	if response.status != http.StatusOK {
		t.Fatalf("expected LinaPro auth to pass, got status=%d body=%s", response.status, response.body)
	}
	if !strings.Contains(response.body, `"userId":7`) || !strings.Contains(response.body, `"username":"lina-user"`) {
		t.Fatalf("expected LinaPro identity in response, got %s", response.body)
	}
	if tietaSvc.calls != 0 {
		t.Fatalf("expected Tieta fallback not to be called, got %d calls", tietaSvc.calls)
	}
}

// TestMediaDualAuthFallsBackToTietaToken verifies invalid LinaPro auth can still pass through Tieta.
func TestMediaDualAuthFallsBackToTietaToken(t *testing.T) {
	authSvc := &fakeMediaAuthService{}
	tietaSvc := &fakeMediaTietaAuthService{users: map[string]*mediasvc.TietaUser{
		"tieta-ok": {Id: 13, Username: "tieta-user"},
	}}
	baseURL, shutdown := startMediaAuthTestServer(t, authSvc, tietaSvc)
	defer shutdown()

	response := doMediaAuthRequest(t, http.MethodGet, baseURL+"/media/auth-test", "Bearer tieta-ok", "")
	if response.status != http.StatusOK {
		t.Fatalf("expected Tieta fallback to pass, got status=%d body=%s", response.status, response.body)
	}
	if !strings.Contains(response.body, `"userId":13`) || !strings.Contains(response.body, `"username":"tieta-user"`) {
		t.Fatalf("expected Tieta identity in response, got %s", response.body)
	}
}

// TestMediaDualAuthFallsBackWhenLinaProPermissionDenied verifies permission denial still tries Tieta.
func TestMediaDualAuthFallsBackWhenLinaProPermissionDenied(t *testing.T) {
	authSvc := &fakeMediaAuthService{identities: map[string]*contract.AuthenticatedIdentity{
		"shared-token": {UserID: 7, Username: "lina-user", Permissions: []string{"other:permission"}},
	}}
	tietaSvc := &fakeMediaTietaAuthService{users: map[string]*mediasvc.TietaUser{
		"shared-token": {Id: 13, Username: "tieta-user"},
	}}
	baseURL, shutdown := startMediaAuthTestServer(t, authSvc, tietaSvc)
	defer shutdown()

	response := doMediaAuthRequest(t, http.MethodGet, baseURL+"/media/auth-test", "Bearer shared-token", "")
	if response.status != http.StatusOK {
		t.Fatalf("expected permission failure to fall back to Tieta, got status=%d body=%s", response.status, response.body)
	}
	if !strings.Contains(response.body, `"userId":13`) {
		t.Fatalf("expected Tieta identity after permission fallback, got %s", response.body)
	}
}

// TestMediaDualAuthRejectsWhenBothFail verifies rejection happens only after both channels fail.
func TestMediaDualAuthRejectsWhenBothFail(t *testing.T) {
	baseURL, shutdown := startMediaAuthTestServer(t, &fakeMediaAuthService{}, &fakeMediaTietaAuthService{})
	defer shutdown()

	response := doMediaAuthRequest(t, http.MethodGet, baseURL+"/media/auth-test", "Bearer invalid", "")
	if response.status != http.StatusUnauthorized {
		t.Fatalf("expected both auth failures to reject with 401, got status=%d body=%s", response.status, response.body)
	}
}

// TestMediaDualAuthReadsTietaBodyToken verifies HotGo-compatible body token fallback.
func TestMediaDualAuthReadsTietaBodyToken(t *testing.T) {
	authSvc := &fakeMediaAuthService{}
	tietaSvc := &fakeMediaTietaAuthService{users: map[string]*mediasvc.TietaUser{
		"body-token": {Id: 17, Username: "body-user"},
	}}
	baseURL, shutdown := startMediaAuthTestServer(t, authSvc, tietaSvc)
	defer shutdown()

	response := doMediaAuthRequest(t, http.MethodPost, baseURL+"/route/auth-test", "", `{"token":"body-token"}`)
	if response.status != http.StatusOK {
		t.Fatalf("expected Tieta body token to pass, got status=%d body=%s", response.status, response.body)
	}
	if !strings.Contains(response.body, `"userId":17`) || !strings.Contains(response.body, `"username":"body-user"`) {
		t.Fatalf("expected body-token identity in response, got %s", response.body)
	}
}

// startMediaAuthTestServer starts one ephemeral GoFrame server for auth middleware tests.
func startMediaAuthTestServer(
	t *testing.T,
	authSvc contract.AuthService,
	tietaSvc mediaTietaAuthenticator,
) (string, func()) {
	t.Helper()

	server := g.Server(fmt.Sprintf("media-auth-test-%d", time.Now().UnixNano()))
	server.SetDumpRouterMap(false)
	server.SetPort(0)
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(mediaAuthTestResponse, mediaDualAuthMiddleware(authSvc, tietaSvc))
		group.Bind(&mediaAuthTestHandler{})
	})
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	return fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort()), func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown server: %v", err)
		}
	}
}

// mediaAuthTestResponse writes a compact JSON envelope for middleware tests.
func mediaAuthTestResponse(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}
	if err := r.GetError(); err != nil {
		status := r.Response.Status
		if status == 0 {
			status = http.StatusInternalServerError
		}
		r.Response.Status = status
		r.Response.WriteJson(g.Map{"message": err.Error()})
		return
	}
	r.Response.WriteJson(r.GetHandlerResponse())
}

// mediaAuthHTTPResponse stores one HTTP response snapshot.
type mediaAuthHTTPResponse struct {
	status int
	body   string
}

// doMediaAuthRequest sends one request to the auth test server.
func doMediaAuthRequest(
	t *testing.T,
	method string,
	targetURL string,
	authorization string,
	body string,
) mediaAuthHTTPResponse {
	t.Helper()

	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	request, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Fatalf("close response body: %v", closeErr)
		}
	}()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return mediaAuthHTTPResponse{status: response.StatusCode, body: string(bodyBytes)}
}

// fakeMediaAuthService implements contract.AuthService for middleware tests.
type fakeMediaAuthService struct {
	identities map[string]*contract.AuthenticatedIdentity
}

// SelectTenant is unused by media auth tests.
func (s *fakeMediaAuthService) SelectTenant(
	context.Context,
	contract.SelectTenantInput,
) (*contract.TenantTokenOutput, error) {
	return nil, fmt.Errorf("not implemented")
}

// SwitchTenant is unused by media auth tests.
func (s *fakeMediaAuthService) SwitchTenant(
	context.Context,
	contract.SwitchTenantInput,
) (*contract.TenantTokenOutput, error) {
	return nil, fmt.Errorf("not implemented")
}

// AuthenticateBearer returns the configured fake identity for token.
func (s *fakeMediaAuthService) AuthenticateBearer(
	_ context.Context,
	bearerToken string,
) (*contract.AuthenticatedIdentity, error) {
	token := normalizeMediaBearerToken(bearerToken)
	if s != nil && s.identities != nil {
		if identity := s.identities[token]; identity != nil {
			return identity, nil
		}
	}
	return nil, fmt.Errorf("invalid LinaPro token")
}

// fakeMediaTietaAuthService implements mediaTietaAuthenticator for middleware tests.
type fakeMediaTietaAuthService struct {
	users map[string]*mediasvc.TietaUser
	calls int
}

// AuthenticateTietaToken returns the configured fake Tieta user for token.
func (s *fakeMediaTietaAuthService) AuthenticateTietaToken(
	_ context.Context,
	token string,
) (*mediasvc.TietaUser, error) {
	if s != nil {
		s.calls++
		if user := s.users[normalizeMediaBearerToken(token)]; user != nil {
			return user, nil
		}
	}
	return nil, fmt.Errorf("invalid Tieta token")
}
