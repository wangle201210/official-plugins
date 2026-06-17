// Package backend wires the media source plugin into the host plugin registry.
package backend

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
	mediaplugin "lina-plugin-media"
	mediacontroller "lina-plugin-media/backend/internal/controller/media"
	mediaopencontroller "lina-plugin-media/backend/internal/controller/mediaopen"
	collectionsvc "lina-plugin-media/backend/internal/service/collection"
	cronsvc "lina-plugin-media/backend/internal/service/cron"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// media plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "media"
)

// sharedCollectionSvc owns the media data collection TCP server lifecycle.
var sharedCollectionSvc = collectionsvc.New()

// sharedCronMu protects sharedCronSvc initialization across startup hooks.
var sharedCronMu sync.Mutex

// sharedCronSvc owns media plugin maintenance jobs.
var sharedCronSvc cronsvc.Service

// init registers the media source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(mediaplugin.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointSystemStarted,
		pluginhost.CallbackExecutionModeBlocking,
		startCollectionServer,
	); err != nil {
		panic(err)
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// startCollectionServer starts the plugin-owned net-flux compatible TCP server after host startup.
func startCollectionServer(ctx context.Context, payload pluginhost.HookPayload) error {
	if payload == nil || payload.Services() == nil {
		return gerror.New("media collection server requires host services")
	}
	plugins := payload.Services().Plugins()
	if plugins == nil {
		return gerror.New("media collection server requires host plugin config service")
	}
	configSvc := plugins.Config()
	if configSvc == nil {
		return gerror.New("media collection server requires host plugin config service")
	}
	cacheSvc := payload.Services().Cache()
	if cacheSvc == nil {
		return gerror.New("media collection server requires host cache service")
	}
	if err := sharedCollectionSvc.Start(ctx, configSvc, cacheSvc); err != nil {
		return err
	}
	mediaSvc, err := mediasvc.New(mediaBizCtxWithTietaOverlay(payload.Services().BizCtx()), cacheSvc)
	if err != nil {
		return err
	}
	return startMediaCron(ctx, mediaSvc)
}

// startMediaCron starts plugin maintenance jobs once for the shared source plugin instance.
func startMediaCron(ctx context.Context, mediaSvc mediasvc.Service) error {
	sharedCronMu.Lock()
	defer sharedCronMu.Unlock()
	if sharedCronSvc == nil {
		cronSvc, err := cronsvc.New(mediaSvc)
		if err != nil {
			return err
		}
		sharedCronSvc = cronSvc
	}
	return sharedCronSvc.Start(ctx)
}

// registerRoutes binds mediaopen routes through InnerApiAuth and management
// routes through the media-scoped LinaPro/Tieta dual-auth chain.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.Services()
	if hostServices == nil || hostServices.BizCtx() == nil {
		return gerror.New("media routes require host bizctx service")
	}
	cacheSvc := hostServices.Cache()
	if cacheSvc == nil {
		return gerror.New("media routes require host cache service")
	}
	plugins := hostServices.Plugins()
	if plugins == nil {
		return gerror.New("media routes require host plugin config service")
	}
	configSvc := plugins.Config()
	if configSvc == nil {
		return gerror.New("media routes require host plugin config service")
	}
	mediaSvc, err := mediasvc.New(mediaBizCtxWithTietaOverlay(hostServices.BizCtx()), cacheSvc)
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
		mediaRegisterCommonMiddlewares(group, middlewares)
		mediaRegisterAPIDocRoutes(group, mediaSvc)
		group.Middleware(mediaInnerAPIAuthMiddleware(configSvc))
		group.Bind(publicController)
	})
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		mediaRegisterCommonMiddlewares(group, middlewares)
		group.Middleware(
			mediaDualAuthMiddleware(middlewares.Auth(), mediaSvc),
			mediaSkipWhenTietaAuthenticated(middlewares.Tenancy()),
			mediaSkipWhenTietaAuthenticated(middlewares.Permission()),
			mediaSkipWhenTietaAuthenticated(mediaMarkHostGatePassed),
		)
		group.Bind(protectedController)
	})
	return nil
}

// mediaRegisterCommonMiddlewares installs host-published request middleware shared by media routes.
func mediaRegisterCommonMiddlewares(group pluginhost.RouteGroup, middlewares pluginhost.RouteMiddlewares) {
	group.Middleware(
		middlewares.NeverDoneCtx(),
		middlewares.HandlerResponse(),
		middlewares.CORS(),
		middlewares.RequestBodyLimit(),
		middlewares.Ctx(),
	)
}
