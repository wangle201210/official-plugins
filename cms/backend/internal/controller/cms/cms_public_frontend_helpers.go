// This file contains CMS public frontend pagination, path, date, and query helpers.

package cms

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"html/template"
	cmsplugin "lina-plugin-cms"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// buildPublicFrontendPagination builds pagination links and counters for public lists.
func buildPublicFrontendPagination(r *ghttp.Request, total int, pageSize int, current int) *publicFrontendPagination {
	if pageSize <= 0 {
		pageSize = publicFrontendPageSize
	}
	current = normalizePublicFrontendPage(current)
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}
	if current > totalPages {
		current = totalPages
	}
	return &publicFrontendPagination{Rows: total, Current: current, PageSize: pageSize, TotalPages: totalPages, IndexHref: publicFrontendPageHref(r, 1), PreHref: publicFrontendPageHref(r, publicFrontendClampPageTarget(current-1, totalPages)), NextHref: publicFrontendPageHref(r, publicFrontendClampPageTarget(current+1, totalPages)), LastHref: publicFrontendPageHref(r, totalPages), NumBar: publicFrontendNumBar(r, current, totalPages)}
}

// publicFrontendNumBar builds the numbered-page HTML fragment for imported templates.
func publicFrontendNumBar(r *ghttp.Request, current int, totalPages int) template.HTML {
	if totalPages <= 0 {
		return ""
	}
	window := 2
	start := current - window
	if start < 1 {
		start = 1
	}
	end := current + window
	if end > totalPages {
		end = totalPages
	}
	var builder strings.Builder
	for page := start; page <= end; page++ {
		if page == current {
			builder.WriteString(`<span class="page-num-current">`)
			builder.WriteString(strconv.Itoa(page))
			builder.WriteString(`</span>`)
			continue
		}
		builder.WriteString(`<a href="`)
		builder.WriteString(template.HTMLEscapeString(publicFrontendPageHref(r, page)))
		builder.WriteString(`" class="page-num">`)
		builder.WriteString(strconv.Itoa(page))
		builder.WriteString(`</a>`)
	}
	return template.HTML(builder.String())
}

// publicFrontendPageHref builds a page link while preserving request filters.
func publicFrontendPageHref(r *ghttp.Request, page int) string {
	if page < 1 {
		page = 1
	}
	values := url.Values{}
	basePath := "/cms-site"
	if r != nil && r.URL != nil {
		if cleaned := publicFrontendNormalizeRequestURLPath(r.URL.Path); cleaned != "" {
			basePath = cleaned
		}
		for key, items := range r.URL.Query() {
			for _, item := range items {
				values.Add(key, item)
			}
		}
	}
	if page <= 1 {
		values.Del("page")
		values.Del("pageNum")
	} else {
		values.Set("page", strconv.Itoa(page))
		values.Del("pageNum")
	}
	encoded := values.Encode()
	if encoded == "" {
		return basePath
	}
	return basePath + "?" + encoded
}

// publicFrontendRequestPath extracts the CMS public route path from a request.
func publicFrontendRequestPath(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	if value := strings.TrimSpace(r.GetRouter("path").String()); value != "" {
		return publicFrontendNormalizeCategoryPath(value)
	}
	if r.URL == nil {
		return ""
	}
	return publicFrontendNormalizeCategoryPath(strings.TrimPrefix(r.URL.Path, "/cms-site"))
}

// publicFrontendCategoryPathHref converts a category path into a public link.
func publicFrontendCategoryPathHref(value string) string {
	normalized := publicFrontendNormalizeCategoryPath(value)
	if normalized == "" {
		return ""
	}
	return "/cms-site/" + normalized + "/"
}

// publicFrontendNormalizeCategoryPath sanitizes category paths used in public URLs.
func publicFrontendNormalizeCategoryPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if trimmed == "/cms-site" {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "/cms-site/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return ""
	}
	cleaned := path.Clean("/" + trimmed)
	cleaned = strings.Trim(cleaned, "/")
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}

// publicFrontendNormalizeRequestURLPath normalizes a request URL path for route matching.
func publicFrontendNormalizeRequestURLPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !strings.HasPrefix(trimmed, "/cms-site") {
		return ""
	}
	cleaned := path.Clean("/" + strings.TrimPrefix(trimmed, "/"))
	if cleaned == "/" || cleaned == "." {
		return "/cms-site"
	}
	if !strings.HasPrefix(cleaned, "/cms-site") {
		return "/cms-site"
	}
	if strings.HasSuffix(trimmed, "/") && cleaned != "/cms-site" {
		return cleaned + "/"
	}
	return cleaned
}

// publicFrontendCategoryListTemplate resolves the list template for a category.
func publicFrontendCategoryListTemplate(category *publicFrontendCategory) string {
	if category == nil {
		return publicFrontendListName
	}
	if category.Type == cmssvc.CategoryTypeSingle {
		return publicFrontendCategoryContentTemplate(category, publicFrontendSingleName)
	}
	return publicFrontendTemplateFileName(category.ListTemplate, publicFrontendListName)
}

// publicFrontendCategoryContentTemplate resolves the content template for a category.
func publicFrontendCategoryContentTemplate(category *publicFrontendCategory, fallback string) string {
	if category == nil {
		return fallback
	}
	return publicFrontendTemplateFileName(category.ContentTemplate, fallback)
}

// publicFrontendTemplateFileName validates a requested template file name.
func publicFrontendTemplateFileName(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	name := path.Base(trimmed)
	if name == "." || name == "/" || strings.Contains(name, "..") {
		return fallback
	}
	if !strings.HasSuffix(name, ".html") {
		name += ".html"
	}
	if _, ok := publicFrontendPageTemplates[name]; !ok {
		return fallback
	}
	if !publicFrontendTemplateFileExists(name) {
		return fallback
	}
	return name
}

// publicFrontendTemplateFileExists reports whether an embedded public template exists.
func publicFrontendTemplateFileExists(fileName string) bool {
	cleaned := path.Base(strings.TrimSpace(fileName))
	if cleaned == "" || cleaned == "." || strings.Contains(cleaned, "..") {
		return false
	}
	_, err := cmsplugin.EmbeddedFiles.ReadFile(path.Join("public/templates", cleaned))
	return err == nil
}

// publicFrontendClampedPage clamps a requested page number to available results.
func publicFrontendClampedPage(page int, total int, pageSize int) int {
	if pageSize <= 0 {
		pageSize = publicFrontendPageSize
	}
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}
	return publicFrontendClampPageTarget(page, totalPages)
}

// publicFrontendClampPageTarget returns the nearest valid page number.
func publicFrontendClampPageTarget(page int, totalPages int) int {
	if page < 1 {
		return 1
	}
	if totalPages > 0 && page > totalPages {
		return totalPages
	}
	return page
}

// mapPublicFrontendSlides converts service slides into public template models.
func mapPublicFrontendSlides(items []*cmssvc.SlideItem) []*publicFrontendSlide {
	slides := make([]*publicFrontendSlide, 0, len(items))
	for index, item := range items {
		if item == nil {
			continue
		}
		slides = append(slides, &publicFrontendSlide{Index: index + 1, Title: item.Title, Image: publicFrontendNormalizeAssetPath(item.Image), Link: publicFrontendNormalizeAssetPath(item.Link), Subtitle: item.Subtitle, Group: publicFrontendSlideGroup(item.GroupCode)})
	}
	return slides
}

// publicFrontendSlideGroup returns slides belonging to a template group.
func publicFrontendSlideGroup(groupCode string) string {
	if strings.TrimSpace(groupCode) == "" {
		return "1"
	}
	return strings.TrimSpace(groupCode)
}

// firstPublicFrontendSlide returns the first slide used by legacy templates.
func firstPublicFrontendSlide(items []*cmssvc.SlideItem) *publicFrontendSlide {
	if len(items) == 0 || items[0] == nil {
		return &publicFrontendSlide{}
	}
	return &publicFrontendSlide{Index: 1, Title: items[0].Title, Image: publicFrontendNormalizeAssetPath(items[0].Image), Link: publicFrontendNormalizeAssetPath(items[0].Link), Subtitle: items[0].Subtitle, Group: publicFrontendSlideGroup(items[0].GroupCode)}
}

// publicFrontendNormalizeContentHTML prepares trusted article HTML for templates.
func publicFrontendNormalizeContentHTML(content string) string {
	return strings.ReplaceAll(content, publicFrontendStaticPath, "/cms-site/assets/static/")
}

// publicFrontendNormalizeAssetPath maps stored asset paths to public site URLs.
func publicFrontendNormalizeAssetPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, publicFrontendStaticPath) {
		return "/cms-site/assets" + trimmed
	}
	return trimmed
}

// publicFrontendDate formats an optional CMS time for public templates.
func publicFrontendDate(value *gtime.Time) string {
	if value == nil {
		return "未发布"
	}
	return value.Format("Y-m-d")
}

// publicFrontendTimestamp returns a Unix millisecond timestamp for public templates.
func publicFrontendTimestamp(value *gtime.Time) int64 {
	if value == nil {
		return 0
	}
	return value.Timestamp()
}

// publicFrontendTextLength clips text by display width with a configurable suffix.
func publicFrontendTextLength(value string, length int, more string) string {
	if length <= 0 {
		return value
	}
	runes := []rune(value)
	width := 0.0
	cut := 0
	for index, item := range runes {
		width += publicFrontendRuneDisplayWidth(item)
		if width > float64(length) {
			break
		}
		cut = index + 1
	}
	if cut >= len(runes) {
		return value
	}
	return string(runes[:cut]) + more
}

// publicFrontendRuneDisplayWidth estimates terminal-style display width for CJK-aware clipping.
func publicFrontendRuneDisplayWidth(value rune) float64 {
	if value <= 127 {
		return 0.5
	}
	return 1
}

// queryInt64 reads an int64 query parameter from a public request.
func queryInt64(r *ghttp.Request, key string) int64 {
	if r == nil {
		return 0
	}
	value, err := strconv.ParseInt(strings.TrimSpace(r.GetQuery(key).String()), 10, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

// queryInt reads an int query parameter from a public request.
func queryInt(r *ghttp.Request, key string) int {
	if r == nil {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(r.GetQuery(key).String()))
	if err != nil || value < 0 {
		return 0
	}
	return value
}

// normalizePublicFrontendPage applies the default first page for public requests.
func normalizePublicFrontendPage(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

// writePublicFrontendStatus writes a plain HTTP error response and stops routing.
func writePublicFrontendStatus(r *ghttp.Request, status int, message string) {
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.WriteStatus(status, "<!doctype html><html lang=\"zh-CN\"><body><main>"+template.HTMLEscapeString(message)+"</main></body></html>")
	r.ExitAll()
}
