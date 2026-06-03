// player_phone.go defines the request and response DTOs for sicau-niu player
// phone-number authorization binding (one-phone-one-account).

package v1

import "github.com/gogf/gf/v2/frame/g"

// BindPhoneReq is the request for binding a WeChat-authorized phone number to
// the current player.
type BindPhoneReq struct {
	g.Meta            `path:"/plugins/sicau-niu/player/phone" method:"post" tags:"Sicau Niu Player" summary:"Bind player phone number" dc:"Bind a WeChat-authorized phone number to the current player under the one-phone-one-account constraint, recording a device fingerprint. Requires a valid player token."`
	Code              string `json:"code" dc:"New-style getPhoneNumber authorization code; provide either code or encryptedData+iv. In mock mode the code or phone field carries the plain phone number." eg:"e2c1f3a..."`
	EncryptedData     string `json:"encryptedData" dc:"Legacy encrypted phone payload from getPhoneNumber; optional when code is provided" eg:""`
	Iv                string `json:"iv" dc:"Legacy decryption initialization vector paired with encryptedData; optional when code is provided" eg:""`
	Phone             string `json:"phone" dc:"Explicit phone number consumed only by the mock gateway in development/test; ignored by the real gateway" eg:"13800138000"`
	DeviceFingerprint string `json:"deviceFingerprint" dc:"Lightweight device fingerprint captured for risk control; optional" eg:"fp-7f3a9c"`
}

// BindPhoneRes is the response for binding a phone number. It is intentionally
// empty; success is conveyed by the absence of an error.
type BindPhoneRes struct{}
