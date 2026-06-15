// This file implements the CMS album delete controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// AlbumDelete deletes one CMS album together with its image rows.
func (c *ControllerV1) AlbumDelete(ctx context.Context, req *v1.AlbumDeleteReq) (res *v1.AlbumDeleteRes, err error) {
	if err = c.cmsSvc.DeleteAlbum(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.AlbumDeleteRes{}, nil
}
