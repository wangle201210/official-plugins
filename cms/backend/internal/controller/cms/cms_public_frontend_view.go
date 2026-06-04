// This file builds CMS public frontend view models from service data and request context.

package cms

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"lina-core/pkg/bizerr"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
	"strings"
	"time"
)

// buildPublicFrontendView builds the request-specific public page view model.
func (c *ControllerV1) buildPublicFrontendView(ctx context.Context, r *ghttp.Request) (*publicFrontendView, error) {
	categoryID := queryInt64(r, "categoryId")
	categoryCode := strings.TrimSpace(r.GetQuery("category").String())
	slug := strings.TrimSpace(r.GetQuery("article").String())
	routePath := publicFrontendRequestPath(r)
	pageNum := queryInt(r, "page")
	if pageNum <= 0 {
		pageNum = queryInt(r, "pageNum")
	}
	pageNum = normalizePublicFrontendPage(pageNum)
	keyword := strings.TrimSpace(r.GetQuery("keyword").String())
	isSearchPage := routePath == "search"
	templateName := publicFrontendIndexName
	activeCategoryID := categoryID
	var currentArticle *cmssvc.ArticleItem
	if categoryID > 0 {
		templateName = publicFrontendListName
	}
	if isSearchPage {
		templateName = publicFrontendSearchName
		activeCategoryID = 0
	}
	if slug != "" {
		templateName = publicFrontendDetailName
		article, err := c.cmsSvc.GetPublicArticleBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		currentArticle = article
		activeCategoryID = article.CategoryId
	}
	view, err := c.buildPublicFrontendBaseView(ctx, r, templateName, activeCategoryID)
	if err != nil {
		return nil, err
	}
	if activeCategoryID > 0 {
		markPublicFrontendActiveCategory(view.NavCategories, activeCategoryID)
		markPublicFrontendActiveCategory(view.Categories, activeCategoryID)
		view.CurrentCategory = findPublicFrontendCategory(view.Categories, activeCategoryID)
	}
	if categoryID <= 0 && categoryCode == "" && slug == "" && routePath != "" && !isSearchPage {
		if category := findPublicFrontendCategoryByPath(view.Categories, routePath); category != nil {
			categoryID = category.Id
			activeCategoryID = category.Id
			templateName = publicFrontendCategoryListTemplate(category)
			view.CurrentCategory = category
			markPublicFrontendActiveCategory(view.NavCategories, activeCategoryID)
			markPublicFrontendActiveCategory(view.Categories, activeCategoryID)
		} else {
			return nil, bizerr.NewCode(cmssvc.CodePublicContentNotFound)
		}
	}
	if categoryID <= 0 && categoryCode != "" {
		if category := findPublicFrontendCategoryByCode(view.Categories, categoryCode); category != nil {
			categoryID = category.Id
			activeCategoryID = category.Id
			templateName = publicFrontendCategoryListTemplate(category)
			view.CurrentCategory = category
			markPublicFrontendActiveCategory(view.NavCategories, activeCategoryID)
			markPublicFrontendActiveCategory(view.Categories, activeCategoryID)
		}
	}
	if view.CurrentCategory != nil && slug == "" {
		templateName = publicFrontendCategoryListTemplate(view.CurrentCategory)
	}
	listScope := publicFrontendListScope
	if isSearchPage {
		listScope = publicFrontendSearchScope
	}
	listAttrs := publicFrontendTemplateArticleAttrs(templateName, listScope)
	listPageSize := publicFrontendLoopLimit(listScope, listAttrs)
	searchCategoryResolved := true
	if isSearchPage && listAttrs.Code != "" {
		category := findPublicFrontendCategoryByCode(view.Categories, listAttrs.Code)
		searchCategoryResolved = category != nil
		if category != nil {
			categoryID = category.Id
		}
	}
	articlePage := &cmssvc.ArticleListOutput{}
	if (!isSearchPage || keyword != "") && searchCategoryResolved {
		var err error
		articlePage, err = c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{PageNum: pageNum, PageSize: listPageSize, CategoryId: categoryID, Keyword: keyword, Order: listAttrs.Order, IncludeHiddenCategories: view.CurrentCategory != nil || isSearchPage})
		if err != nil {
			return nil, err
		}
	}
	if categoryID > 0 {
		clampedPage := publicFrontendClampedPage(pageNum, articlePage.Total, listPageSize)
		if clampedPage != pageNum {
			pageNum = clampedPage
			articlePage, err = c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{PageNum: pageNum, PageSize: listPageSize, CategoryId: categoryID, Keyword: keyword, Order: listAttrs.Order, IncludeHiddenCategories: true})
			if err != nil {
				return nil, err
			}
		}
	}
	slides, err := c.cmsSvc.ListPublicSlides(ctx)
	if err != nil {
		return nil, err
	}
	allArticlePage, err := c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{PageNum: 1, PageSize: publicFrontendMaxLoopSize, IncludeHiddenCategories: true})
	if err != nil {
		return nil, err
	}
	if categoryID > 0 {
		view.Articles = mapPublicFrontendArticles(articlePage.List, false)
		view.Pagination = buildPublicFrontendPagination(r, articlePage.Total, listPageSize, pageNum)
	} else if isSearchPage {
		view.Articles = mapPublicFrontendSearchArticles(articlePage.List, keyword)
		view.Pagination = buildPublicFrontendPagination(r, articlePage.Total, listPageSize, pageNum)
	} else {
		view.Articles = mapPublicFrontendArticles(allArticlePage.List, false)
	}
	view.PrimarySlide = firstPublicFrontendSlide(slides)
	view.Slides = mapPublicFrontendSlides(slides)
	if slug == "" {
		if isSearchPage {
			view.TemplateName = strings.TrimSuffix(publicFrontendSearchName, ".html")
			view.PageTitle = publicFrontendSearchPageTitle(keyword)
			return view, nil
		}
		if view.CurrentCategory != nil {
			view.TemplateName = strings.TrimSuffix(templateName, ".html")
			view.PageTitle = view.CurrentCategory.Name
			if view.CurrentCategory.Type == cmssvc.CategoryTypeSingle {
				view.TemplateName = strings.TrimSuffix(publicFrontendCategoryContentTemplate(view.CurrentCategory, publicFrontendSingleName), ".html")
				singlePage, err := c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{PageNum: 1, PageSize: 1, CategoryId: view.CurrentCategory.Id, Order: listAttrs.Order, IncludeHiddenCategories: true})
				if err != nil {
					return nil, err
				}
				if len(singlePage.List) > 0 {
					view.CurrentArticle = mapPublicFrontendArticle(singlePage.List[0], true)
				}
			}
			return view, nil
		}
		return view, nil
	}
	view.CurrentArticle = mapPublicFrontendArticle(currentArticle, true)
	view.PageTitle = currentArticle.Title
	view.CurrentCategory = findPublicFrontendCategory(view.Categories, currentArticle.CategoryId)
	if view.CurrentCategory != nil {
		view.TemplateName = strings.TrimSuffix(publicFrontendCategoryContentTemplate(view.CurrentCategory, publicFrontendDetailName), ".html")
	}
	previousArticle, nextArticle, err := c.publicFrontendAdjacentArticles(ctx, currentArticle)
	if err != nil {
		return nil, err
	}
	view.PreviousArticle = previousArticle
	view.NextArticle = nextArticle
	markPublicFrontendActiveCategory(view.NavCategories, currentArticle.CategoryId)
	markPublicFrontendActiveCategory(view.Categories, currentArticle.CategoryId)
	return view, nil
}

// publicFrontendAdjacentArticles resolves previous and next article links within a category.
func (c *ControllerV1) publicFrontendAdjacentArticles(ctx context.Context, current *cmssvc.ArticleItem) (*publicFrontendArticle, *publicFrontendArticle, error) {
	if current == nil || current.CmsArticle == nil {
		return nil, nil, nil
	}
	page, err := c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{PageNum: 1, PageSize: publicFrontendMaxLoopSize, CategoryId: current.CategoryId, Order: cmssvc.PublicArticleOrderDate, IncludeHiddenCategories: true})
	if err != nil {
		return nil, nil, err
	}
	var previous *publicFrontendArticle
	var next *publicFrontendArticle
	for index, item := range page.List {
		if item == nil || item.CmsArticle == nil || item.Id != current.Id {
			continue
		}
		if index > 0 {
			previous = mapPublicFrontendArticle(page.List[index-1], false)
		}
		if index+1 < len(page.List) {
			next = mapPublicFrontendArticle(page.List[index+1], false)
		}
		break
	}
	return previous, next, nil
}

// buildPublicFrontendBaseView loads site-wide public categories, links, messages, and template metadata.
func (c *ControllerV1) buildPublicFrontendBaseView(ctx context.Context, r *ghttp.Request, templateFileName string, activeCategoryID int64) (*publicFrontendView, error) {
	site, err := c.cmsSvc.GetSite(ctx, true)
	if err != nil {
		return nil, err
	}
	site.Logo = publicFrontendNormalizeAssetPath(site.Logo)
	site.Weixin = publicFrontendNormalizeAssetPath(site.Weixin)
	categories, err := c.cmsSvc.ListCategories(ctx, cmssvc.CategoryListInput{})
	if err != nil {
		return nil, err
	}
	publicCategories, err := c.cmsSvc.ListCategories(ctx, cmssvc.CategoryListInput{PublicOnly: true})
	if err != nil {
		return nil, err
	}
	links, err := c.cmsSvc.ListPublicLinks(ctx)
	if err != nil {
		return nil, err
	}
	navCategories := buildPublicFrontendCategoryTree(publicCategories, activeCategoryID)
	allCategories := buildPublicFrontendCategoryTree(categories, activeCategoryID)
	flatCategories := flattenPublicFrontendCategoryTree(allCategories)
	messageState := r.GetQuery("message").String()
	return &publicFrontendView{Site: site, Categories: flatCategories, NavCategories: navCategories, CurrentCategory: findPublicFrontendCategory(flatCategories, activeCategoryID), Links: mapPublicFrontendLinks(links), TemplateName: strings.TrimSuffix(templateFileName, ".html"), Keyword: strings.TrimSpace(r.GetQuery("keyword").String()), FirstCategoryHref: firstPublicFrontendCategoryHref(navCategories), CompanyWeixin: site.Weixin, Submitted: messageState == "submitted", InvalidMessage: messageState == "invalid", MessageError: messageState == "error", ShowMessages: site.ShowMessages == cmssvc.StatusEnabled, Year: time.Now().Year()}, nil
}

// publicFrontendSearchPageTitle formats the search result page title.
func publicFrontendSearchPageTitle(keyword string) string {
	if strings.TrimSpace(keyword) == "" {
		return "搜索结果"
	}
	return strings.TrimSpace(keyword) + "-搜索结果"
}
