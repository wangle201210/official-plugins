// Package tenantadapter adapts linapro-tenant-core internal services to the
// framework tenant capability provider contract.
package tenantadapter

import (
	"github.com/gogf/gf/v2/errors/gerror"

	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-plugin-linapro-tenant-core/backend/internal/service/membership"
	"lina-plugin-linapro-tenant-core/backend/internal/service/provider"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	"lina-plugin-linapro-tenant-core/backend/internal/service/tenantplugin"
)

// New creates the tenant framework capability provider from host-published services.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	pluginLifecycleSvc plugincontract.PluginLifecycleService,
) (tenantcap.Provider, error) {
	if bizCtxSvc == nil {
		return nil, gerror.New("linapro-tenant-core provider requires host bizctx service")
	}
	membershipSvc := membership.New(bizCtxSvc)
	resolverConfigSvc := resolverconfig.New()
	tenantPluginSvc := tenantplugin.New(bizCtxSvc, pluginLifecycleSvc)
	resolverSvc := resolver.New(bizCtxSvc, membershipSvc)
	return provider.New(membershipSvc, resolverSvc, resolverConfigSvc, tenantPluginSvc)
}
