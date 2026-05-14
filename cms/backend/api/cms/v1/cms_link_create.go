// This file declares the CMS friendly link create API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LinkCreateReq defines the request for creating a CMS friendly link.
type LinkCreateReq struct {
	g.Meta    `path:"/cms/links" method:"post" tags:"CMS Links" summary:"Create CMS friendly link" dc:"Create a CMS friendly link for public footer rendering." permission:"cms:link:add"`
	GroupCode string `json:"groupCode" dc:"Display group code" eg:"1"`
	Name      string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Link name" eg:"LinaPro"`
	Url       string `json:"url" v:"required#gf.gvalid.rule.required" dc:"Link URL" eg:"https://linapro.ai"`
	Logo      string `json:"logo" dc:"Logo URL" eg:"/uploads/link.png"`
	Sort      int    `json:"sort" dc:"Display order" eg:"1"`
	Status    int    `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
}

// LinkCreateRes defines the response for creating a CMS friendly link.
type LinkCreateRes struct {
	Id int64 `json:"id" dc:"Link ID" eg:"1"`
}
