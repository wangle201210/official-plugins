// This file declares the CMS message list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MessageListReq defines the request for listing visitor messages.
type MessageListReq struct {
	g.Meta   `path:"/cms/messages" method:"get" tags:"CMS Messages" summary:"Get visitor message list" dc:"Query CMS visitor messages by page." permission:"cms:message:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Status   *int   `json:"status" dc:"Filter by status: 0=pending, 1=approved, 2=rejected" eg:"0"`
	Keyword  string `json:"keyword" dc:"Filter by visitor name, email, mobile, or content" eg:"alice"`
}

// MessageListRes defines the response for listing visitor messages.
type MessageListRes struct {
	List  []*MessageItem `json:"list" dc:"Visitor message list"`
	Total int            `json:"total" dc:"Total number of messages" eg:"20"`
}
