// This file verifies the ULink SMS gateway client contract against a local
// HTTP stub without external dependencies.

package uidentity

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLegacySMSGatewayDisabledWithoutConfig keeps the plugin-local record-only
// mode when no gateway address is configured.
func TestLegacySMSGatewayDisabledWithoutConfig(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	cfg, err := service.legacySMSGatewayConfig(context.Background())
	if err != nil {
		t.Fatalf("read gateway config: %v", err)
	}
	if cfg.enabled() {
		t.Fatal("expected gateway to be disabled without config")
	}
}

// TestSendLegacySMSGatewayKeepsOldULinkContract verifies the old query field
// names and OK-prefix success rule.
func TestSendLegacySMSGatewayKeepsOldULinkContract(t *testing.T) {
	t.Parallel()

	responseBody := "OK:1234567890"
	var gotQuery map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = map[string]string{}
		for key := range r.URL.Query() {
			gotQuery[key] = r.URL.Query().Get(key)
		}
		_, _ = fmt.Fprint(w, responseBody)
	}))
	t.Cleanup(server.Close)

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, fmt.Sprintf(`
legacy:
  sms:
    gatewayAddr: %s
    loginName: YAND002
    password: gateway-secret
    signName: 测试签名
`, server.URL))}
	cfg, err := service.legacySMSGatewayConfig(context.Background())
	if err != nil {
		t.Fatalf("read gateway config: %v", err)
	}
	if !cfg.enabled() {
		t.Fatal("expected configured gateway to be enabled")
	}
	if cfg.content("123456") != "你的验证码为123456" {
		t.Fatalf("gateway content template mismatch: %q", cfg.content("123456"))
	}

	body, err := sendLegacySMSGateway(context.Background(), cfg, "13800000000", cfg.content("123456"))
	if err != nil {
		t.Fatalf("gateway send failed: %v", err)
	}
	if body != responseBody {
		t.Fatalf("gateway response body = %q", body)
	}
	if gotQuery["LoginName"] != "YAND002" || gotQuery["Pwd"] != "gateway-secret" ||
		gotQuery["FeeType"] != "2" || gotQuery["Mobile"] != "13800000000" ||
		gotQuery["Content"] != "你的验证码为123456" || gotQuery["SignName"] != "测试签名" {
		t.Fatalf("gateway query fields mismatch: %#v", gotQuery)
	}
	if _, ok := gotQuery["TimingDate"]; !ok {
		t.Fatalf("gateway query must keep old empty fields: %#v", gotQuery)
	}

	responseBody = "ERR:balance"
	if _, err := sendLegacySMSGateway(context.Background(), cfg, "13800000000", "x"); err == nil {
		t.Fatal("expected non-OK gateway body to fail")
	}
}
