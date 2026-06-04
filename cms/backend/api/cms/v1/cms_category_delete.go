// This file declares the CMS category delete API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CategoryDeleteReq defines the request for deleting a CMS category.
type CategoryDeleteReq struct {
	g.Meta `path:"/cms/categories/{id}" method:"delete" tags:"CMS Categories" summary:"Delete CMS category" dc:"Delete a CMS category if it has no children or articles." permission:"cms:category:remove"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Category ID" eg:"1"`
}

// CategoryDeleteRes defines the response for deleting a CMS category.
type CategoryDeleteRes struct{}
