// Package backend wires the linapro-org-core source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	hostorgcap "lina-core/pkg/orgcap"
	"lina-core/pkg/pluginhost"
	orgcenter "lina-plugin-linapro-org-core"
	deptcontroller "lina-plugin-linapro-org-core/backend/internal/controller/dept"
	postcontroller "lina-plugin-linapro-org-core/backend/internal/controller/post"
	deptsvc "lina-plugin-linapro-org-core/backend/internal/service/dept"
	postsvc "lina-plugin-linapro-org-core/backend/internal/service/post"
	"lina-plugin-linapro-org-core/backend/provider/orgcapadapter"
)

// linapro-org-core plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-org-core"
)

// init registers the linapro-org-core source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(orgcenter.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerRoutes binds department and post management routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.I18n() == nil || hostServices.TenantFilter() == nil {
		return gerror.New("linapro-org-core routes require host i18n and tenant-filter services")
	}
	hostorgcap.RegisterProvider(orgcapadapter.New(hostServices.TenantFilter()))
	deptSvc := deptsvc.New(hostServices.TenantFilter())
	postSvc := postsvc.New(hostServices.I18n(), hostServices.TenantFilter())
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			group.Bind(
				deptcontroller.NewV1(deptSvc),
				postcontroller.NewV1(postSvc),
			)
		})
	})
	return nil
}
