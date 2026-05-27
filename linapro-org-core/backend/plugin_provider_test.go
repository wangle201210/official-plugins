// This file verifies the linapro-org-core provider adapter construction boundary.

package backend

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/capability/orgcap"
)

// TestProvideOrgUsesTypedProviderEnv verifies provider construction only needs
// the narrow orgcap.ProviderEnv published for organization capability.
func TestProvideOrgUsesTypedProviderEnv(t *testing.T) {
	provider, err := provideOrg(context.Background(), orgcap.ProviderEnv{
		PluginID:     pluginID,
		TenantFilter: fakeTenantFilter{},
	})
	if err != nil {
		t.Fatalf("expected typed org provider env to construct provider: %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider instance")
	}
}

// TestProvideOrgRejectsMissingTenantFilter verifies provider construction does
// not silently create an adapter without the host-published tenant filter.
func TestProvideOrgRejectsMissingTenantFilter(t *testing.T) {
	provider, err := provideOrg(context.Background(), orgcap.ProviderEnv{
		PluginID: pluginID,
	})
	if err == nil {
		t.Fatal("expected missing tenant filter to fail provider construction")
	}
	if provider != nil {
		t.Fatal("expected nil provider when construction fails")
	}
}

// fakeTenantFilter is the minimal host-published tenant filter required by
// linapro-org-core provider construction tests.
type fakeTenantFilter struct{}

// Context returns a deterministic tenant context for construction-only tests.
func (fakeTenantFilter) Context(context.Context) plugincontract.TenantFilterContext {
	return plugincontract.TenantFilterContext{TenantID: 1}
}

// Apply returns the input model unchanged because these tests do not execute queries.
func (fakeTenantFilter) Apply(_ context.Context, model *gdb.Model, _ string) *gdb.Model {
	return model
}
