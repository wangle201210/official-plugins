// niu_list.go declares the protected cattle-list request/response DTOs for the
// sicau-niu sample plugin API. The list returns a fixed, small in-memory dataset,
// so its size is a bounded constant rather than an unbounded query result.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq is the request for querying the sicau-niu sample cattle list.
type ListReq struct {
	g.Meta  `path:"/plugins/sicau-niu/cattle" method:"get" tags:"Sicau Niu Demo" summary:"Query sicau-niu sample cattle list" dc:"Return a bounded static list of sample cattle records used to verify that a source plugin menu page can read protected backend data. The dataset is a fixed small constant and triggers no database access." permission:"sicau-niu:example:view"`
	Keyword string `json:"keyword" in:"query" dc:"Optional case-insensitive fuzzy filter applied to the cattle name; empty returns the full sample list." eg:"daisy"`
}

// ListRes is the response for querying the sicau-niu sample cattle list.
type ListRes struct {
	List  []*CattleItem `json:"list" dc:"Sample cattle records matching the optional keyword filter."`
	Total int           `json:"total" dc:"Total number of matched sample cattle records." eg:"3"`
}

// CattleItem is one sample cattle record projected to the API response boundary.
type CattleItem struct {
	ID        int64  `json:"id" dc:"Sample cattle identifier." eg:"1"`
	Name      string `json:"name" dc:"Sample cattle name." eg:"Daisy"`
	Breed     string `json:"breed" dc:"Sample cattle breed." eg:"Simmental"`
	WeightKg  int    `json:"weightKg" dc:"Sample cattle live weight in kilograms." eg:"520"`
	CreatedAt int64  `json:"createdAt" dc:"Record creation time. Unix timestamp in milliseconds." eg:"1717372800000"`
}
