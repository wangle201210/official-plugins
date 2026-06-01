// This file declares legacy-compatible CAS runtime endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeAccountDetail exposes account profile fields used by CAS clients.
type RuntimeAccountDetail struct {
	Birthday string `json:"birthday" dc:"Date-only birthday in YYYY-MM-DD format" eg:"2000-01-01"`
	Email    string `json:"email" dc:"Account email address" eg:"user@example.com"`
	Gender   int    `json:"gender" dc:"Gender code stored by the plugin account detail" eg:"1"`
	Qq       string `json:"qq" dc:"QQ number" eg:"10001"`
	Wechat   string `json:"wechat" dc:"Wechat union ID bound to the account" eg:"unionid_001"`
	Idcard   string `json:"idcard" dc:"Identity card or certificate number" eg:"510000200001010000"`
	Avatar   string `json:"avatar" dc:"Avatar URL or storage key" eg:"https://example.com/avatar.png"`
	Face     string `json:"face" dc:"Face verification marker or image reference" eg:"verified"`
}

// RuntimeAccount exposes an account projection for runtime CAS clients.
type RuntimeAccount struct {
	Id            int64                 `json:"id" dc:"Account ID" eg:"1"`
	Number        string                `json:"number" dc:"Stable account number" eg:"A001"`
	Name          string                `json:"name" dc:"Account display name" eg:"Alice"`
	Phone         string                `json:"phone" dc:"Mobile phone number" eg:"13800000000"`
	Status        int                   `json:"status" dc:"Account status: 0=not active, 1=normal, 2=locked" eg:"1"`
	PassLevel     int                   `json:"passLevel" dc:"Password strength level" eg:"3"`
	ContainerId   int64                 `json:"containerId" dc:"Container ID" eg:"1"`
	ContainerName string                `json:"containerName" dc:"Container display name" eg:"students"`
	UnitId        int64                 `json:"unitId" dc:"Primary unit ID" eg:"1"`
	UnitName      string                `json:"unitName" dc:"Primary unit display name" eg:"College of Computing"`
	ExpireAt      *int64                `json:"expireAt" dc:"Account expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
	Groups        []string              `json:"groups" dc:"Group names attached to the account" eg:"[\"default\"]"`
	Detail        *RuntimeAccountDetail `json:"detail" dc:"Account detail projection" eg:"{}"`
}

// RuntimeApplication exposes an application projection for runtime clients.
type RuntimeApplication struct {
	Id          int64  `json:"id" dc:"Application ID" eg:"1"`
	Name        string `json:"name" dc:"Application name" eg:"Campus Portal"`
	Alias       string `json:"alias" dc:"Application alias" eg:"portal"`
	ClientId    string `json:"clientId" dc:"Stable client ID" eg:"portal"`
	AccessModel string `json:"accessModel" dc:"Application access model, for example cas/oauth/ldap" eg:"cas"`
	CallbackUrl string `json:"callbackUrl" dc:"CAS callback URL" eg:"https://example.com/callback"`
}

// CasPasswordLoginReq defines account-password CAS login.
type CasPasswordLoginReq struct {
	g.Meta   `path:"/uidentity/cas/password-logins" method:"post" tags:"UIdentity CAS Runtime" summary:"Login with account password" dc:"Validate a plugin account password for one application, issue CAS TGT and ST tickets, return the accessible account projections, and record a CAS login log."`
	ClientId string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	Number   string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Password string `json:"password" v:"required" dc:"Plaintext password submitted by the runtime client" eg:"S3cure@2026"`
}

// CasPhoneLoginReq defines phone and SMS-code CAS login.
type CasPhoneLoginReq struct {
	g.Meta   `path:"/uidentity/cas/phone-logins" method:"post" tags:"UIdentity CAS Runtime" summary:"Login with phone and SMS code" dc:"Validate a phone SMS code for one application, issue CAS TGT and ST tickets, return the accessible account projections, and record a CAS login log."`
	ClientId string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	Phone    string `json:"phone" v:"required" dc:"Mobile phone number" eg:"13800000000"`
	Code     string `json:"code" v:"required" dc:"SMS verification code" eg:"123456"`
}

// CasUnionIDLoginReq defines direct union ID CAS login.
type CasUnionIDLoginReq struct {
	g.Meta   `path:"/uidentity/cas/union-id-logins" method:"post" tags:"UIdentity CAS Runtime" summary:"Login with union ID" dc:"Resolve a Wechat union ID to a plugin account, issue CAS TGT and ST tickets, return the accessible account projections, and record a CAS login log."`
	ClientId string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	UnionId  string `json:"unionId" v:"required" dc:"Wechat union ID" eg:"unionid_001"`
}

// CasServiceTicketReq defines service-ticket issue from a TGT.
type CasServiceTicketReq struct {
	g.Meta    `path:"/uidentity/cas/service-tickets" method:"post" tags:"UIdentity CAS Runtime" summary:"Issue CAS service ticket" dc:"Issue a new CAS service ticket from an existing ticket-granting ticket and optionally select an authorized delegated account."`
	ClientId  string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	Tgt       string `json:"tgt" v:"required" dc:"Ticket-granting ticket" eg:"TGT_abcdef"`
	AccountId int64  `json:"accountId" dc:"Optional selected account ID, defaults to the TGT owner" eg:"1"`
}

// CasServiceValidateReq defines service-ticket validation.
type CasServiceValidateReq struct {
	g.Meta `path:"/uidentity/cas/service-validations" method:"post" tags:"UIdentity CAS Runtime" summary:"Validate CAS service ticket" dc:"Consume and validate a CAS service ticket, enforce selected-account authorization and application access, and return account and application projections."`
	Ticket string `json:"ticket" v:"required" dc:"CAS service ticket" eg:"ST_abcdef"`
	UserId int64  `json:"userId" dc:"Optional selected account ID to validate delegated access" eg:"1"`
}

// CasTicketLogoutReq defines CAS TGT/ST deletion.
type CasTicketLogoutReq struct {
	g.Meta `path:"/uidentity/cas/tickets/{ticket}" method:"delete" tags:"UIdentity CAS Runtime" summary:"Delete CAS ticket" dc:"Delete a CAS ticket-granting ticket or service ticket from plugin runtime storage."`
	Ticket string `json:"ticket" v:"required" dc:"CAS TGT or ST value" eg:"TGT_abcdef"`
}

// CasRuntimeLoginRes returns CAS login ticket and account data.
type CasRuntimeLoginRes struct {
	CallbackUrl string            `json:"callbackUrl" dc:"Application callback URL with ticket query parameter appended when configured" eg:"https://example.com/callback?ticket=ST_abcdef"`
	Tgt         string            `json:"tgt" dc:"Ticket-granting ticket" eg:"TGT_abcdef"`
	St          string            `json:"st" dc:"Service ticket" eg:"ST_abcdef"`
	User        *RuntimeAccount   `json:"user" dc:"Primary account projection for the login account" eg:"{}"`
	Users       []*RuntimeAccount `json:"users" dc:"Accounts currently accessible for the application, including delegated accounts" eg:"[]"`
}

// CasPasswordLoginRes returns password-login ticket and account data.
type CasPasswordLoginRes = CasRuntimeLoginRes

// CasPhoneLoginRes returns phone-login ticket and account data.
type CasPhoneLoginRes = CasRuntimeLoginRes

// CasUnionIDLoginRes returns union-ID-login ticket and account data.
type CasUnionIDLoginRes = CasRuntimeLoginRes

// CasServiceValidateRes returns the validated account and application.
type CasServiceValidateRes struct {
	Ticket  string              `json:"ticket" dc:"Consumed service ticket" eg:"ST_abcdef"`
	User    *RuntimeAccount     `json:"user" dc:"Selected account projection" eg:"{}"`
	App     *RuntimeApplication `json:"app" dc:"Application projection" eg:"{}"`
	Success bool                `json:"success" dc:"Whether validation succeeded" eg:"true"`
}

// CasServiceTicketRes returns a newly issued ST from a TGT.
type CasServiceTicketRes struct {
	St          string `json:"st" dc:"New service ticket" eg:"ST_abcdef"`
	CallbackUrl string `json:"callbackUrl" dc:"Application callback URL with ticket query parameter appended when configured" eg:"https://example.com/callback?ticket=ST_abcdef"`
}

// CasTicketLogoutRes is an empty CAS logout response.
type CasTicketLogoutRes struct{}
