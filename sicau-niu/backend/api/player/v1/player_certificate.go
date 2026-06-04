// player_certificate.go defines the request and response DTOs for the player
// electronic-certificate generation: the holder requests their own granted
// certificate honor and receives the personalized PNG (base64) plus the structured
// certificate fields. Time fields are returned as Unix milliseconds.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CertificateReq is the request for the player electronic certificate.
type CertificateReq struct {
	g.Meta  `path:"/plugins/sicau-niu/player/certificates" method:"get" tags:"Sicau Niu Player" summary:"Get player electronic certificate" dc:"Return the personalized electronic certificate for a certificate honor the authenticated player already holds: the base64-encoded PNG plus the holder nickname, honor name/code and grant time. The honor must be a certificate type and the player must already hold it, otherwise a business error is returned. Requires a valid player token."`
	HonorId int64 `json:"honorId" v:"required|min:1" dc:"Certificate honor definition ID the player holds" eg:"7"`
}

// CertificateRes is the response for the player electronic certificate.
type CertificateRes struct {
	ImageBase64 string `json:"imageBase64" dc:"Base64-encoded certificate PNG" eg:"iVBORw0KGgoAAAANS..."`
	Nickname    string `json:"nickname" dc:"Holder nickname; empty when unset" eg:"川农牛仔"`
	HonorName   string `json:"honorName" dc:"Certificate honor display name" eg:"川农120图鉴收藏证书"`
	HonorCode   string `json:"honorCode" dc:"Certificate honor unique code" eg:"cert_full_complete"`
	UnlockedAt  *int64 `json:"unlockedAt" dc:"Grant time as Unix milliseconds; null when unset" eg:"1717488000000"`
}
