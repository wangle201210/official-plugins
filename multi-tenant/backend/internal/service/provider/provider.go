// Package provider implements the multi-tenant plugin provider adapter.
package provider

import (
	"github.com/gogf/gf/v2/errors/gerror"

	pkgtenantcap "lina-core/pkg/tenantcap"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/resolver"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
	"lina-plugin-multi-tenant/backend/internal/service/tenantplugin"
)

// Provider is the plugin-owned tenant capability provider. It mirrors the
// host tenantcap contract so registration can be wired once the host seam lands.
type Provider struct {
	membershipSvc     membership.Service
	resolverSvc       resolver.Service
	resolverConfigSvc resolverconfig.Service
	tenantPluginSvc   tenantplugin.Service
}

// Ensure Provider implements the host tenant capability provider contract.
var _ pkgtenantcap.Provider = (*Provider)(nil)
var _ pkgtenantcap.UserMembershipProvider = (*Provider)(nil)
var _ pkgtenantcap.TenantPluginProvisioningProvider = (*Provider)(nil)

// New creates and returns a Provider instance.
func New(
	membershipSvc membership.Service,
	resolverSvc resolver.Service,
	resolverConfigSvc resolverconfig.Service,
	tenantPluginSvc tenantplugin.Service,
) (*Provider, error) {
	if membershipSvc == nil {
		return nil, gerror.New("multi-tenant provider requires membership service")
	}
	if resolverSvc == nil {
		return nil, gerror.New("multi-tenant provider requires resolver service")
	}
	if resolverConfigSvc == nil {
		return nil, gerror.New("multi-tenant provider requires resolver config service")
	}
	if tenantPluginSvc == nil {
		return nil, gerror.New("multi-tenant provider requires tenant plugin service")
	}
	return &Provider{
		membershipSvc:     membershipSvc,
		resolverSvc:       resolverSvc,
		resolverConfigSvc: resolverConfigSvc,
		tenantPluginSvc:   tenantPluginSvc,
	}, nil
}
