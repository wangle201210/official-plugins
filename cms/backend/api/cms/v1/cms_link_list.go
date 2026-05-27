// This file declares the CMS friendly link list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LinkListReq defines the request for listing CMS friendly links.
type LinkListReq struct {
	g.Meta    `path:"/cms/links" method:"get" tags:"CMS Links" summary:"Get CMS friendly link list" dc:"Query CMS friendly links by page." permission:"cms:link:query"`
	PageNum   int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize  int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	GroupCode string `json:"groupCode" dc:"Filter by display group code" eg:"1"`
	Status    *int   `json:"status" dc:"Filter by status: 0=disabled, 1=enabled" eg:"1"`
	Keyword   string `json:"keyword" dc:"Filter by link name or URL" eg:"LinaPro"`
}

// LinkListRes defines the response for listing CMS friendly links.
type LinkListRes struct {
	List  []*LinkItem `json:"list" dc:"Friendly link list" eg:"[]"`
	Total int         `json:"total" dc:"Total number of links" eg:"20"`
}
