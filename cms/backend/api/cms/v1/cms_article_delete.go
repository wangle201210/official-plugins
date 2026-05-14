// This file declares the CMS article delete API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ArticleDeleteReq defines the request for deleting one CMS article.
type ArticleDeleteReq struct {
	g.Meta `path:"/cms/articles/{id}" method:"delete" tags:"CMS Articles" summary:"Delete CMS article" dc:"Delete one CMS article by ID." permission:"cms:article:remove"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Article ID" eg:"1"`
}

// ArticleDeleteRes defines the response for deleting one CMS article.
type ArticleDeleteRes struct{}
