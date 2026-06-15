// This file implements the CMS album creation controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// AlbumCreate creates a CMS album together with its full image list.
func (c *ControllerV1) AlbumCreate(ctx context.Context, req *v1.AlbumCreateReq) (res *v1.AlbumCreateRes, err error) {
	id, err := c.cmsSvc.CreateAlbum(ctx, cmssvc.AlbumSaveInput{
		CategoryId:  req.CategoryId,
		Name:        req.Name,
		Cover:       req.Cover,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
		Images:      toAlbumImageInputs(req.Images),
	})
	if err != nil {
		return nil, err
	}
	return &v1.AlbumCreateRes{Id: id}, nil
}
