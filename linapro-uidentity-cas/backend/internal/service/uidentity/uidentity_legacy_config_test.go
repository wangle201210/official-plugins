// This file tests legacy static configuration projection from plugin-scoped
// config without requiring host-global configuration fixtures.

package uidentity

import (
	"context"
	"strings"
	"testing"

	configsvc "lina-core/pkg/plugin/capability/config"
	plugincontract "lina-core/pkg/plugin/capability/contract"
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

func newLegacyConfigTestService(t *testing.T, content string) plugincontract.ConfigService {
	t.Helper()

	return configsvc.NewFactory(t.TempDir(), t.TempDir()).
		WithArtifactConfig(legacyConfigTestPluginID, []byte(content)).
		ForPlugin(legacyConfigTestPluginID)
}
