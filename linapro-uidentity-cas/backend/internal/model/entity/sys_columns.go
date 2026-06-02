// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysColumns is the golang structure for table sys_columns.
type SysColumns struct {
	ColumnId           int64      `json:"columnId"           orm:"column_id"             description:""`
	TableId            int64      `json:"tableId"            orm:"table_id"              description:""`
	ColumnName         string     `json:"columnName"         orm:"column_name"           description:""`
	ColumnComment      string     `json:"columnComment"      orm:"column_comment"        description:""`
	ColumnType         string     `json:"columnType"         orm:"column_type"           description:""`
	GoType             string     `json:"goType"             orm:"go_type"               description:""`
	GoField            string     `json:"goField"            orm:"go_field"              description:""`
	JsonField          string     `json:"jsonField"          orm:"json_field"            description:""`
	IsPk               string     `json:"isPk"               orm:"is_pk"                 description:""`
	IsIncrement        string     `json:"isIncrement"        orm:"is_increment"          description:""`
	IsRequired         string     `json:"isRequired"         orm:"is_required"           description:""`
	IsInsert           string     `json:"isInsert"           orm:"is_insert"             description:""`
	IsEdit             string     `json:"isEdit"             orm:"is_edit"               description:""`
	IsList             string     `json:"isList"             orm:"is_list"               description:""`
	IsQuery            string     `json:"isQuery"            orm:"is_query"              description:""`
	QueryType          string     `json:"queryType"          orm:"query_type"            description:""`
	HtmlType           string     `json:"htmlType"           orm:"html_type"             description:""`
	DictType           string     `json:"dictType"           orm:"dict_type"             description:""`
	Sort               int64      `json:"sort"               orm:"sort"                  description:""`
	List               string     `json:"list"               orm:"list"                  description:""`
	Pk                 bool       `json:"pk"                 orm:"pk"                    description:""`
	Required           bool       `json:"required"           orm:"required"              description:""`
	SuperColumn        bool       `json:"superColumn"        orm:"super_column"          description:""`
	UsableColumn       bool       `json:"usableColumn"       orm:"usable_column"         description:""`
	Increment          bool       `json:"increment"          orm:"increment"             description:""`
	Insert             bool       `json:"insert"             orm:"insert"                description:""`
	Edit               bool       `json:"edit"               orm:"edit"                  description:""`
	Query              bool       `json:"query"              orm:"query"                 description:""`
	Remark             string     `json:"remark"             orm:"remark"                description:""`
	FkTableName        string     `json:"fkTableName"        orm:"fk_table_name"         description:""`
	FkTableNameClass   string     `json:"fkTableNameClass"   orm:"fk_table_name_class"   description:""`
	FkTableNamePackage string     `json:"fkTableNamePackage" orm:"fk_table_name_package" description:""`
	FkLabelId          string     `json:"fkLabelId"          orm:"fk_label_id"           description:""`
	FkLabelName        string     `json:"fkLabelName"        orm:"fk_label_name"         description:""`
	CreateBy           int64      `json:"createBy"           orm:"create_by"             description:""`
	UpdateBy           int64      `json:"updateBy"           orm:"update_by"             description:""`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"            description:""`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"            description:""`
	DeletedAt          *time.Time `json:"deletedAt"          orm:"deleted_at"            description:""`
}
