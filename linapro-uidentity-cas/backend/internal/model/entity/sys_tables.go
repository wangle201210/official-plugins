// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysTables is the golang structure for table sys_tables.
type SysTables struct {
	TableId             int64      `json:"tableId"             orm:"table_id"              description:""`
	TableName           string     `json:"tableName"           orm:"table_name"            description:""`
	TableComment        string     `json:"tableComment"        orm:"table_comment"         description:""`
	ClassName           string     `json:"className"           orm:"class_name"            description:""`
	TplCategory         string     `json:"tplCategory"         orm:"tpl_category"          description:""`
	PackageName         string     `json:"packageName"         orm:"package_name"          description:""`
	ModuleName          string     `json:"moduleName"          orm:"module_name"           description:""`
	ModuleFrontName     string     `json:"moduleFrontName"     orm:"module_front_name"     description:""`
	BusinessName        string     `json:"businessName"        orm:"business_name"         description:""`
	FunctionName        string     `json:"functionName"        orm:"function_name"         description:""`
	FunctionAuthor      string     `json:"functionAuthor"      orm:"function_author"       description:""`
	PkColumn            string     `json:"pkColumn"            orm:"pk_column"             description:""`
	PkGoField           string     `json:"pkGoField"           orm:"pk_go_field"           description:""`
	PkJsonField         string     `json:"pkJsonField"         orm:"pk_json_field"         description:""`
	Options             string     `json:"options"             orm:"options"               description:""`
	TreeCode            string     `json:"treeCode"            orm:"tree_code"             description:""`
	TreeParentCode      string     `json:"treeParentCode"      orm:"tree_parent_code"      description:""`
	TreeName            string     `json:"treeName"            orm:"tree_name"             description:""`
	Tree                bool       `json:"tree"                orm:"tree"                  description:""`
	Crud                bool       `json:"crud"                orm:"crud"                  description:""`
	Remark              string     `json:"remark"              orm:"remark"                description:""`
	IsDataScope         int64      `json:"isDataScope"         orm:"is_data_scope"         description:""`
	IsActions           int64      `json:"isActions"           orm:"is_actions"            description:""`
	IsAuth              int64      `json:"isAuth"              orm:"is_auth"               description:""`
	IsLogicalDelete     string     `json:"isLogicalDelete"     orm:"is_logical_delete"     description:""`
	LogicalDelete       bool       `json:"logicalDelete"       orm:"logical_delete"        description:""`
	LogicalDeleteColumn string     `json:"logicalDeleteColumn" orm:"logical_delete_column" description:""`
	CreateBy            int64      `json:"createBy"            orm:"create_by"             description:""`
	UpdateBy            int64      `json:"updateBy"            orm:"update_by"             description:""`
	CreatedAt           *time.Time `json:"createdAt"           orm:"created_at"            description:""`
	UpdatedAt           *time.Time `json:"updatedAt"           orm:"updated_at"            description:""`
	DeletedAt           *time.Time `json:"deletedAt"           orm:"deleted_at"            description:""`
}
