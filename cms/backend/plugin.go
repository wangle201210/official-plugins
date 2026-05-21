// Package backend wires the CMS source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginhost"
	cmsplugin "lina-plugin-cms"
	cmscontroller "lina-plugin-cms/backend/internal/controller/cms"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// CMS plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded CMS plugin.
	pluginID = "cms"
)

// init registers the CMS source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(cmsplugin.EmbeddedFiles)
	plugin.Lifecycle().RegisterUninstallHandler(func(ctx context.Context, input pluginhost.SourcePluginUninstallInput) error {
		if !input.PurgeStorageData() {
			return nil
		}
		return cmssvc.PurgeStorageData(ctx)
	})
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds CMS public and management routes through the published
// host middleware set.
func registerRoutes(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.BizCtx() == nil {
		return gerror.New("cms routes require host bizctx service")
	}
	cmsSvc, err := cmssvc.New(hostServices.BizCtx())
	if err != nil {
		return err
	}
	controller, err := cmscontroller.NewV1(cmsSvc)
	if err != nil {
		return err
	}

	routes.Group("/", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.GET("/cms-site", controller.PublicFrontendPage)
		group.GET("/cms-site/assets/cms-site.css", controller.PublicFrontendStyle)
		group.GET("/cms-site/assets/*file", controller.PublicFrontendAsset)
		group.GET("/static/*file", controller.PublicFrontendStaticAsset)
		group.GET("/cms-site/message", controller.PublicFrontendMessagePage)
		group.POST("/cms-site/messages", controller.PublicFrontendMessage)
		group.GET("/cms-site/*path", controller.PublicFrontendPage)
	})

	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Bind(
				controller.PublicSite,
				controller.PublicCategoryList,
				controller.PublicArticleList,
				controller.PublicArticleGet,
				controller.PublicLinkList,
				controller.PublicSlideList,
				controller.PublicMessageCreate,
				controller.PublicMessageList,
			)
		})
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			group.Bind(
				controller.SiteGet,
				controller.SiteUpdate,
				controller.SiteClearData,
				controller.SiteLoadSampleData,
				controller.CategoryList,
				controller.CategoryCreate,
				controller.CategoryUpdate,
				controller.CategoryDelete,
				controller.ArticleList,
				controller.ArticleGet,
				controller.ArticleCreate,
				controller.ArticleUpdate,
				controller.ArticleDelete,
				controller.MessageList,
				controller.MessageUpdate,
				controller.MessageDelete,
				controller.LinkList,
				controller.LinkCreate,
				controller.LinkUpdate,
				controller.LinkDelete,
				controller.SlideList,
				controller.SlideCreate,
				controller.SlideUpdate,
				controller.SlideDelete,
			)
		})
	})
	return nil
}
