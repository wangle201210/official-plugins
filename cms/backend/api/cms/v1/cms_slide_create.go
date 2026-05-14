// This file declares the CMS slide create API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SlideCreateReq defines the request for creating a CMS slide.
type SlideCreateReq struct {
	g.Meta    `path:"/cms/slides" method:"post" tags:"CMS Slides" summary:"Create CMS slide" dc:"Create a CMS slide for public carousel rendering." permission:"cms:slide:add"`
	GroupCode string `json:"groupCode" dc:"Display group code" eg:"1"`
	Title     string `json:"title" v:"required#gf.gvalid.rule.required" dc:"Slide title" eg:"Welcome"`
	Subtitle  string `json:"subtitle" dc:"Slide subtitle" eg:"First slide"`
	Image     string `json:"image" v:"required#gf.gvalid.rule.required" dc:"Slide image URL" eg:"/uploads/banner.png"`
	Link      string `json:"link" dc:"Click target URL" eg:"/news"`
	Sort      int    `json:"sort" dc:"Display order" eg:"1"`
	Status    int    `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
}

// SlideCreateRes defines the response for creating a CMS slide.
type SlideCreateRes struct {
	Id int64 `json:"id" dc:"Slide ID" eg:"1"`
}
