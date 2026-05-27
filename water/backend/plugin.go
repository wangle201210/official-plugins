// Package backend wires the water source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
	mediastrategy "lina-plugin-media/backend/provider/strategy"
	waterplugin "lina-plugin-water"
	watercontroller "lina-plugin-water/backend/internal/controller/water"
	watersvc "lina-plugin-water/backend/internal/service/water"
)

// water plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "water"
)

// init registers the water source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(waterplugin.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds water routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.Services()
	if hostServices == nil || hostServices.BizCtx() == nil || hostServices.Cache() == nil {
		return gerror.New("water routes require host bizctx and cache services")
	}
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	strategyResolver, err := mediastrategy.NewResolver(hostServices.BizCtx(), hostServices.Cache())
	if err != nil {
		return err
	}
	waterSvc, err := watersvc.New(hostServices.Cache(), strategyResolver)
	if err != nil {
		return err
	}
	controller, err := watercontroller.NewV1(waterSvc)
	if err != nil {
		return err
	}
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
			group.Bind(controller)
		})
	})
	return nil
}
