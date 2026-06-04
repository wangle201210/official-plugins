// record.go defines the request and response DTOs for the sicau-niu activity-record
// read-only queries: feeding, steal, gift, check-in, activation and the grass
// ledger. Every endpoint is operator-facing and protected by the host unified
// Auth+Tenancy+Permission chain (permission sicau-niu:record:list); results are
// DB-side paged and time fields are returned as Unix milliseconds.

package v1

import "github.com/gogf/gf/v2/frame/g"

// FeedingsReq is the request for the feeding-record query.
type FeedingsReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/records/feedings" method:"get" tags:"Sicau Niu Record" summary:"List feeding records" dc:"Read-only paged feeding records with optional cattle and player filtering, newest first, DB-side pagination; player and cattle names are batch-assembled. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	NiuId    int64 `json:"niuId" dc:"Filter by cattle ID; lists all when 0" eg:"1"`
	UserId   int64 `json:"userId" dc:"Filter by player ID; lists all when 0" eg:"0"`
	PageNum  int   `json:"pageNum" dc:"Page number; defaults to 1" eg:"1"`
	PageSize int   `json:"pageSize" dc:"Items per page; defaults to 10, capped at 100" eg:"10"`
}

// FeedingsRes is the response for the feeding-record query.
type FeedingsRes struct {
	List  []*FeedingItem `json:"list" dc:"Feeding records on the current page" eg:"[]"`
	Total int            `json:"total" dc:"Total matched feeding records" eg:"0"`
}

// FeedingItem is one feeding record row.
type FeedingItem struct {
	Id               int64  `json:"id" eg:"1"`
	UserId           int64  `json:"userId" eg:"1"`
	Nickname         string `json:"nickname" dc:"Player nickname" eg:"川农牛仔"`
	NiuId            int64  `json:"niuId" eg:"7"`
	NiuName          string `json:"niuName" dc:"Cattle name" eg:"信工牛"`
	NiuCode          string `json:"niuCode" dc:"Cattle code" eg:"NIU-007"`
	BaseAmount       int    `json:"baseAmount" dc:"Base grass amount fed" eg:"20"`
	CoefficientBasis int    `json:"coefficientBasis" dc:"Effect coefficient basis (×100)" eg:"150"`
	EffectAmount     int    `json:"effectAmount" dc:"Actual feeding effect" eg:"30"`
	IsIronBonus      int    `json:"isIronBonus" dc:"Whether the iron-cow bonus applied: 1=yes, 0=no" eg:"1"`
	FedAt            *int64 `json:"fedAt" dc:"Feeding time, Unix ms; null when unset" eg:"1717488000000"`
	CreatedAt        *int64 `json:"createdAt" dc:"Record time, Unix ms" eg:"1717488000000"`
}

// StealsReq is the request for the steal-record query.
type StealsReq struct {
	g.Meta       `path:"/plugins/sicau-niu/admin/records/steals" method:"get" tags:"Sicau Niu Record" summary:"List steal records" dc:"Read-only paged steal records with optional actor and target filtering; actor and target nicknames are batch-assembled. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	ActorUserId  int64 `json:"actorUserId" dc:"Filter by steal actor player ID; lists all when 0" eg:"0"`
	TargetUserId int64 `json:"targetUserId" dc:"Filter by stolen-from player ID; lists all when 0" eg:"0"`
	PageNum      int   `json:"pageNum" eg:"1"`
	PageSize     int   `json:"pageSize" eg:"10"`
}

// StealsRes is the response for the steal-record query.
type StealsRes struct {
	List  []*StealItem `json:"list" eg:"[]"`
	Total int          `json:"total" eg:"0"`
}

// StealItem is one steal record row.
type StealItem struct {
	Id             int64  `json:"id" eg:"1"`
	ActorUserId    int64  `json:"actorUserId" eg:"1"`
	ActorNickname  string `json:"actorNickname" dc:"Steal actor nickname" eg:"川农牛仔"`
	TargetUserId   int64  `json:"targetUserId" eg:"2"`
	TargetNickname string `json:"targetNickname" dc:"Stolen-from nickname" eg:"水院小张"`
	Amount         int    `json:"amount" dc:"Grass amount stolen" eg:"8"`
	StealDate      string `json:"stealDate" dc:"Steal day, yyyy-mm-dd" eg:"2026-06-01"`
	CreatedAt      *int64 `json:"createdAt" eg:"1717488000000"`
}

// GiftsReq is the request for the gift-record query.
type GiftsReq struct {
	g.Meta     `path:"/plugins/sicau-niu/admin/records/gifts" method:"get" tags:"Sicau Niu Record" summary:"List gift records" dc:"Read-only paged gift records with optional sender and receiver filtering; sender and receiver nicknames are batch-assembled. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	FromUserId int64 `json:"fromUserId" dc:"Filter by gift sender player ID; lists all when 0" eg:"0"`
	ToUserId   int64 `json:"toUserId" dc:"Filter by gift receiver player ID; lists all when 0" eg:"0"`
	PageNum    int   `json:"pageNum" eg:"1"`
	PageSize   int   `json:"pageSize" eg:"10"`
}

// GiftsRes is the response for the gift-record query.
type GiftsRes struct {
	List  []*GiftItem `json:"list" eg:"[]"`
	Total int         `json:"total" eg:"0"`
}

// GiftItem is one gift record row.
type GiftItem struct {
	Id           int64  `json:"id" eg:"1"`
	FromUserId   int64  `json:"fromUserId" eg:"1"`
	FromNickname string `json:"fromNickname" dc:"Sender nickname" eg:"川农牛仔"`
	ToUserId     int64  `json:"toUserId" eg:"5"`
	ToNickname   string `json:"toNickname" dc:"Receiver nickname" eg:"园林小美"`
	Amount       int    `json:"amount" dc:"Grass amount gifted" eg:"12"`
	GiftDate     string `json:"giftDate" dc:"Gift day, yyyy-mm-dd" eg:"2026-06-01"`
	CreatedAt    *int64 `json:"createdAt" eg:"1717488000000"`
}

// CheckinsReq is the request for the check-in-record query.
type CheckinsReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/records/checkins" method:"get" tags:"Sicau Niu Record" summary:"List check-in records" dc:"Read-only paged check-in records with optional player filtering; player nicknames are batch-assembled. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	UserId   int64 `json:"userId" dc:"Filter by player ID; lists all when 0" eg:"0"`
	PageNum  int   `json:"pageNum" eg:"1"`
	PageSize int   `json:"pageSize" eg:"10"`
}

// CheckinsRes is the response for the check-in-record query.
type CheckinsRes struct {
	List  []*CheckinItem `json:"list" eg:"[]"`
	Total int            `json:"total" eg:"0"`
}

// CheckinItem is one check-in record row.
type CheckinItem struct {
	Id          int64  `json:"id" eg:"1"`
	UserId      int64  `json:"userId" eg:"1"`
	Nickname    string `json:"nickname" dc:"Player nickname" eg:"川农牛仔"`
	CheckinDate string `json:"checkinDate" dc:"Check-in day, yyyy-mm-dd" eg:"2026-06-01"`
	Amount      int    `json:"amount" dc:"Grass amount granted" eg:"35"`
	CreatedAt   *int64 `json:"createdAt" eg:"1717488000000"`
}

// ActivationsReq is the request for the activation-record query.
type ActivationsReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/records/activations" method:"get" tags:"Sicau Niu Record" summary:"List activation records" dc:"Read-only paged activation records with optional player and cattle filtering; player and cattle names are batch-assembled. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	UserId   int64 `json:"userId" dc:"Filter by player ID; lists all when 0" eg:"0"`
	NiuId    int64 `json:"niuId" dc:"Filter by cattle ID; lists all when 0" eg:"0"`
	PageNum  int   `json:"pageNum" eg:"1"`
	PageSize int   `json:"pageSize" eg:"10"`
}

// ActivationsRes is the response for the activation-record query.
type ActivationsRes struct {
	List  []*ActivationItem `json:"list" eg:"[]"`
	Total int               `json:"total" eg:"0"`
}

// ActivationItem is one activation record row.
type ActivationItem struct {
	Id           int64  `json:"id" eg:"1"`
	UserId       int64  `json:"userId" eg:"1"`
	Nickname     string `json:"nickname" dc:"Player nickname" eg:"川农牛仔"`
	NiuId        int64  `json:"niuId" eg:"7"`
	NiuName      string `json:"niuName" dc:"Cattle name" eg:"信工牛"`
	NiuCode      string `json:"niuCode" dc:"Cattle code" eg:"NIU-007"`
	ActivityDate string `json:"activityDate" dc:"Activation day, yyyy-mm-dd" eg:"2026-06-01"`
	IsFirst      int    `json:"isFirst" dc:"Whether first activator: 1=yes, 0=no" eg:"1"`
	OrderNo      int    `json:"orderNo" dc:"Arrival order for the cattle" eg:"1"`
	ActivatedAt  *int64 `json:"activatedAt" dc:"Activation time, Unix ms; null when unset" eg:"1717488000000"`
	CreatedAt    *int64 `json:"createdAt" eg:"1717488000000"`
}

// GrassTxnsReq is the request for the grass-ledger query.
type GrassTxnsReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/records/grass-txns" method:"get" tags:"Sicau Niu Record" summary:"List grass ledger" dc:"Read-only paged grass-account ledger entries with optional player filtering; player nicknames are batch-assembled. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	UserId   int64 `json:"userId" dc:"Filter by player ID; lists all when 0" eg:"0"`
	PageNum  int   `json:"pageNum" eg:"1"`
	PageSize int   `json:"pageSize" eg:"10"`
}

// GrassTxnsRes is the response for the grass-ledger query.
type GrassTxnsRes struct {
	List  []*GrassTxnItem `json:"list" eg:"[]"`
	Total int             `json:"total" eg:"0"`
}

// GrassTxnItem is one grass-ledger entry row.
type GrassTxnItem struct {
	Id        int64  `json:"id" eg:"1"`
	UserId    int64  `json:"userId" eg:"1"`
	Nickname  string `json:"nickname" dc:"Player nickname" eg:"川农牛仔"`
	Delta     int64  `json:"delta" dc:"Signed grass change: positive=gain, negative=spend" eg:"35"`
	TxnType   string `json:"txnType" dc:"Ledger entry type, e.g. checkin/feed/steal/gift" eg:"checkin"`
	RefId     int64  `json:"refId" dc:"Reference ID of the action that produced this entry; 0 when none" eg:"12"`
	CreatedAt *int64 `json:"createdAt" eg:"1717488000000"`
}
