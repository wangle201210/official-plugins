// Package backend wires the linapro-uidentity-cas source plugin into the host plugin registry.
package backend

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginhost"
	uidentitycas "lina-plugin-linapro-uidentity-cas"
	uidentitycontroller "lina-plugin-linapro-uidentity-cas/backend/internal/controller/uidentity"
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
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// registerRoutes validates host dependencies before binding UIdentity CAS routes.
func registerRoutes(_ context.Context, registrar pluginhost.HTTPRegistrar) error {
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
	uidentityController := uidentitycontroller.NewV1(uidentitySvc)
	routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
		group.Group("/api/v1", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
			)
			group.POST("/uidentity/password-challenges", uidentityController.AccountPasswordChallenge)
			group.POST("/uidentity/password-challenges/{challengeId}/phone", uidentityController.AccountPasswordPhoneVerify)
			group.PUT("/uidentity/password-challenges/{challengeId}/password", uidentityController.AccountPasswordSelfReset)
			group.POST("/uidentity/cas/login", uidentityController.CasLogin)
			group.POST("/uidentity/cas/password-logins", uidentityController.CasPasswordLogin)
			group.POST("/uidentity/cas/phone-logins", uidentityController.CasPhoneLogin)
			group.POST("/uidentity/cas/union-id-logins", uidentityController.CasUnionIDLogin)
			group.POST("/uidentity/cas/service-tickets", uidentityController.CasServiceTicket)
			group.POST("/uidentity/cas/service-validations", uidentityController.CasServiceValidate)
			group.DELETE("/uidentity/cas/tickets/{ticket}", uidentityController.CasTicketLogout)
			group.POST("/uidentity/runtime-tokens", uidentityController.RuntimeTokenIssue)
			group.GET("/uidentity/runtime-tokens/{accessToken}/user-info", uidentityController.RuntimeTokenInfo)
			group.POST("/uidentity/activations", uidentityController.ActivationStart)
			group.PUT("/uidentity/activations/{challengeId}/face", uidentityController.ActivationFace)
			group.PUT("/uidentity/activations/{challengeId}/password", uidentityController.ActivationPassword)
			group.PUT("/uidentity/activations/{challengeId}/phone", uidentityController.ActivationPhone)
			group.PUT("/uidentity/activations/{challengeId}/wechat", uidentityController.ActivationWechat)
			group.GET("/uidentity/activations/{challengeId}/state", uidentityController.ActivationState)
			group.POST("/uidentity/users/union-id-lookups", uidentityController.UserUnionIDLookup)
			group.POST("/uidentity/users/union-id-bindings", uidentityController.UserUnionIDBind)
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.PUT("/uidentity/accounts/{id}/password", uidentityController.AccountPassword)
				group.POST("/uidentity/oauth/tokens", uidentityController.OAuthIssue)
				group.GET("/uidentity/stats", uidentityController.Stats)
				group.PUT("/uidentity/users/{number}/password", uidentityController.UserPasswordChange)
				group.PUT("/uidentity/users/{number}/phone", uidentityController.UserPhoneChange)
				group.PUT("/uidentity/users/{number}/email", uidentityController.UserEmailChange)
				group.PUT("/uidentity/users/{number}/qq", uidentityController.UserQQChange)
				group.DELETE("/uidentity/users/{number}/wechat", uidentityController.UserWechatUnbind)
				group.GET("/uidentity/users/{number}", uidentityController.UserInfo)
				group.GET("/uidentity/users/{number}/cas-login-logs", uidentityController.UserLoginLogs)
				group.GET("/uidentity/users/{number}/applications", uidentityController.UserApplications)
				group.GET("/uidentity/users/{number}/account-app-roles", uidentityController.UserAppRoles)
				group.POST("/uidentity/users/{number}/account-app-roles", uidentityController.UserAppRoleCreate)
				group.PUT("/uidentity/users/{number}/account-app-roles/{id}", uidentityController.UserAppRoleUpdate)
				group.GET("/uidentity/{resource}", uidentityController.ResourceList)
				group.POST("/uidentity/{resource}", uidentityController.ResourceCreate)
				group.GET("/uidentity/{resource}/{id}", uidentityController.ResourceGet)
				group.PUT("/uidentity/{resource}/{id}", uidentityController.ResourceUpdate)
				group.DELETE("/uidentity/{resource}/{ids}", uidentityController.ResourceDelete)
			})
		})
	})

	return nil
}
