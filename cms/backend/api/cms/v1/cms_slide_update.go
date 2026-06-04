// This file declares the CMS slide update API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SlideUpdateReq defines the request for updating a CMS slide.
type SlideUpdateReq struct {
	g.Meta    `path:"/cms/slides/{id}" method:"put" tags:"CMS Slides" summary:"Update CMS slide" dc:"Update a CMS slide by ID." permission:"cms:slide:edit"`
	Id        int64  `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Slide ID" eg:"1"`
	GroupCode string `json:"groupCode" dc:"Display group code" eg:"1"`
	Title     string `json:"title" v:"required#gf.gvalid.rule.required" dc:"Slide title" eg:"Welcome"`
	Subtitle  string `json:"subtitle" dc:"Slide subtitle" eg:"First slide"`
	Image     string `json:"image" v:"required#gf.gvalid.rule.required" dc:"Slide image URL" eg:"/uploads/banner.png"`
	Link      string `json:"link" dc:"Click target URL" eg:"/news"`
	Sort      int    `json:"sort" dc:"Display order" eg:"1"`
	Status    int    `json:"status" v:"required|in:0,1#gf.gvalid.rule.required|gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
}

// SlideUpdateRes defines the response for updating a CMS slide.
type SlideUpdateRes struct{}
