// This file verifies dependency-light Wechat QR login helper behavior.

package uidentity

import "testing"

func TestWechatLoginAuthorizeURL(t *testing.T) {
	t.Parallel()

	got := wechatLoginAuthorizeURL("https://wechat.example.com/oauth?scope=snsapi_userinfo", "portal", "loginByQr_123", "choose")
	want := "https://wechat.example.com/oauth?appid=portal&cascallback=choose&scope=snsapi_userinfo&state=loginByQr_123"
	if got != want {
		t.Fatalf("unexpected authorize URL:\nwant %s\n got %s", want, got)
	}
	if got := wechatLoginAuthorizeURL("", "portal", "loginByQr_123", ""); got != "" {
		t.Fatalf("expected empty base URL to stay empty, got %s", got)
	}
}

func TestWechatLoginResultProjection(t *testing.T) {
	t.Parallel()

	payload := &wechatLoginStateData{
		State:       "loginByQr_123",
		Status:      wechatLoginStatusBindNeeded,
		RedirectURL: "https://example.com/callback",
		ChallengeID: "uid_123",
		CallbackURL: "https://example.com/bind?challengeId=uid_123",
		ErrorCode:   CodeUnsupportedExternalFlow.RuntimeCode(),
		Message:     CodeUnsupportedExternalFlow.Fallback(),
	}
	result := wechatLoginResult(payload)
	if result == nil || result.State != payload.State || result.Status != payload.Status ||
		result.ChallengeID != payload.ChallengeID || result.CallbackURL != payload.CallbackURL ||
		result.ErrorCode != payload.ErrorCode || result.Message != payload.Message {
		t.Fatalf("unexpected result projection: %#v", result)
	}
	if wechatLoginResult(nil) != nil {
		t.Fatal("expected nil payload projection to be nil")
	}
}
