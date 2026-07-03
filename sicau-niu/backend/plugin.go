// Package backend wires the sicau-niu source plugin into the host plugin
// registry. It registers the embedded plugin assets, binds the plugin HTTP route
// surfaces, and contributes the background IOT locator refresh job. Player
// requests read only local plugin tables; the 1-minute locator cron refreshes
// physical iron-cow coordinates from the external IOT platform on the primary
// node. The plugin owns its WeChat-gateway, player-token, identity, college,
// cattle, card, ranking, honor and locator-refresh services; each service graph
// is constructed once at callback registration time from plugin-scoped
// configuration.
package backend

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginhost"
	pluginsicauniu "lina-plugin-sicau-niu"
	adminctrl "lina-plugin-sicau-niu/backend/internal/controller/admin"
	playerctrl "lina-plugin-sicau-niu/backend/internal/controller/player"
	recordctrl "lina-plugin-sicau-niu/backend/internal/controller/record"
	settlementctrl "lina-plugin-sicau-niu/backend/internal/controller/settlement"
	wallctrl "lina-plugin-sicau-niu/backend/internal/controller/wall"
	"lina-plugin-sicau-niu/backend/internal/middleware"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
	grasssocialsvc "lina-plugin-sicau-niu/backend/internal/service/grasssocial"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
	rankingsvc "lina-plugin-sicau-niu/backend/internal/service/ranking"
	recordsvc "lina-plugin-sicau-niu/backend/internal/service/record"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
	tokensvc "lina-plugin-sicau-niu/backend/internal/service/token"
	wallsvc "lina-plugin-sicau-niu/backend/internal/service/wall"
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
	// configKeyLBSThresholdMeters is the plugin config key for the LBS activation
	// distance threshold in meters.
	configKeyLBSThresholdMeters = "activation.lbsThresholdMeters"
	// defaultLBSThresholdMeters is the fallback LBS activation distance threshold
	// in meters when config is absent.
	defaultLBSThresholdMeters = 50
	// configKeyActivationDailyAttemptLimit is the plugin config key for the daily
	// activation attempt cap, counting failed photo check-ins as well.
	configKeyActivationDailyAttemptLimit = "activation.dailyAttemptLimit"
	// defaultActivationDailyAttemptLimit is the fallback daily activation attempt
	// cap when config is absent.
	defaultActivationDailyAttemptLimit = 20
	// configKeyActivationMaxSpeedMps is the plugin config key for the maximum
	// plausible movement speed between successive check-ins, in meters per second.
	configKeyActivationMaxSpeedMps = "activation.maxSpeedMps"
	// defaultActivationMaxSpeedMps is the fallback movement speed ceiling in
	// meters per second when config is absent.
	defaultActivationMaxSpeedMps = 25
	// configKeyPosterCampusBadge is the plugin config key for the poster campus
	// anniversary badge text.
	configKeyPosterCampusBadge = "poster.campusBadge"
	// configKeyCheckinMinAmount and configKeyCheckinMaxAmount are the plugin config
	// keys for the daily check-in grass grant range.
	configKeyCheckinMinAmount = "checkin.minAmount"
	configKeyCheckinMaxAmount = "checkin.maxAmount"
	// defaultCheckinMinAmount and defaultCheckinMaxAmount are the fallback daily
	// check-in grant bounds when config is absent.
	defaultCheckinMinAmount = 20
	defaultCheckinMaxAmount = 50
	// configKeyStealDailyTargets, configKeyStealDailyLimit, configKeyStealMinAmount
	// and configKeyStealMaxAmount are the plugin config keys for the steal feature.
	configKeyStealDailyTargets = "steal.dailyTargets"
	configKeyStealDailyLimit   = "steal.dailyLimit"
	configKeyStealMinAmount    = "steal.minAmount"
	configKeyStealMaxAmount    = "steal.maxAmount"
	// defaultStealDailyTargets, defaultStealDailyLimit, defaultStealMinAmount and
	// defaultStealMaxAmount are the fallback steal settings when config is absent.
	defaultStealDailyTargets = 12
	defaultStealDailyLimit   = 5
	defaultStealMinAmount    = 5
	defaultStealMaxAmount    = 20
	// configKeyGiftDailyLimit and configKeyGiftMinAmount are the plugin config keys
	// for the gift feature.
	configKeyGiftDailyLimit = "gift.dailyLimit"
	configKeyGiftMinAmount  = "gift.minAmount"
	// defaultGiftDailyLimit and defaultGiftMinAmount are the fallback gift settings
	// when config is absent.
	defaultGiftDailyLimit = 12
	defaultGiftMinAmount  = 12
	// configKeyIronBonusThresholdMeters is the plugin config key for the feeding
	// iron-cow proximity bonus distance threshold in meters.
	configKeyIronBonusThresholdMeters = "ironBonus.thresholdMeters"
	// defaultIronBonusThresholdMeters is the fallback iron-bonus distance threshold
	// in meters when config is absent.
	defaultIronBonusThresholdMeters = 12
	// configKeyRankingTopN is the plugin config key for the leaderboard Top-N cap.
	configKeyRankingTopN = "ranking.topN"
	// defaultRankingTopN is the fallback leaderboard Top-N cap when config is absent.
	defaultRankingTopN = 100
	// configKeyMiniappURL is the plugin config key for the H5 wall's return-to-
	// mini-program URL.
	configKeyMiniappURL = "miniapp.url"
	// configKeyAnomalyFeedDaily, configKeyAnomalyStealDaily and configKeyAnomalyLimit
	// are the plugin config keys for the settlement anomaly alert thresholds and cap.
	configKeyAnomalyFeedDaily  = "anomaly.feedDailyThreshold"
	configKeyAnomalyStealDaily = "anomaly.stealDailyThreshold"
	configKeyAnomalyLimit      = "anomaly.listLimit"
	// defaultAnomalyFeedDaily, defaultAnomalyStealDaily and defaultAnomalyListLimit
	// are the fallback anomaly thresholds and cap when config is absent.
	defaultAnomalyFeedDaily  = 100
	defaultAnomalyStealDaily = 5
	defaultAnomalyListLimit  = 200
	// configKeyIOTLocatorBaseURL is the plugin config key for the external IOT
	// positioning platform base URL.
	configKeyIOTLocatorBaseURL = "niu.baseUrl"
	// configKeyIOTLocatorKey is the plugin config key for the external IOT key.
	configKeyIOTLocatorKey = "niu.key"
	// configKeyIOTLocatorSecret is the plugin config key for the external IOT secret.
	configKeyIOTLocatorSecret = "niu.secret"
	// configKeyIOTLocatorPageSize is the plugin config key for locator-list page size.
	configKeyIOTLocatorPageSize = "niu.pageSize"
	// configKeyIOTLocatorRefreshInterval is the plugin config key for locator
	// refresh interval.
	configKeyIOTLocatorRefreshInterval = "niu.refreshInterval"
	// configKeyIOTLocatorTokenTTL is the plugin config key for the external token TTL.
	configKeyIOTLocatorTokenTTL = "niu.tokenTTL"
	// defaultIOTLocatorRefreshInterval is the required default iron-cow location
	// refresh cadence.
	defaultIOTLocatorRefreshInterval = time.Minute
	// defaultIOTLocatorTokenTTL follows the IOT platform document's three-day token validity.
	defaultIOTLocatorTokenTTL = 72 * time.Hour
	// defaultIOTLocatorPageSize bounds one external locator-list request.
	defaultIOTLocatorPageSize = 100
	// defaultIOTLocatorHTTPTimeout bounds one external IOT HTTP request so a slow
	// platform response does not overlap the next 1-minute refresh indefinitely.
	defaultIOTLocatorHTTPTimeout = 8 * time.Second
	// ironLocationRefreshJobName identifies the iron-cow IOT refresh job declaration.
	ironLocationRefreshJobName = "sicau-niu-iron-location-refresh"
	// ironLocationRefreshJobDisplayName is the English source title for the locator refresh job.
	ironLocationRefreshJobDisplayName = "Sicau Niu Iron Location Refresh"
	// ironLocationRefreshJobDescription is the English source description for the locator refresh job.
	ironLocationRefreshJobDescription = "Refreshes registered iron-cow locator coordinates from the IOT positioning platform."
)

// init registers the embedded sicau-niu source plugin and its route callbacks.
func init() {
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(pluginsicauniu.EmbeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	); err != nil {
		panic(err)
	}
	if err := plugin.Jobs().RegisterJobs(
		pluginhost.ExtensionPointJobsRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerIronLocationJob,
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
	if services == nil {
		return gerror.New("sicau-niu routes require host plugin config service")
	}
	plugins := services.Plugins()
	if plugins == nil {
		return gerror.New("sicau-niu routes require host plugin config service")
	}
	configSvc := plugins.Config()
	if configSvc == nil {
		return gerror.New("sicau-niu routes require host plugin config service")
	}

	tokenService, gateway, err := buildAuthDependencies(ctx, configSvc)
	if err != nil {
		return err
	}

	activationConfig, err := buildActivationConfig(ctx, configSvc)
	if err != nil {
		return err
	}

	grassConfig, feedingConfig, grassSocialConfig, err := buildGrassConfigs(ctx, configSvc)
	if err != nil {
		return err
	}

	rankingConfig, err := buildRankingConfig(ctx, configSvc)
	if err != nil {
		return err
	}

	miniappURL, err := configSvc.String(ctx, configKeyMiniappURL, "")
	if err != nil {
		return gerror.Wrap(err, "sicau-niu read miniapp url failed")
	}
	settlementConfig, err := buildSettlementConfig(ctx, configSvc)
	if err != nil {
		return err
	}
	dailyAttemptLimit, err := configSvc.Int(ctx, configKeyActivationDailyAttemptLimit, defaultActivationDailyAttemptLimit)
	if err != nil {
		return gerror.Wrap(err, "sicau-niu read activation daily attempt limit failed")
	}
	maxSpeedMps, err := configSvc.Int(ctx, configKeyActivationMaxSpeedMps, defaultActivationMaxSpeedMps)
	if err != nil {
		return gerror.Wrap(err, "sicau-niu read activation max speed failed")
	}
	rulesService := rulessvc.New(&rulessvc.RuleSet{
		ActivationLBSThresholdMeters: int(activationConfig.LBSThresholdMeters),
		ActivationDailyAttemptLimit:  dailyAttemptLimit,
		ActivationMaxSpeedMps:        maxSpeedMps,
		PosterCampusBadge:            activationConfig.CampusBadge,
		CheckinMinAmount:             grassConfig.CheckinMinAmount,
		CheckinMaxAmount:             grassConfig.CheckinMaxAmount,
		StealDailyTargets:            grassSocialConfig.StealDailyTargets,
		StealDailyLimit:              grassSocialConfig.StealDailyLimit,
		StealMinAmount:               grassSocialConfig.StealMinAmount,
		StealMaxAmount:               grassSocialConfig.StealMaxAmount,
		GiftDailyLimit:               grassSocialConfig.GiftDailyLimit,
		GiftMinAmount:                grassSocialConfig.GiftMinAmount,
		IronBonusThresholdMeters:     int(feedingConfig.IronBonusThresholdMeters),
		RankingTopN:                  rankingConfig.TopN,
		AnomalyFeedDailyThreshold:    settlementConfig.FeedDailyThreshold,
		AnomalyStealDailyThreshold:   settlementConfig.StealDailyThreshold,
		AnomalyListLimit:             settlementConfig.AnomalyLimit,
		MiniappURL:                   miniappURL,
	})

	collegeService := collegesvc.New()
	identityService := identitysvc.New(gateway, tokenService, collegeService)
	cattleService := cattlesvc.New(collegeService)
	cardService := cardsvc.New(cattleService)
	activationService := activationsvc.New(
		identityService,
		activationsvc.NewBasicPosterRenderer(),
		rulesService,
		activationConfig,
	)
	grassService := grasssvc.New(rulesService, grassConfig)
	feedingService := feedingsvc.New(grassService, feedingsvc.NewStoredIronLocation(), rulesService, feedingConfig)
	grassSocialService := grasssocialsvc.New(grassService, rulesService, grassSocialConfig)
	rankingService := rankingsvc.New(rulesService, rankingConfig)
	honorService := honorsvc.New(
		honorsvc.NewBasicCertRenderer(),
		rulesService,
		honorsvc.Config{CampusBadge: activationConfig.CampusBadge},
	)
	wallService := wallsvc.New(rulesService, wallsvc.Config{MiniappURL: miniappURL})
	settlementService := settlementsvc.New(rulesService, settlementConfig)
	recordService := recordsvc.New()
	playerAuth := middleware.NewPlayerAuth(tokenService)
	playerController := playerctrl.NewV1(
		identityService,
		collegeService,
		activationService,
		grassService,
		feedingService,
		grassSocialService,
		rankingService,
		honorService,
	)
	adminController := adminctrl.NewV1(collegeService, identityService, cattleService, cardService, honorService)
	wallController := wallctrl.NewV1(wallService)
	settlementController := settlementctrl.NewV1(settlementService, rankingService, rulesService)
	recordController := recordctrl.NewV1(recordService)

	routes.Group(routes.APIPrefix(), func(group pluginhost.RouteGroup) {
		group.Group("/api/v1", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.NeverDoneCtx(),
				middlewares.HandlerResponse(),
				middlewares.CORS(),
				middlewares.RequestBodyLimit(),
				middlewares.Ctx(),
			)

			// Public surface: WeChat player login and the C6 public memorial wall,
			// no authentication. The wall endpoints expose only nicknames and
			// activity information and never return privacy fields.
			group.Group("/", func(group pluginhost.RouteGroup) {
				group.Bind(playerController.Login)
				group.Bind(
					wallController.FirstActivators,
					wallController.Highlights,
					wallController.Stats,
					wallController.Config,
				)
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
					playerController.VisibleNiu,
					playerController.Activate,
					playerController.Collection,
					playerController.Poster,
					playerController.Checkin,
					playerController.GrassAccount,
					playerController.Feed,
					playerController.FeedingTrail,
					playerController.StealTargets,
					playerController.Steal,
					playerController.Gift,
					playerController.Messages,
					playerController.MarkMessageRead,
					playerController.FeedRanking,
					playerController.CollegeRanking,
					playerController.FriendRanking,
					playerController.PlayerHonors,
					playerController.Certificate,
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
					adminController.ListNiu,
					adminController.GetNiu,
					adminController.CreateNiu,
					adminController.UpdateNiu,
					adminController.DeleteNiu,
					adminController.ListIron,
					adminController.CreateIron,
					adminController.UpdateIron,
					adminController.DeleteIron,
					adminController.ListCard,
					adminController.GetCard,
					adminController.CreateCard,
					adminController.UpdateCard,
					adminController.DeleteCard,
					adminController.ListQuote,
					adminController.CreateQuote,
					adminController.UpdateQuote,
					adminController.DeleteQuote,
					adminController.ListHonor,
					adminController.GetHonor,
					adminController.CreateHonor,
					adminController.UpdateHonor,
					adminController.DeleteHonor,
				)
				group.Bind(
					settlementController.Dashboard,
					settlementController.ExportPlayers,
					settlementController.CertificateOptions,
					settlementController.IssueCertificates,
					settlementController.RiskDeviceClusters,
					settlementController.CreateArchive,
					settlementController.ListArchives,
					settlementController.Activity,
					settlementController.RiskAnomalies,
					settlementController.Rules,
					settlementController.UpdateRules,
					settlementController.FeedRanking,
					settlementController.FriendRanking,
					settlementController.CollegeRanking,
				)
				group.Bind(
					recordController.Feedings,
					recordController.Steals,
					recordController.Gifts,
					recordController.Checkins,
					recordController.Activations,
					recordController.ActivationAttempts,
					recordController.GrassTxns,
				)
			})
		})
	})
	return nil
}

// registerIronLocationJob contributes the primary-node IOT locator refresh job
// when niu.key and niu.secret are configured. Missing credentials keep
// development and test deployments on stored/mock coordinates without registering
// an external polling job.
func registerIronLocationJob(ctx context.Context, registrar pluginhost.JobsRegistrar) error {
	if registrar == nil {
		return gerror.New("sicau-niu iron-location job requires registrar")
	}
	configSvc, err := pluginConfigFromServices(registrar.Services(), "sicau-niu iron-location job")
	if err != nil {
		return err
	}
	enabled, refresher, interval, err := buildIronLocationRefresh(ctx, configSvc)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	return registrar.AddWithMetadata(
		ctx,
		"@every "+interval.String(),
		ironLocationRefreshJobName,
		ironLocationRefreshJobDisplayName,
		ironLocationRefreshJobDescription,
		func(ctx context.Context) error {
			return refreshIronLocations(ctx, registrar.IsPrimaryNode(), refresher)
		},
	)
}

// refreshIronLocations runs one primary-node refresh cycle. The job registrar
// already guards disabled plugins; this function keeps primary-node gating and
// refresher dependency checks testable without touching host scheduler state.
func refreshIronLocations(ctx context.Context, primaryNode bool, refresher feedingsvc.IronLocationRefresher) error {
	if !primaryNode {
		return nil
	}
	if refresher == nil {
		return gerror.New("sicau-niu iron-location refresh requires refresher")
	}
	_, err := refresher.Refresh(ctx)
	return err
}

// pluginConfigFromServices extracts the plugin-scoped ConfigService from host
// callback services and returns explicit setup errors for invalid callback wiring.
func pluginConfigFromServices(services capability.Services, purpose string) (plugincap.ConfigService, error) {
	if services == nil {
		return nil, gerror.New(purpose + " requires plugin config service")
	}
	plugins := services.Plugins()
	if plugins == nil {
		return nil, gerror.New(purpose + " requires plugin config service")
	}
	configSvc := plugins.Config()
	if configSvc == nil {
		return nil, gerror.New(purpose + " requires plugin config service")
	}
	return configSvc, nil
}

// buildIronLocationRefresh reads IOT locator config and constructs the background
// refresher. It returns enabled=false when no credentials are configured.
func buildIronLocationRefresh(
	ctx context.Context,
	config plugincap.ConfigService,
) (enabled bool, refresher feedingsvc.IronLocationRefresher, interval time.Duration, err error) {
	key, err := config.String(ctx, configKeyIOTLocatorKey, "")
	if err != nil {
		return false, nil, 0, gerror.Wrap(err, "sicau-niu read IOT locator key failed")
	}
	secret, err := config.String(ctx, configKeyIOTLocatorSecret, "")
	if err != nil {
		return false, nil, 0, gerror.Wrap(err, "sicau-niu read IOT locator secret failed")
	}
	if key == "" && secret == "" {
		return false, nil, 0, nil
	}

	baseURL, err := config.String(ctx, configKeyIOTLocatorBaseURL, "")
	if err != nil {
		return false, nil, 0, gerror.Wrap(err, "sicau-niu read IOT locator baseUrl failed")
	}
	pageSize, err := config.Int(ctx, configKeyIOTLocatorPageSize, defaultIOTLocatorPageSize)
	if err != nil {
		return false, nil, 0, gerror.Wrap(err, "sicau-niu read IOT locator pageSize failed")
	}
	interval, err = config.Duration(ctx, configKeyIOTLocatorRefreshInterval, defaultIOTLocatorRefreshInterval)
	if err != nil {
		return false, nil, 0, gerror.Wrap(err, "sicau-niu read IOT locator refreshInterval failed")
	}
	if interval < defaultIOTLocatorRefreshInterval {
		interval = defaultIOTLocatorRefreshInterval
	}
	tokenTTL, err := config.Duration(ctx, configKeyIOTLocatorTokenTTL, defaultIOTLocatorTokenTTL)
	if err != nil {
		return false, nil, 0, gerror.Wrap(err, "sicau-niu read IOT locator tokenTTL failed")
	}

	refresher, err = feedingsvc.NewIOTIronLocationRefresher(feedingsvc.IronLocationConfig{
		BaseURL:  baseURL,
		Key:      key,
		Secret:   secret,
		PageSize: pageSize,
		TokenTTL: tokenTTL,
	}, &http.Client{Timeout: defaultIOTLocatorHTTPTimeout})
	if err != nil {
		return false, nil, 0, err
	}
	return true, refresher, interval, nil
}

// buildAuthDependencies constructs the player token service and WeChat gateway
// from the plugin-scoped configuration. Missing token.secret is allowed so
// unrelated source-plugin deployments can still start; player-token operations
// continue returning the token-secret business error until the secret is set.
func buildAuthDependencies(
	ctx context.Context,
	config plugincap.ConfigService,
) (tokensvc.Service, wechatsvc.Gateway, error) {
	tokenSecret, err := config.String(ctx, configKeyTokenSecret, "")
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read token secret failed")
	}
	tokenTTL, err := config.Duration(ctx, configKeyTokenTTL, defaultTokenTTL)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "sicau-niu read token ttl failed")
	}
	var tokenService tokensvc.Service
	if strings.TrimSpace(tokenSecret) == "" {
		logger.Warningf(ctx, "sicau-niu %s is not configured; player login and player-authenticated routes will return configuration errors until it is set", configKeyTokenSecret)
		tokenService = tokensvc.NewUnconfigured()
	} else {
		tokenService, err = tokensvc.New(tokensvc.Config{Secret: tokenSecret, TTL: tokenTTL})
		if err != nil {
			return nil, nil, err
		}
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

// buildActivationConfig reads the plain-value activation configuration (LBS
// distance threshold and poster campus badge) from the plugin-scoped
// configuration. It returns an error when the configuration cannot be read so the
// failure surfaces at startup.
func buildActivationConfig(
	ctx context.Context,
	config plugincap.ConfigService,
) (activationsvc.Config, error) {
	thresholdMeters, err := config.Int(ctx, configKeyLBSThresholdMeters, defaultLBSThresholdMeters)
	if err != nil {
		return activationsvc.Config{}, gerror.Wrap(err, "sicau-niu read lbs threshold failed")
	}
	campusBadge, err := config.String(ctx, configKeyPosterCampusBadge, "")
	if err != nil {
		return activationsvc.Config{}, gerror.Wrap(err, "sicau-niu read poster campus badge failed")
	}
	return activationsvc.Config{
		LBSThresholdMeters: float64(thresholdMeters),
		CampusBadge:        campusBadge,
	}, nil
}

// buildRankingConfig reads the plain-value C5 leaderboard configuration (the
// Top-N cap) from the plugin-scoped configuration. It returns an error when the
// configuration cannot be read so the failure surfaces at startup rather than at
// request time. A non-positive Top-N falls back to the service default.
func buildRankingConfig(
	ctx context.Context,
	config plugincap.ConfigService,
) (rankingsvc.Config, error) {
	topN, err := config.Int(ctx, configKeyRankingTopN, defaultRankingTopN)
	if err != nil {
		return rankingsvc.Config{}, gerror.Wrap(err, "sicau-niu read ranking topN failed")
	}
	return rankingsvc.Config{TopN: topN}, nil
}

// buildSettlementConfig reads the plain-value settlement anomaly configuration (the
// per-day feed/steal thresholds and the alert list cap) from the plugin-scoped
// configuration. It returns an error when the configuration cannot be read so the
// failure surfaces at startup. Non-positive values fall back to the service defaults.
func buildSettlementConfig(
	ctx context.Context,
	config plugincap.ConfigService,
) (settlementsvc.Config, error) {
	feedDaily, err := config.Int(ctx, configKeyAnomalyFeedDaily, defaultAnomalyFeedDaily)
	if err != nil {
		return settlementsvc.Config{}, gerror.Wrap(err, "sicau-niu read anomaly feedDailyThreshold failed")
	}
	stealDaily, err := config.Int(ctx, configKeyAnomalyStealDaily, defaultAnomalyStealDaily)
	if err != nil {
		return settlementsvc.Config{}, gerror.Wrap(err, "sicau-niu read anomaly stealDailyThreshold failed")
	}
	listLimit, err := config.Int(ctx, configKeyAnomalyLimit, defaultAnomalyListLimit)
	if err != nil {
		return settlementsvc.Config{}, gerror.Wrap(err, "sicau-niu read anomaly listLimit failed")
	}
	return settlementsvc.Config{
		FeedDailyThreshold:  feedDaily,
		StealDailyThreshold: stealDaily,
		AnomalyLimit:        listLimit,
	}, nil
}

// buildGrassConfigs reads the plain-value C4 configuration for the grass, feeding
// and grass-social capabilities from the plugin-scoped configuration. It returns
// an error when any configuration value cannot be read so the failure surfaces at
// startup rather than at request time.
func buildGrassConfigs(
	ctx context.Context,
	config plugincap.ConfigService,
) (grasssvc.Config, feedingsvc.Config, grasssocialsvc.Config, error) {
	checkinMin, err := config.Int(ctx, configKeyCheckinMinAmount, defaultCheckinMinAmount)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read checkin minAmount failed")
	}
	checkinMax, err := config.Int(ctx, configKeyCheckinMaxAmount, defaultCheckinMaxAmount)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read checkin maxAmount failed")
	}

	ironBonusThreshold, err := config.Int(ctx, configKeyIronBonusThresholdMeters, defaultIronBonusThresholdMeters)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read iron bonus threshold failed")
	}

	stealDailyTargets, err := config.Int(ctx, configKeyStealDailyTargets, defaultStealDailyTargets)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read steal dailyTargets failed")
	}
	stealDailyLimit, err := config.Int(ctx, configKeyStealDailyLimit, defaultStealDailyLimit)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read steal dailyLimit failed")
	}
	stealMin, err := config.Int(ctx, configKeyStealMinAmount, defaultStealMinAmount)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read steal minAmount failed")
	}
	stealMax, err := config.Int(ctx, configKeyStealMaxAmount, defaultStealMaxAmount)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read steal maxAmount failed")
	}

	giftDailyLimit, err := config.Int(ctx, configKeyGiftDailyLimit, defaultGiftDailyLimit)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read gift dailyLimit failed")
	}
	giftMin, err := config.Int(ctx, configKeyGiftMinAmount, defaultGiftMinAmount)
	if err != nil {
		return grasssvc.Config{}, feedingsvc.Config{}, grasssocialsvc.Config{}, gerror.Wrap(err, "sicau-niu read gift minAmount failed")
	}

	grassConfig := grasssvc.Config{
		CheckinMinAmount: checkinMin,
		CheckinMaxAmount: checkinMax,
	}
	feedingConfig := feedingsvc.Config{
		IronBonusThresholdMeters: float64(ironBonusThreshold),
	}
	grassSocialConfig := grasssocialsvc.Config{
		StealDailyTargets: stealDailyTargets,
		StealDailyLimit:   stealDailyLimit,
		StealMinAmount:    stealMin,
		StealMaxAmount:    stealMax,
		GiftDailyLimit:    giftDailyLimit,
		GiftMinAmount:     giftMin,
	}
	return grassConfig, feedingConfig, grassSocialConfig, nil
}
