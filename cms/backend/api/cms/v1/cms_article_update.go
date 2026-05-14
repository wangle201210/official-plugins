// This file declares the CMS article update API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ArticleUpdateReq defines the request for updating a CMS article.
type ArticleUpdateReq struct {
	g.Meta      `path:"/cms/articles/{id}" method:"put" tags:"CMS Articles" summary:"Update CMS article" dc:"Update a CMS article." permission:"cms:article:edit"`
	Id          int64  `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Article ID" eg:"1"`
	CategoryId  int64  `json:"categoryId" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Category ID" eg:"1"`
	Title       string `json:"title" v:"required#gf.gvalid.rule.required" dc:"Article title" eg:"Welcome to LinaPro CMS"`
	Subtitle    string `json:"subtitle" dc:"Article subtitle" eg:"First article"`
	Slug        string `json:"slug" v:"required#gf.gvalid.rule.required" dc:"Public URL slug" eg:"welcome-to-linapro-cms"`
	Summary     string `json:"summary" dc:"Article summary" eg:"CMS introduction"`
	Cover       string `json:"cover" dc:"Cover image URL" eg:"/uploads/cover.png"`
	Author      string `json:"author" dc:"Author name" eg:"LinaPro"`
	Source      string `json:"source" dc:"Content source" eg:"CMS Plugin"`
	Content     string `json:"content" v:"required#gf.gvalid.rule.required" dc:"Article body HTML" eg:"<p>Hello</p>"`
	Tags        string `json:"tags" dc:"Comma-separated tag names" eg:"CMS,LinaPro"`
	Keywords    string `json:"keywords" dc:"SEO keywords" eg:"CMS,LinaPro"`
	Description string `json:"description" dc:"SEO description" eg:"LinaPro CMS article"`
	Sort        int    `json:"sort" dc:"Display order" eg:"1"`
	Status      int    `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=draft, 1=published" eg:"1"`
	IsTop       int    `json:"isTop" v:"in:0,1#gf.gvalid.rule.in" dc:"Top flag: 0=no, 1=yes" eg:"0"`
	IsRecommend int    `json:"isRecommend" v:"in:0,1#gf.gvalid.rule.in" dc:"Recommend flag: 0=no, 1=yes" eg:"1"`
}

// ArticleUpdateRes defines the response for updating a CMS article.
type ArticleUpdateRes struct{}
