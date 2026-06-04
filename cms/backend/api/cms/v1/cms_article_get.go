// This file declares the CMS article detail API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ArticleGetReq defines the request for reading CMS article details.
type ArticleGetReq struct {
	g.Meta `path:"/cms/articles/{id}" method:"get" tags:"CMS Articles" summary:"Get CMS article details" dc:"Get CMS article details by ID." permission:"cms:article:query"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Article ID" eg:"1"`
}

// ArticleGetRes defines the response for reading CMS article details.
type ArticleGetRes struct {
	*ArticleItem
}
