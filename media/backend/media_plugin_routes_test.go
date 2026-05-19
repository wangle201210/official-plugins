// This file verifies media plugin route authentication boundaries.

package backend

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/pluginhost"
	"lina-core/pkg/pluginservice/bizctx"
	"lina-core/pkg/pluginservice/contract"
)

// mediaRouteHostServices publishes only the host services required by media route registration.
type mediaRouteHostServices struct {
	pluginhost.HostServices
	bizCtx contract.BizCtxService
	cache  contract.CacheService
}

// BizCtx returns the test host business-context adapter.
func (s *mediaRouteHostServices) BizCtx() contract.BizCtxService {
	return s.bizCtx
}

// Cache returns the test host cache adapter.
func (s *mediaRouteHostServices) Cache() contract.CacheService {
	return s.cache
}

// mediaRouteCache stores route-memory values for route boundary tests.
type mediaRouteCache struct {
	values map[string]string
}

// newMediaRouteCache creates a route cache test double.
func newMediaRouteCache() *mediaRouteCache {
	return &mediaRouteCache{values: make(map[string]string)}
}

// Get returns one stored cache item.
func (c *mediaRouteCache) Get(
	_ context.Context,
	namespace string,
	key string,
) (*contract.CacheItem, bool, error) {
	value, ok := c.values[namespace+"\x00"+key]
	if !ok {
		return nil, false, nil
	}
	return &contract.CacheItem{Key: key, ValueKind: contract.CacheValueKindString, Value: value}, true, nil
}

// Set stores one string cache item.
func (c *mediaRouteCache) Set(
	_ context.Context,
	namespace string,
	key string,
	value string,
	_ time.Duration,
) (*contract.CacheItem, error) {
	c.values[namespace+"\x00"+key] = value
	return &contract.CacheItem{Key: key, ValueKind: contract.CacheValueKindString, Value: value}, nil
}

// Delete removes one cache item.
func (c *mediaRouteCache) Delete(_ context.Context, namespace string, key string) error {
	delete(c.values, namespace+"\x00"+key)
	return nil
}

// Incr implements the unused integer cache operation for interface compliance.
func (c *mediaRouteCache) Incr(
	_ context.Context,
	namespace string,
	key string,
	delta int64,
	_ time.Duration,
) (*contract.CacheItem, error) {
	cacheKey := namespace + "\x00" + key
	current := int64(0)
	if value := strings.TrimSpace(c.values[cacheKey]); value != "" {
		if _, err := fmt.Sscan(value, &current); err != nil {
			return nil, err
		}
	}
	current += delta
	c.values[cacheKey] = fmt.Sprintf("%d", current)
	return &contract.CacheItem{Key: key, ValueKind: contract.CacheValueKindInt, IntValue: current}, nil
}

// Expire reports success for the unused expiration operation.
func (c *mediaRouteCache) Expire(
	context.Context,
	string,
	string,
	time.Duration,
) (bool, *time.Time, error) {
	return true, nil, nil
}

// mediaRouteHTTPResponse stores one HTTP response snapshot.
type mediaRouteHTTPResponse struct {
	status int
	body   string
}

// TestMediaPluginRoutesPreferHostAuth verifies media routes try the LinaPro host chain first.
func TestMediaPluginRoutesPreferHostAuth(t *testing.T) {
	setMediaRouteTietaMock(t, true)

	var (
		authCalls       atomic.Int32
		tenancyCalls    atomic.Int32
		permissionCalls atomic.Int32
	)
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			tenancyCalls.Add(1)
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			permissionCalls.Add(1)
			r.Middleware.Next()
		},
	)

	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/route/get",
		`{"deviceCode":"34020000001320000001","channelCode":"34020000001320000002"}`,
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected host-authenticated media route to pass, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 1 || tenancyCalls.Load() != 1 || permissionCalls.Load() != 1 {
		t.Fatalf(
			"expected host chain once, got auth=%d tenancy=%d permission=%d",
			authCalls.Load(),
			tenancyCalls.Load(),
			permissionCalls.Load(),
		)
	}
}

// TestMediaPluginRoutesFallbackToTietaWhenHostAuthFails verifies Tieta fallback stays inside media routes.
func TestMediaPluginRoutesFallbackToTietaWhenHostAuthFails(t *testing.T) {
	setMediaRouteTietaMock(t, true)

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
			r.Response.Status = http.StatusUnauthorized
			r.Response.Write("unauthorized")
		},
		func(r *ghttp.Request) {
			t.Fatalf("tenancy middleware should be skipped after Tieta fallback")
		},
		func(r *ghttp.Request) {
			t.Fatalf("permission middleware should be skipped after Tieta fallback")
		},
	)

	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodGet,
		baseURL+"/api/v1/media/strategies?pageNum=1&pageSize=1",
		"",
		map[string]string{"Authorization": "Bearer anything"},
	)
	if response.status == http.StatusUnauthorized {
		t.Fatalf("expected Tieta fallback to pass media management auth, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 1 {
		t.Fatalf("expected host Auth to be tried once before Tieta fallback, got %d calls", authCalls.Load())
	}
}

// TestMediaPluginRoutesFallbackToTietaAfterHostPermissionFailure verifies fallback after LinaPro permission denial.
func TestMediaPluginRoutesFallbackToTietaAfterHostPermissionFailure(t *testing.T) {
	setMediaRouteTietaMock(t, true)

	var (
		authCalls       atomic.Int32
		tenancyCalls    atomic.Int32
		permissionCalls atomic.Int32
	)
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			tenancyCalls.Add(1)
			r.Middleware.Next()
		},
		func(r *ghttp.Request) {
			permissionCalls.Add(1)
			r.Response.Status = http.StatusForbidden
			r.Response.Write("forbidden")
		},
	)

	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/route/get",
		`{"deviceCode":"34020000001320000001","channelCode":"34020000001320000002"}`,
		map[string]string{"Authorization": "Bearer fallback-token"},
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected Tieta fallback after permission failure, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 1 || tenancyCalls.Load() != 1 || permissionCalls.Load() != 1 {
		t.Fatalf(
			"expected host chain to fail once before fallback, got auth=%d tenancy=%d permission=%d",
			authCalls.Load(),
			tenancyCalls.Load(),
			permissionCalls.Load(),
		)
	}
}

// TestMediaPluginRoutesRejectWhenBothAuthPathsFail verifies media routes fail only after both auth paths fail.
func TestMediaPluginRoutesRejectWhenBothAuthPathsFail(t *testing.T) {
	setMediaRouteTietaMock(t, true)

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
			r.Response.Status = http.StatusUnauthorized
			r.Response.Write("unauthorized")
		},
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	)

	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/route/get",
		`{"deviceCode":"34020000001320000001","channelCode":"34020000001320000002"}`,
	)
	if response.status != http.StatusUnauthorized {
		t.Fatalf("expected media route auth failure, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 1 {
		t.Fatalf("expected host Auth to be tried once, got %d calls", authCalls.Load())
	}
}

// setMediaRouteTietaMock installs a test config adapter with deterministic Tieta mock mode.
func setMediaRouteTietaMock(t *testing.T, enabled bool) {
	t.Helper()

	content := fmt.Sprintf(`
tieta:
  mock: %t
  baseUrl: "http://tieta.invalid"
  timeout: "3s"
`, enabled)
	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create config adapter: %v", err)
	}
	cfg := g.Cfg()
	previous := cfg.GetAdapter()
	cfg.SetAdapter(adapter)
	t.Cleanup(func() {
		cfg.SetAdapter(previous)
	})
}

// mediaRouteNoOpMiddleware continues the route middleware chain in tests.
func mediaRouteNoOpMiddleware(r *ghttp.Request) {
	r.Middleware.Next()
}

// startMediaRouteTestServer starts an ephemeral GoFrame server with media plugin routes.
func startMediaRouteTestServer(t *testing.T, middlewares pluginhost.RouteMiddlewares) (string, func()) {
	t.Helper()

	server := g.Server(fmt.Sprintf("media-route-test-%d", time.Now().UnixNano()))
	server.SetDumpRouterMap(false)
	server.SetPort(0)
	hostServices := &mediaRouteHostServices{
		bizCtx: bizctx.New(nil),
		cache:  newMediaRouteCache(),
	}
	server.Group("/", func(group *ghttp.RouterGroup) {
		registrar := pluginhost.NewHTTPRegistrar(
			server,
			group,
			pluginID,
			func(context.Context, string) bool { return true },
			middlewares,
			hostServices,
		)
		if err := registerRoutes(context.Background(), registrar); err != nil {
			t.Fatalf("register media routes: %v", err)
		}
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

// mediaRouteTestResponse writes controller results for route boundary tests.
func mediaRouteTestResponse(r *ghttp.Request) {
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
	response := r.GetHandlerResponse()
	if response != nil {
		r.Response.WriteJson(response)
	}
}

// doMediaRouteRequest sends one JSON request to the media route test server.
func doMediaRouteRequest(
	t *testing.T,
	method string,
	targetURL string,
	body string,
	headers ...map[string]string,
) mediaRouteHTTPResponse {
	t.Helper()

	requestBody := strings.NewReader(body)
	request, err := http.NewRequest(method, targetURL, requestBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for _, headerSet := range headers {
		for key, value := range headerSet {
			request.Header.Set(key, value)
		}
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
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return mediaRouteHTTPResponse{status: response.StatusCode, body: string(responseBody)}
}
