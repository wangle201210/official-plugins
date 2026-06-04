// This file implements the CMS article list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ArticleList returns paged CMS articles.
func (c *ControllerV1) ArticleList(ctx context.Context, req *v1.ArticleListReq) (res *v1.ArticleListRes, err error) {
	out, err := c.cmsSvc.ListArticles(ctx, cmssvc.ArticleListInput{
		PageNum:         req.PageNum,
		PageSize:        req.PageSize,
		CategoryId:      req.CategoryId,
		CategoryType:    req.CategoryType,
		IncludeChildren: req.IncludeChildren,
		Status:          req.Status,
		Title:           req.Title,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ArticleListRes{List: toAPIArticles(out.List), Total: out.Total}, nil
}
