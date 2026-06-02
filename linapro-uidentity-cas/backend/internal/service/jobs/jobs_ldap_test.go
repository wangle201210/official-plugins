// This file verifies LDAP request construction for the migrated legacy jobs
// without requiring a live directory server.

package jobs

import (
	"testing"

	"github.com/go-ldap/ldap/v3"

	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

func TestBuildLDAPAddRequestOmitsOpenUID(t *testing.T) {
	req, err := buildLDAPAddRequest(ldapJobConfig{
		baseDN:          "dc=example,dc=test",
		objectClass:     []string{"top", "person"},
		defaultPassword: defaultLDAPPassword,
	}, &entity.Account{
		Number: "2026001",
		Name:   "张三",
		Phone:  "13800138000",
		Status: 1,
	}, &entity.AccountDetail{
		Email:  "zhangsan@example.test",
		Wechat: "open-id",
	}, &entity.Container{Name: legacyContainerStudent})
	if err != nil {
		t.Fatalf("buildLDAPAddRequest returned error: %v", err)
	}
	if ldapAddAttr(req, "securityEmail") != "zhangsan@example.test" {
		t.Fatalf("expected securityEmail to be copied, attrs=%#v", req.Attributes)
	}
	if got := ldapAddAttr(req, "openUID"); got != "" {
		t.Fatalf("expected LDAP add request to omit openUID, got %q", got)
	}
}

func TestBuildLDAPModifyStillSyncsOpenUID(t *testing.T) {
	req := ldap.NewModifyRequest("uid=2026001,ou=bzks,ou=People,dc=example,dc=test", nil)
	changed := buildLDAPModify(req, &ldap.Entry{DN: req.DN, Attributes: []*ldap.EntryAttribute{
		{Name: "openUID", Values: []string{"old-open-id"}},
	}}, &entity.Account{
		Number: "2026001",
		Name:   "张三",
		Phone:  "13800138000",
		Status: 1,
	}, &entity.AccountDetail{
		Email:  "zhangsan@example.test",
		Wechat: "new-open-id",
	}, &entity.Container{Name: legacyContainerStudent})
	if !changed {
		t.Fatal("expected LDAP modify request to change openUID")
	}
	if got := ldapModifyAttr(req, "openUID"); got != "new-open-id" {
		t.Fatalf("expected LDAP modify request to replace openUID, got %q changes=%#v", got, req.Changes)
	}
	if got := ldapModifyAttr(req, "usertype"); got != "" {
		t.Fatalf("expected ordinary LDAP modify request to leave usertype to container moves, got %q", got)
	}
}

func ldapAddAttr(req *ldap.AddRequest, name string) string {
	for _, attr := range req.Attributes {
		if attr.Type == name && len(attr.Vals) > 0 {
			return attr.Vals[0]
		}
	}
	return ""
}

func ldapModifyAttr(req *ldap.ModifyRequest, name string) string {
	for _, change := range req.Changes {
		if change.Modification.Type == name && len(change.Modification.Vals) > 0 {
			return change.Modification.Vals[0]
		}
	}
	return ""
}
