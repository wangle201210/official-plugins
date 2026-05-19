// Package backend wires the linapro-tenant-core source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginhost"
	pkgtenantcap "lina-core/pkg/tenantcap"
	multitenant "lina-plugin-linapro-tenant-core"
	authcontroller "lina-plugin-linapro-tenant-core/backend/internal/controller/auth"
	platformcontroller "lina-plugin-linapro-tenant-core/backend/internal/controller/platform"
	tenantcontroller "lina-plugin-linapro-tenant-core/backend/internal/controller/tenant"
	"lina-plugin-linapro-tenant-core/backend/internal/service/impersonate"
	"lina-plugin-linapro-tenant-core/backend/internal/service/lifecycleprecondition"
	"lina-plugin-linapro-tenant-core/backend/internal/service/membership"
	"lina-plugin-linapro-tenant-core/backend/internal/service/provider"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	tenantsvc "lina-plugin-linapro-tenant-core/backend/internal/service/tenant"
	"lina-plugin-linapro-tenant-core/backend/internal/service/tenantplugin"
)

// linapro-tenant-core plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-tenant-core"
)

// init registers the linapro-tenant-core source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(multitenant.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterBeforeDisableHandler(beforeDisable); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterBeforeUninstallHandler(beforeUninstall); err != nil {
		panic(err)
	}
	if err := plugin.Lifecycle().RegisterBeforeTenantDeleteHandler(beforeTenantDelete); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// beforeDisable enforces linapro-tenant-core plugin disable preconditions.
func beforeDisable(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, lifecycleprecondition.ReasonDisableTenantsExist, err
	}
	return precondition.BeforeDisable(ctx, input)
}

// beforeUninstall enforces linapro-tenant-core plugin uninstall preconditions.
func beforeUninstall(ctx context.Context, input pluginhost.SourcePluginLifecycleInput) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, lifecycleprecondition.ReasonUninstallTenantsExist, err
	}
	return precondition.BeforeUninstall(ctx, input)
}

// beforeTenantDelete runs linapro-tenant-core tenant deletion preconditions.
func beforeTenantDelete(
	ctx context.Context,
	input pluginhost.SourcePluginTenantLifecycleInput,
) (bool, string, error) {
	precondition, err := newLifecyclePrecondition()
	if err != nil {
		return false, "", err
	}
	return precondition.BeforeTenantDelete(ctx, input)
}

// newLifecyclePrecondition creates the plugin-owned lifecycle precondition
// checker from the tenant-counting dependency it requires.
func newLifecyclePrecondition() (*lifecycleprecondition.Checker, error) {
	return lifecycleprecondition.New(tenantsvc.ExistingCounter{})
}

// registerRoutes binds linapro-tenant-core routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil ||
		hostServices.Auth() == nil ||
		hostServices.BizCtx() == nil ||
		hostServices.Config() == nil {
		return gerror.New("linapro-tenant-core routes require host auth, bizctx, and config services")
	}
	var (
		membershipSvc     = membership.New(hostServices.BizCtx())
		resolverConfigSvc = resolverconfig.New()
		tenantPluginSvc   = tenantplugin.New(hostServices.BizCtx(), hostServices.PluginLifecycle())
		tenantSvc         = tenantsvc.New(hostServices.BizCtx(), resolverConfigSvc, tenantPluginSvc, hostServices.PluginLifecycle())
		resolverSvc       = resolver.New(hostServices.BizCtx(), membershipSvc)
	)
	providerSvc, err := provider.New(membershipSvc, resolverSvc, resolverConfigSvc, tenantPluginSvc)
	if err != nil {
		return err
	}
	pkgtenantcap.RegisterProvider(providerSvc)
	var (
		impersonateSvc = impersonate.New(hostServices.BizCtx(), hostServices.Config(), tenantSvc)
		routes         = registrar.Routes()
		middlewares    = routes.Middlewares()
	)
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		authCtrl := authcontroller.NewV1(hostServices.Auth(), membershipSvc, providerSvc)
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Bind(
				authCtrl.SelectTenant,
			)
		})
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			group.Bind(
				authCtrl.LoginTenants,
				authCtrl.SwitchTenant,
				platformcontroller.NewV1(tenantSvc, impersonateSvc),
				tenantcontroller.NewV1(tenantPluginSvc),
			)
		})
	})
	return nil
}
