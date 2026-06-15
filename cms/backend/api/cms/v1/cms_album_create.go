// This file declares the CMS album create API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// AlbumCreateReq defines the request for creating a CMS album with its images.
type AlbumCreateReq struct {
	g.Meta      `path:"/cms/albums" method:"post" tags:"CMS Albums" summary:"Create CMS album" dc:"Create a CMS album together with its full image list. Images are saved as a whole in one transaction; at most 100 images per album." permission:"cms:album:add"`
	CategoryId  int64             `json:"categoryId" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Category ID" eg:"1"`
	Name        string            `json:"name" v:"required#gf.gvalid.rule.required" dc:"Album name" eg:"Campus gallery"`
	Cover       string            `json:"cover" dc:"Cover image URL" eg:"/uploads/cover.png"`
	Description string            `json:"description" dc:"Album description" eg:"Campus photos"`
	Sort        int               `json:"sort" dc:"Display order" eg:"1"`
	Status      int               `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
	Images      []*AlbumImageItem `json:"images" dc:"Full album image list replaced as a whole on save; at most 100 images per album, larger lists are rejected" eg:"[]"`
}

// AlbumCreateRes defines the response for creating a CMS album.
type AlbumCreateRes struct {
	Id int64 `json:"id" dc:"Album ID" eg:"1"`
}
