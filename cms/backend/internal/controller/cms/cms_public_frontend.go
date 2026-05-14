// This file implements the CMS public HTML/CSS frontend handlers.

package cms

import (
	"bytes"
	"context"
	"html"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	cmsplugin "lina-plugin-cms"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

const (
	publicFrontendTemplateGlob = "public/templates/*.html"
	publicFrontendStylePath    = "public/cms-site.css"
	publicFrontendAssetPrefix  = "public/assets"
	publicFrontendStaticPath   = "/static/"
	publicFrontendIndexName    = "index.html"
	publicFrontendListName     = "list.html"
	publicFrontendSearchName   = "search.html"
	publicFrontendSingleName   = "single.html"
	publicFrontendDetailName   = "detail.html"
	publicFrontendMessageName  = "message.html"
	publicFrontendPageSize     = 12
	publicFrontendMaxLoopSize  = 100
	publicFrontendPreviewRunes = 96
	publicFrontendPreviewLead  = 36
	publicFrontendTextMore     = "···"
)

var (
	publicFrontendIncludePattern    = regexp.MustCompile(`\{include\s+file=([^}]+)\}`)
	publicFrontendLoopStartPattern  = regexp.MustCompile(`\{cms:(nav|children|grandchildren|list|search|slide|link|category)((?:[^{}]|\{(?:category|article|nav|child|grandchild|list|search):[^}]+\})*)\}`)
	publicFrontendIfStartPattern    = regexp.MustCompile(`\{cms:if\(([^)]*)\)\}`)
	publicFrontendScriptPattern     = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	publicFrontendStylePattern      = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	publicFrontendHTMLTagPattern    = regexp.MustCompile(`(?s)<[^>]+>`)
	publicFrontendWhitespacePattern = regexp.MustCompile(`\s+`)
)

var publicFrontendPageTemplates = map[string]struct{}{
	publicFrontendIndexName:   {},
	publicFrontendListName:    {},
	"list-card.html":          {},
	publicFrontendSearchName:  {},
	publicFrontendSingleName:  {},
	publicFrontendDetailName:  {},
	publicFrontendMessageName: {},
}

var publicFrontendTemplateCache = struct {
	once sync.Once
	tpl  *template.Template
	err  error
}{}

// publicFrontendTemplateScope identifies the template tag replacement context.
type publicFrontendTemplateScope string

const (
	publicFrontendRootScope       publicFrontendTemplateScope = "root"
	publicFrontendNavScope        publicFrontendTemplateScope = "nav"
	publicFrontendChildNavScope   publicFrontendTemplateScope = "child"
	publicFrontendGrandchildScope publicFrontendTemplateScope = "grandchild"
	publicFrontendListScope       publicFrontendTemplateScope = "list"
	publicFrontendSearchScope     publicFrontendTemplateScope = "search"
	publicFrontendSlideScope      publicFrontendTemplateScope = "slide"
	publicFrontendLinkScope       publicFrontendTemplateScope = "link"
	publicFrontendCategoryScope   publicFrontendTemplateScope = "category"
)

// publicFrontendLoopAttrs contains parsed CMS loop tag attributes.
type publicFrontendLoopAttrs struct {
	Limit    int                       // Limit controls the number of items rendered by one loop.
	Code     string                    // Code filters a content loop by category code.
	Parent   string                    // Parent filters a navigation loop by parent category code.
	Group    string                    // Group filters grouped slides and friendly links.
	Order    cmssvc.PublicArticleOrder // Order captures the template-requested order strategy.
	ParentID int64                     // ParentID is the numeric form of Parent when present.
}

// publicFrontendTextAttrs contains parsed text tag modifiers.
type publicFrontendTextAttrs struct {
	Length int    // Length truncates by Chinese display-width units.
	More   string // More is the suffix appended after truncated text.
}

// publicFrontendView is the data model rendered into the public HTML template.
type publicFrontendView struct {
	Site              *cmssvc.SiteItem
	Categories        []*publicFrontendCategory
	NavCategories     []*publicFrontendCategory
	Articles          []*publicFrontendArticle
	CurrentCategory   *publicFrontendCategory
	CurrentArticle    *publicFrontendArticle
	PrimarySlide      *publicFrontendSlide
	Slides            []*publicFrontendSlide
	Links             []*publicFrontendLink
	Pagination        *publicFrontendPagination
	TemplateName      string
	PageTitle         string
	Keyword           string
	FirstCategoryHref string
	CompanyWeixin     string
	PreviousArticle   *publicFrontendArticle
	NextArticle       *publicFrontendArticle
	Submitted         bool
	InvalidMessage    bool
	MessageError      bool
	Year              int
}

// publicFrontendCategory is a category projection for navigation and sidebars.
type publicFrontendCategory struct {
	Id              int64
	Code            string
	Name            string
	Path            string
	Href            string
	Depth           int
	Active          bool
	External        bool
	Type            int
	ListTemplate    string
	ContentTemplate string
	Title           string
	Keywords        string
	Description     string
	Parent          *publicFrontendCategory
	Children        []*publicFrontendCategory
}

// publicFrontendArticle is an article projection prepared for safe templates.
type publicFrontendArticle struct {
	Id              int64
	CategoryId      int64
	Index           int
	Title           string
	Subtitle        string
	Summary         string
	Cover           string
	Author          string
	Source          string
	Keywords        string
	Description     string
	CategoryName    string
	PublishedAt     string
	PublishedAtUnix int64
	Href            string
	Views           int64
	Sort            int
	IsTop           int
	IsRecommend     int
	SearchPreview   template.HTML
	ContentHTML     template.HTML
}

// publicFrontendPagination is the public list pagination projection.
type publicFrontendPagination struct {
	Rows       int
	Current    int
	PageSize   int
	TotalPages int
	IndexHref  string
	PreHref    string
	NextHref   string
	LastHref   string
	NumBar     template.HTML
}

// publicFrontendSlide is the public hero slide projection.
type publicFrontendSlide struct {
	Index    int
	Title    string
	Image    string
	Link     string
	Subtitle string
	Group    string
}

// publicFrontendLink is the public friendly-link projection.
type publicFrontendLink struct {
	Name  string
	Url   string
	Logo  string
	Group string
}

// PublicFrontendPage renders the CMS public site as a standalone HTML page.
func (c *ControllerV1) PublicFrontendPage(r *ghttp.Request) {
	if r == nil {
		return
	}
	view, err := c.buildPublicFrontendView(r.GetCtx(), r)
	if err != nil {
		if bizerr.Is(err, cmssvc.CodePublicContentNotFound) {
			writePublicFrontendStatus(r, http.StatusNotFound, "CMS public content was not found.")
			return
		}
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS public site is temporarily unavailable.")
		return
	}
	tpl, err := publicFrontendTemplate()
	if err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS public template is temporarily unavailable.")
		return
	}

	var buffer bytes.Buffer
	if err = tpl.ExecuteTemplate(&buffer, view.TemplateName+".html", view); err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS public page could not be rendered.")
		return
	}
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.Write(buffer.Bytes())
	r.ExitAll()
}

// PublicFrontendMessagePage renders the standalone CMS public message template.
func (c *ControllerV1) PublicFrontendMessagePage(r *ghttp.Request) {
	if r == nil {
		return
	}
	view, err := c.buildPublicFrontendBaseView(r.GetCtx(), r, publicFrontendMessageName, 0)
	if err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS public message page is temporarily unavailable.")
		return
	}
	view.TemplateName = strings.TrimSuffix(publicFrontendMessageName, ".html")
	view.PageTitle = "提交留言"
	tpl, err := publicFrontendTemplate()
	if err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS public template is temporarily unavailable.")
		return
	}
	var buffer bytes.Buffer
	if err = tpl.ExecuteTemplate(&buffer, publicFrontendMessageName, view); err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS public message page could not be rendered.")
		return
	}
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.Write(buffer.Bytes())
	r.ExitAll()
}

// PublicFrontendStyle serves the CMS public site CSS asset.
func (c *ControllerV1) PublicFrontendStyle(r *ghttp.Request) {
	if r == nil {
		return
	}
	content, err := cmsplugin.EmbeddedFiles.ReadFile(publicFrontendStylePath)
	if err != nil {
		writePublicFrontendStatus(r, http.StatusNotFound, "CMS public stylesheet was not found.")
		return
	}
	r.Response.Header().Set("Cache-Control", "public, max-age=60")
	r.Response.Header().Set("Content-Type", "text/css; charset=utf-8")
	r.Response.Write(content)
	r.ExitAll()
}

// PublicFrontendAsset serves embedded CMS public template assets.
func (c *ControllerV1) PublicFrontendAsset(r *ghttp.Request) {
	if r == nil {
		return
	}
	fileName := path.Clean(strings.TrimPrefix(r.GetRouter("file").String(), "/"))
	c.servePublicFrontendAsset(r, fileName)
}

// PublicFrontendStaticAsset serves restored public static asset paths used in content HTML.
func (c *ControllerV1) PublicFrontendStaticAsset(r *ghttp.Request) {
	if r == nil {
		return
	}
	fileName := path.Join("static", path.Clean(strings.TrimPrefix(r.GetRouter("file").String(), "/")))
	c.servePublicFrontendAsset(r, fileName)
}

// servePublicFrontendAsset writes one embedded public asset response.
func (c *ControllerV1) servePublicFrontendAsset(r *ghttp.Request, fileName string) {
	cleaned := path.Clean(strings.TrimPrefix(fileName, "/"))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		writePublicFrontendStatus(r, http.StatusNotFound, "CMS public asset was not found.")
		return
	}
	assetPath := path.Join(publicFrontendAssetPrefix, cleaned)
	content, err := cmsplugin.EmbeddedFiles.ReadFile(assetPath)
	if err != nil {
		writePublicFrontendStatus(r, http.StatusNotFound, "CMS public asset was not found.")
		return
	}
	r.Response.Header().Set("Cache-Control", "public, max-age=300")
	r.Response.ServeContent(path.Base(cleaned), time.Time{}, bytes.NewReader(content))
	r.ExitAll()
}

// PublicFrontendMessage handles the public HTML visitor-message form.
func (c *ControllerV1) PublicFrontendMessage(r *ghttp.Request) {
	if r == nil {
		return
	}
	name := strings.TrimSpace(r.GetForm("name").String())
	email := strings.TrimSpace(r.GetForm("email").String())
	mobile := strings.TrimSpace(r.GetForm("mobile").String())
	content := strings.TrimSpace(r.GetForm("content").String())
	if name == "" || content == "" {
		r.Response.RedirectTo("/cms-site/message?message=invalid", http.StatusSeeOther)
		r.ExitAll()
		return
	}
	if _, err := c.cmsSvc.CreatePublicMessage(r.GetCtx(), cmssvc.PublicMessageCreateInput{
		Name:      name,
		Mobile:    mobile,
		Email:     email,
		Content:   content,
		UserIp:    r.GetClientIp(),
		UserAgent: r.Header.Get("User-Agent"),
	}); err != nil {
		r.Response.RedirectTo("/cms-site/message?message=error", http.StatusSeeOther)
		r.ExitAll()
		return
	}
	r.Response.RedirectTo("/cms-site/message?message=submitted", http.StatusSeeOther)
	r.ExitAll()
}

// buildPublicFrontendView loads CMS data and maps it to the public template.
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
		articlePage, err = c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{
			PageNum:                 pageNum,
			PageSize:                listPageSize,
			CategoryId:              categoryID,
			Keyword:                 keyword,
			Order:                   listAttrs.Order,
			IncludeHiddenCategories: view.CurrentCategory != nil || isSearchPage,
		})
		if err != nil {
			return nil, err
		}
	}
	if categoryID > 0 {
		clampedPage := publicFrontendClampedPage(pageNum, articlePage.Total, listPageSize)
		if clampedPage != pageNum {
			pageNum = clampedPage
			articlePage, err = c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{
				PageNum:                 pageNum,
				PageSize:                listPageSize,
				CategoryId:              categoryID,
				Keyword:                 keyword,
				Order:                   listAttrs.Order,
				IncludeHiddenCategories: true,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	slides, err := c.cmsSvc.ListPublicSlides(ctx)
	if err != nil {
		return nil, err
	}
	allArticlePage, err := c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{
		PageNum:                 1,
		PageSize:                publicFrontendMaxLoopSize,
		IncludeHiddenCategories: true,
	})
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
				singlePage, err := c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{
					PageNum:                 1,
					PageSize:                1,
					CategoryId:              view.CurrentCategory.Id,
					Order:                   listAttrs.Order,
					IncludeHiddenCategories: true,
				})
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

// publicFrontendAdjacentArticles returns the neighboring articles in the same
// category and public ordering used by list pages.
func (c *ControllerV1) publicFrontendAdjacentArticles(
	ctx context.Context,
	current *cmssvc.ArticleItem,
) (*publicFrontendArticle, *publicFrontendArticle, error) {
	if current == nil || current.CmsArticle == nil {
		return nil, nil, nil
	}
	page, err := c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{
		PageNum:                 1,
		PageSize:                publicFrontendMaxLoopSize,
		CategoryId:              current.CategoryId,
		Order:                   cmssvc.PublicArticleOrderDate,
		IncludeHiddenCategories: true,
	})
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

// buildPublicFrontendBaseView loads common data shared by all public templates.
func (c *ControllerV1) buildPublicFrontendBaseView(
	ctx context.Context,
	r *ghttp.Request,
	templateFileName string,
	activeCategoryID int64,
) (*publicFrontendView, error) {
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
	return &publicFrontendView{
		Site:              site,
		Categories:        flatCategories,
		NavCategories:     navCategories,
		CurrentCategory:   findPublicFrontendCategory(flatCategories, activeCategoryID),
		Links:             mapPublicFrontendLinks(links),
		TemplateName:      strings.TrimSuffix(templateFileName, ".html"),
		Keyword:           strings.TrimSpace(r.GetQuery("keyword").String()),
		FirstCategoryHref: firstPublicFrontendCategoryHref(navCategories),
		CompanyWeixin:     site.Weixin,
		Submitted:         messageState == "submitted",
		InvalidMessage:    messageState == "invalid",
		MessageError:      messageState == "error",
		Year:              time.Now().Year(),
	}, nil
}

// publicFrontendSearchPageTitle returns the page title for public search
// results.
func publicFrontendSearchPageTitle(keyword string) string {
	if strings.TrimSpace(keyword) == "" {
		return "搜索结果"
	}
	return strings.TrimSpace(keyword) + "-搜索结果"
}

// publicFrontendTemplate parses and returns the embedded public HTML template.
func publicFrontendTemplate() (*template.Template, error) {
	publicFrontendTemplateCache.once.Do(func() {
		matches, err := fs.Glob(cmsplugin.EmbeddedFiles, publicFrontendTemplateGlob)
		if err != nil {
			publicFrontendTemplateCache.err = err
			return
		}
		tpl := template.New("")
		tpl = tpl.Funcs(template.FuncMap{
			"cmsCategoryArticles": publicFrontendCategoryArticles,
			"cmsCategoryByCode":   publicFrontendCategoryByCode,
			"cmsCategoryChildren": publicFrontendCategoryChildren,
			"cmsGroupedLinks":     publicFrontendGroupedLinks,
			"cmsGroupedSlides":    publicFrontendGroupedSlides,
			"cmsLimit":            publicFrontendLimit,
			"cmsOrderArticles":    publicFrontendOrderArticles,
			"cmsRootCategory":     publicFrontendRootCategory,
			"cmsTextLength":       publicFrontendTextLength,
			"safeHTML":            func(value template.HTML) template.HTML { return value },
		})
		for _, match := range matches {
			content, err := cmsplugin.EmbeddedFiles.ReadFile(match)
			if err != nil {
				publicFrontendTemplateCache.err = err
				return
			}
			compiled := compilePublicFrontendTemplate(string(content), publicFrontendRootScope)
			if _, err = tpl.New(path.Base(match)).Parse(compiled); err != nil {
				publicFrontendTemplateCache.err = err
				return
			}
		}
		publicFrontendTemplateCache.tpl = tpl
	})
	return publicFrontendTemplateCache.tpl, publicFrontendTemplateCache.err
}

// compilePublicFrontendTemplate translates CMS tags into Go html/template syntax.
func compilePublicFrontendTemplate(content string, scope publicFrontendTemplateScope) string {
	compiled := replacePublicFrontendIncludes(content)
	compiled = replacePublicFrontendLoops(compiled)
	compiled = replacePublicFrontendIfs(compiled, scope)
	compiled = replacePublicFrontendRootTags(compiled)
	compiled = replacePublicFrontendScopedTags(compiled, scope)
	return compiled
}

// replacePublicFrontendLoops compiles block loop tags from inside out.
func replacePublicFrontendLoops(content string) string {
	for {
		loc := publicFrontendLoopStartPattern.FindStringSubmatchIndex(content)
		if loc == nil {
			return content
		}
		name := content[loc[2]:loc[3]]
		attrs := publicFrontendParseLoopAttrs(content[loc[4]:loc[5]])
		closeTag := "{/cms:" + name + "}"
		closeIndex := strings.Index(content[loc[1]:], closeTag)
		if closeIndex < 0 {
			return content
		}
		bodyStart := loc[1]
		bodyEnd := loc[1] + closeIndex
		loopScope := publicFrontendLoopScope(name)
		body := compilePublicFrontendTemplate(content[bodyStart:bodyEnd], loopScope)
		compiled := publicFrontendLoopTemplate(loopScope, attrs, body)
		content = content[:loc[0]] + compiled + content[bodyEnd+len(closeTag):]
	}
}

// replacePublicFrontendIfs compiles conditional block tags from inside out.
func replacePublicFrontendIfs(content string, scope publicFrontendTemplateScope) string {
	for {
		matches := publicFrontendIfStartPattern.FindAllStringSubmatchIndex(content, -1)
		if len(matches) == 0 {
			return content
		}
		loc := matches[len(matches)-1]
		expression := content[loc[2]:loc[3]]
		closeTag := "{/cms:if}"
		closeIndex := strings.Index(content[loc[1]:], closeTag)
		if closeIndex < 0 {
			return content
		}
		bodyStart := loc[1]
		bodyEnd := loc[1] + closeIndex
		condition := publicFrontendIfCondition(expression, scope)
		ifTrue, ifFalse, hasElse := strings.Cut(content[bodyStart:bodyEnd], "{else}")
		body := compilePublicFrontendTemplate(ifTrue, scope)
		if hasElse {
			body = body + "{{else}}" + compilePublicFrontendTemplate(ifFalse, scope)
		}
		content = content[:loc[0]] + "{{if " + condition + "}}" + body + "{{end}}" + content[bodyEnd+len(closeTag):]
	}
}

// replacePublicFrontendIncludes maps CMS include tags to template calls.
func replacePublicFrontendIncludes(content string) string {
	return publicFrontendIncludePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := publicFrontendIncludePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		fileName := strings.Trim(parts[1], " \"'")
		switch path.Base(fileName) {
		case "head.html":
			return `{{template "head" .}}`
		case "foot.html":
			return `{{template "foot" .}}`
		case "sidebar.html", "categorynav.html":
			return `{{template "sidebar" .}}`
		default:
			name := strings.TrimSuffix(path.Base(fileName), path.Ext(fileName))
			if name == "" {
				return match
			}
			return `{{template "` + name + `" .}}`
		}
	})
}

// publicFrontendLoopScope returns the item scope used inside a CMS loop.
func publicFrontendLoopScope(name string) publicFrontendTemplateScope {
	switch name {
	case "nav":
		return publicFrontendNavScope
	case "children":
		return publicFrontendChildNavScope
	case "grandchildren":
		return publicFrontendGrandchildScope
	case "list":
		return publicFrontendListScope
	case "search":
		return publicFrontendSearchScope
	case "slide":
		return publicFrontendSlideScope
	case "link":
		return publicFrontendLinkScope
	case "category":
		return publicFrontendCategoryScope
	default:
		return publicFrontendRootScope
	}
}

// publicFrontendParseLoopAttrs parses CMS loop attributes such as limit=5.
func publicFrontendParseLoopAttrs(raw string) publicFrontendLoopAttrs {
	attrs := publicFrontendLoopAttrs{}
	for _, field := range strings.Fields(strings.TrimSpace(raw)) {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch key {
		case "limit":
			limit, err := strconv.Atoi(value)
			if err == nil && limit > 0 {
				attrs.Limit = publicFrontendClampLoopLimit(limit)
			}
		case "code":
			attrs.Code = strings.TrimSpace(value)
		case "parent":
			attrs.Parent = strings.TrimSpace(value)
			if parentID, err := strconv.ParseInt(attrs.Parent, 10, 64); err == nil && parentID > 0 {
				attrs.ParentID = parentID
			}
		case "group":
			attrs.Group = strings.TrimSpace(value)
		case "order":
			attrs.Order = cmssvc.NormalizePublicArticleOrder(value)
		default:
			continue
		}
	}
	return attrs
}

// publicFrontendParseTextAttrs parses CMS text modifiers such as length=18.
func publicFrontendParseTextAttrs(raw string) publicFrontendTextAttrs {
	attrs := publicFrontendTextAttrs{More: publicFrontendTextMore}
	for _, field := range strings.Fields(strings.TrimSpace(raw)) {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch key {
		case "length":
			length, err := strconv.Atoi(value)
			if err == nil && length > 0 {
				attrs.Length = length
			}
		case "more":
			attrs.More = value
		default:
			continue
		}
	}
	return attrs
}

// publicFrontendClampLoopLimit caps template loop limits to a safe maximum.
func publicFrontendClampLoopLimit(value int) int {
	if value < 1 {
		return 0
	}
	if value > publicFrontendMaxLoopSize {
		return publicFrontendMaxLoopSize
	}
	return value
}

// publicFrontendLoopLimit returns the effective loop item limit.
func publicFrontendLoopLimit(scope publicFrontendTemplateScope, attrs publicFrontendLoopAttrs) int {
	if attrs.Limit > 0 {
		return attrs.Limit
	}
	switch scope {
	case publicFrontendListScope, publicFrontendSearchScope:
		return publicFrontendPageSize
	default:
		return 0
	}
}

// publicFrontendTemplateArticleAttrs returns the attrs declared by the first
// article loop in one template file, matching common CMS list-template behavior.
func publicFrontendTemplateArticleAttrs(
	templateFileName string,
	scope publicFrontendTemplateScope,
) publicFrontendLoopAttrs {
	content, err := cmsplugin.EmbeddedFiles.ReadFile(path.Join("public/templates", templateFileName))
	if err != nil {
		return publicFrontendLoopAttrs{Limit: publicFrontendPageSize}
	}
	matches := publicFrontendLoopStartPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) < 3 || publicFrontendLoopScope(match[1]) != scope {
			continue
		}
		return publicFrontendParseLoopAttrs(match[2])
	}
	return publicFrontendLoopAttrs{Limit: publicFrontendPageSize}
}

// publicFrontendTemplateListAttrs returns the attrs declared by the first list
// loop in one template file, matching common CMS list-template behavior.
func publicFrontendTemplateListAttrs(templateFileName string) publicFrontendLoopAttrs {
	return publicFrontendTemplateArticleAttrs(templateFileName, publicFrontendListScope)
}

// publicFrontendLoopTemplate wraps a compiled CMS loop body.
func publicFrontendLoopTemplate(scope publicFrontendTemplateScope, attrs publicFrontendLoopAttrs, body string) string {
	limit := strconv.Itoa(publicFrontendLoopLimit(scope, attrs))
	switch scope {
	case publicFrontendNavScope:
		if attrs.Parent != "" && attrs.Parent != "0" {
			if attrs.Parent == "{category:topcode}" {
				return "{{range cmsCategoryChildren $ (cmsRootCategory $.CurrentCategory) " + limit + "}}" + body + "{{end}}"
			}
			return "{{range cmsCategoryChildren $ " + publicFrontendCategoryArg(attrs.Parent) + " " + limit + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit .NavCategories " + limit + "}}" + body + "{{end}}"
	case publicFrontendChildNavScope, publicFrontendGrandchildScope:
		return "{{range cmsLimit .Children " + limit + "}}" + body + "{{end}}"
	case publicFrontendListScope, publicFrontendSearchScope:
		if attrs.Code != "" {
			return "{{range cmsCategoryArticles $ " + publicFrontendStringLiteral(attrs.Code) + " " + limit + " " + publicFrontendStringLiteral(string(attrs.Order)) + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit (cmsOrderArticles .Articles " + publicFrontendStringLiteral(string(attrs.Order)) + ") " + limit + "}}" + body + "{{end}}"
	case publicFrontendSlideScope:
		if attrs.Group != "" {
			return "{{range cmsGroupedSlides $.Slides " + publicFrontendStringLiteral(attrs.Group) + " " + limit + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit .Slides " + limit + "}}" + body + "{{end}}"
	case publicFrontendLinkScope:
		if attrs.Group != "" {
			return "{{range cmsGroupedLinks $.Links " + publicFrontendStringLiteral(attrs.Group) + " " + limit + "}}" + body + "{{end}}"
		}
		return "{{range cmsLimit .Links " + limit + "}}" + body + "{{end}}"
	case publicFrontendCategoryScope:
		if attrs.Code != "" {
			if attrs.Code == "{category:topcode}" {
				return "{{with cmsRootCategory $.CurrentCategory}}" + body + "{{end}}"
			}
			if attrs.Code == "{category:code}" {
				return "{{with $.CurrentCategory}}" + body + "{{end}}"
			}
			return "{{with cmsCategoryByCode $ " + publicFrontendStringLiteral(attrs.Code) + "}}" + body + "{{end}}"
		}
		return "{{with .CurrentCategory}}" + body + "{{end}}"
	default:
		return body
	}
}

// publicFrontendStringLiteral safely quotes a template string argument.
func publicFrontendStringLiteral(value string) string {
	return strconv.Quote(strings.TrimSpace(value))
}

// publicFrontendCategoryArg maps one category attribute value to a template expression.
func publicFrontendCategoryArg(value string) string {
	switch strings.TrimSpace(value) {
	case "[nav:code]", "[child:code]", "[grandchild:code]":
		return "."
	case "{category:code}":
		return "$.CurrentCategory"
	default:
		return publicFrontendStringLiteral(value)
	}
}

// publicFrontendLimit returns at most limit items from a slice-like value.
func publicFrontendLimit(items any, limit int) any {
	if items == nil || limit <= 0 {
		return items
	}
	value := reflect.ValueOf(items)
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() <= limit {
			return items
		}
		return value.Slice(0, limit).Interface()
	default:
		return items
	}
}

// publicFrontendIfCondition maps common CMS-style conditions.
func publicFrontendIfCondition(expression string, scope publicFrontendTemplateScope) string {
	expr := strings.TrimSpace(expression)
	switch {
	case strings.Contains(expr, "0=='{category:code}'"),
		strings.Contains(expr, "'0'=='{category:code}'"):
		return "not .CurrentCategory"
	case expr == "{message:submitted}":
		return ".Submitted"
	case expr == "{message:invalid}":
		return ".InvalidMessage"
	case expr == "{message:error}":
		return ".MessageError"
	case expr == "{slide:firstimage}":
		return ".PrimarySlide.Image"
	case expr == "{site:logo}":
		return ".Site.Logo"
	case strings.Contains(expr, "page:total"):
		return "and .Pagination (gt .Pagination.Rows 0)"
	case expr == "{article:image}":
		return "and .CurrentArticle .CurrentArticle.Cover"
	case strings.Contains(expr, "childcount"):
		return ".Children"
	case strings.Contains(expr, "[slide:index]"):
		return "eq .Index 1"
	case expr == "[list:image]":
		return ".Cover"
	case expr == "[search:image]":
		return ".Cover"
	case strings.Contains(expr, "[nav:code]") && strings.Contains(expr, "{category:topcode}"):
		return ".Active"
	case strings.Contains(expr, "[nav:code]") && strings.Contains(expr, "{category:code}"):
		return ".Active"
	case strings.Contains(expr, "[child:code]") && strings.Contains(expr, "{category:code}"):
		return ".Active"
	case strings.Contains(expr, "[grandchild:code]") && strings.Contains(expr, "{category:code}"):
		return ".Active"
	case strings.Contains(expr, "nav:active"),
		strings.Contains(expr, "child:active"),
		strings.Contains(expr, "grandchild:active"),
		strings.Contains(expr, "category:active"):
		return ".Active"
	case strings.Contains(expr, "article:image"):
		return "and .CurrentArticle .CurrentArticle.Cover"
	default:
		return "false"
	}
}

// replacePublicFrontendRootTags maps global site and page tags.
func replacePublicFrontendRootTags(content string) string {
	replacements := []string{
		"{site:path}", "/cms-site",
		"{site:assets}", "/cms-site/assets",
		"{search:action}", "/cms-site/search",
		"{message:action}", "/cms-site/messages",
		"{message:page}", "/cms-site/message",
		"{site:title}", "{{.Site.Name}}",
		"{site:name}", "{{.Site.Name}}",
		"{site:subtitle}", "{{.Site.Slogan}}",
		"{site:slogan}", "{{.Site.Slogan}}",
		"{site:logo}", "{{.Site.Logo}}",
		"{site:keywords}", "{{.Site.Keywords}}",
		"{site:description}", "{{.Site.Description}}",
		"{site:address}", "{{.Site.Address}}",
		"{site:phone}", "{{.Site.Phone}}",
		"{site:email}", "{{.Site.Email}}",
		"{site:contact}", "{{.Site.Contact}}",
		"{site:icp}", "{{.Site.Icp}}",
		"{site:wechat}", "{{.CompanyWeixin}}",
		"{search:keyword}", "{{.Keyword}}",
		"{site:year}", "{{.Year}}",
		"{category:firstlink}", "{{.FirstCategoryHref}}",
		"{slide:firstimage}", "{{.PrimarySlide.Image}}",
		"{slide:firsttitle}", "{{.PrimarySlide.Title}}",
		"{page:title}", "{{if .PageTitle}}{{.PageTitle}}-{{end}}{{.Site.Name}}",
		"{page:keywords}", "{{if .CurrentArticle}}{{.CurrentArticle.Keywords}}{{else if .CurrentCategory}}{{.CurrentCategory.Keywords}}{{else}}{{.Site.Keywords}}{{end}}",
		"{page:description}", "{{if .CurrentArticle}}{{.CurrentArticle.Description}}{{else if .CurrentCategory}}{{.CurrentCategory.Description}}{{else}}{{.Site.Description}}{{end}}",
		"{category:name}", "{{if .CurrentCategory}}{{.CurrentCategory.Name}}{{end}}",
		"{category:link}", "{{if .CurrentCategory}}{{.CurrentCategory.Href}}{{end}}",
		"{category:code}", "{{if .CurrentCategory}}{{.CurrentCategory.Code}}{{end}}",
		"{category:topcode}", "{{with cmsRootCategory .CurrentCategory}}{{.Code}}{{end}}",
		"{category:description}", "{{if .CurrentCategory}}{{.CurrentCategory.Description}}{{end}}",
		"{article:title}", "{{if .CurrentArticle}}{{.CurrentArticle.Title}}{{end}}",
		"{article:subtitle}", "{{if .CurrentArticle}}{{.CurrentArticle.Subtitle}}{{end}}",
		"{article:summary}", "{{if .CurrentArticle}}{{.CurrentArticle.Summary}}{{end}}",
		"{article:image}", "{{if .CurrentArticle}}{{.CurrentArticle.Cover}}{{end}}",
		"{article:author}", "{{if .CurrentArticle}}{{.CurrentArticle.Author}}{{end}}",
		"{article:source}", "{{if .CurrentArticle}}{{.CurrentArticle.Source}}{{end}}",
		"{article:date}", "{{if .CurrentArticle}}{{.CurrentArticle.PublishedAt}}{{end}}",
		"{article:views}", "{{if .CurrentArticle}}{{.CurrentArticle.Views}}{{end}}",
		"{article:content}", "{{if .CurrentArticle}}{{.CurrentArticle.ContentHTML}}{{end}}",
		"{article:previous}", "{{if .PreviousArticle}}<a href=\"{{.PreviousArticle.Href}}\">{{.PreviousArticle.Title}}</a>{{else}}无{{end}}",
		"{article:next}", "{{if .NextArticle}}<a href=\"{{.NextArticle.Href}}\">{{.NextArticle.Title}}</a>{{else}}无{{end}}",
		"{page:total}", "{{if .Pagination}}{{.Pagination.Rows}}{{else}}0{{end}}",
		"{page:first}", "{{if .Pagination}}{{.Pagination.IndexHref}}{{end}}",
		"{page:previous}", "{{if .Pagination}}{{.Pagination.PreHref}}{{end}}",
		"{page:next}", "{{if .Pagination}}{{.Pagination.NextHref}}{{end}}",
		"{page:last}", "{{if .Pagination}}{{.Pagination.LastHref}}{{end}}",
		"{page:numbers}", "{{if .Pagination}}{{.Pagination.NumBar}}{{end}}",
	}
	replaced := strings.NewReplacer(replacements...).Replace(content)
	replaced = regexp.MustCompile(`\{article:date\s+style=[^}]+\}`).ReplaceAllString(replaced, "{{if .CurrentArticle}}{{.CurrentArticle.PublishedAt}}{{end}}")
	replaced = regexp.MustCompile(`\{page:breadcrumb[^}]*\}`).ReplaceAllString(replaced, `<a href="/cms-site">首页</a>{{if .CurrentCategory}}<span class="sep">&gt;</span>{{with cmsRootCategory .CurrentCategory}}<a href="{{.Href}}">{{.Name}}</a>{{end}}{{if ne .CurrentCategory.Id (cmsRootCategory .CurrentCategory).Id}}<span class="sep">&gt;</span><span>{{.CurrentCategory.Name}}</span>{{end}}{{end}}`)
	return replaced
}

// replacePublicFrontendScopedTags maps tags available inside CMS loops.
func replacePublicFrontendScopedTags(content string, scope publicFrontendTemplateScope) string {
	switch scope {
	case publicFrontendNavScope:
		return replacePublicFrontendCategoryTags(content, "nav")
	case publicFrontendChildNavScope:
		return replacePublicFrontendCategoryTags(content, "child")
	case publicFrontendGrandchildScope:
		return replacePublicFrontendCategoryTags(content, "grandchild")
	case publicFrontendListScope:
		return replacePublicFrontendArticleTags(content, "list")
	case publicFrontendSearchScope:
		return replacePublicFrontendArticleTags(content, "search")
	case publicFrontendSlideScope:
		return replacePublicFrontendSlideTags(content)
	case publicFrontendLinkScope:
		return replacePublicFrontendLinkTags(content)
	case publicFrontendCategoryScope:
		return replacePublicFrontendCategoryTags(content, "category")
	default:
		return content
	}
}

// replacePublicFrontendCategoryTags maps nav item tags.
func replacePublicFrontendCategoryTags(content string, prefix string) string {
	replaced := strings.NewReplacer(
		"["+prefix+":link]", "{{.Href}}",
		"["+prefix+":code]", "{{.Code}}",
		"["+prefix+":name]", "{{.Name}}",
		"["+prefix+":title]", "{{.Title}}",
		"["+prefix+":keywords]", "{{.Keywords}}",
		"["+prefix+":description]", "{{.Description}}",
		"["+prefix+":childcount]", "{{len .Children}}",
		"["+prefix+":active]", "{{.Active}}",
	).Replace(content)
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "name", ".Name")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "title", ".Title")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "description", ".Description")
	return replaced
}

// replacePublicFrontendArticleTags maps list item tags.
func replacePublicFrontendArticleTags(content string, prefix string) string {
	replaced := strings.NewReplacer(
		"["+prefix+":id]", "{{.Id}}",
		"["+prefix+":index]", "{{.Index}}",
		"["+prefix+":link]", "{{.Href}}",
		"["+prefix+":title]", "{{.Title}}",
		"["+prefix+":subtitle]", "{{.Subtitle}}",
		"["+prefix+":summary]", "{{.Summary}}",
		"["+prefix+":content]", "{{.Summary}}",
		"["+prefix+":preview]", "{{.SearchPreview}}",
		"["+prefix+":image]", "{{.Cover}}",
		"["+prefix+":date]", "{{.PublishedAt}}",
		"["+prefix+":views]", "{{.Views}}",
		"["+prefix+":category]", "{{.CategoryName}}",
	).Replace(content)
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "title", ".Title")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "subtitle", ".Subtitle")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "summary", ".Summary")
	replaced = replacePublicFrontendTextParamTags(replaced, prefix, "content", ".Summary")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:title\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.Title}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:summary\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.Summary}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:content\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.Summary}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:preview\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.SearchPreview}}")
	replaced = regexp.MustCompile(`\[`+regexp.QuoteMeta(prefix)+`:date\s+[^\]]+\]`).ReplaceAllString(replaced, "{{.PublishedAt}}")
	return replaced
}

// replacePublicFrontendSlideTags maps slide loop tags.
func replacePublicFrontendSlideTags(content string) string {
	replaced := strings.NewReplacer(
		"[slide:index]", "{{.Index}}",
		"[slide:url]", "{{.Link}}",
		"[slide:image]", "{{.Image}}",
	).Replace(content)
	replaced = replacePublicFrontendTextParamTags(replaced, "slide", "title", ".Title")
	replaced = replacePublicFrontendTextParamTags(replaced, "slide", "subtitle", ".Subtitle")
	replaced = strings.NewReplacer(
		"[slide:title]", "{{.Title}}",
		"[slide:subtitle]", "{{.Subtitle}}",
	).Replace(replaced)
	return replaced
}

// replacePublicFrontendLinkTags maps friendly link loop tags.
func replacePublicFrontendLinkTags(content string) string {
	return strings.NewReplacer(
		"[link:url]", "{{.Url}}",
		"[link:name]", "{{.Name}}",
		"[link:logo]", "{{.Logo}}",
	).Replace(content)
}

// replacePublicFrontendTextParamTags maps text tags with length/more modifiers
// to template helper calls.
func replacePublicFrontendTextParamTags(content string, prefix string, name string, expression string) string {
	pattern := regexp.MustCompile(`\[` + regexp.QuoteMeta(prefix) + `:` + regexp.QuoteMeta(name) + `\s+([^\]]+)\]`)
	return pattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		attrs := publicFrontendParseTextAttrs(parts[1])
		if attrs.Length > 0 {
			return "{{cmsTextLength " + expression + " " + strconv.Itoa(attrs.Length) + " " + publicFrontendStringLiteral(attrs.More) + "}}"
		}
		return "{{" + expression + "}}"
	})
}

// buildPublicFrontendCategoryTree returns a hierarchy-preserving category tree.
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

// mapPublicFrontendCategory maps one service category and all its children.
func mapPublicFrontendCategory(
	item *cmssvc.CategoryItem,
	activeID int64,
	depth int,
	parent *publicFrontendCategory,
) *publicFrontendCategory {
	if item == nil || item.CmsCategory == nil {
		return nil
	}
	category := &publicFrontendCategory{
		Id:              item.Id,
		Code:            item.Code,
		Name:            item.Name,
		Path:            item.Path,
		Href:            publicFrontendCategoryHref(item),
		Depth:           depth,
		Active:          item.Id == activeID,
		External:        item.Type == cmssvc.CategoryTypeExternal,
		Type:            item.Type,
		ListTemplate:    item.ListTemplate,
		ContentTemplate: item.ContentTemplate,
		Title:           item.Title,
		Keywords:        item.Keywords,
		Description:     item.Description,
		Parent:          parent,
		Children:        make([]*publicFrontendCategory, 0, len(item.Children)),
	}
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

// flattenPublicFrontendCategoryTree returns a depth-aware list for sidebars.
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

// markPublicFrontendActiveCategory refreshes active flags after category-code routing.
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

// publicFrontendCategoryHref returns the navigation href for one category.
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

// findPublicFrontendCategory finds one flattened category by ID.
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

// findPublicFrontendCategoryByPath finds one flattened category by public path.
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

// findPublicFrontendCategoryByCode finds one flattened category by stable code.
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

// publicFrontendCategoryByCode finds a category in a view by stable code.
func publicFrontendCategoryByCode(view *publicFrontendView, code string) *publicFrontendCategory {
	if view == nil {
		return nil
	}
	return findPublicFrontendCategoryByCode(view.Categories, code)
}

// publicFrontendRootCategory returns the top-level ancestor for one category.
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

// publicFrontendCategoryChildren returns children for a category code or item.
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

// publicFrontendCategoryArticles filters current article projection by category code.
func publicFrontendCategoryArticles(
	view *publicFrontendView,
	code string,
	limit int,
	order string,
) []*publicFrontendArticle {
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

// publicFrontendOrderArticles sorts public articles by a CMS template order.
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

// publicFrontendArticleDefaultLess compares articles by the default public
// ordering used when templates request date order or omit an order.
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

// publicFrontendArticleDateLess compares articles by the date-first CMS order.
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

// publicFrontendReindexArticles refreshes rendered loop indexes after ordering.
func publicFrontendReindexArticles(items []*publicFrontendArticle) []*publicFrontendArticle {
	for index, item := range items {
		if item == nil {
			continue
		}
		item.Index = index + 1
	}
	return items
}

// publicFrontendGroupedSlides filters slides by group marker.
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

// publicFrontendGroupedLinks filters links by group marker.
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

// publicFrontendLimitCategories returns category items limited by template attr.
func publicFrontendLimitCategories(items []*publicFrontendCategory, limit int) []*publicFrontendCategory {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// publicFrontendLimitArticles returns article items limited by template attr.
func publicFrontendLimitArticles(items []*publicFrontendArticle, limit int) []*publicFrontendArticle {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// publicFrontendLimitSlides returns slide items limited by template attr.
func publicFrontendLimitSlides(items []*publicFrontendSlide, limit int) []*publicFrontendSlide {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// publicFrontendLimitLinks returns link items limited by template attr.
func publicFrontendLimitLinks(items []*publicFrontendLink, limit int) []*publicFrontendLink {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// firstPublicFrontendCategoryHref returns the first internal category link.
func firstPublicFrontendCategoryHref(categories []*publicFrontendCategory) string {
	for _, category := range categories {
		if category != nil && !category.External {
			return category.Href
		}
	}
	return "/cms-site"
}

// mapPublicFrontendArticles maps service articles for template rendering.
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

// mapPublicFrontendSearchArticles maps service articles and prepares search
// preview fragments for the public search result template.
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

// mapPublicFrontendArticle maps one service article for template rendering.
func mapPublicFrontendArticle(item *cmssvc.ArticleItem, includeContent bool) *publicFrontendArticle {
	if item == nil || item.CmsArticle == nil {
		return &publicFrontendArticle{}
	}
	values := url.Values{}
	values.Set("article", item.Slug)
	article := &publicFrontendArticle{
		Id:              item.Id,
		CategoryId:      item.CategoryId,
		Title:           item.Title,
		Subtitle:        item.Subtitle,
		Summary:         item.Summary,
		Cover:           publicFrontendNormalizeAssetPath(item.Cover),
		Author:          item.Author,
		Source:          item.Source,
		Keywords:        item.Keywords,
		Description:     item.Description,
		CategoryName:    item.CategoryName,
		PublishedAt:     publicFrontendDate(item.PublishedAt),
		PublishedAtUnix: publicFrontendTimestamp(item.PublishedAt),
		Href:            "/cms-site?" + values.Encode(),
		Views:           item.Views,
		Sort:            item.Sort,
		IsTop:           item.IsTop,
		IsRecommend:     item.IsRecommend,
	}
	if includeContent {
		article.ContentHTML = template.HTML(publicFrontendNormalizeContentHTML(item.Content))
	}
	return article
}

// publicFrontendSearchPreview returns a safe highlighted excerpt for one
// public search result.
func publicFrontendSearchPreview(item *cmssvc.ArticleItem, keyword string) template.HTML {
	if item == nil || item.CmsArticle == nil {
		return ""
	}
	keyword = strings.TrimSpace(keyword)
	candidates := []string{
		item.Summary,
		item.Description,
		item.Subtitle,
		item.Content,
		item.Tags,
		item.Keywords,
		item.Title,
	}
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

// publicFrontendPlainText converts CMS HTML-ish text into compact plain text.
func publicFrontendPlainText(value string) string {
	text := html.UnescapeString(strings.TrimSpace(value))
	text = publicFrontendScriptPattern.ReplaceAllString(text, " ")
	text = publicFrontendStylePattern.ReplaceAllString(text, " ")
	text = publicFrontendHTMLTagPattern.ReplaceAllString(text, " ")
	text = publicFrontendWhitespacePattern.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// publicFrontendHighlightPreview clips text around a keyword and highlights the
// first hit using safe HTML.
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

// publicFrontendFindKeywordRunes returns the first case-insensitive keyword
// hit as rune offsets.
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

// publicFrontendClipRunes returns a compact plain preview when there is no
// keyword hit in the selected fallback text.
func publicFrontendClipRunes(text string, limit int) string {
	runes := []rune(text)
	if limit <= 0 || len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + publicFrontendTextMore
}

// mapPublicFrontendLinks maps enabled friendly links for template loops.
func mapPublicFrontendLinks(items []*cmssvc.LinkItem) []*publicFrontendLink {
	links := make([]*publicFrontendLink, 0, len(items))
	for index, item := range items {
		if item == nil {
			continue
		}
		links = append(links, &publicFrontendLink{
			Name:  item.Name,
			Url:   publicFrontendNormalizeAssetPath(item.Url),
			Logo:  publicFrontendNormalizeAssetPath(item.Logo),
			Group: publicFrontendLinkGroup(item.GroupCode, index),
		})
	}
	return links
}

// publicFrontendLinkGroup maps the restored reference-site links into four footer groups.
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

// buildPublicFrontendPagination returns CMS-style pagination links for the
// current public list request.
func buildPublicFrontendPagination(
	r *ghttp.Request,
	total int,
	pageSize int,
	current int,
) *publicFrontendPagination {
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
	return &publicFrontendPagination{
		Rows:       total,
		Current:    current,
		PageSize:   pageSize,
		TotalPages: totalPages,
		IndexHref:  publicFrontendPageHref(r, 1),
		PreHref:    publicFrontendPageHref(r, publicFrontendClampPageTarget(current-1, totalPages)),
		NextHref:   publicFrontendPageHref(r, publicFrontendClampPageTarget(current+1, totalPages)),
		LastHref:   publicFrontendPageHref(r, totalPages),
		NumBar:     publicFrontendNumBar(r, current, totalPages),
	}
}

// publicFrontendNumBar renders the compact numeric pagination bar.
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

// publicFrontendPageHref returns the current route with an adjusted page query.
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

// publicFrontendRequestPath returns a normalized wildcard category route path.
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

// publicFrontendCategoryPathHref returns the public href for a configured path.
func publicFrontendCategoryPathHref(value string) string {
	normalized := publicFrontendNormalizeCategoryPath(value)
	if normalized == "" {
		return ""
	}
	return "/cms-site/" + normalized + "/"
}

// publicFrontendNormalizeCategoryPath canonicalizes a stored or requested path.
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

// publicFrontendNormalizeRequestURLPath preserves the current public route base.
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

// publicFrontendCategoryListTemplate returns the list template for a category.
func publicFrontendCategoryListTemplate(category *publicFrontendCategory) string {
	if category == nil {
		return publicFrontendListName
	}
	if category.Type == cmssvc.CategoryTypeSingle {
		return publicFrontendCategoryContentTemplate(category, publicFrontendSingleName)
	}
	return publicFrontendTemplateFileName(category.ListTemplate, publicFrontendListName)
}

// publicFrontendCategoryContentTemplate returns the detail template for a category.
func publicFrontendCategoryContentTemplate(category *publicFrontendCategory, fallback string) string {
	if category == nil {
		return fallback
	}
	return publicFrontendTemplateFileName(category.ContentTemplate, fallback)
}

// publicFrontendTemplateFileName validates a configured template file name.
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

// publicFrontendTemplateFileExists reports whether an embedded template exists.
func publicFrontendTemplateFileExists(fileName string) bool {
	cleaned := path.Base(strings.TrimSpace(fileName))
	if cleaned == "" || cleaned == "." || strings.Contains(cleaned, "..") {
		return false
	}
	_, err := cmsplugin.EmbeddedFiles.ReadFile(path.Join("public/templates", cleaned))
	return err == nil
}

// publicFrontendClampedPage clamps a requested page to the available page
// count calculated from the current result total.
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

// publicFrontendClampPageTarget clamps a pagination link target.
func publicFrontendClampPageTarget(page int, totalPages int) int {
	if page < 1 {
		return 1
	}
	if totalPages > 0 && page > totalPages {
		return totalPages
	}
	return page
}

// mapPublicFrontendSlides maps enabled slides for template loops.
func mapPublicFrontendSlides(items []*cmssvc.SlideItem) []*publicFrontendSlide {
	slides := make([]*publicFrontendSlide, 0, len(items))
	for index, item := range items {
		if item == nil {
			continue
		}
		slides = append(slides, &publicFrontendSlide{
			Index:    index + 1,
			Title:    item.Title,
			Image:    publicFrontendNormalizeAssetPath(item.Image),
			Link:     publicFrontendNormalizeAssetPath(item.Link),
			Subtitle: item.Subtitle,
			Group:    publicFrontendSlideGroup(item.GroupCode),
		})
	}
	return slides
}

// publicFrontendSlideGroup returns the restored slide group marker.
func publicFrontendSlideGroup(groupCode string) string {
	if strings.TrimSpace(groupCode) == "" {
		return "1"
	}
	return strings.TrimSpace(groupCode)
}

// firstPublicFrontendSlide returns the first enabled slide for the hero panel.
func firstPublicFrontendSlide(items []*cmssvc.SlideItem) *publicFrontendSlide {
	if len(items) == 0 || items[0] == nil {
		return &publicFrontendSlide{}
	}
	return &publicFrontendSlide{
		Index:    1,
		Title:    items[0].Title,
		Image:    publicFrontendNormalizeAssetPath(items[0].Image),
		Link:     publicFrontendNormalizeAssetPath(items[0].Link),
		Subtitle: items[0].Subtitle,
		Group:    publicFrontendSlideGroup(items[0].GroupCode),
	}
}

// publicFrontendNormalizeContentHTML maps public static URLs to CMS asset URLs.
func publicFrontendNormalizeContentHTML(content string) string {
	return strings.ReplaceAll(content, publicFrontendStaticPath, "/cms-site/assets/static/")
}

// publicFrontendNormalizeAssetPath maps restored public static URLs to CMS asset URLs.
func publicFrontendNormalizeAssetPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, publicFrontendStaticPath) {
		return "/cms-site/assets" + trimmed
	}
	return trimmed
}

// publicFrontendDate formats CMS publication dates for the template.
func publicFrontendDate(value *gtime.Time) string {
	if value == nil {
		return "未发布"
	}
	return value.Format("Y-m-d")
}

// publicFrontendTimestamp returns a comparable timestamp for public ordering.
func publicFrontendTimestamp(value *gtime.Time) int64 {
	if value == nil {
		return 0
	}
	return value.Timestamp()
}

// publicFrontendTextLength truncates text using Chinese-width display units.
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

// publicFrontendRuneDisplayWidth returns Chinese-width units for one rune.
func publicFrontendRuneDisplayWidth(value rune) float64 {
	if value <= 127 {
		return 0.5
	}
	return 1
}

// queryInt64 reads one int64 query parameter.
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

// queryInt reads one integer query parameter.
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

// normalizePublicFrontendPage clamps a public page number.
func normalizePublicFrontendPage(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

// writePublicFrontendStatus writes a minimal HTML status response.
func writePublicFrontendStatus(r *ghttp.Request, status int, message string) {
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.WriteStatus(status, "<!doctype html><html lang=\"zh-CN\"><body><main>"+template.HTMLEscapeString(message)+"</main></body></html>")
	r.ExitAll()
}
