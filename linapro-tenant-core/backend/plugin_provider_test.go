// This file verifies the linapro-tenant-core provider adapter construction boundary.

package backend

import (
	"context"
	"testing"

	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// TestProvideTenantUsesTypedProviderEnv verifies provider construction only
// needs the narrow tenantcap.ProviderEnv published for tenant capability.
func TestProvideTenantUsesTypedProviderEnv(t *testing.T) {
	provider, err := provideTenant(context.Background(), tenantcap.ProviderEnv{
		PluginID: pluginID,
		BizCtx:   fakeBizCtx{},
	})
	if err != nil {
		t.Fatalf("expected typed tenant provider env to construct provider: %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider instance")
	}
}

// TestProvideTenantRejectsMissingBizCtx verifies provider construction does
// not silently create an adapter without the host-published bizctx service.
func TestProvideTenantRejectsMissingBizCtx(t *testing.T) {
	provider, err := provideTenant(context.Background(), tenantcap.ProviderEnv{
		PluginID: pluginID,
	})
	if err == nil {
		t.Fatal("expected missing bizctx service to fail provider construction")
	}
	if provider != nil {
		t.Fatal("expected nil provider when construction fails")
	}
}

// fakeBizCtx is the minimal host-published business context required by
// linapro-tenant-core provider construction tests.
type fakeBizCtx struct{}

// Current returns a platform context because these tests only cover construction.
func (fakeBizCtx) Current(context.Context) plugincontract.CurrentContext {
	return plugincontract.CurrentContext{PlatformBypass: true}
}
