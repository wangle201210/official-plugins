// This file declares shared CMS API response projections.

package v1

import "github.com/gogf/gf/v2/os/gtime"

// SiteItem defines CMS site settings returned by management and public APIs.
type SiteItem struct {
	Id          int64       `json:"id" dc:"Site ID" eg:"1"`
	SiteKey     string      `json:"siteKey" dc:"Stable site key" eg:"default"`
	Name        string      `json:"name" dc:"Site name" eg:"LinaPro CMS"`
	Logo        string      `json:"logo" dc:"Site logo URL" eg:"/uploads/logo.png"`
	Weixin      string      `json:"weixin" dc:"WeChat QR code image URL" eg:"/uploads/weixin.png"`
	Domain      string      `json:"domain" dc:"Primary site domain" eg:"https://example.com"`
	Slogan      string      `json:"slogan" dc:"Site slogan" eg:"AI-native content delivery"`
	Keywords    string      `json:"keywords" dc:"SEO keywords" eg:"LinaPro,CMS"`
	Description string      `json:"description" dc:"SEO description" eg:"LinaPro CMS site"`
	Icp         string      `json:"icp" dc:"ICP record number" eg:"ICP 00000000"`
	Contact     string      `json:"contact" dc:"Contact person" eg:"Admin"`
	Phone       string      `json:"phone" dc:"Contact phone" eg:"13800000000"`
	Email       string      `json:"email" dc:"Contact email" eg:"hello@example.com"`
	Address     string      `json:"address" dc:"Contact address" eg:"Shanghai"`
	Status      int         `json:"status" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
	CreatedBy   int64       `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy   int64       `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt   *gtime.Time `json:"createdAt" dc:"Creation time"`
	UpdatedAt   *gtime.Time `json:"updatedAt" dc:"Update time"`
}

// CategoryItem defines a CMS category node returned by management and public APIs.
type CategoryItem struct {
	Id              int64           `json:"id" dc:"Category ID" eg:"1"`
	ParentId        int64           `json:"parentId" dc:"Parent category ID" eg:"0"`
	Code            string          `json:"code" dc:"Stable category code" eg:"news"`
	Name            string          `json:"name" dc:"Category name" eg:"News"`
	Type            int             `json:"type" dc:"Category type: 1=list, 2=single page, 3=external link" eg:"1"`
	Path            string          `json:"path" dc:"Public category path" eg:"/news"`
	ListTemplate    string          `json:"listTemplate" dc:"Public list template file" eg:"list.html"`
	ContentTemplate string          `json:"contentTemplate" dc:"Public content/detail template file" eg:"detail.html"`
	Cover           string          `json:"cover" dc:"Category cover image URL" eg:"/uploads/news.png"`
	Outlink         string          `json:"outlink" dc:"External link URL" eg:"https://example.com"`
	Title           string          `json:"title" dc:"SEO title" eg:"News"`
	Keywords        string          `json:"keywords" dc:"SEO keywords" eg:"news,linapro"`
	Description     string          `json:"description" dc:"SEO description" eg:"Latest news"`
	Sort            int             `json:"sort" dc:"Display order" eg:"1"`
	Status          int             `json:"status" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
	CreatedBy       int64           `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy       int64           `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt       *gtime.Time     `json:"createdAt" dc:"Creation time"`
	UpdatedAt       *gtime.Time     `json:"updatedAt" dc:"Update time"`
	Children        []*CategoryItem `json:"children" dc:"Child category tree"`
}

// ArticleItem defines a CMS article returned by management and public APIs.
type ArticleItem struct {
	Id           int64       `json:"id" dc:"Article ID" eg:"1"`
	CategoryId   int64       `json:"categoryId" dc:"Category ID" eg:"1"`
	CategoryName string      `json:"categoryName" dc:"Category name" eg:"News"`
	Title        string      `json:"title" dc:"Article title" eg:"Welcome to LinaPro CMS"`
	Subtitle     string      `json:"subtitle" dc:"Article subtitle" eg:"First article"`
	Slug         string      `json:"slug" dc:"Public URL slug" eg:"welcome-to-linapro-cms"`
	Summary      string      `json:"summary" dc:"Article summary" eg:"CMS introduction"`
	Cover        string      `json:"cover" dc:"Cover image URL" eg:"/uploads/cover.png"`
	Author       string      `json:"author" dc:"Author name" eg:"LinaPro"`
	Source       string      `json:"source" dc:"Content source" eg:"CMS Plugin"`
	Content      string      `json:"content" dc:"Article body HTML" eg:"<p>Hello</p>"`
	Tags         string      `json:"tags" dc:"Comma-separated tag names" eg:"CMS,LinaPro"`
	Keywords     string      `json:"keywords" dc:"SEO keywords" eg:"CMS,LinaPro"`
	Description  string      `json:"description" dc:"SEO description" eg:"LinaPro CMS article"`
	Sort         int         `json:"sort" dc:"Display order" eg:"1"`
	Status       int         `json:"status" dc:"Status: 0=draft, 1=published" eg:"1"`
	IsTop        int         `json:"isTop" dc:"Top flag: 0=no, 1=yes" eg:"0"`
	IsRecommend  int         `json:"isRecommend" dc:"Recommend flag: 0=no, 1=yes" eg:"1"`
	Views        int64       `json:"views" dc:"View count" eg:"100"`
	PublishedAt  *gtime.Time `json:"publishedAt" dc:"Publication time"`
	CreatedBy    int64       `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy    int64       `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt    *gtime.Time `json:"createdAt" dc:"Creation time"`
	UpdatedAt    *gtime.Time `json:"updatedAt" dc:"Update time"`
}

// MessageItem defines a CMS visitor message returned by management APIs.
type MessageItem struct {
	Id        int64       `json:"id" dc:"Message ID" eg:"1"`
	Name      string      `json:"name" dc:"Visitor name" eg:"Alice"`
	Mobile    string      `json:"mobile" dc:"Visitor mobile" eg:"13800000000"`
	Email     string      `json:"email" dc:"Visitor email" eg:"alice@example.com"`
	Content   string      `json:"content" dc:"Message content" eg:"Please contact me"`
	Reply     string      `json:"reply" dc:"Reply content" eg:"Thanks"`
	Status    int         `json:"status" dc:"Status: 0=pending, 1=approved, 2=rejected" eg:"1"`
	UserIp    string      `json:"userIp" dc:"Visitor IP" eg:"127.0.0.1"`
	UserAgent string      `json:"userAgent" dc:"Visitor user agent" eg:"Mozilla/5.0"`
	CreatedBy int64       `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy int64       `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time"`
}

// LinkItem defines a public CMS friendly link.
type LinkItem struct {
	Id        int64       `json:"id" dc:"Link ID" eg:"1"`
	GroupCode string      `json:"groupCode" dc:"Display group code" eg:"1"`
	Name      string      `json:"name" dc:"Link name" eg:"LinaPro"`
	Url       string      `json:"url" dc:"Link URL" eg:"https://linapro.ai"`
	Logo      string      `json:"logo" dc:"Logo URL" eg:"/uploads/link.png"`
	Sort      int         `json:"sort" dc:"Display order" eg:"1"`
	Status    int         `json:"status" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
	CreatedBy int64       `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy int64       `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time"`
}

// SlideItem defines a public CMS slide.
type SlideItem struct {
	Id        int64       `json:"id" dc:"Slide ID" eg:"1"`
	GroupCode string      `json:"groupCode" dc:"Display group code" eg:"1"`
	Title     string      `json:"title" dc:"Slide title" eg:"Welcome"`
	Subtitle  string      `json:"subtitle" dc:"Slide subtitle" eg:"First slide"`
	Image     string      `json:"image" dc:"Slide image URL" eg:"/uploads/banner.png"`
	Link      string      `json:"link" dc:"Click target URL" eg:"/news"`
	Sort      int         `json:"sort" dc:"Display order" eg:"1"`
	Status    int         `json:"status" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
	CreatedBy int64       `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy int64       `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt *gtime.Time `json:"createdAt" dc:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" dc:"Update time"`
}
