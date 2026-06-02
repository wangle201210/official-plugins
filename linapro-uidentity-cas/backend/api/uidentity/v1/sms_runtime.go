// This file declares legacy-compatible SMS verification endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SmsSendReq sends or records one SMS verification code.
type SmsSendReq struct {
	g.Meta `path:"/uidentity/sms-codes" method:"post" tags:"UIdentity SMS Runtime" summary:"Send SMS verification code" dc:"Create a plugin-local SMS verification code for CAS login, activation, phone binding, or password reset. When no external SMS gateway is configured, the code is recorded locally for deterministic verification."`
	Type   string `json:"type" v:"required" dc:"SMS scenario type: login, active, bind, pwd_change" eg:"login"`
	Phone  string `json:"phone" v:"required" dc:"Mobile phone number receiving the verification code" eg:"13800000000"`
	Code   string `json:"code" dc:"Captcha code required by the legacy SMS send contract; legacy common-pass code is accepted when enabled" eg:"1234"`
	UUID   string `json:"uuid" dc:"Captcha UUID required when the supplied code is not the enabled legacy common-pass code" eg:"captcha_uuid"`
}

// SmsSendRes carries the recorded SMS code metadata.
type SmsSendRes struct {
	Id int64 `json:"id" dc:"Plugin SMS record ID" eg:"1"`
}
