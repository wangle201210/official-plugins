// This file declares the CMS album get and delete APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// AlbumGetReq defines the request for reading one CMS album.
type AlbumGetReq struct {
	g.Meta `path:"/cms/albums/{id}" method:"get" tags:"CMS Albums" summary:"Get CMS album detail" dc:"Read one CMS album by ID including its ordered images." permission:"cms:album:query"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Album ID" eg:"1"`
}

// AlbumGetRes defines the response for reading one CMS album.
type AlbumGetRes struct {
	*AlbumItem `dc:"Album detail with images"`
}

// AlbumDeleteReq defines the request for deleting one CMS album.
type AlbumDeleteReq struct {
	g.Meta `path:"/cms/albums/{id}" method:"delete" tags:"CMS Albums" summary:"Delete CMS album" dc:"Delete one CMS album by ID together with all of its image rows." permission:"cms:album:remove"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Album ID" eg:"1"`
}

// AlbumDeleteRes defines the response for deleting one CMS album.
type AlbumDeleteRes struct{}
