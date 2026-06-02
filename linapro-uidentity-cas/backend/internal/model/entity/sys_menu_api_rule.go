// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysMenuApiRule is the golang structure for table sys_menu_api_rule.
type SysMenuApiRule struct {
	SysMenuMenuId int64 `json:"sysMenuMenuId" orm:"sys_menu_menu_id" description:""`
	SysApiId      int64 `json:"sysApiId"      orm:"sys_api_id"       description:""`
}
