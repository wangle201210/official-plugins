// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MediaStrategyDeviceTenant is the golang structure for table media_strategy_device_tenant.
type MediaStrategyDeviceTenant struct {
	TenantId   string `json:"tenantId"   orm:"tenant_id"   description:"租户id"`
	DeviceId   string `json:"deviceId"   orm:"device_id"   description:"设备id（对应device_code）"`
	StrategyId int64  `json:"strategyId" orm:"strategy_id" description:"策略ID"`
}
