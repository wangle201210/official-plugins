// This file declares the CMS product get and delete APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ProductGetReq defines the request for reading one CMS product.
type ProductGetReq struct {
	g.Meta `path:"/cms/products/{id}" method:"get" tags:"CMS Products" summary:"Get CMS product detail" dc:"Read one CMS product by ID including its gallery images." permission:"cms:product:query"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Product ID" eg:"1"`
}

// ProductGetRes defines the response for reading one CMS product.
type ProductGetRes struct {
	*ProductItem `dc:"Product detail"`
}

// ProductDeleteReq defines the request for deleting one CMS product.
type ProductDeleteReq struct {
	g.Meta `path:"/cms/products/{id}" method:"delete" tags:"CMS Products" summary:"Delete CMS product" dc:"Delete one CMS product by ID." permission:"cms:product:remove"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Product ID" eg:"1"`
}

// ProductDeleteRes defines the response for deleting one CMS product.
type ProductDeleteRes struct{}
