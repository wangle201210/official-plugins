// This file implements the CMS public album detail controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// PublicAlbumGet returns one publicly visible CMS album with its images.
func (c *ControllerV1) PublicAlbumGet(ctx context.Context, req *v1.PublicAlbumGetReq) (res *v1.PublicAlbumGetRes, err error) {
	item, err := c.cmsSvc.GetPublicAlbum(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.PublicAlbumGetRes{AlbumItem: toAPIAlbum(item)}, nil
}
