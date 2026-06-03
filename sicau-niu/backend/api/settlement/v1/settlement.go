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
