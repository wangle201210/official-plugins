// This file verifies CMS public frontend template compilation behavior.

package cms

import (
	"bytes"
	"html/template"
	"strconv"
	"strings"
	"testing"

	cmsplugin "lina-plugin-cms"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// TestPublicFrontendListLimitLimitsRenderedItems verifies cms:list limit controls
// rendered article items instead of falling back to the default page size.
func TestPublicFrontendListLimitLimitsRenderedItems(t *testing.T) {
	const source = `{cms:list limit=5 order=date}<span data-testid="item">[list:title]</span>{/cms:list}`

	compiled := compilePublicFrontendTemplate(source, publicFrontendRootScope)
	tpl, err := template.New("case").Funcs(template.FuncMap{
		"cmsLimit":         publicFrontendLimit,
		"cmsOrderArticles": publicFrontendOrderArticles,
	}).Parse(compiled)
	if err != nil {
		t.Fatalf("parse compiled public frontend template: %v", err)
	}

	articles := make([]*publicFrontendArticle, 0, 6)
	for index := 1; index <= 6; index++ {
		articles = append(articles, &publicFrontendArticle{Title: "Article " + strconv.Itoa(index)})
	}

	var buffer bytes.Buffer
	if err = tpl.Execute(&buffer, &publicFrontendView{Articles: articles}); err != nil {
		t.Fatalf("execute compiled public frontend template: %v", err)
	}
	html := buffer.String()
	if count := strings.Count(html, `data-testid="item"`); count != 5 {
		t.Fatalf("expected 5 rendered articles, got %d in %s", count, html)
	}
	if strings.Contains(html, "Article 6") {
		t.Fatalf("expected article beyond limit to be hidden, got %s", html)
	}
}

// TestPublicFrontendSearchLoopCompiles verifies cms:search renders the public
// search result article tags with the same article projection as list loops.
func TestPublicFrontendSearchLoopCompiles(t *testing.T) {
	const source = `{cms:search limit=2 order=date}<a href="[search:link]" data-testid="item">[search:title]<em>[search:preview]</em></a>{/cms:search}`

	compiled := compilePublicFrontendTemplate(source, publicFrontendRootScope)
	tpl, err := template.New("case").Funcs(template.FuncMap{
		"cmsLimit":         publicFrontendLimit,
		"cmsOrderArticles": publicFrontendOrderArticles,
	}).Parse(compiled)
	if err != nil {
		t.Fatalf("parse compiled public frontend search template: %v", err)
	}

	articles := []*publicFrontendArticle{
		{Title: "First", Href: "/cms-site?article=first", SearchPreview: template.HTML(`命中<mark>关键词</mark>`)},
		{Title: "Second", Href: "/cms-site?article=second", SearchPreview: template.HTML(`第二条预览`)},
		{Title: "Third", Href: "/cms-site?article=third"},
	}
	var buffer bytes.Buffer
	if err = tpl.Execute(&buffer, &publicFrontendView{Articles: articles}); err != nil {
		t.Fatalf("execute compiled public frontend search template: %v", err)
	}
	html := buffer.String()
	if count := strings.Count(html, `data-testid="item"`); count != 2 {
		t.Fatalf("expected 2 rendered search results, got %d in %s", count, html)
	}
	if strings.Contains(html, "Third") {
		t.Fatalf("expected search result beyond limit to be hidden, got %s", html)
	}
	if !strings.Contains(html, `<mark>关键词</mark>`) {
		t.Fatalf("expected search preview markup to render, got %s", html)
	}
}

// TestPublicFrontendCategoryTagsCompile verifies category semantics use
// category-named tags throughout current-category rendering.
func TestPublicFrontendCategoryTagsCompile(t *testing.T) {
	const source = `{cms:category code={category:topcode}}<strong>[category:name]</strong>{/cms:category}<a href="{category:link}">{category:name}</a>{cms:search limit=1}<span>[search:category]</span>{/cms:search}`

	compiled := compilePublicFrontendTemplate(source, publicFrontendRootScope)
	tpl, err := template.New("case").Funcs(template.FuncMap{
		"cmsCategoryByCode":   publicFrontendCategoryByCode,
		"cmsLimit":            publicFrontendLimit,
		"cmsOrderArticles":    publicFrontendOrderArticles,
		"cmsRootCategory":     publicFrontendRootCategory,
		"cmsCategoryArticles": publicFrontendCategoryArticles,
	}).Parse(compiled)
	if err != nil {
		t.Fatalf("parse compiled public frontend category template: %v", err)
	}

	root := &publicFrontendCategory{Code: "root", Href: "/cms-site?category=root", Name: "根栏目"}
	child := &publicFrontendCategory{Code: "child", Href: "/cms-site?category=child", Name: "子栏目", Parent: root}
	root.Children = []*publicFrontendCategory{child}
	view := &publicFrontendView{
		Categories:      []*publicFrontendCategory{root, child},
		CurrentCategory: child,
		Articles:        []*publicFrontendArticle{{CategoryName: "新闻动态"}},
	}

	var buffer bytes.Buffer
	if err = tpl.Execute(&buffer, view); err != nil {
		t.Fatalf("execute compiled public frontend category template: %v", err)
	}
	html := buffer.String()
	if !strings.Contains(html, "<strong>根栏目</strong>") {
		t.Fatalf("expected category loop to render root category name, got %s", html)
	}
	if !strings.Contains(html, `<a href="/cms-site?category=child">子栏目</a>`) {
		t.Fatalf("expected current category tag output, got %s", html)
	}
	if !strings.Contains(html, "<span>新闻动态</span>") {
		t.Fatalf("expected article category name output, got %s", html)
	}
}

// TestPublicFrontendSearchActionTargetsSearchPage verifies public templates
// submit keyword searches to the dedicated search result page.
func TestPublicFrontendSearchActionTargetsSearchPage(t *testing.T) {
	compiled := compilePublicFrontendTemplate(`<form action="{search:action}"></form>`, publicFrontendRootScope)
	if !strings.Contains(compiled, `action="/cms-site/search"`) {
		t.Fatalf("expected search action to target /cms-site/search, got %s", compiled)
	}
}

// TestPublicFrontendSearchPageTitle verifies search pages expose a useful
// document title with and without a keyword.
func TestPublicFrontendSearchPageTitle(t *testing.T) {
	if got, want := publicFrontendSearchPageTitle("算力"), "算力-搜索结果"; got != want {
		t.Fatalf("expected keyword search page title %q, got %q", want, got)
	}
	if got, want := publicFrontendSearchPageTitle("  "), "搜索结果"; got != want {
		t.Fatalf("expected empty search page title %q, got %q", want, got)
	}
}

// TestPublicFrontendEmbeddedTemplatesParse verifies embedded CMS public
// templates compile after template files add or change tags.
func TestPublicFrontendEmbeddedTemplatesParse(t *testing.T) {
	tpl, err := publicFrontendTemplate()
	if err != nil {
		t.Fatalf("parse embedded public frontend templates: %v", err)
	}
	if tpl.Lookup(publicFrontendSearchName) == nil {
		t.Fatalf("expected embedded public frontend template %q to be registered", publicFrontendSearchName)
	}
}

// TestPublicFrontendTemplatesUseOriginalAssets verifies public templates keep
// the original CMS site asset contract.
func TestPublicFrontendTemplatesUseOriginalAssets(t *testing.T) {
	partials, err := cmsplugin.EmbeddedFiles.ReadFile("public/templates/partials.html")
	if err != nil {
		t.Fatalf("read embedded public partial template: %v", err)
	}
	partialSource := string(partials)
	for _, expectedPath := range []string{
		"{site:assets}/css/yx.css",
		"{site:assets}/js/jquery-1.12.4.min.js",
		"{site:assets}/js/yx.js",
	} {
		if !strings.Contains(partialSource, expectedPath) {
			t.Fatalf("expected public partial template to reference %q", expectedPath)
		}
	}
	if strings.Contains(partialSource, `{site:assets}/cms-site.css`) {
		t.Fatalf("expected public partial template to avoid replacement stylesheet")
	}

	for _, file := range []string{
		"public/templates/detail.html",
		"public/templates/list-card.html",
		"public/templates/list.html",
		"public/templates/message.html",
		"public/templates/search.html",
		"public/templates/single.html",
	} {
		content, readErr := cmsplugin.EmbeddedFiles.ReadFile(file)
		if readErr != nil {
			t.Fatalf("read embedded public template %s: %v", file, readErr)
		}
		if !strings.Contains(string(content), `{site:assets}/css/yx-page.css`) {
			t.Fatalf("expected public template %s to reference page stylesheet", file)
		}
	}
}

// TestPublicFrontendOriginalAssetsExist verifies the original CMS public assets
// are embedded and available to the public asset handler.
func TestPublicFrontendOriginalAssetsExist(t *testing.T) {
	for _, file := range []string{
		"public/assets/css/yx.css",
		"public/assets/css/yx-page.css",
		"public/assets/js/jquery-1.12.4.min.js",
		"public/assets/js/yx.js",
		"public/assets/static/logo.svg",
		"public/assets/static/wechat.jpg",
	} {
		content, err := cmsplugin.EmbeddedFiles.ReadFile(file)
		if err != nil {
			t.Fatalf("read embedded public asset %s: %v", file, err)
		}
		if len(content) == 0 {
			t.Fatalf("expected embedded public asset %s to be non-empty", file)
		}
	}
}

// TestPublicFrontendSearchPreviewBuildsSafeHighlightedExcerpt verifies public
// search snippets are plain-text, clipped around the body hit, and highlighted.
func TestPublicFrontendSearchPreviewBuildsSafeHighlightedExcerpt(t *testing.T) {
	item := &cmssvc.ArticleItem{
		CmsArticle: &entitymodel.CmsArticle{
			Title:   "产业动态",
			Summary: "不包含目标词的摘要",
			Content: `<p>这一段正文包含算力网络建设的阶段成果。</p><script>alert(1)</script>`,
		},
	}

	preview := string(publicFrontendSearchPreview(item, "算力网络"))
	if !strings.Contains(preview, `<mark>算力网络</mark>`) {
		t.Fatalf("expected highlighted keyword in search preview, got %s", preview)
	}
	if strings.Contains(preview, "<script") || strings.Contains(preview, "alert") {
		t.Fatalf("expected script content to be stripped from search preview, got %s", preview)
	}
	if strings.Contains(preview, "<p>") {
		t.Fatalf("expected HTML tags to be stripped from search preview, got %s", preview)
	}
}

// TestPublicFrontendStaticAssetPathNormalization verifies seeded static media
// paths resolve through the CMS public asset URL space.
func TestPublicFrontendStaticAssetPathNormalization(t *testing.T) {
	if got, want := publicFrontendNormalizeAssetPath("/static/logo.svg"), "/cms-site/assets/static/logo.svg"; got != want {
		t.Fatalf("expected normalized static path %q, got %q", want, got)
	}
	if got, want := publicFrontendNormalizeContentHTML(`<img src="/static/wechat.jpg">`), `<img src="/cms-site/assets/static/wechat.jpg">`; got != want {
		t.Fatalf("expected normalized static content %q, got %q", want, got)
	}
}
