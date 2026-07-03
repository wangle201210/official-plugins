// Package backend wires the linapro-uidentity-cas source plugin into the host
// plugin registry. It owns plugin route and managed scheduled-job handler
// registration while all UIdentity business behavior remains in plugin services
// instead of changing lina-core host contracts.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
	uidentitycas "lina-plugin-linapro-uidentity-cas"
	uidentitycontroller "lina-plugin-linapro-uidentity-cas/backend/internal/controller/uidentity"
	uidentityjobs "lina-plugin-linapro-uidentity-cas/backend/internal/service/jobs"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-uidentity-cas"
)

// init registers the linapro-uidentity-cas source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(uidentitycas.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Jobs().RegisterJobs(
		pluginhost.ExtensionPointJobsRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerManagedJobs,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerRoutes validates host dependencies before binding UIdentity CAS routes.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
		services    = registrar.Services()
	)
	if services == nil || services.BizCtx() == nil || services.Tenant() == nil {
		return gerror.New("linapro-uidentity-cas routes require host bizctx and tenant-filter services")
	}
	tenantFilter := services.Tenant().Filter()
	if tenantFilter == nil {
		return gerror.New("linapro-uidentity-cas routes require host tenant-filter service")
	}
	plugins := services.Plugins()
	if plugins == nil {
		return gerror.New("linapro-uidentity-cas routes require plugin-scoped config service")
	}
	configSvc := plugins.Config()
	if configSvc == nil {
		return gerror.New("linapro-uidentity-cas routes require plugin-scoped config service")
	}
	uidentitySvc := uidentitysvc.New(
		services.BizCtx(),
		configSvc,
		tenantFilter,
	)
	legacyController := uidentitycontroller.NewLegacy(uidentitySvc)
	registerLegacyRoutes(routes, middlewares, legacyController)
	if err := registerLegacyRouteInterceptors(
		registrar.GlobalMiddlewares(),
		middlewares,
		legacyController,
	); err != nil {
		return err
	}

	return nil
}

// registerManagedJobs contributes old uidentity/admin job registry entries as
// LinaPro managed scheduled-job handlers. Scheduling, persistence, logs, and
// trigger control stay in the host task-management module.
func registerManagedJobs(ctx context.Context, registrar pluginhost.JobsRegistrar) error {
	services := registrar.Services()
	if services == nil || services.BizCtx() == nil || services.Tenant() == nil {
		return gerror.New("linapro-uidentity-cas jobs require host bizctx and tenant-filter services")
	}
	tenantFilter := services.Tenant().Filter()
	if tenantFilter == nil {
		return gerror.New("linapro-uidentity-cas jobs require host tenant-filter service")
	}
	plugins := services.Plugins()
	if plugins == nil {
		return gerror.New("linapro-uidentity-cas jobs require plugin-scoped config service")
	}
	configSvc := plugins.Config()
	if configSvc == nil {
		return gerror.New("linapro-uidentity-cas jobs require plugin-scoped config service")
	}
	jobSvc := uidentityjobs.New(services.BizCtx(), configSvc, tenantFilter)
	return jobSvc.Register(ctx, registrar)
}
