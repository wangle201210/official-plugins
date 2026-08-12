// This file defines operator APIs for cloud-moving team maintenance, statistics and report audit.

package v1

type AdminTransportTeam struct {
	Id                      string `json:"id" dc:"Team ID" eg:"12"`
	Name                    string `json:"name" dc:"Team name" eg:"川农云搬牛一团"`
	Status                  string `json:"status" dc:"Team status: effective or invalid" eg:"effective"`
	CreatorId               string `json:"creatorId" dc:"Creator player ID" eg:"8"`
	CreatorName             string `json:"creatorName" dc:"Creator nickname" eg:"川农校友"`
	MemberCount             int    `json:"memberCount" dc:"Current active member count" eg:"3"`
	TotalContributionMeters int64  `json:"totalContributionMeters" dc:"Team cumulative contribution meters" eg:"1250"`
	LastActiveAt            *int64 `json:"lastActiveAt" dc:"Latest active time as Unix milliseconds" eg:"1776333600000"`
	InvalidatedAt           *int64 `json:"invalidatedAt,omitempty" dc:"Invalidation time as Unix milliseconds" eg:"1776592800000"`
	InvalidReason           string `json:"invalidReason" dc:"Invalidation reason" eg:"inactive_72h"`
	CreatedAt               *int64 `json:"createdAt" dc:"Creation time as Unix milliseconds" eg:"1776333600000"`
}

type AdminTransportReport struct {
	Id                 string   `json:"id" dc:"Report ID" eg:"31"`
	TeamId             string   `json:"teamId" dc:"Team ID" eg:"12"`
	TeamName           string   `json:"teamName" dc:"Team name" eg:"川农云搬牛一团"`
	MemberId           string   `json:"memberId" dc:"Membership ID" eg:"21"`
	UserId             string   `json:"userId" dc:"Player ID" eg:"8"`
	UserName           string   `json:"userName" dc:"Player nickname" eg:"川农校友"`
	ActivityDate       string   `json:"activityDate" dc:"Beijing day key" eg:"2026-08-12"`
	StartLat           *float64 `json:"startLat,omitempty" dc:"Previous latitude; omitted on first report" eg:"30.7058"`
	StartLng           *float64 `json:"startLng,omitempty" dc:"Previous longitude; omitted on first report" eg:"103.8317"`
	EndLat             float64  `json:"endLat" dc:"Current latitude" eg:"30.7059"`
	EndLng             float64  `json:"endLng" dc:"Current longitude" eg:"103.8318"`
	SampledAt          *int64   `json:"sampledAt" dc:"Client sample time as Unix milliseconds" eg:"1776333595000"`
	AcceptedAt         *int64   `json:"acceptedAt" dc:"Server acceptance time as Unix milliseconds" eg:"1776333600000"`
	ContributionMeters int64    `json:"contributionMeters" dc:"Contribution meters" eg:"15"`
	UserTotalMeters    int64    `json:"userTotalMeters" dc:"User total after this report" eg:"450"`
	TeamTotalMeters    int64    `json:"teamTotalMeters" dc:"Team total after this report" eg:"1250"`
}
