// This file declares legacy-compatible account activation endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ActivationStartReq defines basic account activation verification.
type ActivationStartReq struct {
	g.Meta `path:"/uidentity/activations" method:"post" tags:"UIdentity Activation" summary:"Start account activation" dc:"Verify account number, display name and identity card, then create a short-lived activation challenge in plugin token storage."`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Name   string `json:"name" v:"required" dc:"Account display name" eg:"Alice"`
	Idcard string `json:"idcard" v:"required" dc:"Identity card or certificate number" eg:"510000200001010000"`
}

// ActivationFaceReq defines local face verification marker update.
type ActivationFaceReq struct {
	g.Meta      `path:"/uidentity/activations/{challengeId}/face" method:"put" tags:"UIdentity Activation" summary:"Record activation face proof" dc:"Record a local face proof marker or image reference for an activation challenge without requiring an external face service by default."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Activation challenge ID" eg:"act_abcdef"`
	FaceUrl     string `json:"faceUrl" v:"required" dc:"Face image URL, storage key or verification marker" eg:"https://example.com/face.png"`
}

// ActivationPasswordReq defines activation password setup.
type ActivationPasswordReq struct {
	g.Meta      `path:"/uidentity/activations/{challengeId}/password" method:"put" tags:"UIdentity Activation" summary:"Set activation password" dc:"Validate password policy and store the new password hash for the account attached to the activation challenge."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Activation challenge ID" eg:"act_abcdef"`
	Password    string `json:"password" v:"required" dc:"New plaintext password" eg:"S3cure@2026"`
}

// ActivationPhoneReq defines activation phone binding.
type ActivationPhoneReq struct {
	g.Meta      `path:"/uidentity/activations/{challengeId}/phone" method:"put" tags:"UIdentity Activation" summary:"Bind activation phone" dc:"Verify an activation SMS code, bind a phone number to the account attached to the challenge, and mark the account normal."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Activation challenge ID" eg:"act_abcdef"`
	Phone       string `json:"phone" v:"required" dc:"Mobile phone number" eg:"13800000000"`
	Code        string `json:"code" v:"required" dc:"SMS verification code" eg:"123456"`
}

// ActivationWechatReq defines activation Wechat binding.
type ActivationWechatReq struct {
	g.Meta      `path:"/uidentity/activations/{challengeId}/wechat" method:"put" tags:"UIdentity Activation" summary:"Bind activation Wechat" dc:"Bind a Wechat union ID to the account attached to the activation challenge and mark the account normal."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Activation challenge ID" eg:"act_abcdef"`
	UnionId     string `json:"unionId" v:"required" dc:"Wechat union ID" eg:"unionid_001"`
}

// ActivationWechatStateCreateReq defines activation Wechat QR state creation.
type ActivationWechatStateCreateReq struct {
	g.Meta      `path:"/uidentity/activations/{challengeId}/wechat-states" method:"post" tags:"UIdentity Activation" summary:"Create activation Wechat state" dc:"Create a Wechat activation binding state for one activation challenge and return the configured external authorization URL when available."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Activation challenge ID" eg:"act_abcdef"`
	Callback    string `json:"callback" dc:"Optional legacy cascallback value echoed to the configured redirect URL" eg:"active"`
}

// ActivationWechatCallbackReq defines external activation Wechat callback completion.
type ActivationWechatCallbackReq struct {
	g.Meta   `path:"/uidentity/activations/wechat-callbacks" method:"post" tags:"UIdentity Activation" summary:"Complete activation Wechat callback" dc:"Record a Wechat activation callback result. If unionId is supplied, the plugin binds it to the account attached to the state; otherwise it records a structured unsupported-flow result."`
	State    string `json:"state" v:"required" dc:"Activation Wechat state, equal to the activation challenge ID" eg:"act_abcdef"`
	UnionId  string `json:"unionId" dc:"Wechat union ID resolved by an external callback adapter" eg:"unionid_001"`
	Code     string `json:"code" dc:"External Wechat callback code retained for diagnostics when no unionId is supplied" eg:"wx_code"`
	Callback string `json:"callback" dc:"Optional legacy cascallback value echoed to the configured redirect URL" eg:"active"`
}

// ActivationStateReq defines activation state lookup.
type ActivationStateReq struct {
	g.Meta      `path:"/uidentity/activations/{challengeId}/state" method:"get" tags:"UIdentity Activation" summary:"Get activation state" dc:"Read the current activation challenge state and account status."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Activation challenge ID" eg:"act_abcdef"`
}

// ActivationStartRes returns activation challenge metadata.
type ActivationStartRes struct {
	ChallengeId string `json:"challengeId" dc:"Activation challenge ID" eg:"act_abcdef"`
	NeedFace    bool   `json:"needFace" dc:"Whether this account should collect face proof before completion" eg:"true"`
	Status      int    `json:"status" dc:"Current account status: 0=not active, 1=normal, 2=locked" eg:"0"`
}

// ActivationStepRes returns updated activation state.
type ActivationStepRes struct {
	ChallengeId string `json:"challengeId" dc:"Activation challenge ID" eg:"act_abcdef"`
	Success     bool   `json:"success" dc:"Whether the activation step succeeded" eg:"true"`
}

// ActivationFaceRes returns face activation step state.
type ActivationFaceRes = ActivationStepRes

// ActivationPasswordRes returns password activation step state.
type ActivationPasswordRes = ActivationStepRes

// ActivationPhoneRes returns phone activation step state.
type ActivationPhoneRes = ActivationStepRes

// ActivationWechatRes returns Wechat activation step state.
type ActivationWechatRes = ActivationStepRes

// ActivationWechatStateCreateRes returns activation Wechat state metadata.
type ActivationWechatStateCreateRes struct {
	State  string `json:"state" dc:"Activation Wechat state, equal to the activation challenge ID" eg:"act_abcdef"`
	Status string `json:"status" dc:"Activation Wechat status: pending, success, unsupported, failed" eg:"pending"`
	Url    string `json:"url" dc:"External authorization URL when runtime.activationWechatAuthorizeUrl is configured" eg:"https://wechat.example.com/oauth?state=act_abcdef"`
}

// ActivationWechatCallbackRes returns callback completion state.
type ActivationWechatCallbackRes struct {
	State       string `json:"state" dc:"Activation Wechat state" eg:"act_abcdef"`
	Status      string `json:"status" dc:"Activation Wechat status: pending, success, unsupported, failed" eg:"success"`
	Success     bool   `json:"success" dc:"Whether activation Wechat binding has completed successfully" eg:"true"`
	RedirectUrl string `json:"redirectUrl" dc:"Configured redirect URL decorated with status and state, when runtime.activationWechatRedirectUrl is configured" eg:"https://example.com/callback?status=success"`
	ErrorCode   string `json:"errorCode" dc:"Structured business error code when status is unsupported or failed" eg:"UIDENTITY_EXTERNAL_FLOW_UNSUPPORTED"`
	Message     string `json:"message" dc:"Diagnostic fallback message for unsupported or failed states" eg:"External identity flow is not configured"`
}

// ActivationStateRes returns current activation state.
type ActivationStateRes struct {
	ChallengeId  string `json:"challengeId" dc:"Activation challenge ID" eg:"act_abcdef"`
	Success      bool   `json:"success" dc:"Whether the account is currently active" eg:"true"`
	Status       int    `json:"status" dc:"Current account status: 0=not active, 1=normal, 2=locked" eg:"1"`
	Stage        string `json:"stage" dc:"Last completed activation stage" eg:"phone"`
	WechatStatus string `json:"wechatStatus" dc:"Activation Wechat status when a Wechat state was created: pending, success, unsupported, failed" eg:"success"`
	RedirectUrl  string `json:"redirectUrl" dc:"Configured Wechat redirect URL decorated with status and state, when available" eg:"https://example.com/callback?status=success"`
	ErrorCode    string `json:"errorCode" dc:"Structured business error code when Wechat status is unsupported or failed" eg:"UIDENTITY_EXTERNAL_FLOW_UNSUPPORTED"`
	Message      string `json:"message" dc:"Diagnostic fallback message for unsupported or failed Wechat states" eg:"External identity flow is not configured"`
}
