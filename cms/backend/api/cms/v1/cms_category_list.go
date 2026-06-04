// This file declares the CMS category tree API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CategoryListReq defines the request for listing CMS categories.
type CategoryListReq struct {
	g.Meta `path:"/cms/categories" method:"get" tags:"CMS Categories" summary:"Get CMS category tree" dc:"Get CMS category tree for management." permission:"cms:category:query"`
	Status *int `json:"status" dc:"Filter by status: 0=disabled, 1=enabled" eg:"1"`
}

// CategoryListRes defines the response for listing CMS categories.
type CategoryListRes struct {
	List []*CategoryItem `json:"list" dc:"Category tree" eg:"[]"`
}
