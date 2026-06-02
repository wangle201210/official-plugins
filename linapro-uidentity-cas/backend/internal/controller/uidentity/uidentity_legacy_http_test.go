// This file tests small old-contract projection helpers used by the legacy
// uidentity/admin HTTP compatibility layer.

package uidentity

import (
	"encoding/xml"
	"net/url"
	"testing"

	"lina-core/pkg/bizerr"
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
	if payload["callbackUrl"] != "https://app.example.com/callback" {
		t.Fatalf("legacy callback fields missing: %#v", payload)
	}
	if _, ok := payload["callback_url"]; ok {
		t.Fatalf("legacy runtime login payload must not add non-legacy callback_url alias: %#v", payload)
	}
	if payload["tgt"] != "1:TGT:abc" || payload["st"] != "ST-abc" {
		t.Fatalf("legacy ticket fields missing: %#v", payload)
	}
	if _, ok := payload["app"]; ok {
		t.Fatalf("legacy runtime login payload must not add non-legacy app field: %#v", payload)
	}
	if payload["uuid"] != "" {
		t.Fatalf("legacy runtime login payload must keep empty uuid field: %#v", payload)
	}
}

func TestLegacyWechatLoginResultPayloadKeepsEmptyCasLoginFields(t *testing.T) {
	payload := legacyWechatLoginResultPayload(&uidentitysvc.WechatLoginQRResultOutput{
		ChallengeID: "bind-uuid",
	})
	if payload["uuid"] != "bind-uuid" {
		t.Fatalf("legacy QR result uuid missing: %#v", payload)
	}
	for _, field := range []string{"callbackUrl", "tgt", "st", "user", "users"} {
		if _, ok := payload[field]; !ok {
			t.Fatalf("legacy QR result field %q missing: %#v", field, payload)
		}
	}
	if payload["user"] != nil || payload["users"] != nil {
		t.Fatalf("legacy QR result empty user fields must keep struct zero values: %#v", payload)
	}
}

func TestLegacyActivationPayloadsKeepOldFieldSets(t *testing.T) {
	start := legacyActivationStartPayload(&uidentitysvc.ActivationOutput{
		ChallengeID: "act-1",
		NeedFace:    true,
		Status:      1,
	})
	if start["uuid"] != "act-1" || start["face"] != true {
		t.Fatalf("legacy activation start payload mismatch: %#v", start)
	}
	if _, ok := start["challengeId"]; ok {
		t.Fatalf("legacy activation start payload must not add challengeId: %#v", start)
	}

	step := legacyActivationStepPayload(&uidentitysvc.ActivationStepOutput{
		ChallengeID: "act-1",
		Success:     true,
	})
	if step["uuid"] != "act-1" || len(step) != 1 {
		t.Fatalf("legacy activation step payload mismatch: %#v", step)
	}

	state := legacyActivationStatePayload(&uidentitysvc.ActivationStateOutput{
		ChallengeID: "act-1",
		Success:     true,
		Status:      2,
	})
	if state["success"] != true || len(state) != 1 {
		t.Fatalf("legacy activation state payload mismatch: %#v", state)
	}
}

func TestLegacyUploadPayloadUsesOldSnakeCaseFields(t *testing.T) {
	payload := legacyUploadPayload(&uidentitysvc.LegacyUploadOutput{Files: []*uidentitysvc.LegacyUploadFile{{
		Size:     123,
		Path:     "static/uploadfile/a.jpg",
		FullPath: "http://example.com/static/uploadfile/a.jpg",
		Name:     "a.jpg",
		Type:     "image/jpeg",
	}}})
	file, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("legacyUploadPayload() = %#v", payload)
	}
	if file["full_path"] != "http://example.com/static/uploadfile/a.jpg" {
		t.Fatalf("legacy upload full_path missing: %#v", file)
	}
	if _, ok := file["fullPath"]; ok {
		t.Fatalf("legacy upload payload must not add fullPath alias: %#v", file)
	}
}

func TestLegacySMSSendErrorMsgKeepsOldChineseText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "captcha",
			err:  bizerr.NewCode(uidentitysvc.CodeSMSCaptchaInvalid),
			want: "验证码错误",
		},
		{
			name: "type",
			err:  bizerr.NewCode(uidentitysvc.CodeSMSTypeInvalid),
			want: "参数错误",
		},
		{
			name: "rate",
			err:  bizerr.NewCode(uidentitysvc.CodeSMSRateLimited),
			want: "发送频繁",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := legacySMSSendErrorMsg(tc.err); got != tc.want {
				t.Fatalf("legacySMSSendErrorMsg() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCleanupStringsSplitsCSVAndDropsEmptyValues(t *testing.T) {
	got := stringsFromAny("1, 2,,3 ")
	if len(got) != 3 || got[0] != "1" || got[1] != "2" || got[2] != "3" {
		t.Fatalf("stringsFromAny() = %#v", got)
	}
}

func TestLegacyAppendEncodedQueryMatchesOldRedirectShell(t *testing.T) {
	values := url.Values{}
	values.Set("appid", "portal")
	values.Set("cascallback", "choose")
	if got := legacyAppendEncodedQuery("https://sso.example.com/login", values); got != "https://sso.example.com/login?appid=portal&cascallback=choose" {
		t.Fatalf("legacyAppendEncodedQuery() = %s", got)
	}
	if got := legacyAppendEncodedQuery("https://sso.example.com/logout?from=admin", values); got != "https://sso.example.com/logout?from=admin&appid=portal&cascallback=choose" {
		t.Fatalf("legacyAppendEncodedQuery() with query = %s", got)
	}
}

func TestLegacyWechatCallbackTextResponseEchoesText(t *testing.T) {
	inbound := `<xml><ToUserName>server</ToUserName><FromUserName>openid</FromUserName><CreateTime>1</CreateTime><MsgType>text</MsgType><Content>hello</Content></xml>`
	got := legacyWechatCallbackTextResponse([]byte(inbound), 1780372800)
	if got == "" {
		t.Fatal("expected text response")
	}
	var outbound legacyWechatTextMessage
	if err := xml.Unmarshal([]byte(got), &outbound); err != nil {
		t.Fatalf("invalid XML response: %v", err)
	}
	if outbound.ToUserName != "openid" || outbound.FromUserName != "server" || outbound.Content != "hello" || outbound.MsgType != "text" {
		t.Fatalf("unexpected response: %#v", outbound)
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
