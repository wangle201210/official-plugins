// This file declares the CMS product list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ProductListReq defines the request for listing CMS products.
type ProductListReq struct {
	g.Meta     `path:"/cms/products" method:"get" tags:"CMS Products" summary:"Get CMS product list" dc:"Query CMS products by page with optional category, status, and name filters." permission:"cms:product:query"`
	PageNum    int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize   int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page, at most 100" eg:"10"`
	CategoryId int64  `json:"categoryId" dc:"Filter by category ID" eg:"1"`
	Status     *int   `json:"status" dc:"Filter by status: 0=draft, 1=published" eg:"1"`
	Name       string `json:"name" dc:"Filter by product name" eg:"Thermal pad"`
}

// ProductListRes defines the response for listing CMS products.
type ProductListRes struct {
	List  []*ProductItem `json:"list" dc:"Product list" eg:"[]"`
	Total int            `json:"total" dc:"Total number of products" eg:"20"`
}
