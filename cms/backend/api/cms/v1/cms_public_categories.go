// This file declares the public CMS category tree API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublicCategoryListReq defines the public request for reading enabled category tree.
type PublicCategoryListReq struct {
	g.Meta `path:"/cms/public/categories" method:"get" tags:"CMS Public" summary:"Get public CMS categories" dc:"Get enabled CMS category tree without management authentication."`
}

// PublicCategoryListRes defines the public response for reading enabled category tree.
type PublicCategoryListRes struct {
	List []*CategoryItem `json:"list" dc:"Enabled category tree" eg:"[]"`
}
