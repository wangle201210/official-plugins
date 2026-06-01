// This file declares legacy-compatible runtime user self-service endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UserUnionIDLookupReq defines union ID account lookup.
type UserUnionIDLookupReq struct {
	g.Meta  `path:"/uidentity/users/union-id-lookups" method:"post" tags:"UIdentity User Runtime" summary:"Lookup account by union ID" dc:"Resolve a Wechat union ID to an account number or create a short-lived bind challenge for later account binding."`
	UnionId string `json:"unionId" v:"required" dc:"Wechat union ID" eg:"unionid_001"`
}

// UserUnionIDBindReq defines union ID binding.
type UserUnionIDBindReq struct {
	g.Meta      `path:"/uidentity/users/union-id-bindings" method:"post" tags:"UIdentity User Runtime" summary:"Bind union ID to account" dc:"Bind a Wechat union ID challenge to an account using either phone plus SMS code or account number plus password verification."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Union ID bind challenge ID" eg:"uid_abcdef"`
	BindType    int    `json:"bindType" v:"required|in:1,2" dc:"Binding verification type: 1=phone+SMS, 2=number+password" eg:"2"`
	Phone       string `json:"phone" dc:"Phone number for bindType=1" eg:"13800000000"`
	Code        string `json:"code" dc:"SMS verification code for bindType=1" eg:"123456"`
	Number      string `json:"number" dc:"Account number for bindType=2" eg:"A001"`
	Password    string `json:"password" dc:"Plaintext password for bindType=2" eg:"S3cure@2026"`
}

// UserPasswordChangeReq defines runtime password change.
type UserPasswordChangeReq struct {
	g.Meta      `path:"/uidentity/users/{number}/password" method:"put" tags:"UIdentity User Runtime" summary:"Change user password" dc:"Validate active password policy and update one account password from the runtime self-service API." permission:"uidentity:cas:runtime"`
	Number      string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	NewPassword string `json:"newPassword" v:"required" dc:"New plaintext password" eg:"S3cure@2026"`
}

// UserPhoneChangeReq defines runtime phone change.
type UserPhoneChangeReq struct {
	g.Meta `path:"/uidentity/users/{number}/phone" method:"put" tags:"UIdentity User Runtime" summary:"Change user phone" dc:"Verify an SMS bind code and update one account phone number from the runtime self-service API." permission:"uidentity:cas:runtime"`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Phone  string `json:"phone" v:"required" dc:"New mobile phone number" eg:"13800000000"`
	Code   string `json:"code" v:"required" dc:"SMS bind verification code" eg:"123456"`
}

// UserEmailChangeReq defines runtime email change.
type UserEmailChangeReq struct {
	g.Meta `path:"/uidentity/users/{number}/email" method:"put" tags:"UIdentity User Runtime" summary:"Change user email" dc:"Update one account detail email address from the runtime self-service API." permission:"uidentity:cas:runtime"`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Email  string `json:"email" v:"required|email" dc:"New email address" eg:"user@example.com"`
}

// UserQQChangeReq defines runtime QQ change.
type UserQQChangeReq struct {
	g.Meta `path:"/uidentity/users/{number}/qq" method:"put" tags:"UIdentity User Runtime" summary:"Change user QQ" dc:"Update one account detail QQ number from the runtime self-service API." permission:"uidentity:cas:runtime"`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Qq     string `json:"qq" v:"required" dc:"New QQ number" eg:"10001"`
}

// UserWechatUnbindReq defines runtime Wechat unbinding.
type UserWechatUnbindReq struct {
	g.Meta `path:"/uidentity/users/{number}/wechat" method:"delete" tags:"UIdentity User Runtime" summary:"Unbind user Wechat" dc:"Clear the Wechat union ID from one account detail from the runtime self-service API." permission:"uidentity:cas:runtime"`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
}

// UserWechatRebindStateCreateReq defines logged-in Wechat rebind state creation.
type UserWechatRebindStateCreateReq struct {
	g.Meta   `path:"/uidentity/users/{number}/wechat-rebind-states" method:"post" tags:"UIdentity User Runtime" summary:"Create user Wechat rebind state" dc:"Create a short-lived Wechat rebind state for one logged-in runtime account and return the configured external authorization URL when available." permission:"uidentity:cas:runtime"`
	Number   string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Callback string `json:"callback" dc:"Optional legacy cascallback value echoed to the configured redirect URL" eg:"rebind"`
}

// UserWechatRebindCallbackReq defines external Wechat rebind callback completion.
type UserWechatRebindCallbackReq struct {
	g.Meta   `path:"/uidentity/users/wechat-rebind-callbacks" method:"post" tags:"UIdentity User Runtime" summary:"Complete user Wechat rebind callback" dc:"Record a Wechat rebind callback result. If unionId is supplied, the plugin binds it to the account attached to the state; otherwise it records a structured unsupported-flow result."`
	State    string `json:"state" v:"required" dc:"Wechat rebind state" eg:"rebindWechat_abcdef"`
	UnionId  string `json:"unionId" dc:"Wechat union ID resolved by an external callback adapter" eg:"unionid_001"`
	Code     string `json:"code" dc:"External Wechat callback code retained for diagnostics when no unionId is supplied" eg:"wx_code"`
	Callback string `json:"callback" dc:"Optional legacy cascallback value echoed to the configured redirect URL" eg:"rebind"`
}

// UserWechatRebindStateReq defines Wechat rebind state lookup.
type UserWechatRebindStateReq struct {
	g.Meta `path:"/uidentity/users/{number}/wechat-rebind-states/{state}" method:"get" tags:"UIdentity User Runtime" summary:"Get user Wechat rebind state" dc:"Read the current Wechat rebind state for one logged-in runtime account without consuming successful terminal states." permission:"uidentity:cas:runtime"`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	State  string `json:"state" v:"required" dc:"Wechat rebind state" eg:"rebindWechat_abcdef"`
}

// UserInfoReq defines runtime account info lookup.
type UserInfoReq struct {
	g.Meta `path:"/uidentity/users/{number}" method:"get" tags:"UIdentity User Runtime" summary:"Get runtime user info" dc:"Return account, detail, unit, container and group projection for one runtime account." permission:"uidentity:cas:runtime"`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
}

// UserLoginLogsReq defines runtime login-log lookup.
type UserLoginLogsReq struct {
	g.Meta   `path:"/uidentity/users/{number}/cas-login-logs" method:"get" tags:"UIdentity User Runtime" summary:"List user CAS login logs" dc:"Return a bounded paged CAS login log list for one runtime account." permission:"uidentity:cas:runtime"`
	Number   string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize int    `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size with hard maximum 100" eg:"20"`
}

// UserApplicationsReq defines runtime accessible application lookup.
type UserApplicationsReq struct {
	g.Meta   `path:"/uidentity/users/{number}/applications" method:"get" tags:"UIdentity User Runtime" summary:"List user accessible applications" dc:"Return enabled applications not blocked by account or group blacklists for one runtime account." permission:"uidentity:cas:runtime"`
	Number   string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize int    `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size with hard maximum 100" eg:"20"`
}

// UserAppRolesReq defines runtime delegated role lookup.
type UserAppRolesReq struct {
	g.Meta   `path:"/uidentity/users/{number}/account-app-roles" method:"get" tags:"UIdentity User Runtime" summary:"List user delegated application roles" dc:"Return a bounded paged list of account application roles granted by one runtime account." permission:"uidentity:cas:runtime"`
	Number   string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize int    `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size with hard maximum 100" eg:"20"`
}

// UserAppRoleCreateReq defines delegated role creation.
type UserAppRoleCreateReq struct {
	g.Meta          `path:"/uidentity/users/{number}/account-app-roles" method:"post" tags:"UIdentity User Runtime" summary:"Create user delegated application role" dc:"Create a delegated application role from one runtime account to another account." permission:"uidentity:cas:runtime"`
	Number          string `json:"number" v:"required" dc:"Granting account number" eg:"A001"`
	EmpoweredNumber string `json:"empoweredNumber" v:"required" dc:"Delegated account number" eg:"B001"`
	AppId           int64  `json:"appId" v:"required|min:1" dc:"Application ID" eg:"1"`
	ExpireAt        *int64 `json:"expireAt" dc:"Delegation expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
}

// UserAppRoleUpdateReq defines delegated role expiration update.
type UserAppRoleUpdateReq struct {
	g.Meta   `path:"/uidentity/users/{number}/account-app-roles/{id}" method:"put" tags:"UIdentity User Runtime" summary:"Update user delegated application role" dc:"Update delegated application role expiration when the role is granted by the runtime account." permission:"uidentity:cas:runtime"`
	Number   string `json:"number" v:"required" dc:"Granting account number" eg:"A001"`
	Id       int64  `json:"id" v:"required|min:1" dc:"Delegated role ID" eg:"1"`
	ExpireAt *int64 `json:"expireAt" dc:"Delegation expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
}

// UserUnionIDLookupRes returns union ID lookup result.
type UserUnionIDLookupRes struct {
	Number      string `json:"number" dc:"Bound account number when the union ID is already attached" eg:"A001"`
	ChallengeId string `json:"challengeId" dc:"Union ID bind challenge ID when no account is attached" eg:"uid_abcdef"`
	CallbackUrl string `json:"callbackUrl" dc:"Optional front-end callback URL from plugin configuration" eg:"https://example.com/bind?challengeId=uid_abcdef"`
}

// UserUnionIDBindRes returns bind result.
type UserUnionIDBindRes struct {
	Number string `json:"number" dc:"Bound account number" eg:"A001"`
}

// UserMutationRes is an empty mutation response.
type UserMutationRes struct{}

// UserPasswordChangeRes is an empty password-change response.
type UserPasswordChangeRes = UserMutationRes

// UserPhoneChangeRes is an empty phone-change response.
type UserPhoneChangeRes = UserMutationRes

// UserEmailChangeRes is an empty email-change response.
type UserEmailChangeRes = UserMutationRes

// UserQQChangeRes is an empty QQ-change response.
type UserQQChangeRes = UserMutationRes

// UserWechatUnbindRes is an empty Wechat-unbind response.
type UserWechatUnbindRes = UserMutationRes

// UserWechatRebindStateCreateRes returns one rebind state.
type UserWechatRebindStateCreateRes struct {
	State     string `json:"state" dc:"Wechat rebind state" eg:"rebindWechat_abcdef"`
	Status    string `json:"status" dc:"Rebind status: pending, success, unsupported, failed" eg:"pending"`
	Url       string `json:"url" dc:"External authorization URL when runtime.wechatRebindAuthorizeUrl is configured" eg:"https://wechat.example.com/oauth?state=rebindWechat_abcdef"`
	ExpiredAt *int64 `json:"expiredAt" dc:"State expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
}

// UserWechatRebindStateRes returns current rebind state.
type UserWechatRebindStateRes struct {
	State       string `json:"state" dc:"Wechat rebind state" eg:"rebindWechat_abcdef"`
	Status      string `json:"status" dc:"Rebind status: pending, success, unsupported, failed" eg:"success"`
	Success     bool   `json:"success" dc:"Whether Wechat has been rebound successfully" eg:"true"`
	RedirectUrl string `json:"redirectUrl" dc:"Configured redirect URL decorated with status and state, when runtime.wechatRebindRedirectUrl is configured" eg:"https://example.com/callback?status=success"`
	ErrorCode   string `json:"errorCode" dc:"Structured business error code when status is unsupported or failed" eg:"UIDENTITY_EXTERNAL_FLOW_UNSUPPORTED"`
	Message     string `json:"message" dc:"Diagnostic fallback message for unsupported or failed states" eg:"External identity flow is not configured"`
	ExpiredAt   *int64 `json:"expiredAt" dc:"State expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
}

// UserWechatRebindCallbackRes returns callback completion state.
type UserWechatRebindCallbackRes = UserWechatRebindStateRes

// UserAppRoleUpdateRes is an empty delegated-role update response.
type UserAppRoleUpdateRes = UserMutationRes

// UserInfoRes returns account info.
type UserInfoRes struct {
	User *RuntimeAccount `json:"user" dc:"Runtime account projection" eg:"{}"`
}

// UserLoginLogsRes returns user CAS login logs.
type UserLoginLogsRes struct {
	List  []ResourceRecord `json:"list" dc:"Paged CAS login log records" eg:"[]"`
	Total int              `json:"total" dc:"Total number of CAS login logs" eg:"10"`
}

// UserApplicationsRes returns accessible applications.
type UserApplicationsRes struct {
	List  []*RuntimeApplication `json:"list" dc:"Accessible application projections" eg:"[]"`
	Total int                   `json:"total" dc:"Total number of accessible applications" eg:"10"`
}

// UserAppRolesRes returns delegated roles.
type UserAppRolesRes struct {
	List  []ResourceRecord `json:"list" dc:"Paged delegated role records" eg:"[]"`
	Total int              `json:"total" dc:"Total number of delegated roles" eg:"10"`
}

// UserAppRoleCreateRes returns created delegated role ID.
type UserAppRoleCreateRes struct {
	Id int64 `json:"id" dc:"Created delegated role ID" eg:"1"`
}
