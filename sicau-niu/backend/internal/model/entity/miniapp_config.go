// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// MiniappConfig is the golang structure for table miniapp_config.
type MiniappConfig struct {
	Id                 int64      `json:"id"                 orm:"id"                    description:""`
	ConfigKey          string     `json:"configKey"          orm:"config_key"            description:""`
	AssetsVersion      string     `json:"assetsVersion"      orm:"assets_version"        description:""`
	StaticAssetBaseUrl string     `json:"staticAssetBaseUrl" orm:"static_asset_base_url" description:""`
	ActivityPhase      string     `json:"activityPhase"      orm:"activity_phase"        description:""`
	DefaultCampus      string     `json:"defaultCampus"      orm:"default_campus"        description:""`
	Anniversary        string     `json:"anniversary"        orm:"anniversary"           description:""`
	AnniversaryAt      string     `json:"anniversaryAt"      orm:"anniversary_at"        description:""`
	Debug              int        `json:"debug"              orm:"debug"                 description:""`
	CampusesJson       string     `json:"campusesJson"       orm:"campuses_json"         description:"Validated campus map asset and affine calibration JSON"`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"            description:""`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"            description:""`
}
