// This file implements CMS controller response projection helpers.

package cms

import (
	v1 "lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"

	"github.com/gogf/gf/v2/os/gtime"
)

// toAPISite converts a service-layer site into the API response projection.
func toAPISite(item *cmssvc.SiteItem) *v1.SiteItem {
	if item == nil {
		return nil
	}
	return &v1.SiteItem{
		Id:           item.Id,
		SiteKey:      item.SiteKey,
		Name:         item.Name,
		Logo:         item.Logo,
		Weixin:       item.Weixin,
		Domain:       item.Domain,
		Slogan:       item.Slogan,
		Keywords:     item.Keywords,
		Description:  item.Description,
		Icp:          item.Icp,
		Contact:      item.Contact,
		Phone:        item.Phone,
		Email:        item.Email,
		Address:      item.Address,
		Status:       item.Status,
		ShowMessages: item.ShowMessages,
		CreatedBy:    item.CreatedBy,
		UpdatedBy:    item.UpdatedBy,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

// toAPIPublicMessages converts approved visitor messages into public-safe projections.
func toAPIPublicMessages(list []*cmssvc.MessageItem) []*v1.PublicMessageItem {
	items := make([]*v1.PublicMessageItem, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		items = append(items, &v1.PublicMessageItem{
			Id:        item.Id,
			Name:      item.Name,
			Content:   item.Content,
			Reply:     item.Reply,
			CreatedAt: toAPITimeMillis(item.CreatedAt),
			UpdatedAt: toAPITimeMillis(item.UpdatedAt),
		})
	}
	return items
}

// toAPITimeMillis converts GoFrame time values to Unix milliseconds for public APIs.
func toAPITimeMillis(value *gtime.Time) *int64 {
	if value == nil {
		return nil
	}
	millis := value.TimestampMilli()
	return &millis
}

// toAPICategory converts a service-layer category into the API response projection.
func toAPICategory(item *cmssvc.CategoryItem) *v1.CategoryItem {
	if item == nil || item.CmsCategory == nil {
		return nil
	}
	children := make([]*v1.CategoryItem, 0, len(item.Children))
	for _, child := range item.Children {
		children = append(children, toAPICategory(child))
	}
	return &v1.CategoryItem{
		Id:              item.Id,
		ParentId:        item.ParentId,
		Code:            item.Code,
		Name:            item.Name,
		Type:            item.Type,
		Path:            item.Path,
		ListTemplate:    item.ListTemplate,
		ContentTemplate: item.ContentTemplate,
		Cover:           item.Cover,
		Outlink:         item.Outlink,
		Title:           item.Title,
		Keywords:        item.Keywords,
		Description:     item.Description,
		Sort:            item.Sort,
		Status:          item.Status,
		CreatedBy:       item.CreatedBy,
		UpdatedBy:       item.UpdatedBy,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		Children:        children,
	}
}

// toAPICategories converts service-layer category nodes into API response projections.
func toAPICategories(list []*cmssvc.CategoryItem) []*v1.CategoryItem {
	items := make([]*v1.CategoryItem, 0, len(list))
	for _, item := range list {
		items = append(items, toAPICategory(item))
	}
	return items
}

// toAPIArticle converts a service-layer article into the API response projection.
func toAPIArticle(item *cmssvc.ArticleItem) *v1.ArticleItem {
	if item == nil || item.CmsArticle == nil {
		return nil
	}
	return &v1.ArticleItem{
		Id:           item.Id,
		CategoryId:   item.CategoryId,
		CategoryName: item.CategoryName,
		Title:        item.Title,
		Subtitle:     item.Subtitle,
		Slug:         item.Slug,
		Summary:      item.Summary,
		Cover:        item.Cover,
		Author:       item.Author,
		Source:       item.Source,
		Content:      item.Content,
		Tags:         item.Tags,
		Keywords:     item.Keywords,
		Description:  item.Description,
		Sort:         item.Sort,
		Status:       item.Status,
		IsTop:        item.IsTop,
		IsRecommend:  item.IsRecommend,
		Views:        item.Views,
		PublishedAt:  item.PublishedAt,
		CreatedBy:    item.CreatedBy,
		UpdatedBy:    item.UpdatedBy,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

// toAPIArticles converts service-layer articles into API response projections.
func toAPIArticles(list []*cmssvc.ArticleItem) []*v1.ArticleItem {
	items := make([]*v1.ArticleItem, 0, len(list))
	for _, item := range list {
		items = append(items, toAPIArticle(item))
	}
	return items
}

// toAPIMessage converts a service-layer visitor message into the API response projection.
func toAPIMessage(item *cmssvc.MessageItem) *v1.MessageItem {
	if item == nil {
		return nil
	}
	return &v1.MessageItem{
		Id:        item.Id,
		Name:      item.Name,
		Mobile:    item.Mobile,
		Email:     item.Email,
		Content:   item.Content,
		Reply:     item.Reply,
		Status:    item.Status,
		UserIp:    item.UserIp,
		UserAgent: item.UserAgent,
		CreatedBy: item.CreatedBy,
		UpdatedBy: item.UpdatedBy,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

// toAPIMessages converts service-layer visitor messages into API response projections.
func toAPIMessages(list []*cmssvc.MessageItem) []*v1.MessageItem {
	items := make([]*v1.MessageItem, 0, len(list))
	for _, item := range list {
		items = append(items, toAPIMessage(item))
	}
	return items
}

// toAPILinks converts service-layer links into API response projections.
func toAPILinks(list []*cmssvc.LinkItem) []*v1.LinkItem {
	items := make([]*v1.LinkItem, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		items = append(items, &v1.LinkItem{
			Id:        item.Id,
			GroupCode: item.GroupCode,
			Name:      item.Name,
			Url:       item.Url,
			Logo:      item.Logo,
			Sort:      item.Sort,
			Status:    item.Status,
			CreatedBy: item.CreatedBy,
			UpdatedBy: item.UpdatedBy,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return items
}

// toAPISlides converts service-layer slides into API response projections.
func toAPISlides(list []*cmssvc.SlideItem) []*v1.SlideItem {
	items := make([]*v1.SlideItem, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		items = append(items, &v1.SlideItem{
			Id:        item.Id,
			GroupCode: item.GroupCode,
			Title:     item.Title,
			Subtitle:  item.Subtitle,
			Image:     item.Image,
			Link:      item.Link,
			Sort:      item.Sort,
			Status:    item.Status,
			CreatedBy: item.CreatedBy,
			UpdatedBy: item.UpdatedBy,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return items
}
