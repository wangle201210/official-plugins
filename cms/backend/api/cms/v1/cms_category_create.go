// This file declares the CMS category create API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CategoryCreateReq defines the request for creating a CMS category.
type CategoryCreateReq struct {
	g.Meta          `path:"/cms/categories" method:"post" tags:"CMS Categories" summary:"Create CMS category" dc:"Create a CMS category." permission:"cms:category:add"`
	ParentId        int64  `json:"parentId" dc:"Parent category ID" eg:"0"`
	Code            string `json:"code" v:"required#gf.gvalid.rule.required" dc:"Stable category code" eg:"news"`
	Name            string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Category name" eg:"News"`
	Type            int    `json:"type" v:"required|in:1,2,3#gf.gvalid.rule.required|gf.gvalid.rule.in" dc:"Category type: 1=list, 2=single page, 3=external link" eg:"1"`
	Path            string `json:"path" dc:"Public category path" eg:"/news"`
	ListTemplate    string `json:"listTemplate" dc:"Public list template file" eg:"list.html"`
	ContentTemplate string `json:"contentTemplate" dc:"Public content/detail template file" eg:"detail.html"`
	Cover           string `json:"cover" dc:"Category cover image URL" eg:"/uploads/news.png"`
	Outlink         string `json:"outlink" dc:"External link URL" eg:"https://example.com"`
	Title           string `json:"title" dc:"SEO title" eg:"News"`
	Keywords        string `json:"keywords" dc:"SEO keywords" eg:"news,linapro"`
	Description     string `json:"description" dc:"SEO description" eg:"Latest news"`
	Sort            int    `json:"sort" dc:"Display order" eg:"1"`
	Status          int    `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
}

// CategoryCreateRes defines the response for creating a CMS category.
type CategoryCreateRes struct {
	Id int64 `json:"id" dc:"Category ID" eg:"1"`
}
