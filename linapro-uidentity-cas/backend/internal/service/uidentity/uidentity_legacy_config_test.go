// This file tests legacy static configuration projection from plugin-scoped
// config without requiring host-global configuration fixtures.

package uidentity

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"gopkg.in/yaml.v3"
	"lina-core/pkg/plugin/capability/plugincap"
)

const legacyConfigTestPluginID = "linapro-uidentity-cas"

func TestLegacyCASConfigUsesPluginScopedDefaults(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	out, err := service.LegacyCASConfig(context.Background())
	if err != nil {
		t.Fatalf("legacy CAS config: %v", err)
	}
	if out.LoginAddr != defaultLegacyCASLoginAddr || out.RestAddr != defaultLegacyCASRestAddr {
		t.Fatalf("unexpected CAS defaults: %#v", out)
	}
	for _, value := range []string{out.LoginAddr, out.LogoutAddr, out.RestAddr} {
		if strings.Contains(value, "/uidentity") {
			t.Fatalf("legacy CAS default must not expose LinaPro route namespace: %#v", out)
		}
	}
}

func TestLegacyStaticConfigDefaultsUseOldAdminRoutes(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	oauth, err := service.LegacyOAuthConfig(context.Background())
	if err != nil {
		t.Fatalf("legacy OAuth config: %v", err)
	}
	token, err := service.LegacyTokenConfig(context.Background())
	if err != nil {
		t.Fatalf("legacy token config: %v", err)
	}
	got := map[string]string{
		"oauth.authorization": oauth.Authorization,
		"oauth.token":         oauth.GetTokenAddr,
		"oauth.userInfo":      oauth.UserInfoAddr,
		"oauth.ping":          oauth.PingAddr,
		"token.get":           token.GetAddr,
		"token.check":         token.CheckAddr,
	}
	want := map[string]string{
		"oauth.authorization": "/api/v1/oauth/auth",
		"oauth.token":         "/api/v1/oauth/token",
		"oauth.userInfo":      "/api/v1/oauth/test",
		"oauth.ping":          "/api/v1/health",
		"token.get":           "/api/v1/token/get",
		"token.check":         "/api/v1/token/getUserInfoByToken",
	}
	for name, wantValue := range want {
		if got[name] != wantValue {
			t.Fatalf("legacy default %s = %q, want %q", name, got[name], wantValue)
		}
		if strings.Contains(got[name], "/uidentity") {
			t.Fatalf("legacy default %s must not expose LinaPro route namespace: %q", name, got[name])
		}
	}
}

func TestLegacyOAuthConfigReadsPluginScopedOverrides(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, `
legacy:
  static:
    oauth:
      authorization: /custom/auth
      getTokenAddr: /custom/token
      userInfoAddr: /custom/user
      logoutAddr: /custom/logout
      pingAddr: /custom/ping
      docs: /custom/docs
`)}
	out, err := service.LegacyOAuthConfig(context.Background())
	if err != nil {
		t.Fatalf("legacy OAuth config: %v", err)
	}
	if out.Authorization != "/custom/auth" || out.GetTokenAddr != "/custom/token" ||
		out.UserInfoAddr != "/custom/user" || out.LogoutAddr != "/custom/logout" ||
		out.PingAddr != "/custom/ping" || out.Docs != "/custom/docs" {
		t.Fatalf("unexpected OAuth override projection: %#v", out)
	}
}

func newLegacyConfigTestService(t *testing.T, content string) plugincap.ConfigService {
	t.Helper()

	var values map[string]any
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		t.Fatalf("parse legacy config test data: %v", err)
	}
	return legacyConfigTestService{values: values}
}

type legacyConfigTestService struct {
	values map[string]any
}

func (s legacyConfigTestService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value, ok := lookupLegacyConfigTestValue(s.values, key)
	if !ok {
		if defaultValue == nil {
			return nil, nil
		}
		return gvar.New(defaultValue), nil
	}
	return gvar.New(value), nil
}

func (s legacyConfigTestService) Exists(_ context.Context, key string) (bool, error) {
	_, ok := lookupLegacyConfigTestValue(s.values, key)
	return ok, nil
}

func (s legacyConfigTestService) Scan(_ context.Context, key string, target any) error {
	value, ok := lookupLegacyConfigTestValue(s.values, key)
	if !ok {
		return nil
	}
	payload, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(payload, target)
}

func (s legacyConfigTestService) String(ctx context.Context, key string, defaultValue string) (string, error) {
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

func (s legacyConfigTestService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return false, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Bool(), nil
}

func (s legacyConfigTestService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Int(), nil
}

func (s legacyConfigTestService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
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

func lookupLegacyConfigTestValue(values map[string]any, key string) (any, bool) {
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
