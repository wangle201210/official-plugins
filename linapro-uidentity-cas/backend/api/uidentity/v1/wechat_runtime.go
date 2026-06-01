// This file declares legacy-compatible Wechat QR login endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// WechatLoginQRReq defines QR login state creation.
type WechatLoginQRReq struct {
	g.Meta   `path:"/uidentity/cas/wechat-login-qrs" method:"post" tags:"UIdentity Wechat Runtime" summary:"Create Wechat QR login state" dc:"Create a short-lived Wechat QR login state and return the configured external authorization URL when available."`
	ClientId string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	Callback string `json:"callback" dc:"Optional legacy cascallback value echoed to the configured redirect URL" eg:"choose"`
}

// WechatLoginCallbackReq defines QR login callback completion.
type WechatLoginCallbackReq struct {
	g.Meta   `path:"/uidentity/cas/wechat-login-callbacks" method:"post" tags:"UIdentity Wechat Runtime" summary:"Complete Wechat QR login callback" dc:"Record a Wechat QR callback result. If unionId is supplied, the plugin completes login or creates a bind challenge; otherwise it records a structured unsupported-flow result."`
	State    string `json:"state" v:"required" dc:"Wechat QR login state" eg:"loginByQr_abcdef"`
	ClientId string `json:"clientId" dc:"Optional application client ID used to guard callback state" eg:"portal"`
	Code     string `json:"code" dc:"External Wechat callback code retained for diagnostics when no unionId is supplied" eg:"wx_code"`
	UnionId  string `json:"unionId" dc:"Wechat union ID resolved by an external callback adapter" eg:"unionid_001"`
	Callback string `json:"callback" dc:"Optional legacy cascallback value echoed to the configured redirect URL" eg:"choose"`
}

// WechatLoginQRResultReq defines QR login result polling.
type WechatLoginQRResultReq struct {
	g.Meta `path:"/uidentity/cas/wechat-login-qrs/{state}/result" method:"get" tags:"UIdentity Wechat Runtime" summary:"Get Wechat QR login result" dc:"Read the current Wechat QR login state. Terminal states are consumed after this read."`
	State  string `json:"state" v:"required" dc:"Wechat QR login state" eg:"loginByQr_abcdef"`
}

// WechatLoginQRRes returns one QR login state.
type WechatLoginQRRes struct {
	State     string `json:"state" dc:"Wechat QR login state" eg:"loginByQr_abcdef"`
	Url       string `json:"url" dc:"External authorization URL when runtime.wechatLoginAuthorizeUrl is configured" eg:"https://wechat.example.com/oauth?state=loginByQr_abcdef"`
	ExpiredAt *int64 `json:"expiredAt" dc:"State expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
}

// WechatLoginQRResultRes returns QR login result state.
type WechatLoginQRResultRes struct {
	State       string              `json:"state" dc:"Wechat QR login state" eg:"loginByQr_abcdef"`
	Status      string              `json:"status" dc:"QR login status: pending, success, bind_required, unsupported, failed" eg:"success"`
	RedirectUrl string              `json:"redirectUrl" dc:"Configured redirect URL decorated with status and state, when runtime.wechatLoginRedirectUrl is configured" eg:"https://example.com/callback?status=success"`
	ChallengeId string              `json:"challengeId" dc:"Union ID bind challenge ID when status is bind_required" eg:"uid_abcdef"`
	CallbackUrl string              `json:"callbackUrl" dc:"Union ID bind callback URL when a bind challenge was created" eg:"https://example.com/bind?challengeId=uid_abcdef"`
	ErrorCode   string              `json:"errorCode" dc:"Structured business error code when status is unsupported or failed" eg:"UIDENTITY_EXTERNAL_FLOW_UNSUPPORTED"`
	Message     string              `json:"message" dc:"Diagnostic fallback message for unsupported or failed states" eg:"External identity flow is not configured"`
	Login       *CasRuntimeLoginRes `json:"login" dc:"CAS login result when status is success" eg:"{}"`
}

// WechatLoginCallbackRes returns the callback result.
type WechatLoginCallbackRes = WechatLoginQRResultRes
