// This file declares the CMS public album APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublicAlbumListReq defines the request for the public album list.
type PublicAlbumListReq struct {
	g.Meta     `path:"/cms/public/albums" method:"get" tags:"CMS Public" summary:"Get public CMS albums" dc:"Query enabled CMS albums under enabled categories by page. Each row carries its cover and image count." permission:""`
	PageNum    int   `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize   int   `json:"pageSize" d:"12" v:"min:1|max:100" dc:"Number of items per page, at most 100" eg:"12"`
	CategoryId int64 `json:"categoryId" dc:"Filter by category ID" eg:"1"`
}

// PublicAlbumListRes defines the response for the public album list.
type PublicAlbumListRes struct {
	List  []*AlbumItem `json:"list" dc:"Album list" eg:"[]"`
	Total int          `json:"total" dc:"Total number of albums" eg:"5"`
}

// PublicAlbumGetReq defines the request for one public album detail.
type PublicAlbumGetReq struct {
	g.Meta `path:"/cms/public/albums/{id}" method:"get" tags:"CMS Public" summary:"Get public CMS album detail" dc:"Read one enabled CMS album by ID including its ordered images. Disabled albums or albums under disabled categories are treated as not found." permission:""`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Album ID" eg:"1"`
}

// PublicAlbumGetRes defines the response for one public album detail.
type PublicAlbumGetRes struct {
	*AlbumItem `dc:"Album detail with images"`
}
