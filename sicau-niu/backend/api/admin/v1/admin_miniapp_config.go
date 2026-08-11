// admin_miniapp_config.go defines operator runtime mini-program config APIs.
package v1

import "github.com/gogf/gf/v2/frame/g"

type GetMiniappConfigReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/config" method:"get" tags:"Sicau Niu Admin" summary:"查询小程序运行配置" dc:"Return the complete operator-maintained public mini-program configuration plus the live activation radius. Protected by host unified permission check." permission:"sicau-niu:settlement:view"`
}

type GetMiniappConfigRes struct {
	*MiniappConfig `json:",inline" dc:"Current complete mini-program runtime configuration" eg:"{}"`
}

type UpdateMiniappConfigReq struct {
	g.Meta             `path:"/plugins/sicau-niu/admin/config" method:"put" tags:"Sicau Niu Admin" summary:"更新小程序运行配置" dc:"Replace validated public mini-program map, asset and activity display configuration. The next public config request sees the update without restart. Protected by host unified permission check." permission:"sicau-niu:settlement:rules"`
	Campuses           []*MiniappCampus `json:"campuses" v:"required|length:3" dc:"Exactly the cd, djy and ya campus configurations" eg:"[]"`
	DefaultCampus      string           `json:"defaultCampus" v:"required|in:cd,djy,ya" dc:"Default campus ID" eg:"cd"`
	Anniversary        string           `json:"anniversary" v:"required|max-length:100" dc:"Anniversary display name" eg:"120 周年校庆"`
	AnniversaryAt      string           `json:"anniversaryAt" dc:"Anniversary date in YYYY-MM-DD; empty disables countdown" eg:"2026-10-06"`
	Debug              bool             `json:"debug" dc:"Whether mini-program debug helpers are enabled" eg:"false"`
	AssetsVersion      string           `json:"assetsVersion" v:"required|max-length:32" dc:"Static asset version" eg:"2026.08.1"`
	StaticAssetBaseUrl string           `json:"staticAssetBaseUrl" v:"max-length:500" dc:"Optional HTTPS remote asset base URL" eg:"https://assets.example.com/sicau-niu"`
	ActivityPhase      string           `json:"activityPhase" v:"required|in:preview,active,closed" dc:"Activity phase" eg:"active"`
}

type UpdateMiniappConfigRes struct {
	*MiniappConfig `json:",inline" dc:"Updated complete mini-program runtime configuration" eg:"{}"`
}

type MiniappConfig struct {
	Campuses           []*MiniappCampus `json:"campuses" dc:"Supported campus configurations" eg:"[]"`
	DefaultCampus      string           `json:"defaultCampus" dc:"Default campus ID" eg:"cd"`
	ActivateRadiusM    int              `json:"activateRadiusM" dc:"Live activation radius from runtime rules" eg:"50"`
	CountdownDays      int              `json:"countdownDays" dc:"Whole countdown days" eg:"56"`
	Anniversary        string           `json:"anniversary" dc:"Anniversary display name" eg:"120 周年校庆"`
	AnniversaryAt      string           `json:"anniversaryAt" dc:"Anniversary date in YYYY-MM-DD" eg:"2026-10-06"`
	Debug              bool             `json:"debug" dc:"Debug helper flag" eg:"false"`
	AssetsVersion      string           `json:"assetsVersion" dc:"Static asset version" eg:"2026.08.1"`
	StaticAssetBaseUrl string           `json:"staticAssetBaseUrl" dc:"Optional HTTPS asset base URL" eg:"https://assets.example.com/sicau-niu"`
	ActivityPhase      string           `json:"activityPhase" dc:"Activity phase" eg:"active"`
}

type MiniappCampus struct {
	Id          string             `json:"id" dc:"Campus ID" eg:"cd"`
	Name        string             `json:"name" dc:"Campus display name" eg:"成都校区"`
	MapImage    string             `json:"mapImage" dc:"Optional map image path or HTTPS URL" eg:"/maps/cd.jpg"`
	Calibration MiniappCalibration `json:"calibration" dc:"GCJ-02 map affine calibration" eg:"{}"`
}

type MiniappCalibration struct {
	Version   int                    `json:"version" dc:"Calibration version" eg:"1"`
	Origin    MiniappOrigin          `json:"origin" dc:"GCJ-02 origin" eg:"{}"`
	ImageSize MiniappImageSize       `json:"imageSize" dc:"Map pixel dimensions" eg:"{}"`
	Affine    MiniappAffineTransform `json:"affine" dc:"Affine transform" eg:"{}"`
}

type MiniappOrigin struct {
	Lat0 float64 `json:"lat0" dc:"Origin latitude" eg:"30.7058"`
	Lng0 float64 `json:"lng0" dc:"Origin longitude" eg:"103.8318"`
}

type MiniappImageSize struct {
	W int `json:"w" dc:"Image width" eg:"16644"`
	H int `json:"h" dc:"Image height" eg:"8241"`
}

type MiniappAffineTransform struct {
	A  float64 `json:"a" dc:"Affine A" eg:"1.785714"`
	B  float64 `json:"b" dc:"Affine B" eg:"0"`
	C  float64 `json:"c" dc:"Affine C" eg:"0"`
	D  float64 `json:"d" dc:"Affine D" eg:"-1.785714"`
	Tx float64 `json:"tx" dc:"X translation" eg:"8322"`
	Ty float64 `json:"ty" dc:"Y translation" eg:"4120.5"`
}
