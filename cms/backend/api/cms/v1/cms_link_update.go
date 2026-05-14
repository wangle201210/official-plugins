// This file declares the CMS friendly link update API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LinkUpdateReq defines the request for updating a CMS friendly link.
type LinkUpdateReq struct {
	g.Meta    `path:"/cms/links/{id}" method:"put" tags:"CMS Links" summary:"Update CMS friendly link" dc:"Update a CMS friendly link by ID." permission:"cms:link:edit"`
	Id        int64  `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Link ID" eg:"1"`
	GroupCode string `json:"groupCode" dc:"Display group code" eg:"1"`
	Name      string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Link name" eg:"LinaPro"`
	Url       string `json:"url" v:"required#gf.gvalid.rule.required" dc:"Link URL" eg:"https://linapro.ai"`
	Logo      string `json:"logo" dc:"Logo URL" eg:"/uploads/link.png"`
	Sort      int    `json:"sort" dc:"Display order" eg:"1"`
	Status    int    `json:"status" v:"required|in:0,1#gf.gvalid.rule.required|gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
}

// LinkUpdateRes defines the response for updating a CMS friendly link.
type LinkUpdateRes struct{}
