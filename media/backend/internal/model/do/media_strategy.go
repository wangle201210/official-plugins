// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStrategy is the golang structure of table media_strategy for DAO operations like Where/Data.
type MediaStrategy struct {
	g.Meta     `orm:"table:media_strategy, do:true"`
	Id         any         // 策略ID（自增，无符号）
	Name       any         // 策略名称
	Strategy   any         // yaml格式策略内容
	Global     any         // 为1则是全局策略，只能有一个是1，0关闭
	Enable     any         // 1开启，0关闭
	CreatorId  any         // 创建人Id
	CreateTime *gtime.Time // 创建时间
	UpdaterId  any         // 修改人Id
	UpdateTime *gtime.Time // 修改时间
}
