// This file declares the CMS album update API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// AlbumUpdateReq defines the request for updating a CMS album with its images.
type AlbumUpdateReq struct {
	g.Meta      `path:"/cms/albums/{id}" method:"put" tags:"CMS Albums" summary:"Update CMS album" dc:"Update a CMS album and replace its full image list as a whole in one transaction; at most 100 images per album." permission:"cms:album:edit"`
	Id          int64             `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Album ID" eg:"1"`
	CategoryId  int64             `json:"categoryId" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Category ID" eg:"1"`
	Name        string            `json:"name" v:"required#gf.gvalid.rule.required" dc:"Album name" eg:"Campus gallery"`
	Cover       string            `json:"cover" dc:"Cover image URL" eg:"/uploads/cover.png"`
	Description string            `json:"description" dc:"Album description" eg:"Campus photos"`
	Sort        int               `json:"sort" dc:"Display order" eg:"1"`
	Status      int               `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
	Images      []*AlbumImageItem `json:"images" dc:"Full album image list replaced as a whole on save; at most 100 images per album, larger lists are rejected" eg:"[]"`
}

// AlbumUpdateRes defines the response for updating a CMS album.
type AlbumUpdateRes struct{}
