// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStrategy is the golang structure for table media_strategy.
type MediaStrategy struct {
	Id         int64       `json:"id"         orm:"id"          description:"策略ID（自增，无符号）"`
	Name       string      `json:"name"       orm:"name"        description:"策略名称"`
	Strategy   string      `json:"strategy"   orm:"strategy"    description:"yaml格式策略内容"`
	Global     int         `json:"global"     orm:"global"      description:"为1则是全局策略，只能有一个是1，0关闭"`
	Enable     int         `json:"enable"     orm:"enable"      description:"1开启，0关闭"`
	CreatorId  int         `json:"creatorId"  orm:"creator_id"  description:"创建人Id"`
	CreateTime *gtime.Time `json:"createTime" orm:"create_time" description:"创建时间"`
	UpdaterId  int         `json:"updaterId"  orm:"updater_id"  description:"修改人Id"`
	UpdateTime *gtime.Time `json:"updateTime" orm:"update_time" description:"修改时间"`
}
