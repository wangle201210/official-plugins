// rules_defaults.go defines stable rule keys, built-in defaults, value bounds
// and fixed key metadata shared by reads and writes.

package rules

const (
	keyActivationLBSThresholdMeters = "activation.lbsThresholdMeters"
	keyPosterCampusBadge            = "poster.campusBadge"
	keyCheckinMinAmount             = "checkin.minAmount"
	keyCheckinMaxAmount             = "checkin.maxAmount"
	keyStealDailyTargets            = "steal.dailyTargets"
	keyStealDailyLimit              = "steal.dailyLimit"
	keyStealMinAmount               = "steal.minAmount"
	keyStealMaxAmount               = "steal.maxAmount"
	keyGiftDailyLimit               = "gift.dailyLimit"
	keyGiftMinAmount                = "gift.minAmount"
	keyIronBonusThresholdMeters     = "ironBonus.thresholdMeters"
	keyRankingTopN                  = "ranking.topN"
	keyAnomalyFeedDailyThreshold    = "anomaly.feedDailyThreshold"
	keyAnomalyStealDailyThreshold   = "anomaly.stealDailyThreshold"
	keyAnomalyListLimit             = "anomaly.listLimit"
	keyMiniappURL                   = "miniapp.url"
)

const (
	maxRankingTopN      = 500
	maxAnomalyListLimit = 500
	maxStringValueLen   = 255
)

// ruleDefinition is one known persisted runtime rule row.
type ruleDefinition struct {
	key    string
	remark string
}

// ruleDefinitions is the fixed operator-maintained key set. The order is stable
// so the admin page and writes stay deterministic.
var ruleDefinitions = []ruleDefinition{
	{key: keyActivationLBSThresholdMeters, remark: "LBS 激活判距阈值（米）"},
	{key: keyPosterCampusBadge, remark: "激活海报和电子证书上的校庆标识"},
	{key: keyCheckinMinAmount, remark: "每日签到草量下限"},
	{key: keyCheckinMaxAmount, remark: "每日签到草量上限"},
	{key: keyStealDailyTargets, remark: "每日随机可偷名单人数"},
	{key: keyStealDailyLimit, remark: "每日偷草次数上限"},
	{key: keyStealMinAmount, remark: "单次偷草草量下限"},
	{key: keyStealMaxAmount, remark: "单次偷草草量上限"},
	{key: keyGiftDailyLimit, remark: "每日送草次数上限"},
	{key: keyGiftMinAmount, remark: "单次送草草量下限"},
	{key: keyIronBonusThresholdMeters, remark: "铁牛邻近喂草加成阈值（米）"},
	{key: keyRankingTopN, remark: "排行榜返回 Top-N 上限"},
	{key: keyAnomalyFeedDailyThreshold, remark: "单日喂草次数异常阈值"},
	{key: keyAnomalyStealDailyThreshold, remark: "单日偷草次数异常阈值"},
	{key: keyAnomalyListLimit, remark: "异常告警返回上限"},
	{key: keyMiniappURL, remark: "H5 数字纪念墙回跳小程序链接"},
}

// defaultRuleSet returns the built-in campaign defaults used when DB rows or
// startup config values are missing.
func defaultRuleSet() RuleSet {
	return RuleSet{
		ActivationLBSThresholdMeters: 50,
		PosterCampusBadge:            "",
		CheckinMinAmount:             20,
		CheckinMaxAmount:             50,
		StealDailyTargets:            12,
		StealDailyLimit:              5,
		StealMinAmount:               5,
		StealMaxAmount:               20,
		GiftDailyLimit:               12,
		GiftMinAmount:                12,
		IronBonusThresholdMeters:     12,
		RankingTopN:                  100,
		AnomalyFeedDailyThreshold:    100,
		AnomalyStealDailyThreshold:   5,
		AnomalyListLimit:             200,
		MiniappURL:                   "",
	}
}

// ruleKeys returns the fixed key list for one bounded rule query.
func ruleKeys() []string {
	keys := make([]string, 0, len(ruleDefinitions))
	for _, definition := range ruleDefinitions {
		keys = append(keys, definition.key)
	}
	return keys
}

// ruleRemark returns the operator-facing remark for a known key.
func ruleRemark(key string) string {
	for _, definition := range ruleDefinitions {
		if definition.key == key {
			return definition.remark
		}
	}
	return ""
}
