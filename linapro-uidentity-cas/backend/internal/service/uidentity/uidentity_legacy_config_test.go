// This file tests legacy static configuration projection from plugin-scoped
// config without requiring host-global configuration fixtures.

package uidentity

import (
	"context"
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
