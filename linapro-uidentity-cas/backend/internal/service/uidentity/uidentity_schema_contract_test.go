package uidentity

import (
	"reflect"
	"testing"

	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
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
