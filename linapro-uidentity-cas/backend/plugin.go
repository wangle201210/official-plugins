// Package backend wires the linapro-uidentity-cas source plugin into the host
// plugin registry. It owns plugin route and managed scheduled-job handler
// registration while all UIdentity business behavior remains in plugin services
// instead of changing lina-core host contracts.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability"
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
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(uidentitycas.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Cron().RegisterCron(
		pluginhost.ExtensionPointCronRegister,
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
	if services == nil || services.BizCtx() == nil || services.TenantFilter() == nil {
		return gerror.New("linapro-uidentity-cas routes require host bizctx and tenant-filter services")
	}
	scopedServices := capability.ServicesForPlugin(services, pluginID)
	if scopedServices.Config() == nil {
		return gerror.New("linapro-uidentity-cas routes require plugin-scoped config service")
	}
	uidentitySvc := uidentitysvc.New(
		scopedServices.BizCtx(),
		scopedServices.Config(),
		services.TenantFilter(),
	)
	legacyController := uidentitycontroller.NewLegacy(uidentitySvc)
	registerLegacyRoutes(routes, middlewares, legacyController)

	return nil
}

// registerManagedJobs contributes old uidentity/admin job registry entries as
// LinaPro managed scheduled-job handlers. Scheduling, persistence, logs, and
// trigger control stay in the host task-management module.
func registerManagedJobs(ctx context.Context, registrar pluginhost.CronRegistrar) error {
	services := registrar.Services()
	if services == nil || services.Config() == nil {
		return gerror.New("linapro-uidentity-cas jobs require plugin-scoped config service")
	}
	jobSvc := uidentityjobs.New(services.BizCtx(), services.Config(), services.TenantFilter())
	return jobSvc.Register(ctx, registrar)
}
