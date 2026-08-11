// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// MiniappConfig is the golang structure of table plugin_sicau_niu_miniapp_config for DAO operations like Where/Data.
type MiniappConfig struct {
	g.Meta             `orm:"table:plugin_sicau_niu_miniapp_config, do:true"`
	Id                 any        //
	ConfigKey          any        //
	AssetsVersion      any        //
	StaticAssetBaseUrl any        //
	ActivityPhase      any        //
	DefaultCampus      any        //
	Anniversary        any        //
	AnniversaryAt      any        //
	Debug              any        //
	CampusesJson       any        // Validated campus map asset and affine calibration JSON
	CreatedAt          *time.Time //
	UpdatedAt          *time.Time //
}
