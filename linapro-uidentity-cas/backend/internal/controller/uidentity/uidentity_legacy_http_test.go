// This file tests small old-contract projection helpers used by the legacy
// uidentity/admin HTTP compatibility layer.

package uidentity

import (
	"testing"

	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

func TestLegacyRuntimeAccountPayloadKeepsOldFieldNames(t *testing.T) {
	expireAt := int64(1780372800000)
	payload := legacyRuntimeAccountPayload(&uidentitysvc.RuntimeAccount{
		ID:            7,
		Number:        "A001",
		Name:          "Legacy User",
		Phone:         "13800000000",
		Status:        1,
		PassLevel:     3,
		ContainerID:   11,
		ContainerName: "main",
		UnitID:        22,
		UnitName:      "unit",
		ExpireAt:      &expireAt,
		Groups:        []string{"staff"},
		Detail: &uidentitysvc.RuntimeAccountDetail{
			Email:  "legacy@example.com",
			QQ:     "10001",
			Wechat: "wx",
			Idcard: "510000000000000000",
		},
	})
	if payload["pass_level"] != 3 || payload["passLevel"] != 3 {
		t.Fatalf("legacy pass level fields missing: %#v", payload)
	}
	if payload["unit"] != "unit" || payload["unit_name"] != "unit" || payload["unitName"] != "unit" {
		t.Fatalf("legacy unit fields missing: %#v", payload)
	}
	if payload["idcard"] != "510000000000000000" || payload["idCard"] != "510000000000000000" {
		t.Fatalf("legacy idcard fields missing: %#v", payload)
	}
}

func TestLegacyRuntimeLoginPayloadKeepsTicketFields(t *testing.T) {
	payload := legacyRuntimeLoginPayload(&uidentitysvc.RuntimeLoginOutput{
		CallbackURL: "https://app.example.com/callback",
		TGT:         "1:TGT:abc",
		ST:          "ST-abc",
		User:        &uidentitysvc.RuntimeAccount{Number: "A001"},
		App:         &uidentitysvc.RuntimeApplication{ClientID: "portal"},
	})
	if payload["callbackUrl"] != "https://app.example.com/callback" || payload["callback_url"] != "https://app.example.com/callback" {
		t.Fatalf("legacy callback fields missing: %#v", payload)
	}
	if payload["tgt"] != "1:TGT:abc" || payload["st"] != "ST-abc" {
		t.Fatalf("legacy ticket fields missing: %#v", payload)
	}
}

func TestCleanupStringsSplitsCSVAndDropsEmptyValues(t *testing.T) {
	got := stringsFromAny("1, 2,,3 ")
	if len(got) != 3 || got[0] != "1" || got[1] != "2" || got[2] != "3" {
		t.Fatalf("stringsFromAny() = %#v", got)
	}
}

func TestLegacySysJobSnapshotsCoverOldExecutableJobs(t *testing.T) {
	records := legacySysJobSnapshots()
	if len(records) != 9 {
		t.Fatalf("legacySysJobSnapshots() returned %d jobs, want 9", len(records))
	}
	names := map[string]struct{}{}
	for _, record := range records {
		names[gconvString(record["jobName"])] = struct{}{}
	}
	for _, name := range []string{"SyncMysql2Ldap", "SyncStudent", "SyncStudentYJS", "SyncStudentWJ", "SyncDept", "SyncJzg", "ChangeContainer", "NewContainerAccount", "WannaT"} {
		if _, ok := names[name]; !ok {
			t.Fatalf("legacy sysjob snapshot missing %s", name)
		}
	}
}

func gconvString(value any) string {
	return value.(string)
}
