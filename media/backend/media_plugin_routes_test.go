// This file verifies media plugin route authentication boundaries.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/pluginhost"
	"lina-core/pkg/pluginservice/bizctx"
	pluginconfig "lina-core/pkg/pluginservice/config"
	"lina-core/pkg/pluginservice/contract"
	mediaopenv1 "lina-plugin-media/backend/api/mediaopen/v1"
)

// mediaRouteHostServices publishes only the host services required by media route registration.
type mediaRouteHostServices struct {
	pluginhost.HostServices
	bizCtx contract.BizCtxService
	cache  contract.CacheService
	config contract.ConfigService
}

// BizCtx returns the test host business-context adapter.
func (s *mediaRouteHostServices) BizCtx() contract.BizCtxService {
	return s.bizCtx
}

// Cache returns the test host cache adapter.
func (s *mediaRouteHostServices) Cache() contract.CacheService {
	return s.cache
}

// Config returns the test host configuration adapter.
func (s *mediaRouteHostServices) Config() contract.ConfigService {
	return s.config
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

// TestMediaOpenRoutesUseInnerAPIAuth verifies mediaopen routes use the HotGo-compatible inner API key gate.
func TestMediaOpenRoutesUseInnerAPIAuth(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})

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
		},
		func(r *ghttp.Request) {
			tenancyCalls.Add(1)
		},
		func(r *ghttp.Request) {
			permissionCalls.Add(1)
		},
	)

	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/route/get",
		`{"deviceCode":"34020000001320000001","channelCode":"34020000001320000002"}`,
		map[string]string{mediaInnerAPIKeyHeader: "media"},
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected inner-api-authenticated mediaopen route to pass, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 0 || tenancyCalls.Load() != 0 || permissionCalls.Load() != 0 {
		t.Fatalf(
			"expected mediaopen route to skip host chain, got auth=%d tenancy=%d permission=%d",
			authCalls.Load(),
			tenancyCalls.Load(),
			permissionCalls.Load(),
		)
	}
}

// TestMediaOpenRoutesExposeOnlyHotGoTokenStrategyEndpoint verifies obsolete token strategy routes stay unpublished.
func TestMediaOpenRoutesExposeOnlyHotGoTokenStrategyEndpoint(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
		},
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	)

	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()

	removedResponse := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/media/strategy-authorizations",
		`{"token":"token-value","deviceId":"34020000001320000001"}`,
		map[string]string{mediaInnerAPIKeyHeader: "media"},
	)
	if removedResponse.status != http.StatusNotFound {
		t.Fatalf("expected obsolete media strategy authorization route to be unpublished, got status=%d body=%s", removedResponse.status, removedResponse.body)
	}

	compatResponse := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/strategy/userDeviceStrategyByToken",
		`{}`,
		map[string]string{mediaInnerAPIKeyHeader: "media"},
	)
	if compatResponse.status == http.StatusNotFound {
		t.Fatalf("expected HotGo-compatible token strategy route to remain published, got body=%s", compatResponse.body)
	}
	if compatResponse.status == http.StatusUnauthorized || compatResponse.status == http.StatusForbidden {
		t.Fatalf("expected HotGo-compatible token strategy route to pass inner API auth, got status=%d body=%s", compatResponse.status, compatResponse.body)
	}
	if authCalls.Load() != 0 {
		t.Fatalf("expected mediaopen strategy routes to avoid host Auth, got %d calls", authCalls.Load())
	}
}

// TestMediaOpenRoutesRejectMissingInnerAPIKey verifies configured inner API auth rejects missing keys.
func TestMediaOpenRoutesRejectMissingInnerAPIKey(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
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
		t.Fatalf("expected missing inner API key to fail, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 0 {
		t.Fatalf("expected mediaopen route to avoid host Auth on inner API failure, got %d calls", authCalls.Load())
	}
}

// TestMediaOpenRoutesRejectInvalidInnerAPIKey verifies configured inner API auth rejects wrong keys.
func TestMediaOpenRoutesRejectInvalidInnerAPIKey(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
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
		map[string]string{mediaInnerAPIKeyHeader: "wrong"},
	)
	if response.status != http.StatusUnauthorized {
		t.Fatalf("expected invalid inner API key to fail, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 0 {
		t.Fatalf("expected mediaopen route to avoid host Auth on inner API failure, got %d calls", authCalls.Load())
	}
}

// TestMediaOpenRoutesUseDefaultInnerAPIKey mirrors HotGo's default inner API key behavior.
func TestMediaOpenRoutesUseDefaultInnerAPIKey(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true})

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
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
		map[string]string{mediaInnerAPIKeyHeader: mediaInnerAPIKeyDefault},
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected default inner API key to pass, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 0 {
		t.Fatalf("expected mediaopen route to avoid host Auth when default inner API key is used, got %d calls", authCalls.Load())
	}
}

// TestMediaOpenRoutesAllowWhenInnerAPIKeyExplicitlyBlank preserves HotGo compatibility for disabled key checks.
func TestMediaOpenRoutesAllowWhenInnerAPIKeyExplicitlyBlank(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, includeInnerAPIKey: true})

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
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
	if response.status != http.StatusOK {
		t.Fatalf("expected explicitly blank inner API key to preserve compatibility, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 0 {
		t.Fatalf("expected mediaopen route to avoid host Auth when inner API key check is disabled, got %d calls", authCalls.Load())
	}
}

// TestMediaOpenRoutesTenantWhiteIPsByTokenReturnsTenantScopedIPs verifies the public whitelist route returns tenant-scoped IPs.
func TestMediaOpenRoutesTenantWhiteIPsByTokenReturnsTenantScopedIPs(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})
	setupMediaRouteSQLite(t)

	var authCalls atomic.Int32
	middlewares := pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		func(r *ghttp.Request) {
			authCalls.Add(1)
		},
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	)

	baseURL, shutdown := startMediaRouteTestServer(t, middlewares)
	defer shutdown()

	if _, err := g.DB().Exec(
		context.Background(),
		`INSERT INTO media_tenant_white (tenant_id, ip, enable) VALUES (?, ?, ?)`,
		"1",
		"192.0.2.10",
		1,
	); err != nil {
		t.Fatalf("insert tenant whitelist fixture: %v", err)
	}

	response := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/tenant-whites/ips",
		`{"token":"1"}`,
		map[string]string{mediaInnerAPIKeyHeader: "media"},
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected whitelist IP route to pass, got status=%d body=%s", response.status, response.body)
	}
	var out mediaopenv1.TenantWhiteIPsByTokenRes
	if err := json.Unmarshal([]byte(response.body), &out); err != nil {
		t.Fatalf("expected tenant-scoped JSON response, got body=%s err=%v", response.body, err)
	}
	if out.TenantId != "1" {
		t.Fatalf("expected mock token tenant ID 1, got %q", out.TenantId)
	}
	if len(out.Ips) != 1 || out.Ips[0] != "192.0.2.10" {
		t.Fatalf("expected enabled tenant whitelist IPs, got %#v", out.Ips)
	}
	if authCalls.Load() != 0 {
		t.Fatalf("expected mediaopen whitelist route to avoid host Auth, got %d calls", authCalls.Load())
	}
}

// TestMediaOpenRoutesGetStreamAliasByAliasReturnsConfig verifies the public alias lookup returns one config by alias.
func TestMediaOpenRoutesGetStreamAliasByAliasReturnsConfig(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})
	setupMediaRouteSQLite(t)

	baseURL, shutdown := startMediaRouteTestServer(t, pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	))
	defer shutdown()

	if _, err := g.DB().Exec(
		context.Background(),
		`INSERT INTO media_stream_alias (alias, auto_remove, stream_path, device_id, channel_id) VALUES (?, ?, ?, ?, ?)`,
		"retail-east",
		1,
		"live/retail-east",
		"34020000001320000001",
		"34020000001320000002",
	); err != nil {
		t.Fatalf("insert stream alias fixture: %v", err)
	}

	response := doMediaRouteRequest(
		t,
		http.MethodGet,
		baseURL+"/api/v1/stream-aliases/by-alias?alias=retail-east",
		"",
		map[string]string{mediaInnerAPIKeyHeader: "media"},
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected stream alias route to pass, got status=%d body=%s", response.status, response.body)
	}
	var out mediaopenv1.GetStreamAliasByAliasRes
	if err := json.Unmarshal([]byte(response.body), &out); err != nil {
		t.Fatalf("expected alias JSON response, got body=%s err=%v", response.body, err)
	}
	if out.Alias != "retail-east" || out.StreamPath != "live/retail-east" || out.AutoRemove != 1 {
		t.Fatalf("unexpected alias config: %+v", out)
	}
	if out.DeviceId != "34020000001320000001" || out.ChannelId != "34020000001320000002" {
		t.Fatalf("expected device channel config, got %+v", out)
	}
}

// TestMediaOpenRoutesListAllNodesReturnsUnpagedData verifies the public node route returns every node.
func TestMediaOpenRoutesListAllNodesReturnsUnpagedData(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})
	setupMediaRouteSQLite(t)

	baseURL, shutdown := startMediaRouteTestServer(t, pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	))
	defer shutdown()

	if _, err := g.DB().Exec(
		context.Background(),
		`INSERT INTO media_node (node_num, name, qn_url, basic_url, dn_url, creator_id, updater_id) VALUES
			(2, '节点二', 'https://qn2.example.com', 'https://basic2.example.com', 'https://dn2.example.com', 1, 2),
			(1, '节点一', 'https://qn1.example.com', 'https://basic1.example.com', 'https://dn1.example.com', 1, 2)`,
	); err != nil {
		t.Fatalf("insert node fixtures: %v", err)
	}

	response := doMediaRouteRequest(
		t,
		http.MethodGet,
		baseURL+"/api/v1/nodes/all",
		"",
		map[string]string{mediaInnerAPIKeyHeader: "media"},
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected all-node route to pass, got status=%d body=%s", response.status, response.body)
	}
	var out mediaopenv1.ListAllNodesRes
	if err := json.Unmarshal([]byte(response.body), &out); err != nil {
		t.Fatalf("expected node JSON response, got body=%s err=%v", response.body, err)
	}
	if len(out.List) != 2 {
		t.Fatalf("expected unpaged full node list, got %+v", out.List)
	}
	if out.List[0].NodeNum != 1 || out.List[1].NodeNum != 2 {
		t.Fatalf("expected nodes ordered by node number, got %+v", out.List)
	}
}

// TestMediaOpenRoutesUserDeviceStrategyResponseOmitsAccessFlags verifies the public strategy response is narrowed.
func TestMediaOpenRoutesUserDeviceStrategyResponseOmitsAccessFlags(t *testing.T) {
	responseType := reflect.TypeOf(mediaopenv1.UserDeviceStrategyByTokenRes{})
	if _, ok := responseType.FieldByName("HasAccess"); ok {
		t.Fatal("expected UserDeviceStrategyByTokenRes to omit HasAccess")
	}
	if _, ok := responseType.FieldByName("StrategyId"); ok {
		t.Fatal("expected UserDeviceStrategyByTokenRes to omit top-level StrategyId")
	}
	if _, ok := responseType.FieldByName("Strategy"); !ok {
		t.Fatal("expected UserDeviceStrategyByTokenRes to keep Strategy")
	}
}

// TestMediaOpenRoutesTenantWhiteIPsByTokenRequiresToken verifies the public whitelist route requires token input.
func TestMediaOpenRoutesTenantWhiteIPsByTokenRequiresToken(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})

	baseURL, shutdown := startMediaRouteTestServer(t, pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	))
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodPost,
		baseURL+"/api/v1/tenant-whites/ips",
		`{}`,
		map[string]string{mediaInnerAPIKeyHeader: "media"},
	)
	if !strings.Contains(response.body, "token") {
		t.Fatalf("expected missing token response to mention token, got body=%s", response.body)
	}
}

// TestTenantWhiteIPsByTokenReqOnlyExposesRequiredToken verifies the public request DTO has one required user-token input.
func TestTenantWhiteIPsByTokenReqOnlyExposesRequiredToken(t *testing.T) {
	reqType := reflect.TypeOf(mediaopenv1.TenantWhiteIPsByTokenReq{})
	metaField := reqType.Field(0)
	if !strings.Contains(string(metaField.Tag), `method:"post"`) {
		t.Fatalf("expected TenantWhiteIPsByTokenReq to use POST, got tag=%s", metaField.Tag)
	}
	if !strings.Contains(string(metaField.Tag), `access:"public"`) {
		t.Fatalf("expected TenantWhiteIPsByTokenReq to declare public access, got tag=%s", metaField.Tag)
	}
	tokenField, ok := reqType.FieldByName("Token")
	if !ok {
		t.Fatalf("expected TenantWhiteIPsByTokenReq to expose Token")
	}
	if !strings.Contains(string(tokenField.Tag), `v:"required`) {
		t.Fatalf("expected TenantWhiteIPsByTokenReq.Token to be required, got tag=%s", tokenField.Tag)
	}
	if strings.Contains(string(tokenField.Tag), "Bearer") {
		t.Fatalf("expected TenantWhiteIPsByTokenReq.Token docs to avoid Bearer prefix, got tag=%s", tokenField.Tag)
	}
	if _, ok := reqType.FieldByName("Authorization"); ok {
		t.Fatalf("expected TenantWhiteIPsByTokenReq to avoid Authorization")
	}
}

// TestMediaOpenRequestDTOsDeclarePublicAccess verifies all mediaopen DTOs opt
// out of the host-level JWT Bearer requirement in generated apidocs.
func TestMediaOpenRequestDTOsDeclarePublicAccess(t *testing.T) {
	requests := []interface{}{
		mediaopenv1.SetRouteDataReq{},
		mediaopenv1.GetRouteDataReq{},
		mediaopenv1.DelRouteDataReq{},
		mediaopenv1.UserDeviceStrategyByTokenReq{},
		mediaopenv1.TenantWhiteIPsByTokenReq{},
		mediaopenv1.GetStreamAliasByAliasReq{},
		mediaopenv1.ListAllNodesReq{},
	}
	for _, request := range requests {
		reqType := reflect.TypeOf(request)
		metaField := reqType.Field(0)
		if !strings.Contains(string(metaField.Tag), `access:"public"`) {
			t.Fatalf("expected %s to declare public access, got tag=%s", reqType.Name(), metaField.Tag)
		}
	}

	strategyType := reflect.TypeOf(mediaopenv1.UserDeviceStrategyByTokenReq{})
	tokenField, ok := strategyType.FieldByName("Token")
	if !ok {
		t.Fatalf("expected UserDeviceStrategyByTokenReq to expose Token")
	}
	if strings.Contains(string(tokenField.Tag), "Bearer") {
		t.Fatalf("expected UserDeviceStrategyByTokenReq.Token docs to avoid Bearer prefix, got tag=%s", tokenField.Tag)
	}
}

// TestMediaPluginOpenAPIDocumentOnlyContainsMediaRoutes verifies the plugin-owned
// documentation endpoint does not depend on lina-core's global API document.
func TestMediaPluginOpenAPIDocumentOnlyContainsMediaRoutes(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})

	baseURL, shutdown := startMediaRouteTestServer(t, pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	))
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodGet,
		baseURL+"/api/v1/media/openapi.json",
		"",
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected media OpenAPI route to return 200, got status=%d body=%s", response.status, response.body)
	}

	var document struct {
		OpenAPI    string `json:"openapi"`
		Components struct {
			SecuritySchemes map[string]struct {
				Type string `json:"type"`
				Name string `json:"name"`
				In   string `json:"in"`
			} `json:"securitySchemes"`
		} `json:"components"`
		Security []map[string][]string `json:"security"`
		Paths    map[string]map[string]struct {
			Summary  string                `json:"summary"`
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	if err := json.Unmarshal([]byte(response.body), &document); err != nil {
		t.Fatalf("decode media OpenAPI document: %v\nbody=%s", err, response.body)
	}
	if document.OpenAPI == "" {
		t.Fatal("expected OpenAPI version to be present")
	}
	if _, ok := document.Components.SecuritySchemes["BearerAuth"]; !ok {
		t.Fatalf("expected media OpenAPI document to publish BearerAuth security scheme")
	}
	innerScheme, ok := document.Components.SecuritySchemes["InnerApiKeyAuth"]
	if !ok {
		t.Fatalf("expected media OpenAPI document to publish InnerApiKeyAuth security scheme")
	}
	if innerScheme.Type != "apiKey" || innerScheme.Name != mediaInnerAPIKeyHeader || innerScheme.In != "header" {
		t.Fatalf("expected InnerApiKeyAuth header scheme, got %+v", innerScheme)
	}
	if _, ok := document.Paths["/api/v1/media/strategies"]; !ok {
		t.Fatalf("expected media management routes in media OpenAPI document")
	}
	if _, ok := document.Paths["/api/v1/strategy/userDeviceStrategyByToken"]; !ok {
		t.Fatalf("expected mediaopen routes in media OpenAPI document")
	}
	if _, ok := document.Paths["/api/v1/water/preview"]; ok {
		t.Fatalf("expected media OpenAPI document to exclude water routes")
	}
	if _, ok := document.Paths["/api/v1/user"]; ok {
		t.Fatalf("expected media OpenAPI document to exclude core routes")
	}

	strategyOperation := document.Paths["/api/v1/media/strategies"]["get"]
	if len(strategyOperation.Security) != 0 {
		t.Fatalf("expected media management route to inherit document BearerAuth, got %+v", strategyOperation.Security)
	}
	mediaOpenOperation := document.Paths["/api/v1/strategy/userDeviceStrategyByToken"]["post"]
	if len(mediaOpenOperation.Security) != 1 {
		t.Fatalf("expected mediaopen route to declare one security requirement, got %+v", mediaOpenOperation.Security)
	}
	if _, ok := mediaOpenOperation.Security[0]["InnerApiKeyAuth"]; !ok {
		t.Fatalf("expected mediaopen route to require InnerApiKeyAuth, got %+v", mediaOpenOperation.Security)
	}
}

// TestMediaPluginAPIDocsPageLoadsMediaDocument verifies the plugin-owned
// Stoplight page points to the plugin-owned OpenAPI JSON document.
func TestMediaPluginAPIDocsPageLoadsMediaDocument(t *testing.T) {
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: true, innerAPIKey: "media", includeInnerAPIKey: true})

	baseURL, shutdown := startMediaRouteTestServer(t, pluginhost.NewRouteMiddlewares(
		mediaRouteNoOpMiddleware,
		mediaRouteTestResponse,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
		mediaRouteNoOpMiddleware,
	))
	defer shutdown()

	response := doMediaRouteRequest(
		t,
		http.MethodGet,
		baseURL+"/api/v1/media/apidocs.html?token=jwt-token&innerApiKey=media",
		"",
	)
	if response.status != http.StatusOK {
		t.Fatalf("expected media apidocs page to return 200, got status=%d body=%s", response.status, response.body)
	}

	requiredFragments := []string{
		`document.createElement('elements-api')`,
		`/api/v1/media/openapi.json`,
		`/stoplight/web-components.min.js`,
		`TryIt_securitySchemeValues`,
		`BearerAuth`,
		`InnerApiKeyAuth`,
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(response.body, fragment) {
			t.Fatalf("expected media apidocs page to contain %q", fragment)
		}
	}
	if strings.Contains(response.body, `'/api.json?`) || strings.Contains(response.body, `"/api.json?`) {
		t.Fatalf("expected media apidocs page to avoid host-wide /api.json")
	}
}

// TestMediaManagementRoutesPreferHostAuth verifies management routes try the LinaPro host chain first.
func TestMediaManagementRoutesPreferHostAuth(t *testing.T) {
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
		http.MethodGet,
		baseURL+"/api/v1/media/strategies?pageNum=1&pageSize=1",
		"",
	)
	if response.status == http.StatusUnauthorized || response.status == http.StatusForbidden {
		t.Fatalf("expected host-authenticated media management route to reach handler, got status=%d body=%s", response.status, response.body)
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
		http.MethodGet,
		baseURL+"/api/v1/media/strategies?pageNum=1&pageSize=1",
		"",
		map[string]string{"Authorization": "Bearer fallback-token"},
	)
	if response.status == http.StatusUnauthorized || response.status == http.StatusForbidden {
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
		http.MethodGet,
		baseURL+"/api/v1/media/strategies?pageNum=1&pageSize=1",
		"",
	)
	if response.status != http.StatusUnauthorized {
		t.Fatalf("expected media route auth failure, got status=%d body=%s", response.status, response.body)
	}
	if authCalls.Load() != 1 {
		t.Fatalf("expected host Auth to be tried once, got %d calls", authCalls.Load())
	}
}

// mediaRouteTestConfig controls the request-time configuration used by route tests.
type mediaRouteTestConfig struct {
	tietaMock          bool
	innerAPIKey        string
	includeInnerAPIKey bool
}

// setMediaRouteTietaMock installs a test config adapter with deterministic Tieta mock mode.
func setMediaRouteTietaMock(t *testing.T, enabled bool) {
	t.Helper()
	setMediaRouteConfig(t, mediaRouteTestConfig{tietaMock: enabled})
}

// setMediaRouteConfig installs a test config adapter with media route options.
func setMediaRouteConfig(t *testing.T, config mediaRouteTestConfig) {
	t.Helper()

	content := fmt.Sprintf(`
tieta:
  mock: %t
  baseUrl: "http://tieta.invalid"
  timeout: "3s"
`, config.tietaMock)
	if config.includeInnerAPIKey {
		content += fmt.Sprintf(`
innerapi:
  apiKey: %q
`, config.innerAPIKey)
	}
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
		config: pluginconfig.New(),
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

// setupMediaRouteSQLite installs a minimal media schema for route-level controller tests.
func setupMediaRouteSQLite(t *testing.T) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	dbPath := t.TempDir() + "/media-route.db"
	if err := gdb.SetConfig(gdb.Config{
		"default": {
			{Link: "sqlite::@file(" + dbPath + ")"},
		},
	}); err != nil {
		t.Fatalf("set sqlite config: %v", err)
	}
	t.Cleanup(func() {
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Fatalf("restore db config: %v", err)
		}
	})

	statements := []string{
		`CREATE TABLE media_strategy (id INTEGER PRIMARY KEY AUTOINCREMENT)`,
		`CREATE TABLE media_strategy_device (device_id TEXT PRIMARY KEY, strategy_id INTEGER NOT NULL)`,
		`CREATE TABLE media_strategy_tenant (tenant_id TEXT PRIMARY KEY, strategy_id INTEGER NOT NULL)`,
		`CREATE TABLE media_strategy_device_tenant (tenant_id TEXT NOT NULL, device_id TEXT NOT NULL, strategy_id INTEGER NOT NULL, PRIMARY KEY (tenant_id, device_id))`,
		`CREATE TABLE media_device_node (device_id TEXT NOT NULL, channel_id TEXT NOT NULL, node_num INTEGER NOT NULL)`,
		`CREATE TABLE media_node (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			node_num INTEGER NOT NULL,
			name TEXT NOT NULL,
			qn_url TEXT NOT NULL,
			basic_url TEXT NOT NULL,
			dn_url TEXT NOT NULL,
			creator_id INTEGER,
			create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updater_id INTEGER,
			update_time TEXT
		)`,
		`CREATE TABLE media_tenant_stream_config (tenant_id TEXT PRIMARY KEY)`,
		`CREATE TABLE media_tenant_white (tenant_id TEXT NOT NULL, ip TEXT NOT NULL, enable INTEGER NOT NULL, PRIMARY KEY (tenant_id, ip))`,
		`CREATE TABLE media_stream_alias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alias TEXT NOT NULL,
			auto_remove INTEGER NOT NULL DEFAULT 0,
			stream_path TEXT NOT NULL DEFAULT '',
			device_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, statement := range statements {
		if _, err := g.DB().Exec(context.Background(), statement); err != nil {
			t.Fatalf("exec sqlite schema: %v", err)
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
