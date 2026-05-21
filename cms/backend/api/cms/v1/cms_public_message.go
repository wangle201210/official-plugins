// This file declares the public CMS visitor message submission API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublicMessageCreateReq defines the public request for submitting a visitor message.
type PublicMessageCreateReq struct {
	g.Meta  `path:"/cms/public/messages" method:"post" tags:"CMS Public" summary:"Submit public CMS visitor message" dc:"Submit a public visitor message without management authentication."`
	Name    string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Visitor name" eg:"Alice"`
	Mobile  string `json:"mobile" dc:"Visitor mobile" eg:"13800000000"`
	Email   string `json:"email" dc:"Visitor email" eg:"alice@example.com"`
	Content string `json:"content" v:"required#gf.gvalid.rule.required" dc:"Message content" eg:"Please contact me"`
}

// PublicMessageCreateRes defines the public response for submitting a visitor message.
type PublicMessageCreateRes struct {
	Id int64 `json:"id" dc:"Message ID" eg:"1"`
}

// PublicMessageListReq defines the public request for approved visitor messages.
type PublicMessageListReq struct {
	g.Meta   `path:"/cms/public/messages" method:"get" tags:"CMS Public" summary:"List approved CMS visitor messages" dc:"List approved visitor messages and replies when site settings allow public message display."`
	PageNum  int `json:"pageNum" dc:"Page number" eg:"1"`
	PageSize int `json:"pageSize" dc:"Page size" eg:"10"`
}

// PublicMessageListRes defines the public approved visitor-message list response.
type PublicMessageListRes struct {
	List  []*PublicMessageItem `json:"list" dc:"Approved visitor messages"`
	Total int                  `json:"total" dc:"Total approved visitor messages" eg:"1"`
}
