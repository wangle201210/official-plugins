// This file declares the CMS category update API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CategoryUpdateReq defines the request for updating a CMS category.
type CategoryUpdateReq struct {
	g.Meta          `path:"/cms/categories/{id}" method:"put" tags:"CMS Categories" summary:"Update CMS category" dc:"Update a CMS category." permission:"cms:category:edit"`
	Id              int64  `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Category ID" eg:"1"`
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

// CategoryUpdateRes defines the response for updating a CMS category.
type CategoryUpdateRes struct{}
