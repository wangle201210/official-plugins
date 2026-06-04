// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaTenantStreamConfig is the golang structure for table media_tenant_stream_config.
type MediaTenantStreamConfig struct {
	TenantId      string      `json:"tenantId"      orm:"tenant_id"      description:"租户id"`
	MaxConcurrent int         `json:"maxConcurrent" orm:"max_concurrent" description:"最大并发数, 0 禁止访问 （1个会话算1个并发）"`
	NodeNum       int         `json:"nodeNum"       orm:"node_num"       description:"节点编号"`
	Enable        int         `json:"enable"        orm:"enable"         description:"1开启，0关闭"`
	CreatorId     int         `json:"creatorId"     orm:"creator_id"     description:"创建人Id"`
	CreateTime    *gtime.Time `json:"createTime"    orm:"create_time"    description:"创建时间"`
	UpdaterId     int         `json:"updaterId"     orm:"updater_id"     description:"修改人Id"`
	UpdateTime    *gtime.Time `json:"updateTime"    orm:"update_time"    description:"修改时间"`
}
