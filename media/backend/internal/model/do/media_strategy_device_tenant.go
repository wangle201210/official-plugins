// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MediaStrategyDeviceTenant is the golang structure of table media_strategy_device_tenant for DAO operations like Where/Data.
type MediaStrategyDeviceTenant struct {
	g.Meta     `orm:"table:media_strategy_device_tenant, do:true"`
	TenantId   any // 租户id
	DeviceId   any // 设备id（对应device_code）
	StrategyId any // 策略ID
}
