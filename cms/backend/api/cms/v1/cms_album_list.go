// This file declares the CMS album list API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// AlbumListReq defines the request for listing CMS albums.
type AlbumListReq struct {
	g.Meta     `path:"/cms/albums" method:"get" tags:"CMS Albums" summary:"Get CMS album list" dc:"Query CMS albums by page with optional category, status, and name filters. Each row carries its image count." permission:"cms:album:query"`
	PageNum    int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize   int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page, at most 100" eg:"10"`
	CategoryId int64  `json:"categoryId" dc:"Filter by category ID" eg:"1"`
	Status     *int   `json:"status" dc:"Filter by status: 0=disabled, 1=enabled" eg:"1"`
	Name       string `json:"name" dc:"Filter by album name" eg:"Campus"`
}

// AlbumListRes defines the response for listing CMS albums.
type AlbumListRes struct {
	List  []*AlbumItem `json:"list" dc:"Album list" eg:"[]"`
	Total int          `json:"total" dc:"Total number of albums" eg:"5"`
}
