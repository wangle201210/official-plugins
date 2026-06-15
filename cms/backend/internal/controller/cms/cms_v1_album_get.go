// This file implements the CMS album detail controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// AlbumGet returns one CMS album with its ordered images.
func (c *ControllerV1) AlbumGet(ctx context.Context, req *v1.AlbumGetReq) (res *v1.AlbumGetRes, err error) {
	item, err := c.cmsSvc.GetAlbum(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.AlbumGetRes{AlbumItem: toAPIAlbum(item)}, nil
}
