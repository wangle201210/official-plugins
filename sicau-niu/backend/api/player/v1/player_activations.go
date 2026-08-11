// player_activations.go defines the request and response DTOs for the LBS
// activation action: the current player reports their check-in location and an
// optional photo path, then the server matches a nearby unactivated cattle; the
// response carries the matched cattle ID, first-activator flag, arrival order,
// activation time and the cattle main card (issued on activation).

package v1

import "github.com/gogf/gf/v2/frame/g"

// ActivateReq is the request for activating a nearby cattle by GPS check-in.
type ActivateReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/activations" method:"post" tags:"寻牛小程序" summary:"拍照打卡激活附近牛只" dc:"Activate the nearest visible cattle not yet visited by the current player using a GCJ-02 check-in. The first visitor flips shared state to active; later visitors still receive their own arrival order and main card. photoPath must be an owned, unused photoId and is consumed atomically on success. Each player may activate at most one cattle per Beijing-time natural day, is bounded by a daily attempt quota, and implausibly movement is rejected without location hints. Requires a valid player token."`
	Lat       float64 `json:"lat" v:"required" dc:"Player reported GPS latitude in GCJ-02" eg:"30.123456"`
	Lng       float64 `json:"lng" v:"required" dc:"Player reported GPS longitude in GCJ-02" eg:"103.123456"`
	PhotoPath string  `json:"photoPath" v:"required" dc:"Opaque player-owned photoId returned by the upload endpoint; consumed once on successful activation" eg:"550e8400-e29b-41d4-a716-446655440000"`
	RequestId string  `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key; retries return the first successful activation response" eg:"activation-1b2a3c4d"`
}

// ActivateRes is the response for a successful cattle activation.
type ActivateRes struct {
	Seq         int64           `json:"seq" dc:"Stable global activation sequence identifier" eg:"128"`
	NiuId       int64           `json:"niuId" dc:"Server matched and activated cattle ID" eg:"1"`
	NiuName     string          `json:"niuName" dc:"Cattle display name, falling back to its code" eg:"川农牛 001"`
	Skin        string          `json:"skin" dc:"Mini-program cattle skin" eg:"normal"`
	Quote       string          `json:"quote" dc:"School-history quote selected for the result page" eg:"追求真理、造福社会、自强不息"`
	IsFirst     bool            `json:"isFirst" dc:"Whether the current player is the cattle first-activator" eg:"true"`
	OrderNo     int             `json:"orderNo" dc:"Arrival order for this cattle, starting at 1" eg:"1"`
	ActivatedAt *int64          `json:"activatedAt" dc:"Activation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	Card        *ActivationCard `json:"card" dc:"The cattle main card issued on activation; null when the cattle has no main card" eg:"null"`
}

// ActivationCard defines the cattle main card returned on activation.
type ActivationCard struct {
	Category  string `json:"category" dc:"Card category: person=人物, event=事件, research=科研, college=院系, spirit=精神" eg:"spirit"`
	Title     string `json:"title" dc:"Card title" eg:"川农大精神"`
	Content   string `json:"content" dc:"Card content text" eg:"爱国敬业、艰苦奋斗、团结拼搏、求实创新"`
	ImagePath string `json:"imagePath" dc:"Card image storage path; empty when none" eg:"sicau-niu/card/1.jpg"`
}
