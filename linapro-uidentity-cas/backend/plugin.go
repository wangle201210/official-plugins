// Package backend wires the linapro-uidentity-cas source plugin into the host
// plugin registry. It owns plugin route registration and starts the plugin-local
// legacy sys_job scheduler, while all UIdentity business behavior remains in plugin
// services instead of changing lina-core host contracts.
package backend

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/pkg/plugin/capability"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/pluginhost"
	uidentitycas "lina-plugin-linapro-uidentity-cas"
	uidentitycontroller "lina-plugin-linapro-uidentity-cas/backend/internal/controller/uidentity"
	uidentitycron "lina-plugin-linapro-uidentity-cas/backend/internal/service/cron"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "linapro-uidentity-cas"
	// legacyJobBootstrapName is the plugin-local GoFrame cron entry that waits
	// until the plugin is enabled before loading legacy sys_job rows.
	legacyJobBootstrapName = "linapro-uidentity-cas-legacy-job-bootstrap"
	// legacyJobBootstrapPattern controls how often the plugin-local bootstrap retries.
	legacyJobBootstrapPattern = "@every 30s"
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

var (
	// uidentityJobCron is the shared plugin-local GoFrame scheduler.
	uidentityJobCron = uidentitycron.New(pluginID)
	// uidentityJobCronStartMu protects plugin-local scheduler startup state.
	uidentityJobCronStartMu sync.Mutex
	// uidentityJobCronStarted reports whether enabled legacy jobs were reloaded.
	uidentityJobCronStarted bool
	// uidentityJobCronStarting reports an in-flight legacy job reload.
	uidentityJobCronStarting bool
	// uidentityJobCronBootstrapRegistered reports whether the retry bootstrap was registered.
	uidentityJobCronBootstrapRegistered bool
)

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
	if err := ensureLegacyJobScheduler(ctx, scopedServices.PluginState()); err != nil {
		return err
	}
	uidentitySvc := uidentitysvc.New(
		scopedServices.BizCtx(),
		scopedServices.Config(),
		services.TenantFilter(),
		uidentityJobCron,
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
			group.POST("/uidentity/sms-codes", uidentityController.SmsSend)
			group.GET("/uidentity/legacy/health", uidentityController.LegacyHealth)
			group.POST("/uidentity/cas/login", uidentityController.CasLogin)
			group.POST("/uidentity/cas/password-logins", uidentityController.CasPasswordLogin)
			group.POST("/uidentity/cas/phone-logins", uidentityController.CasPhoneLogin)
			group.POST("/uidentity/cas/union-id-logins", uidentityController.CasUnionIDLogin)
			group.POST("/uidentity/cas/wechat-login-qrs", uidentityController.WechatLoginQR)
			group.POST("/uidentity/cas/wechat-login-callbacks", uidentityController.WechatLoginCallback)
			group.GET("/uidentity/cas/wechat-login-qrs/{state}/result", uidentityController.WechatLoginQRResult)
			group.POST("/uidentity/cas/service-tickets", uidentityController.CasServiceTicket)
			group.POST("/uidentity/cas/service-validations", uidentityController.CasServiceValidate)
			group.POST("/uidentity/legacy/cas/service-validations.xml", uidentityController.LegacyCASServiceValidateXML)
			group.DELETE("/uidentity/cas/tickets/{ticket}", uidentityController.CasTicketLogout)
			group.POST("/uidentity/runtime-tokens", uidentityController.RuntimeTokenIssue)
			group.GET("/uidentity/runtime-tokens/{accessToken}/user-info", uidentityController.RuntimeTokenInfo)
			group.POST("/uidentity/oauth/authorization-codes", uidentityController.OAuthAuthorizationCode)
			group.POST("/uidentity/oauth/access-tokens", uidentityController.OAuthAccessToken)
			group.GET("/uidentity/oauth/access-tokens/{accessToken}/user-info", uidentityController.OAuthAccessTokenInfo)
			group.POST("/uidentity/activations", uidentityController.ActivationStart)
			group.PUT("/uidentity/activations/{challengeId}/face", uidentityController.ActivationFace)
			group.PUT("/uidentity/activations/{challengeId}/password", uidentityController.ActivationPassword)
			group.PUT("/uidentity/activations/{challengeId}/phone", uidentityController.ActivationPhone)
			group.PUT("/uidentity/activations/{challengeId}/wechat", uidentityController.ActivationWechat)
			group.POST("/uidentity/activations/{challengeId}/wechat-states", uidentityController.ActivationWechatStateCreate)
			group.POST("/uidentity/activations/wechat-callbacks", uidentityController.ActivationWechatCallback)
			group.GET("/uidentity/activations/{challengeId}/state", uidentityController.ActivationState)
			group.POST("/uidentity/users/union-id-lookups", uidentityController.UserUnionIDLookup)
			group.POST("/uidentity/users/union-id-bindings", uidentityController.UserUnionIDBind)
			group.POST("/uidentity/users/wechat-rebind-callbacks", uidentityController.UserWechatRebindCallback)
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Middleware(
					middlewares.Auth(),
					middlewares.Tenancy(),
					middlewares.Permission(),
				)
				group.PUT("/uidentity/accounts/{id}/password", uidentityController.AccountPassword)
				group.POST("/uidentity/accounts/password-unlocks", uidentityController.AccountPasswordUnlock)
				group.POST("/uidentity/accounts/import-checks", uidentityController.AccountImportCheck)
				group.POST("/uidentity/accounts/imports", uidentityController.AccountImport)
				group.POST("/uidentity/legacy/uploads", uidentityController.LegacyUpload)
				group.GET("/uidentity/legacy/config/cas", uidentityController.LegacyCASConfig)
				group.GET("/uidentity/legacy/config/ldap", uidentityController.LegacyLDAPConfig)
				group.GET("/uidentity/legacy/config/oauth", uidentityController.LegacyOAuthConfig)
				group.GET("/uidentity/legacy/config/token", uidentityController.LegacyTokenConfig)
				group.GET("/uidentity/legacy/server-monitor", uidentityController.LegacyServerMonitor)
				group.GET("/uidentity/legacy/log-snapshots", uidentityController.LegacyLogSnapshot)
				group.POST("/uidentity/legacy/external-actions", uidentityController.LegacyExternalAction)
				group.POST("/uidentity/oauth/tokens", uidentityController.OAuthIssue)
				group.GET("/uidentity/stats", uidentityController.Stats)
				group.PUT("/uidentity/users/{number}/password", uidentityController.UserPasswordChange)
				group.PUT("/uidentity/users/{number}/phone", uidentityController.UserPhoneChange)
				group.PUT("/uidentity/users/{number}/email", uidentityController.UserEmailChange)
				group.PUT("/uidentity/users/{number}/qq", uidentityController.UserQQChange)
				group.DELETE("/uidentity/users/{number}/wechat", uidentityController.UserWechatUnbind)
				group.POST("/uidentity/users/{number}/wechat-rebind-states", uidentityController.UserWechatRebindStateCreate)
				group.GET("/uidentity/users/{number}/wechat-rebind-states/{state}", uidentityController.UserWechatRebindState)
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

// ensureLegacyJobScheduler starts immediately when possible and otherwise
// registers a plugin-local GoFrame retry entry that waits for enablement.
func ensureLegacyJobScheduler(ctx context.Context, pluginState plugincontract.PluginStateService) error {
	started, err := startLegacyJobScheduler(ctx, pluginState)
	if err != nil || started {
		return err
	}

	uidentityJobCronStartMu.Lock()
	defer uidentityJobCronStartMu.Unlock()
	if uidentityJobCronStarted || uidentityJobCronBootstrapRegistered {
		return nil
	}
	if gcron.Search(legacyJobBootstrapName) != nil {
		uidentityJobCronBootstrapRegistered = true
		return nil
	}
	_, err = gcron.AddSingleton(ctx, legacyJobBootstrapPattern, func(jobCtx context.Context) {
		started, startErr := startLegacyJobScheduler(jobCtx, pluginState)
		if startErr != nil {
			return
		}
		if started {
			gcron.Remove(legacyJobBootstrapName)
		}
	}, legacyJobBootstrapName)
	if err != nil {
		return err
	}
	uidentityJobCronBootstrapRegistered = true
	return nil
}

// startLegacyJobScheduler reloads plugin-owned enabled sys_job rows into the
// shared GoFrame cron instance after host services become available.
func startLegacyJobScheduler(ctx context.Context, pluginState plugincontract.PluginStateService) (bool, error) {
	if pluginState == nil {
		return false, nil
	}
	if !pluginState.IsEnabledAuthoritative(ctx, pluginID) {
		return false, nil
	}
	uidentityJobCronStartMu.Lock()
	if uidentityJobCronStarted || uidentityJobCronStarting {
		uidentityJobCronStartMu.Unlock()
		return uidentityJobCronStarted, nil
	}
	uidentityJobCronStarting = true
	uidentityJobCronStartMu.Unlock()

	err := uidentityJobCron.Start(ctx, pluginState, nil)

	uidentityJobCronStartMu.Lock()
	defer uidentityJobCronStartMu.Unlock()
	uidentityJobCronStarting = false
	if err != nil {
		return false, err
	}
	uidentityJobCronStarted = true
	uidentityJobCronBootstrapRegistered = false
	return true, nil
}
