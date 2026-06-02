// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Oauth2Token is the golang structure for table oauth2_token.
type Oauth2Token struct {
	Id        int64  `json:"id"        orm:"id"         description:""`
	ExpiredAt int64  `json:"expiredAt" orm:"expired_at" description:""`
	Code      string `json:"code"      orm:"code"       description:""`
	Access    string `json:"access"    orm:"access"     description:""`
	Refresh   string `json:"refresh"   orm:"refresh"    description:""`
	Data      string `json:"data"      orm:"data"       description:""`
}
