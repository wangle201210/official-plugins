// This file declares public CMS article APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublicArticleListReq defines the public request for reading published articles.
type PublicArticleListReq struct {
	g.Meta     `path:"/cms/public/articles" method:"get" tags:"CMS Public" summary:"Get public CMS articles" dc:"Get published CMS articles without management authentication."`
	PageNum    int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize   int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	CategoryId int64  `json:"categoryId" dc:"Filter by category ID" eg:"1"`
	Keyword    string `json:"keyword" dc:"Filter by title, subtitle, summary, body, tags, SEO keywords, or SEO description" eg:"LinaPro"`
}

// PublicArticleListRes defines the public response for reading published articles.
type PublicArticleListRes struct {
	List  []*ArticleItem `json:"list" dc:"Published article list"`
	Total int            `json:"total" dc:"Total number of published articles" eg:"20"`
}

// PublicArticleGetReq defines the public request for reading one published article.
type PublicArticleGetReq struct {
	g.Meta `path:"/cms/public/articles/{slug}" method:"get" tags:"CMS Public" summary:"Get public CMS article details" dc:"Get one published CMS article by slug without management authentication."`
	Slug   string `json:"slug" v:"required#gf.gvalid.rule.required" dc:"Public URL slug" eg:"welcome-to-linapro-cms"`
}

// PublicArticleGetRes defines the public response for reading one published article.
type PublicArticleGetRes struct {
	*ArticleItem
}
