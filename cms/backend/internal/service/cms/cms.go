// Package cms implements CMS site, category, article, message, and public-content services.
package cms

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/bizctxcap"
)

// DictTypeCategoryType groups the dictionary type keys used by CMS status and category metadata.
const (
	DictTypeCategoryType  = "cms_category_type"
	DictTypeArticleStatus = "cms_article_status"
	DictTypeMessageStatus = "cms_message_status"
	DictTypeStatus        = "cms_status"
	DictTypeYesNo         = "cms_yes_no"
)

// CategoryTypeList groups the supported CMS category type constants.
const (
	CategoryTypeList     = 1
	CategoryTypeSingle   = 2
	CategoryTypeExternal = 3
	CategoryTypeProduct  = 4
	CategoryTypeAlbum    = 5
)

// StatusDisabled groups the common enabled and disabled status constants.
const (
	StatusDisabled = 0
	StatusEnabled  = 1
)

// ArticleStatusDraft groups the supported CMS article publication statuses.
const (
	ArticleStatusDraft     = 0
	ArticleStatusPublished = 1
)

// cmsStarterContentSQLPath points to the embedded optional sample dataset used by LoadSampleData.
const (
	cmsStarterContentSQLPath = "manifest/sql/mock-data/002-cms-starter-content.sql"
)

// MessageStatusPending groups the supported visitor message review statuses.
const (
	MessageStatusPending  = 0
	MessageStatusApproved = 1
	MessageStatusRejected = 2
)

// Service exposes the CMS management and public-content operations used by controllers.
type Service interface {
	GetSite(ctx context.Context, publicOnly bool) (*SiteItem, error)
	UpdateSite(ctx context.Context, in SiteUpdateInput) error
	ClearSiteData(ctx context.Context) error
	LoadSampleData(ctx context.Context) error
	ListCategories(ctx context.Context, in CategoryListInput) ([]*CategoryItem, error)
	CreateCategory(ctx context.Context, in CategorySaveInput) (int64, error)
	UpdateCategory(ctx context.Context, in CategorySaveInput) error
	DeleteCategory(ctx context.Context, id int64) error
	ListArticles(ctx context.Context, in ArticleListInput) (*ArticleListOutput, error)
	GetArticle(ctx context.Context, id int64) (*ArticleItem, error)
	CreateArticle(ctx context.Context, in ArticleSaveInput) (int64, error)
	UpdateArticle(ctx context.Context, in ArticleSaveInput) error
	DeleteArticle(ctx context.Context, id int64) error
	// BatchUpdateArticleStatus publishes or unpublishes the given articles in one
	// transaction. The whole batch is rejected with CodeArticleNotFound when any
	// target is missing. Publishing fills empty publication times with now and
	// keeps existing ones; unpublishing keeps publication times.
	BatchUpdateArticleStatus(ctx context.Context, in ArticleBatchStatusInput) error
	// BatchDeleteArticles soft deletes the given articles with one collection
	// statement. The whole batch is rejected with CodeArticleNotFound when any
	// target is missing.
	BatchDeleteArticles(ctx context.Context, ids []int64) error
	ListMessages(ctx context.Context, in MessageListInput) (*MessageListOutput, error)
	ListPublicMessages(ctx context.Context, in PublicMessageListInput) (*MessageListOutput, error)
	UpdateMessage(ctx context.Context, in MessageUpdateInput) error
	DeleteMessage(ctx context.Context, id int64) error
	ListLinks(ctx context.Context, in LinkListInput) (*LinkListOutput, error)
	CreateLink(ctx context.Context, in LinkSaveInput) (int64, error)
	UpdateLink(ctx context.Context, in LinkSaveInput) error
	DeleteLink(ctx context.Context, id int64) error
	ListSlides(ctx context.Context, in SlideListInput) (*SlideListOutput, error)
	CreateSlide(ctx context.Context, in SlideSaveInput) (int64, error)
	UpdateSlide(ctx context.Context, in SlideSaveInput) error
	DeleteSlide(ctx context.Context, id int64) error
	// GetPublicSeoContent returns the public site settings, enabled non-external
	// categories, and the newest publicly visible articles as bounded
	// projections for sitemap and RSS rendering; articleLimit is clamped to
	// [1, SeoArticleLimitMax].
	GetPublicSeoContent(ctx context.Context, articleLimit int) (*SeoContentOutput, error)
	ListPublicArticles(ctx context.Context, in PublicArticleListInput) (*ArticleListOutput, error)
	GetPublicArticleBySlug(ctx context.Context, slug string) (*ArticleItem, error)
	ListProducts(ctx context.Context, in ProductListInput) (*ProductListOutput, error)
	GetProduct(ctx context.Context, id int64) (*ProductItem, error)
	CreateProduct(ctx context.Context, in ProductSaveInput) (int64, error)
	UpdateProduct(ctx context.Context, in ProductSaveInput) error
	DeleteProduct(ctx context.Context, id int64) error
	// ListPublicProducts returns published, released products under enabled
	// categories; scheduled products stay hidden until their publication time.
	ListPublicProducts(ctx context.Context, in PublicProductListInput) (*ProductListOutput, error)
	// GetPublicProductBySlug returns one publicly visible product by slug and
	// increments its view count; hidden products map to CodePublicContentNotFound.
	GetPublicProductBySlug(ctx context.Context, slug string) (*ProductItem, error)
	ListAlbums(ctx context.Context, in AlbumListInput) (*AlbumListOutput, error)
	GetAlbum(ctx context.Context, id int64) (*AlbumItem, error)
	// CreateAlbum creates one album and saves its full image list in one
	// transaction; image counts above AlbumImageMax are rejected.
	CreateAlbum(ctx context.Context, in AlbumSaveInput) (int64, error)
	// UpdateAlbum updates one album and replaces its full image list as a whole
	// in one transaction.
	UpdateAlbum(ctx context.Context, in AlbumSaveInput) error
	// DeleteAlbum removes one album together with all of its image rows.
	DeleteAlbum(ctx context.Context, id int64) error
	// ListPublicAlbums returns enabled albums under enabled categories with
	// covers and image counts assembled by one grouped aggregate per page.
	ListPublicAlbums(ctx context.Context, in PublicAlbumListInput) (*AlbumListOutput, error)
	// GetPublicAlbum returns one enabled album with its ordered images; hidden
	// albums map to CodePublicContentNotFound.
	GetPublicAlbum(ctx context.Context, id int64) (*AlbumItem, error)
	ListPublicLinks(ctx context.Context) ([]*LinkItem, error)
	ListPublicSlides(ctx context.Context) ([]*SlideItem, error)
	CreatePublicMessage(ctx context.Context, in PublicMessageCreateInput) (int64, error)
	PurgeStorageData(ctx context.Context) error
}

// Interface compliance assertion for the default CMS service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service with the host business context dependency.
type serviceImpl struct{ bizCtxSvc bizctxcap.Service }

// New creates a CMS service with the required host business-context dependency.
func New(bizCtxSvc bizctxcap.Service) (Service, error) {
	if bizCtxSvc == nil {
		return nil, gerror.New("cms service requires host bizctx service")
	}
	return &serviceImpl{bizCtxSvc: bizCtxSvc}, nil
}
