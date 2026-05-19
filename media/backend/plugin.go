// Package backend wires the media source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginhost"
	mediaplugin "lina-plugin-media"
	mediacontroller "lina-plugin-media/backend/internal/controller/media"
	mediaopencontroller "lina-plugin-media/backend/internal/controller/mediaopen"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// media plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "media"
)

// init registers the media source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(mediaplugin.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds media management routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.BizCtx() == nil {
		return gerror.New("media routes require host bizctx service")
	}
	cacheSvc := hostServices.Cache()
	if cacheSvc == nil {
		return gerror.New("media routes require host cache service")
	}
	mediaSvc, err := mediasvc.New(hostServices.BizCtx(), cacheSvc)
	if err != nil {
		return err
	}
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	publicController, err := mediaopencontroller.NewV1(mediaSvc)
	if err != nil {
		return err
	}
	protectedController, err := mediacontroller.NewV1(mediaSvc)
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
			group.Bind(publicController)
		})
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			group.Bind(protectedController)
		})
	})
	return nil
}
