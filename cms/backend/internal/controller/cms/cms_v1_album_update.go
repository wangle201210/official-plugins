// This file implements the CMS album update controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// AlbumUpdate updates a CMS album and replaces its full image list.
func (c *ControllerV1) AlbumUpdate(ctx context.Context, req *v1.AlbumUpdateReq) (res *v1.AlbumUpdateRes, err error) {
	err = c.cmsSvc.UpdateAlbum(ctx, cmssvc.AlbumSaveInput{
		Id:          req.Id,
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
	return &v1.AlbumUpdateRes{}, nil
}
