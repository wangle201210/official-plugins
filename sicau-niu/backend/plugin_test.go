// plugin_test.go covers pure backend plugin wiring helpers without starting the
// host HTTP server or scheduler.

package backend

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"gopkg.in/yaml.v3"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/plugincap"
	playerv1 "lina-plugin-sicau-niu/backend/api/player/v1"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
	tokensvc "lina-plugin-sicau-niu/backend/internal/service/token"
)

type fakeIronLocationRefresher struct {
	called bool
}

func (f *fakeIronLocationRefresher) Refresh(ctx context.Context) (*feedingsvc.IronLocationRefreshResult, error) {
	f.called = true
	return &feedingsvc.IronLocationRefreshResult{}, nil
}

func TestBuildAuthDependenciesAllowsMissingTokenSecret(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
wechat:
  mock: true
`)

	tokenService, gateway, err := buildAuthDependencies(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildAuthDependencies returned error: %v", err)
	}
	if tokenService == nil {
		t.Fatal("expected fallback token service")
	}
	if gateway == nil {
		t.Fatal("expected WeChat gateway")
	}
	token, err := tokenService.Sign(context.Background(), 1)
	if token != "" {
		t.Fatalf("expected empty token from unconfigured service, got %q", token)
	}
	assertTokenSecretMissing(t, err)
	playerID, err := tokenService.Verify(context.Background(), "token")
	if playerID != 0 {
		t.Fatalf("expected player ID 0 from unconfigured service, got %d", playerID)
	}
	assertTokenSecretMissing(t, err)
}

func TestBuildAuthDependenciesConfiguredTokenSignsAndVerifies(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
token:
  secret: "unit-test-secret"
  ttl: 1h
wechat:
  mock: true
`)

	tokenService, gateway, err := buildAuthDependencies(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildAuthDependencies returned error: %v", err)
	}
	if gateway == nil {
		t.Fatal("expected WeChat gateway")
	}
	token, err := tokenService.Sign(context.Background(), 42)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	playerID, err := tokenService.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("verify token failed: %v", err)
	}
	if playerID != 42 {
		t.Fatalf("expected player ID 42, got %d", playerID)
	}
}

func TestBuildIronLocationRefreshDisabledWithoutCredentials(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
wechat:
  mock: true
`)

	enabled, refresher, _, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildIronLocationRefresh returned error: %v", err)
	}
	if enabled || refresher != nil {
		t.Fatalf("expected refresh disabled without credentials")
	}
}

func TestBuildIronLocationRefreshDefaultsToOneMinute(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
niu:
  baseUrl: "https://example.test/iot"
  key: "key-1"
  secret: "secret-1"
`)

	enabled, refresher, interval, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildIronLocationRefresh returned error: %v", err)
	}
	if !enabled || refresher == nil {
		t.Fatalf("expected IOT refresh to be enabled")
	}
	if interval != time.Minute {
		t.Fatalf("expected default interval 1min, got %s", interval)
	}
}

func TestBuildIronLocationRefreshClampsIntervalToOneMinute(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
niu:
  baseUrl: "https://example.test/iot"
  key: "key-1"
  secret: "secret-1"
  refreshInterval: 1s
`)

	enabled, refresher, interval, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildIronLocationRefresh returned error: %v", err)
	}
	if !enabled || refresher == nil {
		t.Fatalf("expected IOT refresh to be enabled")
	}
	if interval != time.Minute {
		t.Fatalf("expected interval clamped to 1min, got %s", interval)
	}
}

func TestBuildIronLocationRefreshReadsBaseUrl(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
niu:
  baseUrl: "not-a-url"
  key: "key-1"
  secret: "secret-1"
`)

	_, _, _, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err == nil {
		t.Fatalf("expected invalid baseUrl to fail")
	}
	if !strings.Contains(err.Error(), "niu.baseUrl") {
		t.Fatalf("expected baseUrl validation error, got %v", err)
	}
}

func TestRefreshIronLocationsSkipsNonPrimaryNode(t *testing.T) {
	refresher := &fakeIronLocationRefresher{}
	if err := refreshIronLocations(context.Background(), false, refresher); err != nil {
		t.Fatalf("refreshIronLocations returned error: %v", err)
	}
	if refresher.called {
		t.Fatalf("expected non-primary node to skip refresh")
	}
}

func TestRefreshIronLocationsUsesInjectedRefresherOnPrimaryNode(t *testing.T) {
	refresher := &fakeIronLocationRefresher{}
	if err := refreshIronLocations(context.Background(), true, refresher); err != nil {
		t.Fatalf("refreshIronLocations returned error: %v", err)
	}
	if !refresher.called {
		t.Fatalf("expected primary node to call refresher")
	}
}

func TestPlayerRequestDTOsUseMiniProgramAPIDocTag(t *testing.T) {
	requests := []interface{}{
		playerv1.ActivateReq{},
		playerv1.CollectionReq{},
		playerv1.CertificateReq{},
		playerv1.CheckinReq{},
		playerv1.CollegeOptionsReq{},
		playerv1.FeedReq{},
		playerv1.FeedingTrailReq{},
		playerv1.GiftReq{},
		playerv1.GrassAccountReq{},
		playerv1.PlayerHonorsReq{},
		playerv1.LoginReq{},
		playerv1.MessagesReq{},
		playerv1.MarkMessageReadReq{},
		playerv1.VisibleNiuReq{},
		playerv1.BindPhoneReq{},
		playerv1.PosterReq{},
		playerv1.GetProfileReq{},
		playerv1.UpdateProfileReq{},
		playerv1.FeedRankingReq{},
		playerv1.CollegeRankingReq{},
		playerv1.FriendRankingReq{},
		playerv1.StealTargetsReq{},
		playerv1.StealReq{},
	}
	for _, request := range requests {
		reqType := reflect.TypeOf(request)
		metaField := reqType.Field(0)
		metaTag := string(metaField.Tag)
		if !strings.Contains(metaTag, `tags:"寻牛小程序"`) {
			t.Fatalf("expected %s to use mini-program API doc tag, got tag=%s", reqType.Name(), metaTag)
		}
		if strings.Contains(metaTag, `tags:"Sicau Niu Player"`) {
			t.Fatalf("expected %s to avoid legacy player API doc tag, got tag=%s", reqType.Name(), metaTag)
		}
	}
}

func TestAPIDocSummariesUseChinese(t *testing.T) {
	summaryPattern := regexp.MustCompile(`summary:"([^"]*)"`)
	chinesePattern := regexp.MustCompile(`[\p{Han}]`)
	apiRoot := filepath.Join("api")
	err := filepath.WalkDir(apiRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		matches := summaryPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if !chinesePattern.MatchString(match[1]) {
				t.Fatalf("expected %s summary to use Chinese, got %q", path, match[1])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk API files: %v", err)
	}
}

func newPluginTestConfigService(t *testing.T, content string) plugincap.ConfigService {
	t.Helper()
	var values map[string]any
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		t.Fatalf("parse plugin test config: %v", err)
	}
	return pluginTestConfigService{values: values}
}

type pluginTestConfigService struct {
	values map[string]any
}

func (s pluginTestConfigService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value, ok := lookupPluginTestConfigValue(s.values, key)
	if !ok {
		if defaultValue == nil {
			return nil, nil
		}
		return gvar.New(defaultValue), nil
	}
	return gvar.New(value), nil
}

func (s pluginTestConfigService) Exists(_ context.Context, key string) (bool, error) {
	_, ok := lookupPluginTestConfigValue(s.values, key)
	return ok, nil
}

func (s pluginTestConfigService) Scan(_ context.Context, key string, target any) error {
	value, ok := lookupPluginTestConfigValue(s.values, key)
	if !ok {
		return nil
	}
	payload, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(payload, target)
}

func (s pluginTestConfigService) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return "", err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	raw := value.String()
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	return raw, nil
}

func (s pluginTestConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return false, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Bool(), nil
}

func (s pluginTestConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Int(), nil
}

func (s pluginTestConfigService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func lookupPluginTestConfigValue(values map[string]any, key string) (any, bool) {
	if values == nil {
		return nil, false
	}
	current := any(values)
	for _, part := range strings.Split(key, ".") {
		mapped, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = mapped[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func assertTokenSecretMissing(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected token secret missing error")
	}
	if !bizerr.Is(err, tokensvc.CodeTokenSecretMissing) {
		t.Fatalf("expected %s, got %v", tokensvc.CodeTokenSecretMissing.RuntimeCode(), err)
	}
}
