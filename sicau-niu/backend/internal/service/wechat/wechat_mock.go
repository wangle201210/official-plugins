// wechat_mock.go implements the mockable WeChat gateway used in C1 so the full
// player login and phone-binding flow runs without real WeChat credentials.

package wechat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"lina-core/pkg/bizerr"
)

// mockOpenidPrefix prefixes openids derived from a login code so derived values
// are recognizable in logs and test data.
const mockOpenidPrefix = "mock-openid-"

// mockGateway returns deterministic openids and reads the phone directly from
// the input, enabling end-to-end testing of the login/phone flow.
type mockGateway struct {
	fixedOpenid string // fixedOpenid, when set, is returned for every code.
}

// newMockGateway creates a mock gateway. When fixedOpenid is non-empty it is
// returned for every login code; otherwise a stable openid is derived from the
// code so repeated logins with the same code map to the same player.
func newMockGateway(fixedOpenid string) Gateway {
	return &mockGateway{fixedOpenid: strings.TrimSpace(fixedOpenid)}
}

// Code2Session returns the configured fixed openid or a stable openid derived
// from the login code.
func (g *mockGateway) Code2Session(_ context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", bizerr.NewCode(CodeWeChatCodeInvalid)
	}
	if g.fixedOpenid != "" {
		return g.fixedOpenid, nil
	}
	sum := sha256.Sum256([]byte(code))
	return mockOpenidPrefix + hex.EncodeToString(sum[:8]), nil
}

// DecodePhone reads the phone from the input override or the new-style code so
// the mock flow can bind an explicit phone number without WeChat decryption.
func (g *mockGateway) DecodePhone(_ context.Context, in DecodePhoneInput) (string, error) {
	phone := strings.TrimSpace(in.PhoneOverride)
	if phone == "" {
		phone = strings.TrimSpace(in.Code)
	}
	if phone == "" {
		return "", bizerr.NewCode(CodeWeChatPhoneDecodeFailed)
	}
	return phone, nil
}
