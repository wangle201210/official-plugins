// This file declares the CMS slide list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SlideListReq defines the request for listing CMS slides.
type SlideListReq struct {
	g.Meta    `path:"/cms/slides" method:"get" tags:"CMS Slides" summary:"Get CMS slide list" dc:"Query CMS slides by page." permission:"cms:slide:query"`
	PageNum   int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize  int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	GroupCode string `json:"groupCode" dc:"Filter by display group code" eg:"1"`
	Status    *int   `json:"status" dc:"Filter by status: 0=disabled, 1=enabled" eg:"1"`
	Keyword   string `json:"keyword" dc:"Filter by slide title or subtitle" eg:"Welcome"`
}

// SlideListRes defines the response for listing CMS slides.
type SlideListRes struct {
	List  []*SlideItem `json:"list" dc:"Slide list"`
	Total int          `json:"total" dc:"Total number of slides" eg:"20"`
}
