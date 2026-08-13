// player_config.go defines the unauthenticated, non-sensitive mini-program
// runtime configuration contract.

package v1

import "github.com/gogf/gf/v2/frame/g"

type ConfigReq struct {
	g.Meta `path:"/plugins/sicau-niu/config" method:"get" tags:"寻牛小程序" summary:"查询小程序公开运行配置" dc:"Return non-sensitive mini-program campus calibration, activity settings and live gameplay limits. No player token is required and no credential is exposed."`
}

type ConfigRes struct {
	Campuses           []*CampusConfig `json:"campuses" dc:"Supported campus map calibrations" eg:"[]"`
	DefaultCampus      string          `json:"defaultCampus" dc:"Default campus ID" eg:"cd"`
	ActivateRadiusM    int             `json:"activateRadiusM" dc:"Live activation distance in meters" eg:"50"`
	StealDailyLimit    int             `json:"stealDailyLimit" dc:"Live per-player daily steal action limit" eg:"5"`
	GiftDailyLimit     int             `json:"giftDailyLimit" dc:"Live per-player daily gift action limit" eg:"12"`
	GiftMinAmount      int             `json:"giftMinAmount" dc:"Live minimum grass amount for one gift" eg:"12"`
	CountdownDays      int             `json:"countdownDays" dc:"Whole days until the configured anniversary date" eg:"56"`
	Anniversary        string          `json:"anniversary" dc:"Anniversary display name" eg:"120 周年校庆"`
	Debug              bool            `json:"debug" dc:"Whether mini-program debug helpers are enabled" eg:"false"`
	AssetsVersion      string          `json:"assetsVersion" dc:"Static asset version used for client cache invalidation" eg:"2026.08.1"`
	StaticAssetBaseUrl string          `json:"staticAssetBaseUrl" dc:"Optional HTTPS base URL for remote mini-program map assets" eg:"https://assets.example.com/sicau-niu"`
	ActivityPhase      string          `json:"activityPhase" dc:"Activity phase: preview, active or closed" eg:"active"`
}

type CampusConfig struct {
	Id          string            `json:"id" dc:"Campus ID: cd, djy, ya" eg:"cd"`
	Name        string            `json:"name" dc:"Campus display name" eg:"成都校区"`
	MapImage    string            `json:"mapImage" dc:"Optional absolute, HTTPS or static-base-relative map image path" eg:"/maps/cd.jpg"`
	Calibration CampusCalibration `json:"calibration" dc:"GCJ-02 to map-pixel affine calibration" eg:"{}"`
}

type CampusCalibration struct {
	Version   int               `json:"version" dc:"Calibration schema version" eg:"1"`
	Origin    CalibrationOrigin `json:"origin" dc:"GCJ-02 calibration origin" eg:"{}"`
	ImageSize CalibrationSize   `json:"imageSize" dc:"Map image pixel dimensions" eg:"{}"`
	Affine    CalibrationAffine `json:"affine" dc:"Affine transform coefficients" eg:"{}"`
}

type CalibrationOrigin struct {
	Lat0 float64 `json:"lat0" dc:"GCJ-02 origin latitude" eg:"30.7058"`
	Lng0 float64 `json:"lng0" dc:"GCJ-02 origin longitude" eg:"103.8318"`
}

type CalibrationSize struct {
	W int `json:"w" dc:"Image width in pixels" eg:"16644"`
	H int `json:"h" dc:"Image height in pixels" eg:"8241"`
}

type CalibrationAffine struct {
	A  float64 `json:"a" dc:"Affine coefficient A" eg:"1.785714"`
	B  float64 `json:"b" dc:"Affine coefficient B" eg:"0"`
	C  float64 `json:"c" dc:"Affine coefficient C" eg:"0"`
	D  float64 `json:"d" dc:"Affine coefficient D" eg:"-1.785714"`
	Tx float64 `json:"tx" dc:"Affine X translation" eg:"8322"`
	Ty float64 `json:"ty" dc:"Affine Y translation" eg:"4120.5"`
}
