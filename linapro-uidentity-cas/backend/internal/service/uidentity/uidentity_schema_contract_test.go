package uidentity

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

func TestOauth2TokenLegacySchemaContract(t *testing.T) {
	cols := dao.Oauth2Token.Columns()
	got := []string{cols.Id, cols.ExpiredAt, cols.Code, cols.Access, cols.Refresh, cols.Data}
	want := []string{"id", "expired_at", "code", "access", "refresh", "data"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("oauth2_token columns = %#v, want %#v", got, want)
	}

	columnType := reflect.TypeOf(cols)
	if columnType.NumField() != len(want) {
		t.Fatalf("oauth2_token generated column count = %d, want %d", columnType.NumField(), len(want))
	}
	for _, field := range []string{"CreatedAt", "UpdatedAt", "DeletedAt", "CreateBy", "UpdateBy"} {
		if _, ok := columnType.FieldByName(field); ok {
			t.Fatalf("oauth2_token generated columns still expose legacy-incompatible field %s", field)
		}
	}

	def := (&serviceImpl{}).oauthTokenResource()
	for _, apiName := range []string{"createdAt", "updatedAt", "deletedAt", "createBy", "updateBy"} {
		if _, ok := def.apiToColumn[apiName]; ok {
			t.Fatalf("oauth-token resource still exposes legacy-incompatible API field %s", apiName)
		}
	}
}

func TestLegacySchemaColumnNamesAndGeneratedTypes(t *testing.T) {
	detailCols := dao.AccountDetails.Columns()
	detailColumns := []string{
		detailCols.AccountId, detailCols.Birthday, detailCols.Email, detailCols.Gender, detailCols.Qq,
		detailCols.Wechat, detailCols.Idcard, detailCols.Avatar, detailCols.Source, detailCols.Nj,
		detailCols.Xymc, detailCols.Xydm, detailCols.Xq, detailCols.Xz, detailCols.Yjbysj,
		detailCols.Zymc, detailCols.Bjmc, detailCols.Face, detailCols.CreatedAt, detailCols.UpdatedAt,
		detailCols.DeletedAt, detailCols.CreateBy, detailCols.UpdateBy,
	}
	wantDetailColumns := []string{
		"account_id", "birthday", "email", "gender", "qq", "wechat", "idcard", "avatar", "source", "nj",
		"xymc", "xydm", "xq", "xz", "yjbysj", "zymc", "bjmc", "face", "created_at", "updated_at",
		"deleted_at", "create_by", "update_by",
	}
	if !reflect.DeepEqual(detailColumns, wantDetailColumns) {
		t.Fatalf("account_details columns = %#v, want %#v", detailColumns, wantDetailColumns)
	}

	detailType := reflect.TypeOf(entity.AccountDetails{})
	for field, want := range map[string]reflect.Type{
		"Gender": reflect.TypeOf(int64(0)),
		"Nj":     reflect.TypeOf(int64(0)),
		"Xz":     reflect.TypeOf(int64(0)),
		"Yjbysj": reflect.TypeOf(int64(0)),
		"Face":   reflect.TypeOf(int64(0)),
	} {
		assertEntityFieldType(t, detailType, field, want)
	}

	accountUnitCols := dao.AccountUnit.Columns()
	accountUnitColumns := []string{accountUnitCols.Id, accountUnitCols.AccountId, accountUnitCols.UnitId}
	wantAccountUnitColumns := []string{"id", "account_id", "unit_id"}
	if !reflect.DeepEqual(accountUnitColumns, wantAccountUnitColumns) {
		t.Fatalf("account_unit columns = %#v, want %#v", accountUnitColumns, wantAccountUnitColumns)
	}
	accountUnitType := reflect.TypeOf(accountUnitCols)
	for _, field := range []string{"UnitsId", "CreatedAt", "UpdatedAt", "DeletedAt", "CreateBy", "UpdateBy"} {
		if _, ok := accountUnitType.FieldByName(field); ok {
			t.Fatalf("account_unit generated columns still expose legacy-incompatible field %s", field)
		}
	}

	assertEntityFieldType(t, reflect.TypeOf(entity.AccountChangeLog{}), "ErrNumber", reflect.TypeOf(""))
}

func TestLegacySchemaTableSuffixesAndColumnSets(t *testing.T) {
	const tablePrefix = "plugin_linapro_uidentity_cas_"

	type schemaContract struct {
		name    string
		table   string
		columns any
		want    []string
	}

	contracts := []schemaContract{
		{
			name:    "account",
			table:   dao.Account.Table(),
			columns: dao.Account.Columns(),
			want: []string{
				"id", "number", "name", "phone", "effect_at", "expire_at", "group_id", "pass_level",
				"container_id", "unit_id", "status", "created_at", "updated_at", "deleted_at", "create_by", "update_by",
			},
		},
		{
			name:    "account_details",
			table:   dao.AccountDetails.Table(),
			columns: dao.AccountDetails.Columns(),
			want: []string{
				"account_id", "birthday", "email", "gender", "qq", "wechat", "idcard", "avatar", "source", "nj",
				"xymc", "xydm", "xq", "xz", "yjbysj", "zymc", "bjmc", "face", "created_at", "updated_at",
				"deleted_at", "create_by", "update_by",
			},
		},
		{
			name:    "groups",
			table:   dao.Groups.Table(),
			columns: dao.Groups.Columns(),
			want:    []string{"id", "name", "alias", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "units",
			table:   dao.Units.Table(),
			columns: dao.Units.Columns(),
			want:    []string{"id", "name", "alias", "code", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "containers",
			table:   dao.Containers.Table(),
			columns: dao.Containers.Columns(),
			want:    []string{"id", "name", "alias", "account_count", "admin_count", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "applications",
			table:   dao.Applications.Table(),
			columns: dao.Applications.Columns(),
			want: []string{
				"id", "name", "alias", "client_id", "secret_key", "access_model", "status", "callback_url",
				"whitelist", "created_at", "updated_at", "deleted_at", "create_by", "update_by",
			},
		},
		{
			name:    "account_group",
			table:   dao.AccountGroup.Table(),
			columns: dao.AccountGroup.Columns(),
			want:    []string{"account_id", "groups_id"},
		},
		{
			name:    "account_unit",
			table:   dao.AccountUnit.Table(),
			columns: dao.AccountUnit.Columns(),
			want:    []string{"id", "account_id", "unit_id"},
		},
		{
			name:    "account_app_role",
			table:   dao.AccountAppRole.Table(),
			columns: dao.AccountAppRole.Columns(),
			want:    []string{"id", "give_account_id", "empowered_account_id", "app_id", "expire_at", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "account_app_blacklist",
			table:   dao.AccountAppBlacklist.Table(),
			columns: dao.AccountAppBlacklist.Columns(),
			want:    []string{"id", "name", "app_id", "account_id", "effect_at", "expire_at", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "group_app_blacklist",
			table:   dao.GroupAppBlacklist.Table(),
			columns: dao.GroupAppBlacklist.Columns(),
			want:    []string{"id", "name", "app_id", "group_id", "effect_at", "expire_at", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "pass_ruler",
			table:   dao.PassRuler.Table(),
			columns: dao.PassRuler.Columns(),
			want: []string{
				"id", "name", "capital", "lower", "number", "symbol", "length", "interval",
				"interval_status", "status", "created_at", "updated_at", "deleted_at", "create_by", "update_by",
			},
		},
		{
			name:    "sms",
			table:   dao.Sms.Table(),
			columns: dao.Sms.Columns(),
			want:    []string{"id", "phone", "type", "content", "status", "resp_msg", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "cas_login_log",
			table:   dao.CasLoginLog.Table(),
			columns: dao.CasLoginLog.Columns(),
			want: []string{
				"id", "account_id", "choice_account_id", "app_id", "ipaddr", "login_location", "browser", "os",
				"platform", "login_time", "remark", "msg", "login_type", "created_at", "updated_at",
				"deleted_at", "create_by", "update_by",
			},
		},
		{
			name:    "oauth_log",
			table:   dao.OauthLog.Table(),
			columns: dao.OauthLog.Columns(),
			want:    []string{"id", "user_id", "app_id", "redirect_uri", "scope", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "oauth2_token",
			table:   dao.Oauth2Token.Table(),
			columns: dao.Oauth2Token.Columns(),
			want:    []string{"id", "expired_at", "code", "access", "refresh", "data"},
		},
		{
			name:    "account_change_log",
			table:   dao.AccountChangeLog.Table(),
			columns: dao.AccountChangeLog.Columns(),
			want:    []string{"id", "account_id", "table_name", "action", "data_old", "data_new", "err_msg", "err_number", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
		},
		{
			name:    "account_active_log",
			table:   dao.AccountActiveLog.Table(),
			columns: dao.AccountActiveLog.Columns(),
			want:    []string{"id", "number", "phone", "wechat", "created_at", "type"},
		},
	}

	for _, contract := range contracts {
		t.Run(contract.name, func(t *testing.T) {
			if !strings.HasPrefix(contract.table, tablePrefix) {
				t.Fatalf("table %q does not use plugin prefix %q", contract.table, tablePrefix)
			}
			if suffix := strings.TrimPrefix(contract.table, tablePrefix); suffix != contract.name {
				t.Fatalf("table suffix = %q, want %q", suffix, contract.name)
			}
			if got := columnValues(contract.columns); !reflect.DeepEqual(got, contract.want) {
				t.Fatalf("%s columns = %#v, want %#v", contract.name, got, contract.want)
			}
		})
	}
}

func TestLegacySystemSchemaSQLTableSuffixesAndColumnSets(t *testing.T) {
	tables := legacySystemSQLTables(t)
	const tablePrefix = "plugin_linapro_uidentity_cas_"
	contracts := map[string][]string{
		"sys_dept": {
			"dept_id", "parent_id", "dept_path", "dept_name", "sort", "leader", "phone", "email",
			"status", "create_by", "update_by", "created_at", "updated_at", "deleted_at",
		},
		"sys_config": {
			"id", "config_name", "config_key", "config_value", "config_type", "is_frontend",
			"remark", "create_by", "update_by", "created_at", "updated_at", "deleted_at",
		},
		"sys_tables": {
			"table_id", "table_name", "table_comment", "class_name", "tpl_category", "package_name",
			"module_name", "module_front_name", "business_name", "function_name", "function_author",
			"pk_column", "pk_go_field", "pk_json_field", "options", "tree_code", "tree_parent_code",
			"tree_name", "tree", "crud", "remark", "is_data_scope", "is_actions", "is_auth",
			"is_logical_delete", "logical_delete", "logical_delete_column", "create_by", "update_by",
			"created_at", "updated_at", "deleted_at",
		},
		"sys_columns": {
			"column_id", "table_id", "column_name", "column_comment", "column_type", "go_type",
			"go_field", "json_field", "is_pk", "is_increment", "is_required", "is_insert",
			"is_edit", "is_list", "is_query", "query_type", "html_type", "dict_type", "sort",
			"list", "pk", "required", "super_column", "usable_column", "increment", "insert",
			"edit", "query", "remark", "fk_table_name", "fk_table_name_class", "fk_table_name_package",
			"fk_label_id", "fk_label_name", "create_by", "update_by", "created_at", "updated_at",
			"deleted_at",
		},
		"sys_menu": {
			"menu_id", "menu_name", "title", "icon", "path", "paths", "menu_type", "action",
			"permission", "parent_id", "no_cache", "breadcrumb", "component", "sort", "visible",
			"is_frame", "create_by", "update_by", "created_at", "updated_at", "deleted_at",
		},
		"sys_login_log": {
			"id", "username", "status", "ipaddr", "login_location", "browser", "os", "platform",
			"login_time", "remark", "msg", "created_at", "updated_at", "create_by", "update_by",
		},
		"sys_opera_log": {
			"id", "title", "business_type", "business_types", "method", "request_method",
			"operator_type", "oper_name", "dept_name", "oper_url", "oper_ip", "oper_location",
			"oper_param", "status", "oper_time", "json_result", "remark", "latency_time",
			"user_agent", "created_at", "updated_at", "create_by", "update_by",
		},
		"sys_role_dept": {"role_id", "dept_id"},
		"sys_user": {
			"user_id", "username", "password", "nick_name", "phone", "role_id", "salt", "avatar",
			"sex", "email", "dept_id", "post_id", "remark", "status", "create_by", "update_by",
			"created_at", "updated_at", "deleted_at",
		},
		"sys_role": {
			"role_id", "role_name", "status", "role_key", "role_sort", "flag", "remark", "admin",
			"data_scope", "create_by", "update_by", "created_at", "updated_at", "deleted_at",
		},
		"sys_role_menu": {"role_id", "menu_id"},
		"sys_post": {
			"post_id", "post_name", "post_code", "sort", "status", "remark", "create_by",
			"update_by", "created_at", "updated_at", "deleted_at",
		},
		"sys_dict_data": {
			"dict_code", "dict_sort", "dict_label", "dict_value", "dict_type", "css_class",
			"list_class", "is_default", "status", "default", "remark", "create_by", "update_by",
			"created_at", "updated_at", "deleted_at",
		},
		"sys_dict_type": {
			"dict_id", "dict_name", "dict_type", "status", "remark", "create_by", "update_by",
			"created_at", "updated_at", "deleted_at",
		},
		"sys_api": {
			"id", "handle", "title", "path", "type", "action", "create_by", "update_by",
			"created_at", "updated_at", "deleted_at",
		},
		"sys_menu_api_rule": {"sys_menu_menu_id", "sys_api_id"},
		"sys_job": {
			"job_id", "job_name", "job_group", "job_type", "cron_expression", "invoke_target",
			"args", "misfire_policy", "concurrent", "status", "entry_id", "create_by", "update_by",
			"created_at", "updated_at", "deleted_at",
		},
		"job_log": {
			"id", "job_id", "job_name", "start_at", "end_at", "create_num", "update_num",
			"delete_num", "err_num", "create_by", "update_by", "created_at", "updated_at", "deleted_at",
		},
		"sys_casbin_rule": {"id", "ptype", "v0", "v1", "v2", "v3", "v4", "v5"},
		"tb_demo":         {"id", "name", "created_at", "updated_at", "deleted_at", "create_by", "update_by"},
	}

	for suffix, want := range contracts {
		t.Run(suffix, func(t *testing.T) {
			columns, ok := tables[tablePrefix+suffix]
			if !ok {
				t.Fatalf("missing legacy system table %s%s", tablePrefix, suffix)
			}
			if !reflect.DeepEqual(columns, want) {
				t.Fatalf("%s columns = %#v, want %#v", suffix, columns, want)
			}
		})
	}
}

func assertEntityFieldType(t *testing.T, entityType reflect.Type, fieldName string, want reflect.Type) {
	t.Helper()
	field, ok := entityType.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s missing field %s", entityType.Name(), fieldName)
	}
	if field.Type != want {
		t.Fatalf("%s.%s type = %s, want %s", entityType.Name(), fieldName, field.Type, want)
	}
}

func columnValues(columns any) []string {
	value := reflect.ValueOf(columns)
	result := make([]string, 0, value.NumField())
	for i := 0; i < value.NumField(); i++ {
		result = append(result, value.Field(i).String())
	}
	return result
}

func legacySystemSQLTables(t *testing.T) map[string][]string {
	t.Helper()
	sqlPath := filepath.Join("..", "..", "..", "..", "manifest", "sql", "001-implement-uidentity-cas-backend.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read plugin install SQL: %v", err)
	}
	createTableRe := regexp.MustCompile(`(?is)CREATE TABLE IF NOT EXISTS\s+([a-z0-9_]+)\s*\((.*?)\);`)
	result := make(map[string][]string)
	for _, match := range createTableRe.FindAllStringSubmatch(string(content), -1) {
		table := strings.TrimSpace(match[1])
		result[table] = legacySQLColumnNames(match[2])
	}
	return result
}

func legacySQLColumnNames(body string) []string {
	lines := strings.Split(body, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimSuffix(line, ","))
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "primary key") ||
			strings.HasPrefix(lower, "unique") ||
			strings.HasPrefix(lower, "constraint") {
			continue
		}
		column := strings.Trim(strings.Fields(line)[0], `"`)
		result = append(result, column)
	}
	return result
}
