// player_activations.go defines the request and response DTOs for the LBS
// activation action: the current player reports their check-in location and an
// optional photo path, then the server matches a nearby unactivated cattle; the
// response carries the matched cattle ID, first-activator flag, arrival order,
// activation time and the cattle main card (issued on activation).

package v1

import "github.com/gogf/gf/v2/frame/g"

// ActivateReq is the request for activating a nearby cattle by GPS check-in.
type ActivateReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/activations" method:"post" tags:"Sicau Niu Player" summary:"Activate a nearby cattle (LBS)" dc:"Activate a nearby unactivated cattle by the player's reported GPS check-in location (GCJ-02). The mini program does not pass a cattle ID. The server finds the nearest currently visible inactive cattle within the configured Haversine threshold, locks and rechecks the matched cattle in a transaction, flips it to active and issues the cattle main card. The photo is evidence only; no image recognition is performed. Each player may activate at most one cattle per Beijing-time natural day, is bounded by a daily attempt quota that counts failed check-ins, and implausibly fast movement between successive check-ins is rejected as a speed anomaly; rejection responses never carry distance or bearing hints. Requires a valid player token."`
	Lat       float64 `json:"lat" v:"required" dc:"Player reported GPS latitude in GCJ-02" eg:"30.123456"`
	Lng       float64 `json:"lng" v:"required" dc:"Player reported GPS longitude in GCJ-02" eg:"103.123456"`
	PhotoPath string  `json:"photoPath" dc:"Optional activation photo storage path (evidence only, no recognition); empty when omitted" eg:"sicau-niu/activation/1.jpg"`
}

// ActivateRes is the response for a successful cattle activation.
type ActivateRes struct {
	NiuId       int64           `json:"niuId" dc:"Server matched and activated cattle ID" eg:"1"`
	IsFirst     bool            `json:"isFirst" dc:"Whether the current player is the cattle first-activator" eg:"true"`
	OrderNo     int             `json:"orderNo" dc:"Arrival order for this cattle, starting at 1" eg:"1"`
	ActivatedAt *int64          `json:"activatedAt" dc:"Activation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	Card        *ActivationCard `json:"card" dc:"The cattle main card issued on activation; null when the cattle has no main card"`
}

// ActivationCard defines the cattle main card returned on activation.
type ActivationCard struct {
	Category  string `json:"category" dc:"Card category: person=人物, event=事件, research=科研, college=院系, spirit=精神" eg:"spirit"`
	Title     string `json:"title" dc:"Card title" eg:"川农大精神"`
	Content   string `json:"content" dc:"Card content text" eg:"爱国敬业、艰苦奋斗、团结拼搏、求实创新"`
	ImagePath string `json:"imagePath" dc:"Card image storage path; empty when none" eg:"sicau-niu/card/1.jpg"`
}
