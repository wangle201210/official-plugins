// This file implements the CMS public album list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// PublicAlbumList returns publicly visible CMS albums with image counts.
func (c *ControllerV1) PublicAlbumList(ctx context.Context, req *v1.PublicAlbumListReq) (res *v1.PublicAlbumListRes, err error) {
	out, err := c.cmsSvc.ListPublicAlbums(ctx, cmssvc.PublicAlbumListInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		CategoryId: req.CategoryId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PublicAlbumListRes{List: toAPIAlbums(out.List), Total: out.Total}, nil
}
