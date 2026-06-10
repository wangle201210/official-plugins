// This file tests small old-contract projection helpers used by the legacy
// uidentity/admin HTTP compatibility layer.

package uidentity

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/mojocn/base64Captcha"

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

func TestLegacyWechatRoutesDelegateToQRCodeRuntime(t *testing.T) {
	service := &legacyWechatHTTPFakeService{}
	baseURL := startLegacyWechatTestServer(t, service)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	loginResp, err := client.Get(baseURL + "/api/v1/wechat/login?appid=portal&cascallback=https%3A%2F%2Fold.example%2Fcallback")
	if err != nil {
		t.Fatalf("call legacy wechat login: %v", err)
	}
	closeHTTPResponse(t, loginResp)
	if service.qrInput.ClientID != "portal" || service.qrInput.Callback != "https://old.example/callback" {
		t.Fatalf("legacy wechat login did not pass old params to QR service: %#v", service.qrInput)
	}

	callbackResp, err := client.Get(baseURL + "/api/v1/wechat/loginCallback?state=qr-state&appid=portal&code=wx-code&unionID=union-1&cascallback=choose")
	if err != nil {
		t.Fatalf("call legacy wechat callback: %v", err)
	}
	closeHTTPResponse(t, callbackResp)
	if callbackResp.StatusCode != http.StatusFound {
		t.Fatalf("legacy wechat callback status = %d, want %d", callbackResp.StatusCode, http.StatusFound)
	}
	if service.callbackInput.State != "qr-state" || service.callbackInput.ClientID != "portal" ||
		service.callbackInput.Code != "wx-code" || service.callbackInput.UnionID != "union-1" ||
		service.callbackInput.Callback != "choose" {
		t.Fatalf("legacy wechat callback did not pass old params to QR callback service: %#v", service.callbackInput)
	}
}

func TestLegacyCaptchaRouteReturnsOldEnvelopeAndStoresAnswer(t *testing.T) {
	baseURL := startLegacyHTTPTestServer(t, "captcha", &legacyHTTPFakeService{}, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.GET("/captcha", controller.Captcha)
	})
	resp, err := http.Get(baseURL + "/api/v1/captcha")
	if err != nil {
		t.Fatalf("call legacy captcha: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode captcha response: %v", err)
	}
	id, ok := payload["id"].(string)
	if !ok || id == "" {
		t.Fatalf("captcha id missing: %#v", payload)
	}
	image, ok := payload["data"].(string)
	if !ok || image == "" {
		t.Fatalf("captcha image missing: %#v", payload)
	}
	if payload["code"] != float64(legacyStatusOK) || payload["msg"] != "success" {
		t.Fatalf("captcha envelope mismatch: %#v", payload)
	}
	answer := base64Captcha.DefaultMemStore.Get(id, false)
	if answer == "" {
		t.Fatalf("captcha answer not stored for id %s", id)
	}
	if !base64Captcha.DefaultMemStore.Verify(id, answer, true) {
		t.Fatalf("stored captcha answer could not be verified")
	}
}

func TestLegacyAdminLoginRouteUsesOldFieldsAndTopLevelToken(t *testing.T) {
	service := &legacyHTTPFakeService{}
	baseURL := startLegacyHTTPTestServer(t, "admin-login", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.POST("/login", controller.AdminLogin)
	})
	captchaID, captchaAnswer, captchaErr := base64Captcha.NewCaptcha(base64Captcha.DefaultDriverDigit, base64Captcha.DefaultMemStore).Generate()
	if captchaErr != nil {
		t.Fatalf("generate captcha for login test: %v", captchaErr)
	}
	resp, err := http.PostForm(baseURL+"/api/v1/login", url.Values{
		"UserName": []string{"A001"},
		"Password": []string{"secret"},
		"appid":    []string{"portal"},
		"UUID":     []string{captchaID},
		"Code":     []string{base64Captcha.DefaultMemStore.Get(captchaID, false)},
	})
	if captchaAnswer == "" {
		t.Fatal("captcha image should not be empty")
	}
	if err != nil {
		t.Fatalf("call legacy admin login: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	if service.passwordLoginInput.ClientID != "portal" || service.passwordLoginInput.Number != "A001" ||
		service.passwordLoginInput.Password != "secret" {
		t.Fatalf("legacy admin login did not pass old fields: %#v", service.passwordLoginInput)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy admin login response: %v", err)
	}
	if payload["token"] != "TGT-1" || payload["expire"] == "" {
		t.Fatalf("legacy admin login top-level token missing: %#v", payload)
	}
}

func TestLegacyCasPasswordLoginRequiresCaptchaLikeOldCasLogin(t *testing.T) {
	service := &legacyHTTPFakeService{loginCaptchaErr: bizerr.NewCode(uidentitysvc.CodeSMSCaptchaInvalid)}
	baseURL := startLegacyHTTPTestServer(t, "cas-login-captcha", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.POST("/cas/login", controller.CasPasswordLogin)
	})
	resp, err := http.PostForm(baseURL+"/api/v1/cas/login", url.Values{
		"number":   []string{"A001"},
		"password": []string{"secret"},
		"appid":    []string{"portal"},
		"uuid":     []string{"captcha-uuid"},
		"code":     []string{"0000"},
	})
	if err != nil {
		t.Fatalf("call legacy cas login: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy cas login response: %v", err)
	}
	if payload["code"] == float64(legacyStatusOK) || payload["msg"] != legacyMsgCaptchaInvalid {
		t.Fatalf("legacy cas login must reject invalid captcha: %#v", payload)
	}
	if service.loginCaptchaCode != "0000" || service.loginCaptchaUUID != "captcha-uuid" {
		t.Fatalf("legacy cas login did not forward captcha fields: %#v", service)
	}
	if service.passwordLoginCalled {
		t.Fatal("legacy cas login must not reach password validation on captcha failure")
	}
}

func TestLegacyCasPasswordLoginPassesCaptchaThenLogsIn(t *testing.T) {
	service := &legacyHTTPFakeService{}
	baseURL := startLegacyHTTPTestServer(t, "cas-login-ok", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.POST("/cas/login", controller.CasPasswordLogin)
	})
	resp, err := http.PostForm(baseURL+"/api/v1/cas/login", url.Values{
		"number":   []string{"A001"},
		"password": []string{"secret"},
		"appid":    []string{"portal"},
		"uuid":     []string{"captcha-uuid"},
		"code":     []string{"1234"},
	})
	if err != nil {
		t.Fatalf("call legacy cas login: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy cas login response: %v", err)
	}
	if payload["code"] != float64(legacyStatusOK) || payload["msg"] != legacyMsgLoginSuccess {
		t.Fatalf("legacy cas login should succeed after captcha: %#v", payload)
	}
	if service.passwordLoginInput.Number != "A001" || service.passwordLoginInput.ClientID != "portal" {
		t.Fatalf("legacy cas login did not pass old fields: %#v", service.passwordLoginInput)
	}
}

func TestLegacyCasPasswordLoginKeepsOldWeakPasswordMessage(t *testing.T) {
	service := &legacyHTTPFakeService{loginErr: bizerr.NewCode(uidentitysvc.CodeLoginPasswordWeak)}
	baseURL := startLegacyHTTPTestServer(t, "cas-login-weak", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.POST("/cas/login", controller.CasPasswordLogin)
	})
	resp, err := http.PostForm(baseURL+"/api/v1/cas/login", url.Values{
		"number":   []string{"A001"},
		"password": []string{"weakpass"},
		"appid":    []string{"portal"},
		"uuid":     []string{"captcha-uuid"},
		"code":     []string{"1234"},
	})
	if err != nil {
		t.Fatalf("call legacy cas login: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy cas login response: %v", err)
	}
	if payload["code"] == float64(legacyStatusOK) || payload["msg"] != legacyMsgLoginPasswordWeak {
		t.Fatalf("legacy cas login must keep the old weak-password message: %#v", payload)
	}
}

func TestLegacyActivationStartRequiresCaptchaLikeOldActivate(t *testing.T) {
	service := &legacyHTTPFakeService{loginCaptchaErr: bizerr.NewCode(uidentitysvc.CodeSMSCaptchaInvalid)}
	baseURL := startLegacyHTTPTestServer(t, "activate-captcha", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.POST("/activate/baseInfo", controller.ActivationStart)
	})
	resp, err := http.PostForm(baseURL+"/api/v1/activate/baseInfo", url.Values{
		"number": []string{"A001"},
		"name":   []string{"Alice"},
		"idcard": []string{"510000200001010000"},
		"uuid":   []string{"captcha-uuid"},
		"code":   []string{"0000"},
	})
	if err != nil {
		t.Fatalf("call legacy activation start: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy activation start response: %v", err)
	}
	if payload["code"] == float64(legacyStatusOK) || payload["msg"] != legacyMsgCaptchaInvalid {
		t.Fatalf("legacy activation start must reject invalid captcha: %#v", payload)
	}
}

func TestLegacyAdminLogoutKeepsOldEnvelopeAndRecordsLogout(t *testing.T) {
	service := &legacyHTTPFakeService{}
	baseURL := startLegacyHTTPTestServer(t, "admin-logout", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.POST("/logout", controller.AdminLogout)
	})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/logout?username=legacy-admin", http.NoBody)
	if err != nil {
		t.Fatalf("create legacy logout request: %v", err)
	}
	req.Header.Set("User-Agent", "legacy-agent")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("call legacy admin logout: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy logout response: %v", err)
	}
	if payload["code"] != float64(legacyStatusOK) || payload["msg"] != legacyMsgLogoutSuccess {
		t.Fatalf("legacy logout envelope mismatch: %#v", payload)
	}
	if _, ok := payload["data"]; ok {
		t.Fatalf("legacy logout must not include data: %#v", payload)
	}
	if _, ok := payload["requestId"]; ok {
		t.Fatalf("legacy logout must not include requestId: %#v", payload)
	}
	if service.logoutInput.Username != "legacy-admin" || service.logoutInput.UserAgent != "legacy-agent" {
		t.Fatalf("legacy logout did not record old audit fields: %#v", service.logoutInput)
	}
}

func TestLegacyRuntimeUserRoutesPreferHeaderNumberOverRequestNumber(t *testing.T) {
	service := &legacyHTTPFakeService{}
	baseURL := startLegacyHTTPTestServer(t, "runtime-number", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.POST("/user/changePhone", controller.UserChangePhone)
		group.POST("/user/getUserInfo", controller.UserInfo)
		group.GET("/user/accountAppList", controller.UserApplications)
	})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/user/changePhone?number=query-user", strings.NewReader(`{"number":"body-user","phone":"13800000000","code":"123456"}`))
	if err != nil {
		t.Fatalf("create legacy changePhone request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("number", "header-user")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("call legacy changePhone: %v", err)
	}
	closeHTTPResponse(t, resp)
	if service.changePhoneInput.Number != "header-user" {
		t.Fatalf("legacy changePhone used %q, want header identity", service.changePhoneInput.Number)
	}

	req, err = http.NewRequest(http.MethodPost, baseURL+"/api/v1/user/getUserInfo?number=query-user", strings.NewReader(`{"number":"body-user"}`))
	if err != nil {
		t.Fatalf("create legacy getUserInfo request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("number", "header-user")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("call legacy getUserInfo: %v", err)
	}
	defer closeHTTPResponse(t, resp)
	if service.runtimeUserInfoNumber != "header-user" {
		t.Fatalf("legacy getUserInfo used %q, want header identity", service.runtimeUserInfoNumber)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read legacy getUserInfo response: %v", err)
	}
	if !strings.Contains(string(body), `"number":"header-user"`) {
		t.Fatalf("legacy getUserInfo response = %s", string(body))
	}

	req, err = http.NewRequest(http.MethodGet, baseURL+"/api/v1/user/accountAppList?pageIndex=9&pageSize=1", http.NoBody)
	if err != nil {
		t.Fatalf("create legacy accountAppList request: %v", err)
	}
	req.Header.Set("number", "header-user")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("call legacy accountAppList: %v", err)
	}
	defer closeHTTPResponse(t, resp)
	if service.applicationListInput.Number != "header-user" || !service.applicationListInput.LegacyNoPage ||
		service.applicationListInput.PageNum != 0 || service.applicationListInput.PageSize != 0 {
		t.Fatalf("legacy accountAppList input mismatch: %#v", service.applicationListInput)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read legacy accountAppList response: %v", err)
	}
	if strings.Contains(string(body), "clientId") || !strings.Contains(string(body), `"name":"Legacy App"`) {
		t.Fatalf("legacy accountAppList response = %s", string(body))
	}
}

func TestLegacyGenMutationRoutesKeepOldEmptyData(t *testing.T) {
	service := &legacyHTTPFakeService{}
	baseURL := startLegacyHTTPTestServer(t, "gen-mutation", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.GET("/gen/toproject/{tableId}", controller.LegacyGenToProject)
		group.GET("/gen/apitofile/{tableId}", controller.LegacyGenAPIToFile)
		group.GET("/gen/todb/{tableId}", controller.LegacyGenToDB)
	})
	cases := []struct {
		path string
		msg  string
	}{
		{path: "/api/v1/gen/toproject/11", msg: "Code generated successfully！"},
		{path: "/api/v1/gen/apitofile/12", msg: "Code generated successfully！"},
		{path: "/api/v1/gen/todb/13", msg: "数据生成成功！"},
	}
	for _, tc := range cases {
		resp, err := http.Get(baseURL + tc.path)
		if err != nil {
			t.Fatalf("call legacy gen route %s: %v", tc.path, err)
		}
		defer closeHTTPResponse(t, resp)

		var payload map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode legacy gen response %s: %v", tc.path, err)
		}
		if payload["code"] != float64(legacyStatusOK) || payload["msg"] != tc.msg || payload["data"] != "" {
			t.Fatalf("legacy gen response mismatch for %s: %#v", tc.path, payload)
		}
	}
	if service.genToProjectTableID != 11 || service.genAPIToFileTableID != 12 || service.genToDBTableID != 13 {
		t.Fatalf("legacy gen routes did not pass table ids: %#v", service)
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

type legacyHTTPFakeService struct {
	uidentitysvc.Service

	passwordLoginInput    uidentitysvc.PasswordLoginInput
	passwordLoginCalled   bool
	loginErr              error
	loginCaptchaCode      string
	loginCaptchaUUID      string
	loginCaptchaErr       error
	logoutInput           uidentitysvc.LegacyAdminLogoutInput
	changePhoneInput      uidentitysvc.ChangePhoneInput
	applicationListInput  uidentitysvc.UserApplicationListInput
	runtimeUserInfoNumber string
	genToProjectTableID   int64
	genAPIToFileTableID   int64
	genToDBTableID        int64
}

func (s *legacyHTTPFakeService) LoginByPassword(_ context.Context, in uidentitysvc.PasswordLoginInput) (*uidentitysvc.RuntimeLoginOutput, error) {
	s.passwordLoginInput = in
	s.passwordLoginCalled = true
	if s.loginErr != nil {
		return nil, s.loginErr
	}
	return &uidentitysvc.RuntimeLoginOutput{
		CallbackURL: "https://app.example/callback?ticket=ST-1",
		TGT:         "TGT-1",
		ST:          "ST-1",
		User:        &uidentitysvc.RuntimeAccount{Number: "A001"},
	}, nil
}

func (s *legacyHTTPFakeService) VerifyLoginCaptcha(_ context.Context, code string, uuid string) error {
	s.loginCaptchaCode = code
	s.loginCaptchaUUID = uuid
	return s.loginCaptchaErr
}

func (s *legacyHTTPFakeService) LegacyRedirectConfig(context.Context) (*uidentitysvc.LegacyRedirectConfigOutput, error) {
	return &uidentitysvc.LegacyRedirectConfigOutput{DefaultAppID: "portal"}, nil
}

func (s *legacyHTTPFakeService) RecordLegacyAdminLogout(_ context.Context, in uidentitysvc.LegacyAdminLogoutInput) error {
	s.logoutInput = in
	return nil
}

func (s *legacyHTTPFakeService) ChangeRuntimePhone(_ context.Context, in uidentitysvc.ChangePhoneInput) error {
	s.changePhoneInput = in
	return nil
}

func (s *legacyHTTPFakeService) GetRuntimeUserInfo(_ context.Context, number string) (*uidentitysvc.RuntimeAccount, error) {
	s.runtimeUserInfoNumber = number
	return &uidentitysvc.RuntimeAccount{Number: number}, nil
}

func (s *legacyHTTPFakeService) ListRuntimeApplications(_ context.Context, in uidentitysvc.UserApplicationListInput) (*uidentitysvc.RuntimeApplicationListOutput, error) {
	s.applicationListInput = in
	return &uidentitysvc.RuntimeApplicationListOutput{
		List:  []*uidentitysvc.RuntimeApplication{{ID: 1, Name: "Legacy App", ClientID: "must-not-leak"}},
		Total: 1,
	}, nil
}

func (s *legacyHTTPFakeService) LegacyGenToProject(_ context.Context, tableID int64) (uidentitysvc.Record, error) {
	s.genToProjectTableID = tableID
	return uidentitysvc.Record{"paths": []string{"must-not-leak"}}, nil
}

func (s *legacyHTTPFakeService) LegacyGenAPIToFile(_ context.Context, tableID int64) (uidentitysvc.Record, error) {
	s.genAPIToFileTableID = tableID
	return uidentitysvc.Record{"path": "must-not-leak"}, nil
}

func (s *legacyHTTPFakeService) LegacyGenToDB(_ context.Context, tableID int64) (uidentitysvc.Record, error) {
	s.genToDBTableID = tableID
	return uidentitysvc.Record{"inserted": 12}, nil
}

type legacyWechatHTTPFakeService struct {
	uidentitysvc.Service

	qrInput       uidentitysvc.WechatLoginQRInput
	callbackInput uidentitysvc.WechatLoginCallbackInput
}

func (s *legacyWechatHTTPFakeService) CreateWechatLoginQR(_ context.Context, in uidentitysvc.WechatLoginQRInput) (*uidentitysvc.WechatLoginQROutput, error) {
	s.qrInput = in
	return &uidentitysvc.WechatLoginQROutput{State: "qr-state", URL: "https://wechat.example/qr"}, nil
}

func (s *legacyWechatHTTPFakeService) CompleteWechatLoginQR(_ context.Context, in uidentitysvc.WechatLoginCallbackInput) (*uidentitysvc.WechatLoginQRResultOutput, error) {
	s.callbackInput = in
	return &uidentitysvc.WechatLoginQRResultOutput{ChallengeID: "bind-uuid"}, nil
}

func (s *legacyWechatHTTPFakeService) LegacyRedirectConfig(context.Context) (*uidentitysvc.LegacyRedirectConfigOutput, error) {
	return &uidentitysvc.LegacyRedirectConfigOutput{WechatLoginRedirect: "https://old.example/wechat-result"}, nil
}

func startLegacyWechatTestServer(t *testing.T, service *legacyWechatHTTPFakeService) string {
	t.Helper()

	server := ghttp.GetServer("legacy-wechat-test-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	controller := NewLegacy(service)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.GET("/wechat/login", controller.WechatLogin)
		group.GET("/wechat/loginCallback", controller.WechatLoginCallback)
	})
	if err := server.Start(); err != nil {
		t.Fatalf("start legacy wechat test server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown legacy wechat test server: %v", err)
		}
	})
	return "http://" + server.GetListenedAddress()
}

func startLegacyHTTPTestServer(t *testing.T, name string, service uidentitysvc.Service, register func(*ghttp.RouterGroup, *LegacyController)) string {
	t.Helper()

	server := ghttp.GetServer("legacy-" + name + "-test-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	controller := NewLegacy(service)
	server.Group("/api/v1", func(group *ghttp.RouterGroup) {
		register(group, controller)
	})
	if err := server.Start(); err != nil {
		t.Fatalf("start legacy %s test server: %v", name, err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown legacy %s test server: %v", name, err)
		}
	})
	return "http://" + server.GetListenedAddress()
}

func closeHTTPResponse(t *testing.T, response *http.Response) {
	t.Helper()

	if err := response.Body.Close(); err != nil {
		t.Fatalf("close HTTP response: %v", err)
	}
}

func gconvString(value any) string {
	return value.(string)
}
