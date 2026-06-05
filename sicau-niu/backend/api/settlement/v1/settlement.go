// settlement.go defines the request and response DTOs for the sicau-niu C7
// operator settlement capability: the operations dashboard, the player roster
// export, the batch certificate issuance, the risk-alert (shared-device) view and
// the settlement archive (create and list). Every endpoint is operator-facing and
// protected by the host unified Auth+Tenancy+Permission chain; per-route
// permission is declared in the g.Meta permission tag. Time fields are returned as
// Unix milliseconds.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DashboardReq is the request for the operations dashboard.
type DashboardReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/dashboard" method:"get" tags:"Sicau Niu Settlement" summary:"Operations dashboard" dc:"Return the activity operations dashboard: player count, activated/total cattle, first-activator count, feeding count and total effect, steal/gift/check-in counts and granted-certificate count. Every figure is aggregated on the database side. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// DashboardRes is the response for the operations dashboard.
type DashboardRes struct {
	PlayerCount             int64 `json:"playerCount" dc:"Number of participating players" eg:"3500"`
	ActivatedNiuCount       int64 `json:"activatedNiuCount" dc:"Number of cattle whose status is active" eg:"42"`
	TotalNiuCount           int64 `json:"totalNiuCount" dc:"Total cattle count" eg:"120"`
	FirstActivatorCount     int64 `json:"firstActivatorCount" dc:"Number of first activators (is_first activations)" eg:"42"`
	FeedingCount            int64 `json:"feedingCount" dc:"Total feeding action count" eg:"18000"`
	FeedTotalEffect         int64 `json:"feedTotalEffect" dc:"Total feeding effect accumulated (SUM of effect_amount)" eg:"260000"`
	StealCount              int64 `json:"stealCount" dc:"Total steal action count" eg:"5400"`
	GiftCount               int64 `json:"giftCount" dc:"Total gift action count" eg:"2100"`
	CheckinCount            int64 `json:"checkinCount" dc:"Total check-in count" eg:"9000"`
	CertificateGrantedCount int64 `json:"certificateGrantedCount" dc:"Number of granted certificate honors" eg:"120"`
}

// ExportPlayersReq is the request for the player roster export.
type ExportPlayersReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/export/players" method:"get" tags:"Sicau Niu Settlement" summary:"Export player roster" dc:"Return a bounded player roster projection (nickname/identity/college/grade/activation count/feeding total) for CSV export. The roster is hard-capped; when more players exist the response is truncated and the truncated flag is set. The per-player activation count and feeding total are batch-assembled to avoid N+1. Protected by host unified permission check." permission:"sicau-niu:settlement:export"`
}

// ExportPlayersRes is the response for the player roster export.
type ExportPlayersRes struct {
	List      []*PlayerExportRow `json:"list" dc:"Bounded player roster rows" eg:"[]"`
	Total     int64              `json:"total" dc:"Total player count before the export cap" eg:"3500"`
	Truncated bool               `json:"truncated" dc:"Whether the roster was truncated at the export cap" eg:"false"`
}

// PlayerExportRow is one player row in the roster export.
type PlayerExportRow struct {
	UserId          int64  `json:"userId" dc:"Player ID" eg:"1"`
	Nickname        string `json:"nickname" dc:"Player nickname; empty when unset" eg:"川农牛仔"`
	IdentityType    string `json:"identityType" dc:"Player identity label" eg:"student"`
	CollegeName     string `json:"collegeName" dc:"College name; empty when none" eg:"信息工程学院"`
	Grade           int    `json:"grade" dc:"Player grade year; 0 when unset" eg:"2023"`
	ActivationCount int64  `json:"activationCount" dc:"Player activation count" eg:"5"`
	FeedTotalEffect int64  `json:"feedTotalEffect" dc:"Player total feeding effect" eg:"1500"`
}

// IssueCertificatesReq is the request for batch certificate issuance.
type IssueCertificatesReq struct {
	g.Meta  `path:"/plugins/sicau-niu/settlement/certificates/issue" method:"post" tags:"Sicau Niu Settlement" summary:"Batch issue certificate" dc:"Batch issue one certificate honor to the players who satisfy its unlock rule. Only certificate honors are accepted; participation/feed_count/activation_count unlock rules are settled here while collection rules (category_complete/full_complete) are rejected. Grants are written idempotently in a transaction; already-granted players are skipped. Protected by host unified permission check." permission:"sicau-niu:settlement:issue"`
	HonorId int64 `json:"honorId" v:"required|min:1" dc:"Certificate honor definition ID to issue" eg:"7"`
}

// IssueCertificatesRes is the response for batch certificate issuance.
type IssueCertificatesRes struct {
	Eligible int64 `json:"eligible" dc:"Number of players who satisfy the certificate's unlock rule" eg:"120"`
	Issued   int64 `json:"issued" dc:"Number of newly granted players in this run" eg:"118"`
	Skipped  int64 `json:"skipped" dc:"Number of eligible players already granted (skipped)" eg:"2"`
}

// RiskDeviceClustersReq is the request for the shared-device risk view.
type RiskDeviceClustersReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/risk/device-clusters" method:"get" tags:"Sicau Niu Settlement" summary:"Shared-device risk view" dc:"Return the shared-device risk view: clusters of players sharing one non-empty device fingerprint (one device, many accounts), with member nicknames, for manual review. Read-only and aggregated on the database side; bounded cluster count. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// RiskDeviceClustersRes is the response for the shared-device risk view.
type RiskDeviceClustersRes struct {
	List []*DeviceCluster `json:"list" dc:"Shared-device clusters ordered by member count descending, bounded" eg:"[]"`
}

// DeviceCluster is one shared-device cluster on the risk view.
type DeviceCluster struct {
	Fingerprint string                 `json:"fingerprint" dc:"Shared device fingerprint" eg:"fp-abc123"`
	Count       int64                  `json:"count" dc:"Number of players sharing the fingerprint" eg:"3"`
	Members     []*DeviceClusterMember `json:"members" dc:"Players in the cluster" eg:"[]"`
}

// DeviceClusterMember is one player in a shared-device cluster.
type DeviceClusterMember struct {
	UserId   int64  `json:"userId" dc:"Player ID" eg:"1"`
	Nickname string `json:"nickname" dc:"Player nickname; empty when unset" eg:"川农牛仔"`
}

// ActivityReq is the request for the dashboard activity metrics.
type ActivityReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/activity" method:"get" tags:"Sicau Niu Settlement" summary:"Dashboard activity metrics" dc:"Return the M5 dashboard activity metrics: the daily active-user (DAU) series for the last N days (N capped at 60) and the next-day / 7-day retention over elapsed registration cohorts. Active on a day means the player took any action (activate/feed/check-in/steal/gift) that day. Every figure is aggregated on the database side. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
	Days   int `json:"days" v:"min:0|max:60" dc:"DAU window size in days; defaults to 14 when 0, capped at 60" eg:"14"`
}

// ActivityRes is the response for the dashboard activity metrics.
type ActivityRes struct {
	Dau         []*DauPoint    `json:"dau" dc:"Per-day active-user counts ordered by date ascending; calendar gaps are zero" eg:"[]"`
	RetentionD1 *RetentionStat `json:"retentionD1" dc:"Next-day retention over elapsed cohorts" eg:"null"`
	RetentionD7 *RetentionStat `json:"retentionD7" dc:"7-day retention over elapsed cohorts" eg:"null"`
}

// DauPoint is one day's de-duplicated active-user count.
type DauPoint struct {
	Date        string `json:"date" dc:"Calendar day, yyyy-mm-dd" eg:"2026-06-01"`
	ActiveUsers int64  `json:"activeUsers" dc:"De-duplicated active players that day" eg:"320"`
}

// RetentionStat is one retention bucket.
type RetentionStat struct {
	CohortUsers   int64   `json:"cohortUsers" dc:"Players whose registration day plus the bucket offset has elapsed" eg:"1000"`
	ReturnedUsers int64   `json:"returnedUsers" dc:"Players in the cohort active on registration day plus the offset" eg:"450"`
	Rate          float64 `json:"rate" dc:"Retention rate in [0,1]; 0 when the cohort is empty" eg:"0.45"`
}

// RiskAnomaliesReq is the request for the risk anomaly alert view.
type RiskAnomaliesReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/risk/anomalies" method:"get" tags:"Sicau Niu Settlement" summary:"Risk anomaly alerts" dc:"Return the M13 risk anomaly alerts: players whose single-day feeding or steal count exceeds the configured threshold, with player, nickname, behaviour type, day, that day's count and the threshold, for manual review. Read-only, aggregated on the database side and bounded. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// RiskAnomaliesRes is the response for the risk anomaly alert view.
type RiskAnomaliesRes struct {
	List []*AnomalyAlert `json:"list" dc:"Anomaly alerts ordered by single-day count descending, bounded" eg:"[]"`
}

// RulesReq is the request for operator runtime rules.
type RulesReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/rules" method:"get" tags:"Sicau Niu Settlement" summary:"Get runtime rules" dc:"Return the operator-maintained runtime rules for activation, posters, check-in, steal, gift, iron bonus, ranking, anomaly alerting and H5 mini-program link. The values are read from the plugin-owned rule_config table and normalized with built-in defaults. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// RulesRes is the response for operator runtime rules.
type RulesRes RuleConfig

// UpdateRulesReq is the request for updating operator runtime rules.
type UpdateRulesReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/rules" method:"put" tags:"Sicau Niu Settlement" summary:"Update runtime rules" dc:"Replace the complete operator-maintained runtime rule set. All numeric fields must be positive, min values must not exceed max values, ranking/anomaly caps are bounded, and values apply to subsequent player and operator requests without process-local cache invalidation. Protected by host unified permission check." permission:"sicau-niu:settlement:rules"`
	RuleConfig
}

// UpdateRulesRes is the response for updating operator runtime rules.
type UpdateRulesRes RuleConfig

// RuleConfig is the complete operator-maintained runtime rule projection.
type RuleConfig struct {
	ActivationLBSThresholdMeters int    `json:"activationLbsThresholdMeters" dc:"LBS activation distance threshold in meters; must be positive" eg:"50"`
	PosterCampusBadge            string `json:"posterCampusBadge" dc:"Campus anniversary badge rendered on activation posters and certificates; max 255 characters" eg:"川农 120 周年"`
	CheckinMinAmount             int    `json:"checkinMinAmount" dc:"Daily check-in grass grant lower bound; must be positive and not exceed checkinMaxAmount" eg:"20"`
	CheckinMaxAmount             int    `json:"checkinMaxAmount" dc:"Daily check-in grass grant upper bound; must be positive and not below checkinMinAmount" eg:"50"`
	StealDailyTargets            int    `json:"stealDailyTargets" dc:"Daily stealable target list size; must be positive" eg:"12"`
	StealDailyLimit              int    `json:"stealDailyLimit" dc:"Daily steal action limit per player; must be positive" eg:"5"`
	StealMinAmount               int    `json:"stealMinAmount" dc:"Per-steal random grass lower bound; must be positive and not exceed stealMaxAmount" eg:"5"`
	StealMaxAmount               int    `json:"stealMaxAmount" dc:"Per-steal random grass upper bound; must be positive and not below stealMinAmount" eg:"20"`
	GiftDailyLimit               int    `json:"giftDailyLimit" dc:"Daily gift action limit per player; must be positive" eg:"12"`
	GiftMinAmount                int    `json:"giftMinAmount" dc:"Per-gift minimum grass amount; must be positive" eg:"12"`
	IronBonusThresholdMeters     int    `json:"ironBonusThresholdMeters" dc:"Iron-cow proximity bonus threshold in meters; must be positive" eg:"12"`
	RankingTopN                  int    `json:"rankingTopN" dc:"Leaderboard Top-N cap; must be between 1 and 500" eg:"100"`
	AnomalyFeedDailyThreshold    int    `json:"anomalyFeedDailyThreshold" dc:"Single-day feeding count anomaly threshold; must be positive" eg:"100"`
	AnomalyStealDailyThreshold   int    `json:"anomalyStealDailyThreshold" dc:"Single-day steal count anomaly threshold; must be positive" eg:"5"`
	AnomalyListLimit             int    `json:"anomalyListLimit" dc:"Maximum anomaly alert rows; must be between 1 and 500" eg:"200"`
	MiniappURL                   string `json:"miniappUrl" dc:"Return-to-mini-program URL for the public H5 wall; max 255 characters, empty hides the entry" eg:"weixin://dl/business/?t=abc"`
}

// FeedRankingReq is the operator request for the feed leaderboard.
type FeedRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/rankings/feed" method:"get" tags:"Sicau Niu Settlement" summary:"Operator feed leaderboard" dc:"Return the Top-N personal feeding leaderboard for operators. Aggregation and Top-N limiting run on the database side and nicknames are batch-assembled to avoid N+1. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// FeedRankingRes is the operator feed leaderboard response.
type FeedRankingRes struct {
	List []*FeedRankItem `json:"list" dc:"Players ranked by total feeding effect, bounded by runtime rankingTopN" eg:"[]"`
}

// FriendRankingReq is the operator request for the SICAU-friend leaderboard.
type FriendRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/rankings/friend" method:"get" tags:"Sicau Niu Settlement" summary:"Operator friend leaderboard" dc:"Return the Top-N SICAU-friend feeding leaderboard for operators. Aggregation and Top-N limiting run on the database side and nicknames are batch-assembled to avoid N+1. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// FriendRankingRes is the operator SICAU-friend leaderboard response.
type FriendRankingRes struct {
	List []*FeedRankItem `json:"list" dc:"SICAU-friend players ranked by total feeding effect, bounded by runtime rankingTopN" eg:"[]"`
}

// CollegeRankingReq is the operator request for the college leaderboard.
type CollegeRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/rankings/college" method:"get" tags:"Sicau Niu Settlement" summary:"Operator college leaderboard" dc:"Return the Top-N college feeding leaderboard for operators. Aggregation and Top-N limiting run on the database side and college names are batch-assembled to avoid N+1. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// CollegeRankingRes is the operator college leaderboard response.
type CollegeRankingRes struct {
	List []*CollegeRankItem `json:"list" dc:"Colleges ranked by enrolled students' total feeding effect, bounded by runtime rankingTopN" eg:"[]"`
}

// FeedRankItem is one operator personal leaderboard row.
type FeedRankItem struct {
	Rank     int    `json:"rank" dc:"1-based rank on the leaderboard" eg:"1"`
	UserId   int64  `json:"userId" dc:"Player user ID" eg:"1"`
	Nickname string `json:"nickname" dc:"Player nickname; empty when unset" eg:"川农牛仔"`
	Total    int64  `json:"total" dc:"Total feeding effect" eg:"3600"`
}

// CollegeRankItem is one operator college leaderboard row.
type CollegeRankItem struct {
	Rank        int    `json:"rank" dc:"1-based rank on the leaderboard" eg:"1"`
	CollegeId   int64  `json:"collegeId" dc:"College ID" eg:"3"`
	CollegeName string `json:"collegeName" dc:"College name; empty when missing" eg:"信息工程学院"`
	Total       int64  `json:"total" dc:"Total feeding effect of enrolled students" eg:"180000"`
}

// AnomalyAlert is one single-day over-threshold behaviour record.
type AnomalyAlert struct {
	UserId    int64  `json:"userId" dc:"Player ID" eg:"1"`
	Nickname  string `json:"nickname" dc:"Player nickname; empty when unset" eg:"川农牛仔"`
	Type      string `json:"type" dc:"Behaviour type: feed=喂草, steal=偷草" eg:"feed"`
	Date      string `json:"date" dc:"Natural day, yyyy-mm-dd" eg:"2026-06-01"`
	Count     int64  `json:"count" dc:"The player's behaviour count that day" eg:"260"`
	Threshold int64  `json:"threshold" dc:"The configured threshold the count exceeded" eg:"100"`
}

// CreateArchiveReq is the request for creating a settlement archive.
type CreateArchiveReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/archives" method:"post" tags:"Sicau Niu Settlement" summary:"Create settlement archive" dc:"Freeze the current dashboard metrics into a persisted settlement snapshot with the given title. Protected by host unified permission check." permission:"sicau-niu:settlement:archive"`
	Title  string `json:"title" v:"required|length:1,128" dc:"Archive title" eg:"寻牛活动结算公示 2026"`
}

// CreateArchiveRes is the response for creating a settlement archive.
type CreateArchiveRes struct {
	Id int64 `json:"id" dc:"The newly created settlement archive ID" eg:"1"`
}

// ListArchivesReq is the request for the settlement archive list.
type ListArchivesReq struct {
	g.Meta `path:"/plugins/sicau-niu/settlement/archives" method:"get" tags:"Sicau Niu Settlement" summary:"List settlement archives" dc:"Return the settlement archives ordered by archive time descending, bounded. Each item carries the frozen dashboard snapshot JSON. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

// ListArchivesRes is the response for the settlement archive list.
type ListArchivesRes struct {
	List []*ArchiveItem `json:"list" dc:"Settlement archives ordered by archive time descending, bounded" eg:"[]"`
}

// ArchiveItem is one settlement archive record.
type ArchiveItem struct {
	Id         int64  `json:"id" dc:"Settlement archive ID" eg:"1"`
	Title      string `json:"title" dc:"Archive title" eg:"寻牛活动结算公示 2026"`
	Snapshot   string `json:"snapshot" dc:"Frozen dashboard metrics serialized as JSON text" eg:"{\"playerCount\":3500}"`
	OperatorId int64  `json:"operatorId" dc:"Operator user ID who created the archive; 0 when unattributed" eg:"0"`
	ArchivedAt *int64 `json:"archivedAt" dc:"Archive time as Unix milliseconds; null when unset" eg:"1717488000000"`
}
