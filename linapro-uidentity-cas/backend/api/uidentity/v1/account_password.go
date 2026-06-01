// This file declares account password-management endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// AccountPasswordReq defines administrator password reset for one account.
type AccountPasswordReq struct {
	g.Meta      `path:"/uidentity/accounts/{id}/password" method:"put" tags:"UIdentity CAS" summary:"Reset account password" dc:"Reset one account password after validating active password policy rules. The password hash and strength level are updated inside plugin-owned account data." permission:"uidentity:cas:write"`
	Id          int64  `json:"id" v:"required|min:1" dc:"Account ID" eg:"1"`
	NewPassword string `json:"newPassword" v:"required|min-length:1" dc:"New plaintext password submitted by the operator" eg:"S3cure@2026"`
}

// AccountPasswordRes is an empty administrator password reset response.
type AccountPasswordRes struct{}

// AccountPasswordChallengeReq defines the request for checking an account before self-service password reset.
type AccountPasswordChallengeReq struct {
	g.Meta `path:"/uidentity/password-challenges" method:"post" tags:"UIdentity CAS" summary:"Create password reset challenge" dc:"Create a short-lived password reset challenge for the account number. The challenge is stored in plugin-owned OAuth token storage and must be verified by phone before password reset."`
	Number string `json:"number" v:"required" dc:"Account number" eg:"A001"`
}

// AccountPasswordChallengeRes returns challenge metadata.
type AccountPasswordChallengeRes struct {
	ChallengeId string `json:"challengeId" dc:"Password reset challenge ID" eg:"b1803e7b5a9f4d3fb3c60ee09f18f045"`
	Status      int    `json:"status" dc:"Account status: 0=not active, 1=normal, 2=locked" eg:"1"`
}

// AccountPasswordPhoneVerifyReq defines the request for verifying reset challenge by phone.
type AccountPasswordPhoneVerifyReq struct {
	g.Meta      `path:"/uidentity/password-challenges/{challengeId}/phone" method:"post" tags:"UIdentity CAS" summary:"Verify password reset phone" dc:"Verify a password reset challenge by account phone and SMS code. This plugin records the verification state; actual SMS gateway validation can be integrated by replacing the plugin-local SMS verifier."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Password reset challenge ID" eg:"b1803e7b5a9f4d3fb3c60ee09f18f045"`
	Phone       string `json:"phone" v:"required" dc:"Account phone number" eg:"13800000000"`
	Code        string `json:"code" v:"required" dc:"SMS verification code" eg:"123456"`
}

// AccountPasswordPhoneVerifyRes returns verified challenge metadata.
type AccountPasswordPhoneVerifyRes struct {
	ChallengeId string `json:"challengeId" dc:"Verified password reset challenge ID" eg:"b1803e7b5a9f4d3fb3c60ee09f18f045"`
}

// AccountPasswordSelfResetReq defines the final self-service password reset request.
type AccountPasswordSelfResetReq struct {
	g.Meta      `path:"/uidentity/password-challenges/{challengeId}/password" method:"put" tags:"UIdentity CAS" summary:"Reset password by verified challenge" dc:"Reset the account password after a password reset challenge has been verified. The challenge is consumed after successful reset."`
	ChallengeId string `json:"challengeId" v:"required" dc:"Verified password reset challenge ID" eg:"b1803e7b5a9f4d3fb3c60ee09f18f045"`
	NewPassword string `json:"newPassword" v:"required|min-length:1" dc:"New plaintext password" eg:"S3cure@2026"`
}

// AccountPasswordSelfResetRes is an empty self-service reset response.
type AccountPasswordSelfResetRes struct{}
