// Package backend wires the sicau-niu source plugin into the host plugin registry.
// It registers the embedded plugin assets and binds the sample HTTP routes: one
// plain-text public portal route, one JSON public ping, and one permission-protected
// cattle list. The plugin owns no database, cron, or lifecycle resources, so this
// entry stays intentionally small.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/pluginhost"
	pluginsicauniu "lina-plugin-sicau-niu"
	niuctrl "lina-plugin-sicau-niu/backend/internal/controller/niu"
	niusvc "lina-plugin-sicau-niu/backend/internal/service/niu"
)

// pluginID is the immutable identifier published by the embedded sicau-niu plugin.
const pluginID = "sicau-niu"

// init registers the embedded sicau-niu source plugin and its route callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginsicauniu.EmbeddedFiles)
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

// registerRoutes binds the sicau-niu HTTP routes using the published host
// middleware directory so plugin traffic follows the same governance chain as
// host-owned APIs.
func registerRoutes(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
	)
	niuController := niuctrl.NewV1(niusvc.New())

	routes.Group("/portal/sicau-niu", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
		)
		group.GET("/ping", servePortalPing)
	})

	routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
		group.Group("/api/v1", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
			)

			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Bind(niuController.Ping)
			})

			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.Bind(niuController.List)
			})
		})
	})
	return nil
}

// servePortalPing returns a plugin-owned public route response outside the
// reserved API namespace so route-boundary checks can verify host fallback does
// not claim source-plugin public paths.
func servePortalPing(request *ghttp.Request) {
	request.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	request.Response.Write("sicau-niu-public-pong")
}
