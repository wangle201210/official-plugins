// This file declares the CMS friendly link delete API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LinkDeleteReq defines the request for deleting a CMS friendly link.
type LinkDeleteReq struct {
	g.Meta `path:"/cms/links/{id}" method:"delete" tags:"CMS Links" summary:"Delete CMS friendly link" dc:"Delete a CMS friendly link by ID." permission:"cms:link:remove"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Link ID" eg:"1"`
}

// LinkDeleteRes defines the response for deleting a CMS friendly link.
type LinkDeleteRes struct{}
