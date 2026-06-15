// This file declares the CMS public product APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublicProductListReq defines the request for the public product list.
type PublicProductListReq struct {
	g.Meta     `path:"/cms/public/products" method:"get" tags:"CMS Public" summary:"Get public CMS products" dc:"Query published CMS products under enabled categories by page. Scheduled products whose publication time has not arrived are hidden." permission:""`
	PageNum    int   `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize   int   `json:"pageSize" d:"12" v:"min:1|max:100" dc:"Number of items per page, at most 100" eg:"12"`
	CategoryId int64 `json:"categoryId" dc:"Filter by category ID" eg:"1"`
}

// PublicProductListRes defines the response for the public product list.
type PublicProductListRes struct {
	List  []*ProductItem `json:"list" dc:"Product list" eg:"[]"`
	Total int            `json:"total" dc:"Total number of products" eg:"20"`
}

// PublicProductGetReq defines the request for one public product detail.
type PublicProductGetReq struct {
	g.Meta `path:"/cms/public/products/{slug}" method:"get" tags:"CMS Public" summary:"Get public CMS product detail" dc:"Read one published CMS product by slug and increment its view count. Scheduled products whose publication time has not arrived are treated as not found." permission:""`
	Slug   string `json:"slug" v:"required#gf.gvalid.rule.required" dc:"Public URL slug" eg:"thermal-interface-pad"`
}

// PublicProductGetRes defines the response for one public product detail.
type PublicProductGetRes struct {
	*ProductItem `dc:"Product detail"`
}
