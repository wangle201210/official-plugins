// Package wechat defines the WeChat mini-program authentication gateway seam
// used by the sicau-niu player identity service, plus a mockable implementation.
// The seam isolates the two external WeChat operations the activity needs:
// exchanging a login `code` for an `openid`, and decoding a phone-number
// authorization payload into a phone number. A mock implementation runs the full
// login/phone flow without real WeChat credentials (development and E2E); the real
// implementation performs the production WeChat open-API calls and only needs the
// configured AppID/Secret, so setting `wechat.mock=false` with real credentials
// switches to the live flow with no code change.
package wechat

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// Gateway defines the WeChat authentication operations consumed by player login
// and phone binding. Implementations must be safe for concurrent use.
type Gateway interface {
	// Code2Session exchanges a WeChat mini-program login code for the player's
	// stable openid. It returns CodeWeChatCodeInvalid when the code is blank or
	// cannot be resolved. The returned openid is never empty on success.
	Code2Session(ctx context.Context, code string) (openid string, err error)
	// DecodePhone decodes a WeChat phone-number authorization input into a plain
	// phone number. It returns CodeWeChatPhoneDecodeFailed when the input cannot
	// be decoded. The returned phone is never empty on success.
	DecodePhone(ctx context.Context, in DecodePhoneInput) (phone string, err error)
}

// DecodePhoneInput carries the WeChat phone-authorization payload. The new
// getPhoneNumber flow supplies a Code; the legacy flow supplies the encrypted
// data and IV. The mock implementation reads PhoneOverride first so tests and
// development can bind an explicit phone without real WeChat decryption.
type DecodePhoneInput struct {
	// Code is the new-style getPhoneNumber authorization code.
	Code string
	// EncryptedData is the legacy encrypted phone payload.
	EncryptedData string
	// IV is the legacy decryption initialization vector.
	IV string
	// PhoneOverride is the explicit phone consumed by the mock gateway.
	PhoneOverride string
}

var (
	// CodeWeChatCodeInvalid reports that a WeChat login code could not be resolved.
	CodeWeChatCodeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_WECHAT_CODE_INVALID",
		"WeChat login code is invalid",
		gcode.CodeNotAuthorized,
	)
	// CodeWeChatPhoneDecodeFailed reports that a WeChat phone payload could not be decoded.
	CodeWeChatPhoneDecodeFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_WECHAT_PHONE_DECODE_FAILED",
		"Failed to decode WeChat phone authorization",
		gcode.CodeInvalidParameter,
	)
	// CodeWeChatRequestFailed reports that a WeChat open-API HTTP request failed.
	CodeWeChatRequestFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_WECHAT_REQUEST_FAILED",
		"Failed to reach the WeChat open API",
		gcode.CodeInternalError,
	)
)

// Config carries the pure-value WeChat gateway configuration consumed by the
// gateway factory.
type Config struct {
	// AppID is the WeChat mini-program AppID for the real gateway.
	AppID string
	// Secret is the WeChat mini-program AppSecret for the real gateway.
	Secret string
	// Mock enables the mock gateway instead of the real WeChat calls.
	Mock bool
	// MockOpenid is the fixed openid returned by the mock gateway when set.
	MockOpenid string
}

// New returns a Gateway selected by cfg: the mock gateway when cfg.Mock is true,
// otherwise the production WeChat gateway that performs the real open-API calls
// using the configured AppID/Secret.
func New(cfg Config) Gateway {
	if cfg.Mock {
		return newMockGateway(cfg.MockOpenid)
	}
	return newRealGateway(cfg.AppID, cfg.Secret)
}
