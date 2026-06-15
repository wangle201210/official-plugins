// This file declares the CMS article batch status update API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ArticleBatchUpdateReq defines the request for changing the status of multiple CMS articles.
type ArticleBatchUpdateReq struct {
	g.Meta `path:"/cms/articles" method:"put" tags:"CMS Articles" summary:"Batch update CMS article status" dc:"Publish or unpublish multiple CMS articles in one transaction. The whole batch is rejected when any target article does not exist. Publishing fills an empty publication time with the current time and keeps existing publication times; unpublishing keeps publication times." permission:"cms:article:edit"`
	Ids    []int64 `json:"ids" v:"required#gf.gvalid.rule.required" dc:"Article ID list; at most 100 IDs per request, larger batches are rejected" eg:"[1,2]"`
	Status int     `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Target status: 0=draft (unpublish), 1=published" eg:"1"`
}

// ArticleBatchUpdateRes defines the response for changing the status of multiple CMS articles.
type ArticleBatchUpdateRes struct{}
