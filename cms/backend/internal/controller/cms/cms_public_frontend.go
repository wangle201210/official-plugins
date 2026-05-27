// This file implements the CMS public HTML/CSS frontend handlers.

package cms

import (
	"bytes"
	"github.com/gogf/gf/v2/net/ghttp"
	"html/template"
	"lina-core/pkg/bizerr"
	cmsplugin "lina-plugin-cms"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

// publicFrontendTemplateGlob groups embedded template and rendering constants for the public site.
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

// publicFrontendIncludePattern groups compiled patterns used by the CMS template compiler.
var (
	publicFrontendIncludePattern    = regexp.MustCompile(`\{include\s+file=([^}]+)\}`)
	publicFrontendLoopStartPattern  = regexp.MustCompile(`\{cms:(nav|children|grandchildren|list|search|slide|link|category|message)((?:[^{}]|\{(?:category|article|nav|child|grandchild|list|search|message):[^}]+\})*)\}`)
	publicFrontendIfStartPattern    = regexp.MustCompile(`\{cms:if\(([^)]*)\)\}`)
	publicFrontendScriptPattern     = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	publicFrontendStylePattern      = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	publicFrontendHTMLTagPattern    = regexp.MustCompile(`(?s)<[^>]+>`)
	publicFrontendWhitespacePattern = regexp.MustCompile(`\s+`)
)

// publicFrontendPageTemplates lists template names that may be selected from public request paths.
var publicFrontendPageTemplates = map[string]struct{}{publicFrontendIndexName: {}, publicFrontendListName: {}, "list-card.html": {}, publicFrontendSearchName: {}, publicFrontendSingleName: {}, publicFrontendDetailName: {}, publicFrontendMessageName: {}}

// publicFrontendTemplateCache stores the compiled embedded public-site template set.
var publicFrontendTemplateCache = struct {
	once sync.Once
	tpl  *template.Template
	err  error
}{}

// publicFrontendTemplateScope identifies loop and tag scopes in CMS public templates.
type publicFrontendTemplateScope string

// publicFrontendRootScope groups template scope constants used by tag replacement.
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
	publicFrontendMessageScope    publicFrontendTemplateScope = "message"
)

// publicFrontendLoopAttrs records parsed attributes for CMS public template loops.
type publicFrontendLoopAttrs struct {
	Limit    int
	Code     string
	Parent   string
	Group    string
	Order    cmssvc.PublicArticleOrder
	ParentID int64
}

// publicFrontendTextAttrs records parsed text clipping attributes from template tags.
type publicFrontendTextAttrs struct {
	Length int
	More   string
}

// publicFrontendView is the complete data model passed to public HTML templates.
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
	ApprovedMessages  []*publicFrontendMessage
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
	ShowMessages      bool
	Year              int
}

// publicFrontendCategory is the public-template projection of a CMS category node.
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

// publicFrontendArticle is the public-template projection of a CMS article.
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

// publicFrontendPagination contains page counts and navigation links for list templates.
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

// publicFrontendSlide is the public-template projection of a CMS slide.
type publicFrontendSlide struct {
	Index    int
	Title    string
	Image    string
	Link     string
	Subtitle string
	Group    string
}

// publicFrontendLink is the public-template projection of a friendly link.
type publicFrontendLink struct {
	Name  string
	Url   string
	Logo  string
	Group string
}

// publicFrontendMessage is the public-template projection of an approved visitor message.
type publicFrontendMessage struct {
	Id        int64
	Name      string
	Content   string
	Reply     string
	CreatedAt string
	UpdatedAt string
}

// PublicFrontendPage renders the CMS public page selected by request path or query parameters.
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

// PublicFrontendMessagePage renders the public visitor message page.
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
	if view.ShowMessages {
		messages, err := c.cmsSvc.ListPublicMessages(r.GetCtx(), cmssvc.PublicMessageListInput{PageNum: 1, PageSize: publicFrontendPageSize})
		if err != nil {
			writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS public messages are temporarily unavailable.")
			return
		}
		view.ApprovedMessages = mapPublicFrontendMessages(messages.List)
	}
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

// PublicFrontendStyle serves the embedded CMS public stylesheet.
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

// PublicFrontendAsset serves embedded public-site assets under the CMS path.
func (c *ControllerV1) PublicFrontendAsset(r *ghttp.Request) {
	if r == nil {
		return
	}
	fileName := path.Clean(strings.TrimPrefix(r.GetRouter("file").String(), "/"))
	c.servePublicFrontendAsset(r, fileName)
}

// PublicFrontendStaticAsset serves compatibility static assets referenced by imported templates.
func (c *ControllerV1) PublicFrontendStaticAsset(r *ghttp.Request) {
	if r == nil {
		return
	}
	fileName := path.Join("static", path.Clean(strings.TrimPrefix(r.GetRouter("file").String(), "/")))
	c.servePublicFrontendAsset(r, fileName)
}

// servePublicFrontendAsset validates an embedded asset path and writes its content type.
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

// PublicFrontendMessage accepts a public visitor message form submission.
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
	if _, err := c.cmsSvc.CreatePublicMessage(r.GetCtx(), cmssvc.PublicMessageCreateInput{Name: name, Mobile: mobile, Email: email, Content: content, UserIp: r.GetClientIp(), UserAgent: r.Header.Get("User-Agent")}); err != nil {
		r.Response.RedirectTo("/cms-site/message?message=error", http.StatusSeeOther)
		r.ExitAll()
		return
	}
	r.Response.RedirectTo("/cms-site/message?message=submitted", http.StatusSeeOther)
	r.ExitAll()
}
