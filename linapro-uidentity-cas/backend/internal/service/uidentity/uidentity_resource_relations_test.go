// This file locks the nested relation response keys to the old gorm model
// JSON contract (untagged relation fields marshal under their Go field names).

package uidentity

import "testing"

func TestLegacyResourceRelationsKeepOldPreloadKeys(t *testing.T) {
	t.Parallel()

	expected := map[string][][3]string{
		"account-app-blacklists": {
			{"App", "applications", "appId"},
			{"Account", "accounts", "accountId"},
		},
		"group-app-blacklists": {
			{"App", "applications", "appId"},
			{"Group", "groups", "groupId"},
		},
		"account-app-roles": {
			{"App", "applications", "appId"},
			{"EmpoweredAccount", "accounts", "empoweredAccountId"},
		},
		"cas-login-logs": {
			{"Account", "accounts", "accountId"},
			{"ChoiceAccount", "accounts", "choiceAccountId"},
			{"Application", "applications", "appId"},
		},
		"oauth-logs": {
			{"Account", "accounts", "userId"},
			{"Application", "applications", "appId"},
		},
	}
	for resource, want := range expected {
		got := legacyResourceRelations(resource)
		if len(got) != len(want) {
			t.Fatalf("%s relations = %d, want %d", resource, len(got), len(want))
		}
		for i, relation := range got {
			if relation.responseKey != want[i][0] || relation.resource != want[i][1] || relation.idField != want[i][2] {
				t.Fatalf("%s relation %d = %+v, want %v", resource, i, relation, want[i])
			}
		}
	}
	if legacyResourceRelations("accounts") != nil {
		t.Fatal("accounts must use the dedicated decorator, not generic relations")
	}
}
