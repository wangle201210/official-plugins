// This file tests Tieta token-based media strategy authorization.

package media

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

var (
	mediaStrategySQLiteOnce           sync.Once
	mediaStrategySQLiteOriginalConfig gdb.Config
	mediaStrategySQLitePath           string
	mediaStrategySQLiteConfigured     bool
	mediaStrategySQLiteSetupErr       error
)

// TestMain restores package-global database state after service tests finish.
func TestMain(m *testing.M) {
	code := m.Run()
	if mediaStrategySQLiteConfigured {
		_ = dao.MediaStrategy.DB().Close(context.Background())
		_ = gdb.SetConfig(mediaStrategySQLiteOriginalConfig)
	}
	if mediaStrategySQLitePath != "" {
		_ = os.Remove(mediaStrategySQLitePath)
	}
	os.Exit(code)
}

// newTestMediaBizCtx returns an empty plugin business-context service for media tests.
func newTestMediaBizCtx() bizctxcap.Service {
	return mediaTestBizCtx{}
}

type mediaTestBizCtx struct{}

// Current returns the static business context configured for media tests.
func (mediaTestBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{}
}

// newTestMediaService creates a media service with an explicit test bizctx adapter.
func newTestMediaService(t *testing.T) Service {
	t.Helper()
	svc, err := newWithRouteMemoryCache(newTestMediaBizCtx(), newMemoryRouteMemoryCache(), newTestMediaConfig())
	if err != nil {
		t.Fatalf("create test media service: %v", err)
	}
	return svc
}

// newTestMediaConfig returns a deterministic plugin config service for media tests.
func newTestMediaConfig() plugincap.ConfigService {
	return mediaTestConfig{
		values: map[string]any{
			"tieta": map[string]any{
				"baseUrl": "http://tieta.invalid",
				"mock":    false,
				"timeout": "3s",
			},
			"innerapi": map[string]any{
				"apiKey": "media",
			},
		},
	}
}

type mediaTestConfig struct {
	values map[string]any
}

// Get returns one raw plugin config value.
func (s mediaTestConfig) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value := gjson.New(s.values).Get(key)
	if value == nil || value.IsNil() {
		if defaultValue == nil {
			return nil, nil
		}
		return gvar.New(defaultValue), nil
	}
	return value, nil
}

// Exists reports whether one plugin config key is present.
func (s mediaTestConfig) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	return value != nil && !value.IsNil(), err
}

// Scan scans one plugin config section into target.
func (s mediaTestConfig) Scan(ctx context.Context, key string, target any) error {
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return err
	}
	return value.Scan(target)
}

// String reads one string plugin config value.
func (s mediaTestConfig) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.String(), nil
}

// Bool reads one bool plugin config value.
func (s mediaTestConfig) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads one int plugin config value.
func (s mediaTestConfig) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration reads one duration plugin config value.
func (s mediaTestConfig) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	duration, err := time.ParseDuration(value.String())
	if err != nil {
		return 0, err
	}
	return duration, nil
}

// fakeTietaClient provides deterministic token and device-permission responses for unit tests.
type fakeTietaClient struct {
	user      *TietaUser
	hasAccess bool
	tokens    []string
}

// UserInfoByToken returns the configured test user.
func (c *fakeTietaClient) UserInfoByToken(ctx context.Context, _ plugincap.ConfigService, token string) (*TietaUser, error) {
	c.tokens = append(c.tokens, token)
	return c.user, nil
}

// CheckTenantHasDevice returns the configured device permission result.
func (c *fakeTietaClient) CheckTenantHasDevice(
	ctx context.Context,
	_ plugincap.ConfigService,
	token string,
	tenantID string,
	deviceID string,
) (bool, error) {
	return c.hasAccess, nil
}

// TestParseTietaTokenUsesMediaClient verifies Tieta token parsing stays inside the media service package.
func TestParseTietaTokenUsesMediaClient(t *testing.T) {
	ctx := context.Background()
	client := &fakeTietaClient{user: &TietaUser{Id: 13, Username: "wj530"}}
	restoreTietaClient := replaceMediaTietaClient(t, client)
	defer restoreTietaClient()

	user, err := parseTietaToken(ctx, newTestMediaConfig(), "Bearer token-value")
	if err != nil {
		t.Fatalf("parse tieta token: %v", err)
	}
	if user == nil || user.Id != 13 || user.Username != "wj530" {
		t.Fatalf("unexpected Tieta user: %+v", user)
	}
	if len(client.tokens) != 1 || client.tokens[0] != "Bearer token-value" {
		t.Fatalf("expected media-local parser to pass raw header once, got %#v", client.tokens)
	}
}

// TestBuildTietaUserMapsCustomerName verifies the upstream authentication payload includes the customer name.
func TestBuildTietaUserMapsCustomerName(t *testing.T) {
	var response tietaUserResponse
	if err := gjson.Unmarshal([]byte(`{"code":200,"data":{"customerName":"公安","phone":"18213268117"}}`), &response); err != nil {
		t.Fatalf("unmarshal Tieta user response: %v", err)
	}
	user := buildTietaUser(response.Data)
	if user == nil || user.CustomerName != "公安" || user.Mobile != "18213268117" {
		t.Fatalf("expected customerName and phone from Tieta response, got %+v", user)
	}
}

// TestAuthenticateTietaTokenCachesUserInfo verifies repeated auth reuses the host cache.
func TestAuthenticateTietaTokenCachesUserInfo(t *testing.T) {
	ctx := context.Background()
	cacheSvc := newMemoryRouteMemoryCache()
	svc, err := newWithRouteMemoryCache(newTestMediaBizCtx(), cacheSvc, newTestMediaConfig())
	if err != nil {
		t.Fatalf("create media service: %v", err)
	}
	client := &fakeTietaClient{user: &TietaUser{Id: 13, Username: "wj530", TenantId: "tenant-a"}}
	restoreTietaClient := replaceMediaTietaClient(t, client)
	defer restoreTietaClient()

	first, err := svc.AuthenticateTietaToken(ctx, "Bearer token-value")
	if err != nil {
		t.Fatalf("authenticate first token: %v", err)
	}
	second, err := svc.AuthenticateTietaToken(ctx, "token-value")
	if err != nil {
		t.Fatalf("authenticate cached token: %v", err)
	}

	if first == nil || second == nil || first.Id != second.Id || second.Username != "wj530" {
		t.Fatalf("expected cached Tieta user to match first result, first=%+v second=%+v", first, second)
	}
	if len(client.tokens) != 1 || client.tokens[0] != "token-value" {
		t.Fatalf("expected upstream Tieta user-info once with normalized token, got %#v", client.tokens)
	}
	if cacheSvc.lastNamespace != tietaUserCacheNamespace {
		t.Fatalf("expected Tieta user cache namespace, got %q", cacheSvc.lastNamespace)
	}
	if cacheSvc.lastKey != tietaUserCacheKey("token-value") {
		t.Fatalf("expected hashed Tieta user cache key, got %q", cacheSvc.lastKey)
	}
	if cacheSvc.lastTTL != time.Minute {
		t.Fatalf("expected one-minute Tieta user cache TTL, got %s", cacheSvc.lastTTL)
	}
}

// TestTietaConfigReadsPluginConfig verifies Tieta HTTP settings come from the media plugin config service.
func TestTietaConfigReadsPluginConfig(t *testing.T) {
	ctx := context.Background()
	configSvc := mediaTestConfig{values: map[string]any{
		"tieta": map[string]any{
			"baseUrl": " http://tieta.example.internal ",
			"mock":    true,
			"timeout": "5s",
		},
	}}

	baseURL, err := tietaBaseURL(ctx, configSvc)
	if err != nil {
		t.Fatalf("read Tieta base URL: %v", err)
	}
	if baseURL != "http://tieta.example.internal" {
		t.Fatalf("expected trimmed plugin config base URL, got %q", baseURL)
	}
	if !isTietaMock(ctx, configSvc) {
		t.Fatal("expected Tieta mock flag from plugin config")
	}
	if timeout := tietaTimeout(ctx, configSvc); timeout != 5*time.Second {
		t.Fatalf("expected Tieta timeout from plugin config, got %s", timeout)
	}
}

// TestTietaBaseURLRejectsMissingPluginConfig verifies missing plugin config still fails before outbound calls.
func TestTietaBaseURLRejectsMissingPluginConfig(t *testing.T) {
	_, err := tietaBaseURL(context.Background(), mediaTestConfig{})
	if err == nil {
		t.Fatal("expected missing Tieta base URL error")
	}
	structured, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected bizerr, got %T", err)
	}
	if structured.RuntimeCode() != "MEDIA_TIETA_BASE_URL_MISSING" {
		t.Fatalf("expected missing base URL code, got %s", structured.RuntimeCode())
	}
}

// TestResolveStrategyByTokenUsesTietaTenantDevicePermission verifies token tenant and device authorization drive strategy resolution.
func TestResolveStrategyByTokenUsesTietaTenantDevicePermission(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, Username: "wj530", RealName: "王杰", Mobile: "18213268117", TenantId: "tenant-a"},
		hasAccess: true,
	})
	defer restoreTietaClient()

	strategyID := insertTestStrategy(t, ctx, "租户设备策略", int(SwitchOff), int(SwitchOn))
	if _, err := dao.MediaStrategyDeviceTenant.Ctx(ctx).Data(do.MediaStrategyDeviceTenant{
		TenantId:   "tenant-a",
		DeviceId:   "34020000001320000001",
		StrategyId: strategyID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant-device binding: %v", err)
	}

	out, err := newTestMediaService(t).ResolveStrategyByToken(ctx, ResolveStrategyByTokenInput{
		Token:    "Bearer token-value",
		TenantId: "tenant-a",
		DeviceId: "34020000001320000001",
	})
	if err != nil {
		t.Fatalf("resolve strategy by token: %v", err)
	}
	if !out.HasAccess {
		t.Fatal("expected Tieta device access")
	}
	if !out.Matched || out.Source != string(StrategySourceTenantDevice) {
		t.Fatalf("expected tenant-device strategy match, got matched=%v source=%s", out.Matched, out.Source)
	}
	if out.UserId != 13 || out.TenantId != "tenant-a" || out.StrategyId != strategyID {
		t.Fatalf("unexpected output: %+v", out)
	}
	if out.UserInfo == nil || out.UserInfo.Username != "wj530" {
		t.Fatalf("expected full Tieta user info, got %+v", out.UserInfo)
	}
}

// TestUserDeviceStrategyByTokenReturnsStrategyContent verifies the HotGo-compatible endpoint returns one strategy content field.
func TestUserDeviceStrategyByTokenReturnsStrategyContent(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, Username: "wj530", RealName: "王杰", Mobile: "18213268117", TenantId: "tenant-a"},
		hasAccess: true,
	})
	defer restoreTietaClient()

	strategyID := insertTestStrategy(t, ctx, "兼容策略", int(SwitchOff), int(SwitchOn))
	if _, err := dao.MediaStrategyDeviceTenant.Ctx(ctx).Data(do.MediaStrategyDeviceTenant{
		TenantId:   "tenant-a",
		DeviceId:   "34020000001320000001",
		StrategyId: strategyID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant-device binding: %v", err)
	}

	out, err := newTestMediaService(t).UserDeviceStrategyByToken(ctx, UserDeviceStrategyByTokenInput{
		Token:    "token-value",
		DeviceId: "34020000001320000001",
		NodeId:   "1",
	})
	if err != nil {
		t.Fatalf("resolve HotGo-compatible strategy by token: %v", err)
	}
	if out.Strategy == nil || out.Strategy.Id != uint64(strategyID) {
		t.Fatalf("unexpected compatibility output: %+v", out)
	}
	if out.UserInfo == nil || out.UserInfo.TenantId != "tenant-a" {
		t.Fatalf("expected Tieta user info, got %+v", out.UserInfo)
	}
	if out.Strategy == nil || out.Strategy.StrategyContent != "record:\n  enabled: true\n" {
		t.Fatalf("expected strategy content field, got %+v", out.Strategy)
	}
}

// TestUserDeviceStrategyByTokenReturnsEmptyStrategyWithoutAccess verifies denied device permission hides strategy details.
func TestUserDeviceStrategyByTokenReturnsEmptyStrategyWithoutAccess(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, Username: "wj530", TenantId: "tenant-a"},
		hasAccess: false,
	})
	defer restoreTietaClient()

	insertTestStrategy(t, ctx, "全局策略", int(SwitchOn), int(SwitchOn))
	out, err := newTestMediaService(t).UserDeviceStrategyByToken(ctx, UserDeviceStrategyByTokenInput{
		Token:    "token-value",
		DeviceId: "34020000001320000001",
		NodeId:   "1",
	})
	if err != nil {
		t.Fatalf("resolve HotGo-compatible strategy by denied token: %v", err)
	}
	if out.UserInfo == nil || out.UserInfo.TenantId != "tenant-a" {
		t.Fatalf("expected Tieta user info, got %+v", out.UserInfo)
	}
	if out.Strategy != nil {
		t.Fatalf("expected empty strategy without device access, got %+v", out.Strategy)
	}
}

// TestUserDeviceStrategyByTokenRejectsWhenTenantNodeLimitReached verifies node-scoped stream limits use active sessions.
func TestUserDeviceStrategyByTokenRejectsWhenTenantNodeLimitReached(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, Username: "wj530", TenantId: "tenant-a"},
		hasAccess: true,
	})
	defer restoreTietaClient()

	strategyID := insertTestStrategy(t, ctx, "节点限流策略", int(SwitchOff), int(SwitchOn))
	if _, err := dao.MediaStrategyDeviceTenant.Ctx(ctx).Data(do.MediaStrategyDeviceTenant{
		TenantId:   "tenant-a",
		DeviceId:   "34020000001320000001",
		StrategyId: strategyID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant-device binding: %v", err)
	}
	if _, err := dao.MediaTenantStreamConfig.Ctx(ctx).Data(do.MediaTenantStreamConfig{
		TenantId:      "tenant-a",
		MaxConcurrent: 2,
		NodeNum:       1,
		Enable:        int(TenantStreamEnabled),
	}).Insert(); err != nil {
		t.Fatalf("insert tenant node stream limit: %v", err)
	}
	now := time.Date(2026, 6, 16, 8, 0, 0, 0, time.UTC)
	startTime := gtime.NewFromTime(now.Add(-time.Minute))
	closeTime := gtime.NewFromTime(now)
	insertDashboardReports(t, ctx, []any{
		do.MediaReportSession{
			SessionId:    "limit-active-a",
			StreamId:     "stream-a",
			TenantId:     "tenant-a",
			ClientType:   int(SessionClientTypePC),
			ProtocolType: "HLS",
			NodeId:       "1",
			StartTime:    startTime,
			ReportTime:   now.UnixMilli(),
		},
		do.MediaReportSession{
			SessionId:    "limit-closed",
			StreamId:     "stream-a",
			TenantId:     "tenant-a",
			ClientType:   int(SessionClientTypePC),
			ProtocolType: "HLS",
			NodeId:       "1",
			StartTime:    startTime,
			CloseTime:    closeTime,
			ReportTime:   now.UnixMilli(),
		},
	})

	out, err := newTestMediaService(t).UserDeviceStrategyByToken(ctx, UserDeviceStrategyByTokenInput{
		Token:    "token-value",
		DeviceId: "34020000001320000001",
		NodeId:   "1",
	})
	if err != nil {
		t.Fatalf("expected closed sessions not to count against limit: %v", err)
	}
	if out.Strategy == nil || out.Strategy.Id != uint64(strategyID) {
		t.Fatalf("expected strategy while active count below limit, got %+v", out)
	}

	insertDashboardReports(t, ctx, []any{
		do.MediaReportSession{
			SessionId:    "limit-active-b",
			StreamId:     "stream-b",
			TenantId:     "tenant-a",
			ClientType:   int(SessionClientTypePC),
			ProtocolType: "HLS",
			NodeId:       "1",
			StartTime:    startTime,
			ReportTime:   now.UnixMilli(),
		},
	})
	_, err = newTestMediaService(t).UserDeviceStrategyByToken(ctx, UserDeviceStrategyByTokenInput{
		Token:    "token-value",
		DeviceId: "34020000001320000001",
		NodeId:   "1",
	})
	if err == nil {
		t.Fatal("expected node stream limit exceeded error")
	}
	structured, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected bizerr, got %T", err)
	}
	if structured.RuntimeCode() != "MEDIA_TENANT_STREAM_LIMIT_EXCEEDED" {
		t.Fatalf("expected tenant stream limit code, got %s", structured.RuntimeCode())
	}
}

// TestListTenantWhiteIPsByTokenReturnsEnabledTenantIPs verifies token tenant resolution drives whitelist lookup.
func TestListTenantWhiteIPsByTokenReturnsEnabledTenantIPs(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user: &TietaUser{Id: 13, Username: "wj530", TenantId: "tenant-a"},
	})
	defer restoreTietaClient()

	insertTestTenantWhite(t, ctx, "tenant-a", "192.0.2.10", int(WhiteEnabled))
	insertTestTenantWhite(t, ctx, "tenant-a", "192.0.2.11", int(WhiteDisabled))
	insertTestTenantWhite(t, ctx, "tenant-b", "192.0.2.12", int(WhiteEnabled))

	out, err := newTestMediaService(t).ListTenantWhiteIPsByToken(ctx, TenantWhiteIPsByTokenInput{
		Token: "token-value",
	})
	if err != nil {
		t.Fatalf("list tenant whitelist IPs by token: %v", err)
	}
	if out.TenantId != "tenant-a" {
		t.Fatalf("expected tenant-a in whitelist lookup output, got %q", out.TenantId)
	}
	if len(out.Ips) != 1 || out.Ips[0] != "192.0.2.10" {
		t.Fatalf("expected only enabled tenant-a whitelist IPs, got %#v", out.Ips)
	}
}

// TestTenantStreamConfigUsesTenantAndNodeKey verifies tenant stream configs are keyed by tenant and node.
func TestTenantStreamConfigUsesTenantAndNodeKey(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	svc := newTestMediaService(t)

	insertTestNode(t, ctx, 1, "节点一")
	insertTestNode(t, ctx, 2, "节点二")

	first, err := svc.CreateTenantStreamConfig(ctx, TenantStreamConfigMutationInput{
		TenantId:      "tenant-a",
		MaxConcurrent: 10,
		NodeNum:       1,
		Enable:        int(TenantStreamEnabled),
	})
	if err != nil {
		t.Fatalf("create first tenant stream config: %v", err)
	}
	if first.TenantId != "tenant-a" || first.NodeNum != 1 {
		t.Fatalf("unexpected first mutation output: %+v", first)
	}

	second, err := svc.CreateTenantStreamConfig(ctx, TenantStreamConfigMutationInput{
		TenantId:      "tenant-a",
		MaxConcurrent: 20,
		NodeNum:       2,
		Enable:        int(TenantStreamEnabled),
	})
	if err != nil {
		t.Fatalf("create second tenant stream config: %v", err)
	}
	if second.TenantId != "tenant-a" || second.NodeNum != 2 {
		t.Fatalf("unexpected second mutation output: %+v", second)
	}

	if _, err = svc.CreateTenantStreamConfig(ctx, TenantStreamConfigMutationInput{
		TenantId:      "tenant-a",
		MaxConcurrent: 30,
		NodeNum:       2,
		Enable:        int(TenantStreamEnabled),
	}); err == nil {
		t.Fatal("expected duplicate tenant stream config error")
	}

	updated, err := svc.UpdateTenantStreamConfig(ctx, "tenant-a", 1, TenantStreamConfigMutationInput{
		TenantId:      "tenant-a",
		MaxConcurrent: 40,
		NodeNum:       1,
		Enable:        int(TenantStreamDisabled),
	})
	if err != nil {
		t.Fatalf("update first tenant stream config: %v", err)
	}
	if updated.TenantId != "tenant-a" || updated.NodeNum != 1 {
		t.Fatalf("unexpected updated mutation output: %+v", updated)
	}

	firstDetail, err := svc.GetTenantStreamConfig(ctx, "tenant-a", 1)
	if err != nil {
		t.Fatalf("get first tenant stream config: %v", err)
	}
	if firstDetail.MaxConcurrent != 40 || firstDetail.Enable != int(TenantStreamDisabled) {
		t.Fatalf("expected first config updated, got %+v", firstDetail)
	}

	secondDetail, err := svc.GetTenantStreamConfig(ctx, "tenant-a", 2)
	if err != nil {
		t.Fatalf("get second tenant stream config: %v", err)
	}
	if secondDetail.MaxConcurrent != 20 || secondDetail.Enable != int(TenantStreamEnabled) {
		t.Fatalf("expected second config unchanged, got %+v", secondDetail)
	}

	if _, err = svc.DeleteTenantStreamConfig(ctx, "tenant-a", 1); err != nil {
		t.Fatalf("delete first tenant stream config: %v", err)
	}
	if _, err = svc.GetTenantStreamConfig(ctx, "tenant-a", 1); err == nil {
		t.Fatal("expected first config to be deleted")
	}
	if _, err = svc.GetTenantStreamConfig(ctx, "tenant-a", 2); err != nil {
		t.Fatalf("expected second config to remain after deleting first: %v", err)
	}
}

// TestResolveStrategyByTokenRejectsTenantMismatch verifies callers cannot override the token tenant.
func TestResolveStrategyByTokenRejectsTenantMismatch(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, TenantId: "tenant-a"},
		hasAccess: true,
	})
	defer restoreTietaClient()

	_, err := newTestMediaService(t).ResolveStrategyByToken(ctx, ResolveStrategyByTokenInput{
		Token:    "token-value",
		TenantId: "tenant-b",
		DeviceId: "34020000001320000001",
	})
	if err == nil {
		t.Fatal("expected tenant mismatch error")
	}
	structured, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected bizerr, got %T", err)
	}
	if structured.RuntimeCode() != "MEDIA_TIETA_TENANT_MISMATCH" {
		t.Fatalf("expected tenant mismatch code, got %s", structured.RuntimeCode())
	}
}

// TestResolveStrategyByTokenDeniesWithoutDevicePermission verifies denied Tieta permission does not return a strategy.
func TestResolveStrategyByTokenDeniesWithoutDevicePermission(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	restoreTietaClient := replaceMediaTietaClient(t, &fakeTietaClient{
		user:      &TietaUser{Id: 13, TenantId: "tenant-a"},
		hasAccess: false,
	})
	defer restoreTietaClient()

	insertTestStrategy(t, ctx, "全局策略", int(SwitchOn), int(SwitchOn))
	out, err := newTestMediaService(t).ResolveStrategyByToken(ctx, ResolveStrategyByTokenInput{
		Authorization: "Bearer token-value",
		DeviceId:      "34020000001320000001",
	})
	if err != nil {
		t.Fatalf("resolve strategy by denied token: %v", err)
	}
	if out.HasAccess || out.Matched || out.Source != string(StrategySourceNone) {
		t.Fatalf("expected no access and no strategy, got %+v", out)
	}
}

// setupMediaStrategySQLite creates the minimal media tables required by strategy resolution.
func setupMediaStrategySQLite(t *testing.T, ctx context.Context) {
	t.Helper()

	mediaStrategySQLiteOnce.Do(func() {
		mediaStrategySQLiteOriginalConfig = gdb.GetAllConfig()
		tmpFile, err := os.CreateTemp("", "linapro-media-service-*.db")
		if err != nil {
			mediaStrategySQLiteSetupErr = err
			return
		}
		mediaStrategySQLitePath = tmpFile.Name()
		if err = tmpFile.Close(); err != nil {
			mediaStrategySQLiteSetupErr = err
			return
		}
		if err = os.Remove(mediaStrategySQLitePath); err != nil {
			mediaStrategySQLiteSetupErr = err
			return
		}
		if err = gdb.SetConfig(gdb.Config{
			"default": {
				{Link: "sqlite::@file(" + mediaStrategySQLitePath + ")"},
			},
		}); err != nil {
			mediaStrategySQLiteSetupErr = err
			return
		}
		mediaStrategySQLiteConfigured = true
	})
	if mediaStrategySQLiteSetupErr != nil {
		t.Fatalf("setup sqlite config: %v", mediaStrategySQLiteSetupErr)
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS media_strategy (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			strategy TEXT NOT NULL,
			global INTEGER NOT NULL,
			enable INTEGER NOT NULL,
			creator_id INTEGER,
			updater_id INTEGER,
			create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS media_strategy_device (
			device_id TEXT PRIMARY KEY,
			strategy_id INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS media_strategy_tenant (
			tenant_id TEXT PRIMARY KEY,
			strategy_id INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS media_strategy_device_tenant (
			tenant_id TEXT NOT NULL,
			device_id TEXT NOT NULL,
			strategy_id INTEGER NOT NULL,
			PRIMARY KEY (tenant_id, device_id)
		)`,
		`CREATE TABLE IF NOT EXISTS media_device_node (device_id TEXT NOT NULL, channel_id TEXT NOT NULL, node_num INTEGER NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS media_node (id INTEGER PRIMARY KEY AUTOINCREMENT, node_num INTEGER NOT NULL, name TEXT NOT NULL, qn_url TEXT NOT NULL, basic_url TEXT NOT NULL, dn_url TEXT NOT NULL, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS media_tenant_stream_config (tenant_id TEXT NOT NULL, max_concurrent INTEGER NOT NULL, node_num INTEGER NOT NULL, enable INTEGER NOT NULL, creator_id INTEGER, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, updater_id INTEGER, update_time TEXT, PRIMARY KEY (tenant_id, node_num))`,
		`CREATE TABLE IF NOT EXISTS media_tenant_white (tenant_id TEXT NOT NULL, ip TEXT NOT NULL, enable INTEGER NOT NULL, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (tenant_id, ip))`,
		`CREATE TABLE IF NOT EXISTS media_stream_alias (id INTEGER PRIMARY KEY AUTOINCREMENT, alias TEXT NOT NULL, auto_remove INTEGER NOT NULL, stream_path TEXT NOT NULL, device_id TEXT NOT NULL, channel_id TEXT NOT NULL, create_time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, statement := range statements {
		if _, err := dao.MediaStrategy.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("exec sqlite schema: %v", err)
		}
	}

	cleanupStatements := []string{
		`DELETE FROM media_stream_alias`,
		`DELETE FROM media_tenant_white`,
		`DELETE FROM media_tenant_stream_config`,
		`DELETE FROM media_device_node`,
		`DELETE FROM media_node`,
		`DELETE FROM media_strategy_device_tenant`,
		`DELETE FROM media_strategy_device`,
		`DELETE FROM media_strategy_tenant`,
		`DELETE FROM media_strategy`,
	}
	for _, statement := range cleanupStatements {
		if _, err := dao.MediaStrategy.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("cleanup sqlite data: %v", err)
		}
	}
}

// insertTestStrategy inserts one enabled test strategy and returns its generated ID.
func insertTestStrategy(t *testing.T, ctx context.Context, name string, global int, enable int) int64 {
	t.Helper()

	id, err := dao.MediaStrategy.Ctx(ctx).Data(do.MediaStrategy{
		Name:     name,
		Strategy: "record:\n  enabled: true\n",
		Global:   global,
		Enable:   enable,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert strategy: %v", err)
	}
	return id
}

// insertTestTenantWhite inserts one tenant whitelist fixture.
func insertTestTenantWhite(t *testing.T, ctx context.Context, tenantID string, ip string, enable int) {
	t.Helper()

	if _, err := dao.MediaTenantWhite.Ctx(ctx).Data(do.MediaTenantWhite{
		TenantId: tenantID,
		Ip:       ip,
		Enable:   enable,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant whitelist: %v", err)
	}
}

// insertTestNode inserts one media node fixture.
func insertTestNode(t *testing.T, ctx context.Context, nodeNum int, name string) {
	t.Helper()

	if _, err := dao.MediaNode.Ctx(ctx).Data(do.MediaNode{
		NodeNum:  nodeNum,
		Name:     name,
		QnUrl:    "https://qn.example.com",
		BasicUrl: "https://basic.example.com",
		DnUrl:    "https://dn.example.com",
	}).Insert(); err != nil {
		t.Fatalf("insert media node: %v", err)
	}
}

// replaceMediaTietaClient swaps the process Tieta client and returns a restore function.
func replaceMediaTietaClient(t *testing.T, client tietaClient) func() {
	t.Helper()

	original := mediaTietaClient
	mediaTietaClient = client
	return func() {
		mediaTietaClient = original
	}
}
