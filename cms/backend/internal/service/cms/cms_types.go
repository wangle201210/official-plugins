// This file defines CMS service data projections, inputs, outputs, and public ordering helpers.

package cms

import (
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
	"strings"
)

// SiteItem aliases the CMS site entity returned by the service layer.
type SiteItem = entitymodel.CmsSite

// CategoryItem wraps a CMS category with recursive child nodes for tree responses.
type CategoryItem struct {
	*entitymodel.CmsCategory
	Children []*CategoryItem
}

// ArticleItem wraps a CMS article with the resolved category display name.
type ArticleItem struct {
	*entitymodel.CmsArticle
	CategoryName string
}

// MessageItem aliases the CMS visitor message entity.
type MessageItem = entitymodel.CmsMessage

// LinkItem aliases the CMS friendly-link entity.
type LinkItem = entitymodel.CmsLink

// SlideItem aliases the CMS slide entity.
type SlideItem = entitymodel.CmsSlide

// PublicArticleOrder describes the supported ordering modes for public article lists.
type PublicArticleOrder string

// PublicArticleOrderDefault groups the public article ordering constants.
const (
	PublicArticleOrderDefault PublicArticleOrder = ""
	PublicArticleOrderID      PublicArticleOrder = "id"
	PublicArticleOrderDate    PublicArticleOrder = "date"
	PublicArticleOrderManual  PublicArticleOrder = "manual"
	PublicArticleOrderViews   PublicArticleOrder = "views"
)

// NormalizePublicArticleOrder normalizes a public article ordering value to a supported mode.
func NormalizePublicArticleOrder(value string) PublicArticleOrder {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(PublicArticleOrderID):
		return PublicArticleOrderID
	case string(PublicArticleOrderDate):
		return PublicArticleOrderDate
	case string(PublicArticleOrderManual):
		return PublicArticleOrderManual
	case string(PublicArticleOrderViews):
		return PublicArticleOrderViews
	default:
		return PublicArticleOrderDefault
	}
}

// SiteUpdateInput carries management changes for the default CMS site settings.
type SiteUpdateInput struct {
	Name         string
	Logo         string
	Weixin       string
	Domain       string
	Slogan       string
	Keywords     string
	Description  string
	Icp          string
	Contact      string
	Phone        string
	Email        string
	Address      string
	Status       int
	ShowMessages int
}

// CategoryListInput carries filters for category tree queries.
type CategoryListInput struct {
	Status     *int
	PublicOnly bool
}

// CategorySaveInput carries category create and update fields.
type CategorySaveInput struct {
	Id              int64
	ParentId        int64
	Code            string
	Name            string
	Type            int
	Path            string
	ListTemplate    string
	ContentTemplate string
	Cover           string
	Outlink         string
	Title           string
	Keywords        string
	Description     string
	Sort            int
	Status          int
}

// ArticleListInput carries management article list filters and pagination.
type ArticleListInput struct {
	PageNum         int
	PageSize        int
	CategoryId      int64
	CategoryType    int
	IncludeChildren bool
	Status          *int
	Title           string
}

// PublicArticleListInput carries public article filters, ordering, and pagination.
type PublicArticleListInput struct {
	PageNum                 int
	PageSize                int
	CategoryId              int64
	Keyword                 string
	Order                   PublicArticleOrder
	IncludeHiddenCategories bool
}

// ArticleListOutput returns a paged article result set.
type ArticleListOutput struct {
	List  []*ArticleItem
	Total int
}

// ArticleSaveInput carries article create and update fields.
type ArticleSaveInput struct {
	Id          int64
	CategoryId  int64
	Title       string
	Subtitle    string
	Slug        string
	Summary     string
	Cover       string
	Author      string
	Source      string
	Content     string
	Tags        string
	Keywords    string
	Description string
	Sort        int
	Status      int
	IsTop       int
	IsRecommend int
	// PublishedAt optionally pins the publication time as a Unix millisecond
	// timestamp; it is only honored when Status is published and may point to
	// the future to schedule the article.
	PublishedAt *int64
}

// ArticleBatchStatusInput carries the targets and the target status for a
// batch publish or unpublish operation.
type ArticleBatchStatusInput struct {
	Ids    []int64
	Status int
}

// MessageListInput carries visitor message list filters and pagination.
type MessageListInput struct {
	PageNum  int
	PageSize int
	Status   *int
	Keyword  string
}

// MessageListOutput returns a paged message result set.
type MessageListOutput struct {
	List  []*MessageItem
	Total int
}

// PublicMessageListInput carries public visitor message pagination.
type PublicMessageListInput struct {
	PageNum  int
	PageSize int
}

// MessageUpdateInput carries review status and reply changes for one message.
type MessageUpdateInput struct {
	Id     int64
	Status int
	Reply  string
}

// LinkListInput carries friendly-link list filters and pagination.
type LinkListInput struct {
	PageNum   int
	PageSize  int
	GroupCode string
	Status    *int
	Keyword   string
}

// LinkListOutput returns a paged friendly-link result set.
type LinkListOutput struct {
	List  []*LinkItem
	Total int
}

// LinkSaveInput carries friendly-link create and update fields.
type LinkSaveInput struct {
	Id        int64
	GroupCode string
	Name      string
	Url       string
	Logo      string
	Sort      int
	Status    int
}

// SlideListInput carries slide list filters and pagination.
type SlideListInput struct {
	PageNum   int
	PageSize  int
	GroupCode string
	Status    *int
	Keyword   string
}

// SlideListOutput returns a paged slide result set.
type SlideListOutput struct {
	List  []*SlideItem
	Total int
}

// SlideSaveInput carries slide create and update fields.
type SlideSaveInput struct {
	Id        int64
	GroupCode string
	Title     string
	Subtitle  string
	Image     string
	Link      string
	Sort      int
	Status    int
}

// PublicMessageCreateInput carries sanitized public visitor message submission fields.
type PublicMessageCreateInput struct {
	Name      string
	Mobile    string
	Email     string
	Content   string
	UserIp    string
	UserAgent string
}

// ProductItem wraps a CMS product with its category name and decoded gallery URLs.
type ProductItem struct {
	*entitymodel.CmsProduct
	CategoryName string
	Gallery      []string
}

// ProductListInput carries management product list filters and pagination.
type ProductListInput struct {
	PageNum    int
	PageSize   int
	CategoryId int64
	Status     *int
	Name       string
}

// PublicProductListInput carries public product list filters and pagination.
type PublicProductListInput struct {
	PageNum    int
	PageSize   int
	CategoryId int64
}

// ProductListOutput returns a paged product result set.
type ProductListOutput struct {
	List  []*ProductItem
	Total int
}

// ProductSaveInput carries product create and update fields.
type ProductSaveInput struct {
	Id          int64
	CategoryId  int64
	Name        string
	Slug        string
	Summary     string
	Cover       string
	Gallery     []string
	Price       string
	Spec        string
	Content     string
	Keywords    string
	Description string
	Sort        int
	Status      int
	IsTop       int
	IsRecommend int
	// PublishedAt optionally pins the publication time as a Unix millisecond
	// timestamp; it is only honored when Status is published and may point to
	// the future to schedule the product.
	PublishedAt *int64
}

// AlbumItem wraps a CMS album with its category name, image count, and
// optionally its ordered images (detail reads only).
type AlbumItem struct {
	*entitymodel.CmsAlbum
	CategoryName string
	ImageCount   int
	Images       []*entitymodel.CmsAlbumImage
}

// AlbumListInput carries management album list filters and pagination.
type AlbumListInput struct {
	PageNum    int
	PageSize   int
	CategoryId int64
	Status     *int
	Name       string
}

// PublicAlbumListInput carries public album list filters and pagination.
type PublicAlbumListInput struct {
	PageNum    int
	PageSize   int
	CategoryId int64
}

// AlbumListOutput returns a paged album result set.
type AlbumListOutput struct {
	List  []*AlbumItem
	Total int
}

// AlbumImageInput carries one ordered album image in whole-album saves.
type AlbumImageInput struct {
	Url   string
	Title string
	Sort  int
}

// AlbumSaveInput carries album create and update fields with the full image list.
type AlbumSaveInput struct {
	Id          int64
	CategoryId  int64
	Name        string
	Cover       string
	Description string
	Sort        int
	Status      int
	Images      []AlbumImageInput
}
