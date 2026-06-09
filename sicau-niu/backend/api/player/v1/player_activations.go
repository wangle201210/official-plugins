// player_activations.go defines the request and response DTOs for the LBS
// activation action: the current player reports their location and an optional
// photo path to activate a cattle; the response carries the first-activator flag,
// arrival order, activation time and the cattle main card (issued on activation).

package v1

import "github.com/gogf/gf/v2/frame/g"

// ActivateReq is the request for activating a cattle at its GPS anchor.
type ActivateReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/activations" method:"post" tags:"Sicau Niu Player" summary:"Activate a cattle (LBS)" dc:"Activate the target cattle only when it is currently visible to the player map. The server checks online time, optional weekday/time windows and the Haversine distance between the reported location and the cattle GPS anchor against the configured threshold (no image recognition; the photo is evidence only). Each player may activate at most one cattle per natural day and may not re-activate the same cattle. The first activator becomes the cattle first-activator and flips it to active (shared-pool visible); later activators get an incremented arrival order. The cattle main card is issued on success. Requires a valid player token."`
	NiuId     int64   `json:"niuId" v:"required" dc:"Target cattle ID to activate" eg:"1"`
	Lat       float64 `json:"lat" v:"required" dc:"Player reported GPS latitude" eg:"30.123456"`
	Lng       float64 `json:"lng" v:"required" dc:"Player reported GPS longitude" eg:"103.123456"`
	PhotoPath string  `json:"photoPath" dc:"Optional activation photo storage path (evidence only, no recognition); empty when omitted" eg:"sicau-niu/activation/1.jpg"`
}

// ActivateRes is the response for a successful cattle activation.
type ActivateRes struct {
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
