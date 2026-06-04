// This file maps CMS service projections into public frontend models and ordering helpers.

package cms

import (
	"html"
	"html/template"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// buildPublicFrontendCategoryTree builds public category trees with active state applied.
func buildPublicFrontendCategoryTree(categories []*cmssvc.CategoryItem, activeID int64) []*publicFrontendCategory {
	result := make([]*publicFrontendCategory, 0, len(categories))
	for _, item := range categories {
		category := mapPublicFrontendCategory(item, activeID, 0, nil)
		if category != nil {
			result = append(result, category)
		}
	}
	return result
}

// mapPublicFrontendCategory converts one service category into a public template node.
func mapPublicFrontendCategory(item *cmssvc.CategoryItem, activeID int64, depth int, parent *publicFrontendCategory) *publicFrontendCategory {
	if item == nil || item.CmsCategory == nil {
		return nil
	}
	category := &publicFrontendCategory{Id: item.Id, Code: item.Code, Name: item.Name, Path: item.Path, Href: publicFrontendCategoryHref(item), Depth: depth, Active: item.Id == activeID, External: item.Type == cmssvc.CategoryTypeExternal, Type: item.Type, ListTemplate: item.ListTemplate, ContentTemplate: item.ContentTemplate, Title: item.Title, Keywords: item.Keywords, Description: item.Description, Parent: parent, Children: make([]*publicFrontendCategory, 0, len(item.Children))}
	for _, child := range item.Children {
		childCategory := mapPublicFrontendCategory(child, activeID, depth+1, category)
		if childCategory == nil {
			continue
		}
		if childCategory.Active {
			category.Active = true
		}
		category.Children = append(category.Children, childCategory)
	}
	return category
}

// flattenPublicFrontendCategoryTree flattens nested categories for template helper lookup.
func flattenPublicFrontendCategoryTree(categories []*publicFrontendCategory) []*publicFrontendCategory {
	result := make([]*publicFrontendCategory, 0)
	var walk func(items []*publicFrontendCategory)
	walk = func(items []*publicFrontendCategory) {
		for _, item := range items {
			if item == nil {
				continue
			}
			result = append(result, item)
			walk(item.Children)
		}
	}
	walk(categories)
	return result
}

// markPublicFrontendActiveCategory marks the active category and its ancestors.
func markPublicFrontendActiveCategory(categories []*publicFrontendCategory, activeID int64) {
	for _, category := range categories {
		if category == nil {
			continue
		}
		category.Active = category.Id == activeID
		markPublicFrontendActiveCategory(category.Children, activeID)
		for _, child := range category.Children {
			if child != nil && child.Active {
				category.Active = true
				break
			}
		}
	}
}

// publicFrontendCategoryHref resolves the public link for a category.
func publicFrontendCategoryHref(category *cmssvc.CategoryItem) string {
	if category.Type == cmssvc.CategoryTypeExternal && strings.TrimSpace(category.Outlink) != "" {
		return publicFrontendNormalizeAssetPath(category.Outlink)
	}
	if href := publicFrontendCategoryPathHref(category.Path); href != "" {
		return href
	}
	values := url.Values{}
	if strings.TrimSpace(category.Code) != "" {
		values.Set("category", strings.TrimSpace(category.Code))
	} else {
		values.Set("categoryId", strconv.FormatInt(category.Id, 10))
	}
	return "/cms-site?" + values.Encode()
}

// findPublicFrontendCategory finds a category node by ID in a tree.
func findPublicFrontendCategory(categories []*publicFrontendCategory, id int64) *publicFrontendCategory {
	if id <= 0 {
		return nil
	}
	for _, category := range categories {
		if category != nil && category.Id == id {
			return category
		}
	}
	return nil
}

// findPublicFrontendCategoryByPath finds a category node by its normalized path.
func findPublicFrontendCategoryByPath(categories []*publicFrontendCategory, routePath string) *publicFrontendCategory {
	normalized := publicFrontendNormalizeCategoryPath(routePath)
	if normalized == "" {
		return nil
	}
	for _, category := range categories {
		if category != nil && publicFrontendNormalizeCategoryPath(category.Path) == normalized {
			return category
		}
	}
	return nil
}

// findPublicFrontendCategoryByCode finds a category node by code.
func findPublicFrontendCategoryByCode(categories []*publicFrontendCategory, code string) *publicFrontendCategory {
	normalized := strings.TrimSpace(code)
	if normalized == "" {
		return nil
	}
	for _, category := range categories {
		if category != nil && category.Code == normalized {
			return category
		}
	}
	return nil
}

// publicFrontendCategoryByCode returns a category node from the current template scope by code.
func publicFrontendCategoryByCode(view *publicFrontendView, code string) *publicFrontendCategory {
	if view == nil {
		return nil
	}
	return findPublicFrontendCategoryByCode(view.Categories, code)
}

// publicFrontendRootCategory returns the first root category from the current scope.
func publicFrontendRootCategory(category *publicFrontendCategory) *publicFrontendCategory {
	if category == nil {
		return nil
	}
	current := category
	for current.Parent != nil {
		current = current.Parent
	}
	return current
}

// publicFrontendCategoryChildren returns children for a template category scope.
func publicFrontendCategoryChildren(view *publicFrontendView, selector any, limit int) []*publicFrontendCategory {
	if view == nil || selector == nil {
		return nil
	}
	var category *publicFrontendCategory
	switch value := selector.(type) {
	case *publicFrontendCategory:
		category = value
	case nil:
		return nil
	case string:
		category = findPublicFrontendCategoryByCode(view.Categories, value)
	case int64:
		category = findPublicFrontendCategory(view.Categories, value)
	case int:
		category = findPublicFrontendCategory(view.Categories, int64(value))
	default:
		return nil
	}
	if category == nil {
		return nil
	}
	return publicFrontendLimitCategories(category.Children, limit)
}

// publicFrontendCategoryArticles returns articles belonging to a template category scope.
func publicFrontendCategoryArticles(view *publicFrontendView, code string, limit int, order string) []*publicFrontendArticle {
	if view == nil {
		return nil
	}
	category := findPublicFrontendCategoryByCode(view.Categories, code)
	if category == nil {
		return nil
	}
	codes := make(map[string]struct{})
	var collect func(item *publicFrontendCategory)
	collect = func(item *publicFrontendCategory) {
		if item == nil {
			return
		}
		codes[strconv.FormatInt(item.Id, 10)] = struct{}{}
		for _, child := range item.Children {
			collect(child)
		}
	}
	collect(category)
	articles := make([]*publicFrontendArticle, 0)
	for _, article := range view.Articles {
		if article == nil {
			continue
		}
		if _, ok := codes[strconv.FormatInt(article.CategoryId, 10)]; !ok {
			continue
		}
		articles = append(articles, article)
	}
	return publicFrontendLimitArticles(publicFrontendOrderArticles(articles, order), limit)
}

// publicFrontendOrderArticles sorts public articles according to template order attributes.
func publicFrontendOrderArticles(items []*publicFrontendArticle, order string) []*publicFrontendArticle {
	if len(items) == 0 {
		return items
	}
	ordered := make([]*publicFrontendArticle, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		cloned := *item
		ordered = append(ordered, &cloned)
	}
	if len(ordered) <= 1 {
		return publicFrontendReindexArticles(ordered)
	}
	switch cmssvc.NormalizePublicArticleOrder(order) {
	case cmssvc.PublicArticleOrderID:
		sort.SliceStable(ordered, func(i int, j int) bool {
			return ordered[i].Id > ordered[j].Id
		})
	case cmssvc.PublicArticleOrderDate:
		sort.SliceStable(ordered, func(i int, j int) bool {
			return publicFrontendArticleDateLess(ordered[i], ordered[j])
		})
	case cmssvc.PublicArticleOrderManual:
		sort.SliceStable(ordered, func(i int, j int) bool {
			if ordered[i].Sort != ordered[j].Sort {
				return ordered[i].Sort < ordered[j].Sort
			}
			return publicFrontendArticleDefaultLess(ordered[i], ordered[j])
		})
	case cmssvc.PublicArticleOrderViews:
		sort.SliceStable(ordered, func(i int, j int) bool {
			if ordered[i].Views != ordered[j].Views {
				return ordered[i].Views > ordered[j].Views
			}
			return publicFrontendArticleDefaultLess(ordered[i], ordered[j])
		})
	default:
		sort.SliceStable(ordered, func(i int, j int) bool {
			return publicFrontendArticleDefaultLess(ordered[i], ordered[j])
		})
	}
	return publicFrontendReindexArticles(ordered)
}

// publicFrontendArticleDefaultLess compares articles using the default ranking order.
func publicFrontendArticleDefaultLess(left *publicFrontendArticle, right *publicFrontendArticle) bool {
	if left == nil || right == nil {
		return right != nil
	}
	if left.IsTop != right.IsTop {
		return left.IsTop > right.IsTop
	}
	if left.IsRecommend != right.IsRecommend {
		return left.IsRecommend > right.IsRecommend
	}
	if left.Sort != right.Sort {
		return left.Sort < right.Sort
	}
	if left.PublishedAtUnix != right.PublishedAtUnix {
		return left.PublishedAtUnix > right.PublishedAtUnix
	}
	return left.Id > right.Id
}

// publicFrontendArticleDateLess compares articles by publication time.
func publicFrontendArticleDateLess(left *publicFrontendArticle, right *publicFrontendArticle) bool {
	if left == nil || right == nil {
		return right != nil
	}
	if left.IsTop != right.IsTop {
		return left.IsTop > right.IsTop
	}
	if left.PublishedAtUnix != right.PublishedAtUnix {
		return left.PublishedAtUnix > right.PublishedAtUnix
	}
	if left.IsRecommend != right.IsRecommend {
		return left.IsRecommend > right.IsRecommend
	}
	if left.Sort != right.Sort {
		return left.Sort < right.Sort
	}
	return left.Id > right.Id
}

// publicFrontendReindexArticles refreshes display indexes after filtering or sorting.
func publicFrontendReindexArticles(items []*publicFrontendArticle) []*publicFrontendArticle {
	for index, item := range items {
		if item == nil {
			continue
		}
		item.Index = index + 1
	}
	return items
}

// publicFrontendGroupedSlides returns slides for the requested template group.
func publicFrontendGroupedSlides(slides []*publicFrontendSlide, group string, limit int) []*publicFrontendSlide {
	filtered := make([]*publicFrontendSlide, 0, len(slides))
	for _, slide := range slides {
		if slide == nil {
			continue
		}
		if group == "" || slide.Group == group {
			filtered = append(filtered, slide)
		}
	}
	return publicFrontendLimitSlides(filtered, limit)
}

// publicFrontendGroupedLinks returns links for the requested template group.
func publicFrontendGroupedLinks(links []*publicFrontendLink, group string, limit int) []*publicFrontendLink {
	filtered := make([]*publicFrontendLink, 0, len(links))
	for _, link := range links {
		if link == nil {
			continue
		}
		if group == "" || link.Group == group {
			filtered = append(filtered, link)
		}
	}
	return publicFrontendLimitLinks(filtered, limit)
}

// publicFrontendLimitCategories limits category collections for template loops.
func publicFrontendLimitCategories(items []*publicFrontendCategory, limit int) []*publicFrontendCategory {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// publicFrontendLimitArticles limits article collections for template loops.
func publicFrontendLimitArticles(items []*publicFrontendArticle, limit int) []*publicFrontendArticle {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// publicFrontendLimitSlides limits slide collections for template loops.
func publicFrontendLimitSlides(items []*publicFrontendSlide, limit int) []*publicFrontendSlide {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// publicFrontendLimitLinks limits link collections for template loops.
func publicFrontendLimitLinks(items []*publicFrontendLink, limit int) []*publicFrontendLink {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// firstPublicFrontendCategoryHref returns the first usable category link for navigation.
func firstPublicFrontendCategoryHref(categories []*publicFrontendCategory) string {
	for _, category := range categories {
		if category != nil && !category.External {
			return category.Href
		}
	}
	return "/cms-site"
}

// mapPublicFrontendArticles converts service articles into public template models.
func mapPublicFrontendArticles(items []*cmssvc.ArticleItem, includeContent bool) []*publicFrontendArticle {
	articles := make([]*publicFrontendArticle, 0, len(items))
	for index, item := range items {
		if item == nil || item.CmsArticle == nil {
			continue
		}
		article := mapPublicFrontendArticle(item, includeContent)
		article.Index = index + 1
		articles = append(articles, article)
	}
	return articles
}

// mapPublicFrontendSearchArticles converts search results and adds highlighted snippets.
func mapPublicFrontendSearchArticles(items []*cmssvc.ArticleItem, keyword string) []*publicFrontendArticle {
	articles := make([]*publicFrontendArticle, 0, len(items))
	for index, item := range items {
		if item == nil || item.CmsArticle == nil {
			continue
		}
		article := mapPublicFrontendArticle(item, false)
		article.Index = index + 1
		article.SearchPreview = publicFrontendSearchPreview(item, keyword)
		articles = append(articles, article)
	}
	return articles
}

// mapPublicFrontendArticle converts one service article into a public template model.
func mapPublicFrontendArticle(item *cmssvc.ArticleItem, includeContent bool) *publicFrontendArticle {
	if item == nil || item.CmsArticle == nil {
		return &publicFrontendArticle{}
	}
	values := url.Values{}
	values.Set("article", item.Slug)
	article := &publicFrontendArticle{Id: item.Id, CategoryId: item.CategoryId, Title: item.Title, Subtitle: item.Subtitle, Summary: item.Summary, Cover: publicFrontendNormalizeAssetPath(item.Cover), Author: item.Author, Source: item.Source, Keywords: item.Keywords, Description: item.Description, CategoryName: item.CategoryName, PublishedAt: publicFrontendDate(item.PublishedAt), PublishedAtUnix: publicFrontendTimestamp(item.PublishedAt), Href: "/cms-site?" + values.Encode(), Views: item.Views, Sort: item.Sort, IsTop: item.IsTop, IsRecommend: item.IsRecommend}
	if includeContent {
		article.ContentHTML = template.HTML(publicFrontendNormalizeContentHTML(item.Content))
	}
	return article
}

// publicFrontendSearchPreview builds a highlighted search snippet for an article.
func publicFrontendSearchPreview(item *cmssvc.ArticleItem, keyword string) template.HTML {
	if item == nil || item.CmsArticle == nil {
		return ""
	}
	keyword = strings.TrimSpace(keyword)
	candidates := []string{item.Summary, item.Description, item.Subtitle, item.Content, item.Tags, item.Keywords, item.Title}
	var fallback string
	for _, candidate := range candidates {
		text := publicFrontendPlainText(candidate)
		if fallback == "" && text != "" {
			fallback = text
		}
		if keyword == "" || !strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
			continue
		}
		return publicFrontendHighlightPreview(text, keyword)
	}
	return publicFrontendHighlightPreview(fallback, keyword)
}

// publicFrontendPlainText strips HTML and scripts from public article content.
func publicFrontendPlainText(value string) string {
	text := html.UnescapeString(strings.TrimSpace(value))
	text = publicFrontendScriptPattern.ReplaceAllString(text, " ")
	text = publicFrontendStylePattern.ReplaceAllString(text, " ")
	text = publicFrontendHTMLTagPattern.ReplaceAllString(text, " ")
	text = publicFrontendWhitespacePattern.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// publicFrontendHighlightPreview highlights keyword matches inside a clipped preview.
func publicFrontendHighlightPreview(text string, keyword string) template.HTML {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return template.HTML(template.HTMLEscapeString(publicFrontendClipRunes(text, publicFrontendPreviewRunes)))
	}
	hitStart, hitEnd := publicFrontendFindKeywordRunes(text, keyword)
	if hitStart < 0 {
		return template.HTML(template.HTMLEscapeString(publicFrontendClipRunes(text, publicFrontendPreviewRunes)))
	}
	runes := []rune(text)
	start := hitStart - publicFrontendPreviewLead
	if start < 0 {
		start = 0
	}
	end := start + publicFrontendPreviewRunes
	if end < hitEnd {
		end = hitEnd
	}
	if end > len(runes) {
		end = len(runes)
		if end-publicFrontendPreviewRunes > 0 {
			start = end - publicFrontendPreviewRunes
		}
	}
	prefix := ""
	if start > 0 {
		prefix = publicFrontendTextMore
	}
	suffix := ""
	if end < len(runes) {
		suffix = publicFrontendTextMore
	}
	relativeStart := hitStart - start
	relativeEnd := hitEnd - start
	visible := runes[start:end]
	var builder strings.Builder
	builder.WriteString(template.HTMLEscapeString(prefix))
	builder.WriteString(template.HTMLEscapeString(string(visible[:relativeStart])))
	builder.WriteString(`<mark>`)
	builder.WriteString(template.HTMLEscapeString(string(visible[relativeStart:relativeEnd])))
	builder.WriteString(`</mark>`)
	builder.WriteString(template.HTMLEscapeString(string(visible[relativeEnd:])))
	builder.WriteString(template.HTMLEscapeString(suffix))
	return template.HTML(builder.String())
}

// publicFrontendFindKeywordRunes finds a keyword match in rune-indexed text.
func publicFrontendFindKeywordRunes(text string, keyword string) (int, int) {
	runes := []rune(text)
	keywordRunes := []rune(keyword)
	if len(keywordRunes) == 0 || len(keywordRunes) > len(runes) {
		return -1, -1
	}
	lowerKeyword := strings.ToLower(keyword)
	for start := 0; start+len(keywordRunes) <= len(runes); start++ {
		end := start + len(keywordRunes)
		if strings.ToLower(string(runes[start:end])) == lowerKeyword {
			return start, end
		}
	}
	return -1, -1
}

// publicFrontendClipRunes clips a rune slice with a suffix when needed.
func publicFrontendClipRunes(text string, limit int) string {
	runes := []rune(text)
	if limit <= 0 || len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + publicFrontendTextMore
}

// mapPublicFrontendLinks converts service links into public template models.
func mapPublicFrontendLinks(items []*cmssvc.LinkItem) []*publicFrontendLink {
	links := make([]*publicFrontendLink, 0, len(items))
	for index, item := range items {
		if item == nil {
			continue
		}
		links = append(links, &publicFrontendLink{Name: item.Name, Url: publicFrontendNormalizeAssetPath(item.Url), Logo: publicFrontendNormalizeAssetPath(item.Logo), Group: publicFrontendLinkGroup(item.GroupCode, index)})
	}
	return links
}

// mapPublicFrontendMessages converts approved messages into public template models.
func mapPublicFrontendMessages(items []*cmssvc.MessageItem) []*publicFrontendMessage {
	messages := make([]*publicFrontendMessage, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		messages = append(messages, &publicFrontendMessage{Id: item.Id, Name: item.Name, Content: item.Content, Reply: item.Reply, CreatedAt: publicFrontendDate(item.CreatedAt), UpdatedAt: publicFrontendDate(item.UpdatedAt)})
	}
	return messages
}

// publicFrontendLinkGroup returns links belonging to a template group.
func publicFrontendLinkGroup(groupCode string, index int) string {
	if strings.TrimSpace(groupCode) != "" {
		return strings.TrimSpace(groupCode)
	}
	switch {
	case index < 3:
		return "1"
	case index < 6:
		return "2"
	case index < 9:
		return "3"
	default:
		return "4"
	}
}
