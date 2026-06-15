// This file declares the CMS article batch delete API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ArticleBatchDeleteReq defines the request for deleting multiple CMS articles.
type ArticleBatchDeleteReq struct {
	g.Meta `path:"/cms/articles" method:"delete" tags:"CMS Articles" summary:"Batch delete CMS articles" dc:"Soft delete multiple CMS articles in one statement. The whole batch is rejected when any target article does not exist." permission:"cms:article:remove"`
	Ids    []int64 `json:"ids" v:"required#gf.gvalid.rule.required" dc:"Article ID list carried in the JSON request body; at most 100 IDs per request, larger batches are rejected" eg:"[1,2]"`
}

// ArticleBatchDeleteRes defines the response for deleting multiple CMS articles.
type ArticleBatchDeleteRes struct{}
