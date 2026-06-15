// This file declares the CMS product create API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ProductCreateReq defines the request for creating a CMS product.
type ProductCreateReq struct {
	g.Meta      `path:"/cms/products" method:"post" tags:"CMS Products" summary:"Create CMS product" dc:"Create a CMS product as draft or published content with optional gallery images." permission:"cms:product:add"`
	CategoryId  int64    `json:"categoryId" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Category ID" eg:"1"`
	Name        string   `json:"name" v:"required#gf.gvalid.rule.required" dc:"Product name" eg:"Thermal interface pad"`
	Slug        string   `json:"slug" v:"required#gf.gvalid.rule.required" dc:"Public URL slug" eg:"thermal-interface-pad"`
	Summary     string   `json:"summary" dc:"Product summary" eg:"High thermal conductivity pad"`
	Cover       string   `json:"cover" dc:"Cover image URL" eg:"/uploads/product.png"`
	Gallery     []string `json:"gallery" dc:"Gallery image URL list; at most 9 images are kept, extra entries are dropped" eg:"[\"/uploads/p1.png\"]"`
	Price       string   `json:"price" dc:"Display price text" eg:"$99 / sqm"`
	Spec        string   `json:"spec" dc:"Specification summary" eg:"8 W/(m·K), 0.2-2.0 mm"`
	Content     string   `json:"content" v:"required#gf.gvalid.rule.required" dc:"Product detail HTML" eg:"<p>Detail</p>"`
	Keywords    string   `json:"keywords" dc:"SEO keywords" eg:"thermal,pad"`
	Description string   `json:"description" dc:"SEO description" eg:"Thermal interface pad"`
	Sort        int      `json:"sort" dc:"Display order" eg:"1"`
	Status      int      `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=draft, 1=published" eg:"1"`
	IsTop       int      `json:"isTop" v:"in:0,1#gf.gvalid.rule.in" dc:"Top flag: 0=no, 1=yes" eg:"0"`
	IsRecommend int      `json:"isRecommend" v:"in:0,1#gf.gvalid.rule.in" dc:"Recommend flag: 0=no, 1=yes" eg:"1"`
	PublishedAt *int64   `json:"publishedAt" v:"min:1" dc:"Publication time as Unix timestamp in milliseconds. Only effective when status is 1=published; a future value schedules the product and hides it from the public site until that time. Omitted: the current time is used on first publish." eg:"1715740800000"`
}

// ProductCreateRes defines the response for creating a CMS product.
type ProductCreateRes struct {
	Id int64 `json:"id" dc:"Product ID" eg:"1"`
}
