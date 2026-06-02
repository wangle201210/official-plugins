package uidentity

import (
	"reflect"
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
