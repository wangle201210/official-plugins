// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaTenantStreamConfig is the golang structure of table media_tenant_stream_config for DAO operations like Where/Data.
type MediaTenantStreamConfig struct {
	g.Meta        `orm:"table:media_tenant_stream_config, do:true"`
	TenantId      any         // 租户id
	MaxConcurrent any         // 最大并发数, 0 禁止访问 （1个会话算1个并发）
	NodeNum       any         // 节点编号
	Enable        any         // 1开启，0关闭
	CreatorId     any         // 创建人Id
	CreateTime    *gtime.Time // 创建时间
	UpdaterId     any         // 修改人Id
	UpdateTime    *gtime.Time // 修改时间
}
