// This file implements the CMS album list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// AlbumList returns paged CMS albums with image counts.
func (c *ControllerV1) AlbumList(ctx context.Context, req *v1.AlbumListReq) (res *v1.AlbumListRes, err error) {
	out, err := c.cmsSvc.ListAlbums(ctx, cmssvc.AlbumListInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		CategoryId: req.CategoryId,
		Status:     req.Status,
		Name:       req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &v1.AlbumListRes{List: toAPIAlbums(out.List), Total: out.Total}, nil
}
