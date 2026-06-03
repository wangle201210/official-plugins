// wechat_real.go holds the real WeChat gateway placeholder. C1 ships the seam
// and the mock implementation; the real jscode2session and phone-decryption
// calls are intentionally left as a TODO until production AppID/Secret and the
// final getPhoneNumber form are confirmed (see openspec design Open Questions).

package wechat

import (
	"context"

	"lina-core/pkg/bizerr"
)

// realGateway is the production WeChat gateway. It holds the credentials needed
// by the real jscode2session call once that integration is implemented.
type realGateway struct {
	appID  string // appID is the WeChat mini-program AppID.
	secret string // secret is the WeChat mini-program AppSecret.
}

// newRealGateway creates the production WeChat gateway placeholder.
func newRealGateway(appID, secret string) Gateway {
	return &realGateway{appID: appID, secret: secret}
}

// Code2Session is the real WeChat code exchange.
//
// TODO(sicau-niu): call WeChat `sns/jscode2session` with appID/secret and the
// login code, parse the `openid` from the JSON response, and map WeChat error
// codes (40029 invalid code, 45011 rate limit, etc.) onto CodeWeChatCodeInvalid.
func (g *realGateway) Code2Session(_ context.Context, _ string) (string, error) {
	return "", bizerr.NewCode(CodeWeChatNotImplemented)
}

// DecodePhone is the real WeChat phone decoding.
//
// TODO(sicau-niu): support the new getPhoneNumber `code` flow (exchange via
// `wxa/business/getuserphonenumber`) and/or the legacy encrypted-data + IV flow
// (AES-CBC decrypt with the session key), then return the decoded phoneNumber.
func (g *realGateway) DecodePhone(_ context.Context, _ DecodePhoneInput) (string, error) {
	return "", bizerr.NewCode(CodeWeChatNotImplemented)
}
