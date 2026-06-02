package uidentity

import "testing"

func TestNormalizeLegacyResourceBodyKeepsOldAccountDetailFields(t *testing.T) {
	body := normalizeLegacyResourceBody("account-details", map[string]any{
		"account_id": 7,
		"idCard":     "510000000000000000",
		"nj":         "2024",
		"xymc":       "College",
		"xydm":       "C001",
		"xq":         "North",
		"xz":         "4",
		"yjbysj":     "2028",
		"zymc":       "Computer Science",
		"bjmc":       "Class 1",
	})
	want := map[string]any{
		"accountId":    7,
		"idcard":       "510000000000000000",
		"grade":        "2024",
		"college":      "College",
		"collegeCode":  "C001",
		"campus":       "North",
		"schoolSystem": "4",
		"graduatedAt":  "2028",
		"major":        "Computer Science",
		"className":    "Class 1",
	}
	for key, value := range want {
		if body[key] != value {
			t.Fatalf("legacy field %s normalized to %#v, want %#v; body=%#v", key, body[key], value, body)
		}
	}
}

func TestNormalizeLegacyResourceBodyKeepsOldPassRulerInterval(t *testing.T) {
	body := normalizeLegacyResourceBody("pass-ruler", map[string]any{"interval": 90, "intervalStatus": 1})
	if body["intervalDays"] != 90 {
		t.Fatalf("interval normalized to %#v, want 90; body=%#v", body["intervalDays"], body)
	}
	if body["intervalStatus"] != 1 {
		t.Fatalf("intervalStatus changed unexpectedly: %#v", body)
	}
}

func TestResourceFilterAPINameMatchesOldDTOAliases(t *testing.T) {
	cases := map[string]string{
		"nj":              "grade",
		"xymc":            "college",
		"xydm":            "collegeCode",
		"xq":              "campus",
		"xz":              "schoolSystem",
		"yjbysj":          "graduatedAt",
		"zymc":            "major",
		"bjmc":            "className",
		"interval":        "intervalDays",
		"createBy":        "createdBy",
		"updateBy":        "updatedBy",
		"account_ids":     "",
		"callbackUrl":     "callbackUrl",
		"choiceAccountId": "choiceAccountId",
	}
	for in, want := range cases {
		got := resourceFilterAPIName(in)
		if got != want {
			t.Fatalf("resourceFilterAPIName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestProjectRecordAddsLegacyResourceAliases(t *testing.T) {
	record := Record{
		"grade":        "2024",
		"college":      "College",
		"collegeCode":  "C001",
		"campus":       "North",
		"schoolSystem": "4",
		"graduatedAt":  "2028",
		"major":        "Computer Science",
		"className":    "Class 1",
		"createdBy":    int64(11),
		"updatedBy":    int64(12),
	}
	addLegacyResourceResponseAliases(record, &resourceDefinition{name: "account-details"})
	if record["nj"] != "2024" || record["xymc"] != "College" || record["xydm"] != "C001" || record["xq"] != "North" {
		t.Fatalf("legacy account detail aliases missing: %#v", record)
	}
	if record["xz"] != "4" || record["yjbysj"] != "2028" || record["zymc"] != "Computer Science" || record["bjmc"] != "Class 1" {
		t.Fatalf("legacy school aliases missing: %#v", record)
	}
	if record["createBy"] != int64(11) || record["updateBy"] != int64(12) {
		t.Fatalf("legacy audit aliases missing: %#v", record)
	}

	passRule := Record{"intervalDays": 30}
	addLegacyResourceResponseAliases(passRule, &resourceDefinition{name: "pass-rules"})
	if passRule["interval"] != 30 {
		t.Fatalf("legacy pass-ruler interval alias missing: %#v", passRule)
	}
}
