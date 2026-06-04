// This file declares the CMS slide delete API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SlideDeleteReq defines the request for deleting a CMS slide.
type SlideDeleteReq struct {
	g.Meta `path:"/cms/slides/{id}" method:"delete" tags:"CMS Slides" summary:"Delete CMS slide" dc:"Delete a CMS slide by ID." permission:"cms:slide:remove"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Slide ID" eg:"1"`
}

// SlideDeleteRes defines the response for deleting a CMS slide.
type SlideDeleteRes struct{}
