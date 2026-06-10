// This file verifies the account write-path LDAP DN layout, object class
// resolution, and the no-config skip behavior without a live directory.

package uidentity

import (
	"context"
	"testing"
)

// TestLegacyLDAPAccountDNMatchesJobLayout verifies the write-path DN matches
// the nightly job's uid=<number>,ou=<container>,ou=People,<baseDN> layout.
func TestLegacyLDAPAccountDNMatchesJobLayout(t *testing.T) {
	t.Parallel()

	baseDN := "dc=sicau,dc=edu,dc=cn"
	if got := legacyLDAPContainerDN(baseDN, "bzks"); got != "ou=bzks,ou=People,dc=sicau,dc=edu,dc=cn" {
		t.Fatalf("container DN = %q", got)
	}
	if got := legacyLDAPAccountDN(baseDN, "bzks", "A001"); got != "uid=A001,ou=bzks,ou=People,dc=sicau,dc=edu,dc=cn" {
		t.Fatalf("account DN = %q", got)
	}
}

// TestRuntimeLDAPObjectClassDefaultsMatchJob verifies the default object class
// list matches the nightly job's.
func TestRuntimeLDAPObjectClassDefaultsMatchJob(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	classes, err := service.runtimeLDAPObjectClass(context.Background())
	if err != nil {
		t.Fatalf("object class read failed: %v", err)
	}
	want := []string{"top", "person", "inetOrgPerson", "organizationalPerson", "xUserObjectClass"}
	if len(classes) != len(want) {
		t.Fatalf("object class = %v, want %v", classes, want)
	}
	for i := range want {
		if classes[i] != want[i] {
			t.Fatalf("object class[%d] = %q, want %q", i, classes[i], want[i])
		}
	}
}

// TestUpsertAccountLDAPSkipsWithoutConfig verifies the write path is a no-op
// when LDAP is not configured, so account writes keep working without LDAP.
func TestUpsertAccountLDAPSkipsWithoutConfig(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	if err := service.upsertAccountLDAPByID(context.Background(), 123); err != nil {
		t.Fatalf("expected no-op without LDAP config, got %v", err)
	}
}
