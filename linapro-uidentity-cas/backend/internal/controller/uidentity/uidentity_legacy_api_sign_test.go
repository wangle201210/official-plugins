// This file verifies the legacy third-party API signature guard middleware
// keeps the old APIMiddleWare boundary and envelope.

package uidentity

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

type legacyAPISignFakeService struct {
	uidentitysvc.Service

	signInput uidentitysvc.LegacyAPISignatureInput
	signErr   error

	changePhoneInput uidentitysvc.ChangePhoneInput
}

func (s *legacyAPISignFakeService) VerifyLegacyAPISignature(_ context.Context, in uidentitysvc.LegacyAPISignatureInput) error {
	s.signInput = in
	return s.signErr
}

func (s *legacyAPISignFakeService) ChangeRuntimePhone(_ context.Context, in uidentitysvc.ChangePhoneInput) error {
	s.changePhoneInput = in
	return nil
}

func startLegacyAPISignTestServer(t *testing.T, service *legacyAPISignFakeService) string {
	t.Helper()
	return startLegacyHTTPTestServer(t, "api-sign", service, func(group *ghttp.RouterGroup, controller *LegacyController) {
		group.Group("/", func(signed *ghttp.RouterGroup) {
			signed.Middleware(controller.LegacyAPISignGuard)
			signed.POST("/user/changePhone", controller.UserChangePhone)
		})
	})
}

// TestLegacyAPISignGuardPassesHeadersAndBodyToVerifier verifies signed
// requests reach the handler with the verified headers forwarded.
func TestLegacyAPISignGuardPassesHeadersAndBodyToVerifier(t *testing.T) {
	service := &legacyAPISignFakeService{}
	baseURL := startLegacyAPISignTestServer(t, service)

	body := `{"phone":"13800000000","code":"123456"}`
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/user/changePhone", strings.NewReader(body))
	if err != nil {
		t.Fatalf("create signed changePhone request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("appid", "portal")
	req.Header.Set("ts", "1700000000")
	req.Header.Set("sign", "expected-sign")
	req.Header.Set("number", "header-user")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("call signed changePhone: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	if service.signInput.AppID != "portal" || service.signInput.Timestamp != "1700000000" ||
		service.signInput.Sign != "expected-sign" {
		t.Fatalf("signature guard did not forward headers: %#v", service.signInput)
	}
	if string(service.signInput.Body) != body {
		t.Fatalf("signature guard did not forward raw body: %q", string(service.signInput.Body))
	}
	if service.changePhoneInput.Number != "header-user" {
		t.Fatalf("signed request did not reach handler: %#v", service.changePhoneInput)
	}
}

// TestLegacyAPISignGuardRejectsWithOldUnauthorizedEnvelope verifies failures
// keep the old HTTP 200 + {code: 401, msg} contract and skip the handler.
func TestLegacyAPISignGuardRejectsWithOldUnauthorizedEnvelope(t *testing.T) {
	service := &legacyAPISignFakeService{
		signErr: bizerr.NewCode(uidentitysvc.CodeAPISignatureInvalid),
	}
	baseURL := startLegacyAPISignTestServer(t, service)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/user/changePhone", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("create unsigned changePhone request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("number", "header-user")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("call unsigned changePhone: %v", err)
	}
	defer closeHTTPResponse(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("legacy unauthorized must keep HTTP 200, got %d", resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode legacy unauthorized response: %v", err)
	}
	if payload["code"] != float64(http.StatusUnauthorized) {
		t.Fatalf("legacy unauthorized envelope mismatch: %#v", payload)
	}
	if service.changePhoneInput.Number != "" {
		t.Fatalf("rejected request must not reach handler: %#v", service.changePhoneInput)
	}
}
