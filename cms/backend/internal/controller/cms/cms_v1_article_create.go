// This file implements the CMS article creation controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ArticleCreate creates a CMS article.
func (c *ControllerV1) ArticleCreate(ctx context.Context, req *v1.ArticleCreateReq) (res *v1.ArticleCreateRes, err error) {
	id, err := c.cmsSvc.CreateArticle(ctx, cmssvc.ArticleSaveInput{
		CategoryId:  req.CategoryId,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Slug:        req.Slug,
		Summary:     req.Summary,
		Cover:       req.Cover,
		Author:      req.Author,
		Source:      req.Source,
		Content:     req.Content,
		Tags:        req.Tags,
		Keywords:    req.Keywords,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
		IsTop:       req.IsTop,
		IsRecommend: req.IsRecommend,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ArticleCreateRes{Id: id}, nil
}
