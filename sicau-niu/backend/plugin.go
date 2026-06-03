// Package backend wires the sicau-niu source plugin into the host plugin
// registry. It registers the embedded plugin assets and binds three HTTP route
// surfaces under the plugin API prefix: a public WeChat player login route, a
// player-token-protected surface (phone binding, profile, college dropdown)
// guarded by the plugin-owned player-auth middleware, and an operator surface
// (college dictionary CRUD, player query) guarded by the host
// Auth+Tenancy+Permission chain. The plugin owns its WeChat-gateway,
// player-token, identity and college services; the service graph is constructed
// once at route-registration time from the plugin-scoped configuration.
package backend

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/pluginhost"
	pluginsicauniu "lina-plugin-sicau-niu"
	adminctrl "lina-plugin-sicau-niu/backend/internal/controller/admin"
	playerctrl "lina-plugin-sicau-niu/backend/internal/controller/player"
	"lina-plugin-sicau-niu/backend/internal/middleware"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
	tokensvc "lina-plugin-sicau-niu/backend/internal/service/token"
	wechatsvc "lina-plugin-sicau-niu/backend/internal/service/wechat"
)

// Plugin configuration keys and defaults.
const (
	// pluginID is the immutable identifier published by the embedded sicau-niu plugin.
	pluginID = "sicau-niu"
	// configKeyWeChatAppID is the plugin config key for the WeChat AppID.
	configKeyWeChatAppID = "wechat.appId"
	// configKeyWeChatSecret is the plugin config key for the WeChat AppSecret.
	configKeyWeChatSecret = "wechat.secret"
	// configKeyWeChatMock is the plugin config key for the WeChat mock switch.
	configKeyWeChatMock = "wechat.mock"
	// configKeyWeChatMockOpenid is the plugin config key for the mock openid.
	configKeyWeChatMockOpenid = "wechat.mockOpenid"
	// configKeyTokenSecret is the plugin config key for the player token secret.
	configKeyTokenSecret = "token.secret"
	// configKeyTokenTTL is the plugin config key for the player token TTL.
	configKeyTokenTTL = "token.ttl"
	// defaultTokenTTL is the fallback player token validity when config is blank.
	defaultTokenTTL = 168 * time.Hour
)

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

// registerRoutes builds the plugin service graph from configuration and binds
// the public, player and operator HTTP route surfaces. It returns an error when
// required services or configuration are missing so the failure surfaces at
// startup rather than at request time.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	services := registrar.Services()
	if services == nil || services.Config() == nil {
		return gerror.New("sicau-niu routes require host plugin config service")
	}

	tokenService, gateway, err := buildAuthDependencies(ctx, services.Config())
	if err != nil {
		return err
	}

	collegeService := collegesvc.New()
	identityService := identitysvc.New(gateway, tokenService, collegeService)
	playerAuth := middleware.NewPlayerAuth(tokenService)
	playerController := playerctrl.NewV1(identityService, collegeService)
	adminController := adminctrl.NewV1(collegeService, identityService)

	routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
		group.Group("/api/v1", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
			)

			// Public surface: WeChat player login, no authentication.
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Bind(playerController.Login)
			})

			// Player surface: guarded by the plugin player-auth middleware so each
			// player only reads and writes their own data.
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(playerAuth.Handle)
				group.Bind(
					playerController.BindPhone,
					playerController.GetProfile,
					playerController.UpdateProfile,
					playerController.CollegeOptions,
				)
			})

			// Operator surface: guarded by the host Auth+Tenancy+Permission chain;
			// per-route permission is declared on each DTO g.Meta tag.
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.Bind(
					adminController.ListColleges,
					adminController.CreateCollege,
					adminController.UpdateCollege,
					adminController.DeleteCollege,
					adminController.ListPlayers,
				)
			})
		})
	})
	return nil
}

// buildAuthDependencies constructs the player token service and WeChat gateway
// from the plugin-scoped configuration. It returns an error when the token
// secret is missing or the configuration cannot be read.
func buildAuthDependencies(
	ctx context.Context,
	config contract.ConfigService,
) (tokensvc.Service, wechatsvc.Gateway, error) {
	tokenSecret, err := config.String(ctx, configKeyTokenSecret, "")
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read token secret failed")
	}
	tokenTTL, err := config.Duration(ctx, configKeyTokenTTL, defaultTokenTTL)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read token ttl failed")
	}
	tokenService, err := tokensvc.New(tokensvc.Config{Secret: tokenSecret, TTL: tokenTTL})
	if err != nil {
		return nil, nil, err
	}

	wechatAppID, err := config.String(ctx, configKeyWeChatAppID, "")
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read wechat appId failed")
	}
	wechatSecret, err := config.String(ctx, configKeyWeChatSecret, "")
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read wechat secret failed")
	}
	wechatMock, err := config.Bool(ctx, configKeyWeChatMock, false)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read wechat mock failed")
	}
	wechatMockOpenid, err := config.String(ctx, configKeyWeChatMockOpenid, "")
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read wechat mockOpenid failed")
	}
	gateway := wechatsvc.New(wechatsvc.Config{
		AppID:      wechatAppID,
		Secret:     wechatSecret,
		Mock:       wechatMock,
		MockOpenid: wechatMockOpenid,
	})
	return tokenService, gateway, nil
}
