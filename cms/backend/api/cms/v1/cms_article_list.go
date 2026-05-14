// This file declares the CMS article list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ArticleListReq defines the request for listing CMS articles.
type ArticleListReq struct {
	g.Meta          `path:"/cms/articles" method:"get" tags:"CMS Articles" summary:"Get CMS article list" dc:"Query CMS articles by page with optional filters." permission:"cms:article:query"`
	PageNum         int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize        int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	CategoryId      int64  `json:"categoryId" dc:"Filter by category ID" eg:"1"`
	CategoryType    int    `json:"categoryType" v:"in:0,1,2,3#gf.gvalid.rule.in" dc:"Filter by category type: 1=list, 2=single page, 3=external link" eg:"1"`
	IncludeChildren bool   `json:"includeChildren" dc:"Include child categories when categoryId is provided" eg:"true"`
	Status          *int   `json:"status" dc:"Filter by status: 0=draft, 1=published" eg:"1"`
	Title           string `json:"title" dc:"Filter by article title" eg:"Welcome"`
}

// ArticleListRes defines the response for listing CMS articles.
type ArticleListRes struct {
	List  []*ArticleItem `json:"list" dc:"Article list"`
	Total int            `json:"total" dc:"Total number of articles" eg:"20"`
}
