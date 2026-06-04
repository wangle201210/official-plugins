// Package backend wires the sicau-niu source plugin into the host plugin
// registry. It registers the embedded plugin assets and binds three HTTP route
// surfaces under the plugin API prefix: a public WeChat player login route, a
// player-token-protected surface (phone binding, profile, college dropdown, the
// C3/C4 gameplay endpoints, and the C5 leaderboards and player honor list)
// guarded by the plugin-owned player-auth middleware, and an operator surface
// (college dictionary CRUD, player query, the C2 content-asset CRUD for cattle,
// iron-cows, cards and quotes, and the C5 honor-definition CRUD) guarded by the
// host Auth+Tenancy+Permission chain. The plugin owns its WeChat-gateway,
// player-token, identity, college, cattle, card, ranking and honor services; the
// service graph is constructed once at route-registration time from the
// plugin-scoped configuration.
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

	activationConfig, err := buildActivationConfig(ctx, services.Config())
	if err != nil {
		return err
	}

	grassConfig, feedingConfig, grassSocialConfig, err := buildGrassConfigs(ctx, services.Config())
	if err != nil {
		return err
	}

	rankingConfig, err := buildRankingConfig(ctx, services.Config())
	if err != nil {
		return err
	}

	collegeService := collegesvc.New()
	identityService := identitysvc.New(gateway, tokenService, collegeService)
	cattleService := cattlesvc.New(collegeService)
	cardService := cardsvc.New(cattleService)
	activationService := activationsvc.New(
		identityService,
		activationsvc.NewBasicPosterRenderer(),
		activationConfig,
	)
	grassService := grasssvc.New(grassConfig)
	feedingService := feedingsvc.New(grassService, feedingsvc.NewMockIronLocation(), feedingConfig)
	grassSocialService := grasssocialsvc.New(grassService, grassSocialConfig)
	rankingService := rankingsvc.New(rankingConfig)
	honorService := honorsvc.New(
		honorsvc.NewBasicCertRenderer(),
		honorsvc.Config{CampusBadge: activationConfig.CampusBadge},
	)
	miniappURL, err := services.Config().String(ctx, configKeyMiniappURL, "")
	if err != nil {
		return gerror.Wrap(err, "sicau-niu read miniapp url failed")
	}
	wallService := wallsvc.New(wallsvc.Config{MiniappURL: miniappURL})
	settlementConfig, err := buildSettlementConfig(ctx, services.Config())
	if err != nil {
		return err
	}
	settlementService := settlementsvc.New(settlementConfig)
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
	settlementController := settlementctrl.NewV1(settlementService)

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
					settlementController.IssueCertificates,
					settlementController.RiskDeviceClusters,
					settlementController.CreateArchive,
					settlementController.ListArchives,
					settlementController.Activity,
					settlementController.RiskAnomalies,
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

// buildActivationConfig reads the plain-value activation configuration (LBS
// distance threshold and poster campus badge) from the plugin-scoped
// configuration. It returns an error when the configuration cannot be read so the
// failure surfaces at startup.
func buildActivationConfig(
	ctx context.Context,
	config contract.ConfigService,
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
	config contract.ConfigService,
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
	config contract.ConfigService,
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
	config contract.ConfigService,
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
