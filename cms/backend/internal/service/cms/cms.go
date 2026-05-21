// Package cms implements the CMS plugin site, category, article, message, and
// public-content services. It owns plugin_cms_* table access and keeps public
// content visibility rules centralized in the service layer.
package cms

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/dialect"
	plugincontract "lina-core/pkg/pluginservice/contract"
	cmsplugin "lina-plugin-cms"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
)

// CMS dictionary type constants.
const (
	DictTypeCategoryType  = "cms_category_type"  // CMS category type dictionary.
	DictTypeArticleStatus = "cms_article_status" // CMS article status dictionary.
	DictTypeMessageStatus = "cms_message_status" // CMS message status dictionary.
	DictTypeStatus        = "cms_status"         // CMS enabled status dictionary.
	DictTypeYesNo         = "cms_yes_no"         // CMS boolean flag dictionary.
)

// CMS category type values.
const (
	CategoryTypeList     = 1 // CategoryTypeList means a list category.
	CategoryTypeSingle   = 2 // CategoryTypeSingle means a single page category.
	CategoryTypeExternal = 3 // CategoryTypeExternal means an external link category.
)

// CMS status values.
const (
	StatusDisabled = 0 // StatusDisabled means disabled or hidden.
	StatusEnabled  = 1 // StatusEnabled means enabled or visible.
)

// CMS article status values.
const (
	ArticleStatusDraft     = 0 // ArticleStatusDraft means not publicly visible.
	ArticleStatusPublished = 1 // ArticleStatusPublished means publicly visible.
)

const (
	// cmsStarterContentSQLPath embeds the starter dataset used by install and runtime reset.
	cmsStarterContentSQLPath = "manifest/sql/003-cms-starter-content.sql"
)

// CMS message status values.
const (
	MessageStatusPending  = 0 // MessageStatusPending means waiting for moderation.
	MessageStatusApproved = 1 // MessageStatusApproved means approved.
	MessageStatusRejected = 2 // MessageStatusRejected means rejected.
)

// Service defines the CMS plugin service contract.
type Service interface {
	// GetSite retrieves CMS site settings.
	GetSite(ctx context.Context, publicOnly bool) (*SiteItem, error)
	// UpdateSite updates CMS site settings.
	UpdateSite(ctx context.Context, in SiteUpdateInput) error
	// ClearSiteData removes all CMS business content and resets the default site.
	ClearSiteData(ctx context.Context) error
	// LoadSampleData replaces current CMS content with the packaged starter dataset.
	LoadSampleData(ctx context.Context) error
	// ListCategories returns the CMS category tree.
	ListCategories(ctx context.Context, in CategoryListInput) ([]*CategoryItem, error)
	// CreateCategory creates a CMS category.
	CreateCategory(ctx context.Context, in CategorySaveInput) (int64, error)
	// UpdateCategory updates a CMS category.
	UpdateCategory(ctx context.Context, in CategorySaveInput) error
	// DeleteCategory deletes a CMS category.
	DeleteCategory(ctx context.Context, id int64) error
	// ListArticles returns paged CMS articles.
	ListArticles(ctx context.Context, in ArticleListInput) (*ArticleListOutput, error)
	// GetArticle retrieves one CMS article by ID.
	GetArticle(ctx context.Context, id int64) (*ArticleItem, error)
	// CreateArticle creates a CMS article.
	CreateArticle(ctx context.Context, in ArticleSaveInput) (int64, error)
	// UpdateArticle updates a CMS article.
	UpdateArticle(ctx context.Context, in ArticleSaveInput) error
	// DeleteArticle deletes one CMS article.
	DeleteArticle(ctx context.Context, id int64) error
	// ListMessages returns paged visitor messages.
	ListMessages(ctx context.Context, in MessageListInput) (*MessageListOutput, error)
	// ListPublicMessages returns approved visitor messages when enabled by site settings.
	ListPublicMessages(ctx context.Context, in PublicMessageListInput) (*MessageListOutput, error)
	// UpdateMessage updates visitor message moderation data.
	UpdateMessage(ctx context.Context, in MessageUpdateInput) error
	// DeleteMessage deletes one visitor message.
	DeleteMessage(ctx context.Context, id int64) error
	// ListLinks returns paged friendly links.
	ListLinks(ctx context.Context, in LinkListInput) (*LinkListOutput, error)
	// CreateLink creates a friendly link.
	CreateLink(ctx context.Context, in LinkSaveInput) (int64, error)
	// UpdateLink updates a friendly link.
	UpdateLink(ctx context.Context, in LinkSaveInput) error
	// DeleteLink deletes one friendly link.
	DeleteLink(ctx context.Context, id int64) error
	// ListSlides returns paged carousel slides.
	ListSlides(ctx context.Context, in SlideListInput) (*SlideListOutput, error)
	// CreateSlide creates a carousel slide.
	CreateSlide(ctx context.Context, in SlideSaveInput) (int64, error)
	// UpdateSlide updates a carousel slide.
	UpdateSlide(ctx context.Context, in SlideSaveInput) error
	// DeleteSlide deletes one carousel slide.
	DeleteSlide(ctx context.Context, id int64) error
	// ListPublicArticles returns published articles only.
	ListPublicArticles(ctx context.Context, in PublicArticleListInput) (*ArticleListOutput, error)
	// GetPublicArticleBySlug retrieves one published article by slug.
	GetPublicArticleBySlug(ctx context.Context, slug string) (*ArticleItem, error)
	// ListPublicLinks returns enabled links.
	ListPublicLinks(ctx context.Context) ([]*LinkItem, error)
	// ListPublicSlides returns enabled slides.
	ListPublicSlides(ctx context.Context) ([]*SlideItem, error)
	// CreatePublicMessage creates a public visitor message.
	CreatePublicMessage(ctx context.Context, in PublicMessageCreateInput) (int64, error)
	// PurgeStorageData clears plugin-owned tables during purge uninstall.
	PurgeStorageData(ctx context.Context) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc plugincontract.BizCtxService // bizCtxSvc reads the current authenticated user.
}

// New creates and returns a new CMS service instance with host context.
func New(bizCtxSvc plugincontract.BizCtxService) (Service, error) {
	if bizCtxSvc == nil {
		return nil, gerror.New("cms service requires host bizctx service")
	}
	return &serviceImpl{bizCtxSvc: bizCtxSvc}, nil
}

// SiteItem is the service-layer CMS site projection.
type SiteItem = entitymodel.CmsSite

// CategoryItem is the service-layer CMS category projection with nested children.
type CategoryItem struct {
	*entitymodel.CmsCategory
	Children []*CategoryItem
}

// ArticleItem is the service-layer CMS article projection.
type ArticleItem struct {
	*entitymodel.CmsArticle
	CategoryName string
}

// MessageItem is the service-layer CMS message projection.
type MessageItem = entitymodel.CmsMessage

// LinkItem is the service-layer CMS link projection.
type LinkItem = entitymodel.CmsLink

// SlideItem is the service-layer CMS slide projection.
type SlideItem = entitymodel.CmsSlide

// PublicArticleOrder is the public CMS article ordering strategy.
type PublicArticleOrder string

const (
	// PublicArticleOrderDefault uses the stable default public ordering.
	PublicArticleOrderDefault PublicArticleOrder = ""
	// PublicArticleOrderID orders newer records first by ID.
	PublicArticleOrderID PublicArticleOrder = "id"
	// PublicArticleOrderDate orders published content by publication time.
	PublicArticleOrderDate PublicArticleOrder = "date"
	// PublicArticleOrderManual orders content by the CMS manual display order.
	PublicArticleOrderManual PublicArticleOrder = "manual"
	// PublicArticleOrderViews orders content by view count.
	PublicArticleOrderViews PublicArticleOrder = "views"
)

// NormalizePublicArticleOrder returns a supported public article order value.
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

// SiteUpdateInput defines CMS site update input.
type SiteUpdateInput struct {
	Name         string // Site name.
	Logo         string // Site logo URL.
	Weixin       string // WeChat QR code image URL.
	Domain       string // Primary site domain.
	Slogan       string // Site slogan.
	Keywords     string // SEO keywords.
	Description  string // SEO description.
	Icp          string // ICP record number.
	Contact      string // Contact person.
	Phone        string // Contact phone.
	Email        string // Contact email.
	Address      string // Contact address.
	Status       int    // Status: 0=disabled, 1=enabled.
	ShowMessages int    // Show approved visitor messages on the public message page.
}

// CategoryListInput defines CMS category list filters.
type CategoryListInput struct {
	Status     *int // Optional status filter.
	PublicOnly bool // PublicOnly restricts results to enabled categories.
}

// CategorySaveInput defines CMS category create/update input.
type CategorySaveInput struct {
	Id              int64  // Category ID for updates.
	ParentId        int64  // Parent category ID.
	Code            string // Stable category code.
	Name            string // Category name.
	Type            int    // Category type.
	Path            string // Public path.
	ListTemplate    string // Public list template file.
	ContentTemplate string // Public content/detail template file.
	Cover           string // Cover image URL.
	Outlink         string // External link URL.
	Title           string // SEO title.
	Keywords        string // SEO keywords.
	Description     string // SEO description.
	Sort            int    // Display order.
	Status          int    // Status.
}

// ArticleListInput defines CMS article list filters.
type ArticleListInput struct {
	PageNum         int    // Page number.
	PageSize        int    // Page size.
	CategoryId      int64  // Category filter.
	CategoryType    int    // Category type filter.
	IncludeChildren bool   // Include child categories when CategoryId is provided.
	Status          *int   // Optional status filter.
	Title           string // Title fuzzy filter.
}

// PublicArticleListInput defines public article list filters.
type PublicArticleListInput struct {
	PageNum                 int                // Page number.
	PageSize                int                // Page size.
	CategoryId              int64              // Category filter.
	Keyword                 string             // Public keyword filter.
	Order                   PublicArticleOrder // Public article ordering strategy.
	IncludeHiddenCategories bool               // Include categories hidden from navigation.
}

// ArticleListOutput defines a paged CMS article list.
type ArticleListOutput struct {
	List  []*ArticleItem // Article list.
	Total int            // Total count.
}

// ArticleSaveInput defines CMS article create/update input.
type ArticleSaveInput struct {
	Id          int64  // Article ID for updates.
	CategoryId  int64  // Category ID.
	Title       string // Article title.
	Subtitle    string // Article subtitle.
	Slug        string // Public slug.
	Summary     string // Article summary.
	Cover       string // Cover image URL.
	Author      string // Author.
	Source      string // Source.
	Content     string // Article HTML body.
	Tags        string // Comma-separated tags.
	Keywords    string // SEO keywords.
	Description string // SEO description.
	Sort        int    // Display order.
	Status      int    // Status.
	IsTop       int    // Top flag.
	IsRecommend int    // Recommend flag.
}

// MessageListInput defines visitor message list filters.
type MessageListInput struct {
	PageNum  int    // Page number.
	PageSize int    // Page size.
	Status   *int   // Optional message status filter.
	Keyword  string // Fuzzy search keyword.
}

// MessageListOutput defines paged visitor messages.
type MessageListOutput struct {
	List  []*MessageItem // Message list.
	Total int            // Total count.
}

// PublicMessageListInput defines public approved visitor message filters.
type PublicMessageListInput struct {
	PageNum  int // Page number.
	PageSize int // Page size.
}

// MessageUpdateInput defines visitor message moderation input.
type MessageUpdateInput struct {
	Id     int64  // Message ID.
	Status int    // Moderation status.
	Reply  string // Reply content.
}

// LinkListInput defines friendly link list filters.
type LinkListInput struct {
	PageNum   int    // Page number.
	PageSize  int    // Page size.
	GroupCode string // Optional display group code filter.
	Status    *int   // Optional status filter.
	Keyword   string // Fuzzy search keyword.
}

// LinkListOutput defines paged friendly links.
type LinkListOutput struct {
	List  []*LinkItem // Link list.
	Total int         // Total count.
}

// LinkSaveInput defines friendly link create/update input.
type LinkSaveInput struct {
	Id        int64  // Link ID for updates.
	GroupCode string // Display group code.
	Name      string // Link name.
	Url       string // Link URL.
	Logo      string // Logo URL.
	Sort      int    // Display order.
	Status    int    // Status.
}

// SlideListInput defines carousel slide list filters.
type SlideListInput struct {
	PageNum   int    // Page number.
	PageSize  int    // Page size.
	GroupCode string // Optional display group code filter.
	Status    *int   // Optional status filter.
	Keyword   string // Fuzzy search keyword.
}

// SlideListOutput defines paged carousel slides.
type SlideListOutput struct {
	List  []*SlideItem // Slide list.
	Total int          // Total count.
}

// SlideSaveInput defines carousel slide create/update input.
type SlideSaveInput struct {
	Id        int64  // Slide ID for updates.
	GroupCode string // Display group code.
	Title     string // Slide title.
	Subtitle  string // Slide subtitle.
	Image     string // Slide image URL.
	Link      string // Click target URL.
	Sort      int    // Display order.
	Status    int    // Status.
}

// PublicMessageCreateInput defines public visitor message creation input.
type PublicMessageCreateInput struct {
	Name      string // Visitor name.
	Mobile    string // Visitor mobile.
	Email     string // Visitor email.
	Content   string // Visitor message.
	UserIp    string // Visitor IP.
	UserAgent string // Visitor user agent.
}

// GetSite retrieves the default CMS site settings.
func (s *serviceImpl) GetSite(ctx context.Context, publicOnly bool) (*SiteItem, error) {
	columns := dao.CmsSite.Columns()
	model := dao.CmsSite.Ctx(ctx).Where(columns.SiteKey, "default")
	if publicOnly {
		model = model.Where(columns.Status, StatusEnabled)
	}

	var site *entitymodel.CmsSite
	if err := model.Scan(&site); err != nil {
		return nil, err
	}
	if site == nil {
		if publicOnly {
			return nil, bizerr.NewCode(CodePublicContentNotFound)
		}
		return nil, bizerr.NewCode(CodeSiteNotFound)
	}
	return site, nil
}

// UpdateSite updates the default CMS site settings.
func (s *serviceImpl) UpdateSite(ctx context.Context, in SiteUpdateInput) error {
	columns := dao.CmsSite.Columns()
	site, err := s.GetSite(ctx, false)
	if err != nil {
		return err
	}
	_, err = dao.CmsSite.Ctx(ctx).
		Where(columns.Id, site.Id).
		Data(do.CmsSite{
			Name:         in.Name,
			Logo:         in.Logo,
			Weixin:       in.Weixin,
			Domain:       in.Domain,
			Slogan:       in.Slogan,
			Keywords:     in.Keywords,
			Description:  in.Description,
			Icp:          in.Icp,
			Contact:      in.Contact,
			Phone:        in.Phone,
			Email:        in.Email,
			Address:      in.Address,
			Status:       in.Status,
			ShowMessages: in.ShowMessages,
			UpdatedBy:    s.currentUserID(ctx),
		}).
		Update()
	return err
}

// ClearSiteData removes all CMS content rows and leaves one blank default site.
func (s *serviceImpl) ClearSiteData(ctx context.Context) error {
	userID := s.currentUserID(ctx)
	return dao.CmsSite.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := clearCMSData(ctx, tx); err != nil {
			return err
		}
		_, err := tx.Model(dao.CmsSite.Table()).Ctx(ctx).Data(do.CmsSite{
			SiteKey:      "default",
			Name:         "LinaPro CMS",
			Status:       StatusEnabled,
			ShowMessages: StatusEnabled,
			CreatedBy:    userID,
			UpdatedBy:    userID,
		}).Insert()
		return err
	})
}

// LoadSampleData resets CMS business data and reloads the packaged starter dataset.
func (s *serviceImpl) LoadSampleData(ctx context.Context) error {
	return dao.CmsSite.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := clearCMSData(ctx, tx); err != nil {
			return bizerr.WrapCode(err, CodeSampleDataLoadFailed)
		}
		if err := executeStarterContentSQL(ctx, tx); err != nil {
			return err
		}
		if err := markSampleDataMaintainer(ctx, tx, s.currentUserID(ctx)); err != nil {
			return bizerr.WrapCode(err, CodeSampleDataLoadFailed)
		}
		return nil
	})
}

// ListCategories returns CMS categories as a tree.
func (s *serviceImpl) ListCategories(ctx context.Context, in CategoryListInput) ([]*CategoryItem, error) {
	columns := dao.CmsCategory.Columns()
	model := dao.CmsCategory.Ctx(ctx).OrderAsc(columns.Sort).OrderAsc(columns.Id)
	if in.PublicOnly {
		model = model.Where(columns.Status, StatusEnabled)
	} else if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}

	list := make([]*entitymodel.CmsCategory, 0)
	if err := model.Scan(&list); err != nil {
		return nil, err
	}
	return buildCategoryTree(list), nil
}

// CreateCategory creates a CMS category.
func (s *serviceImpl) CreateCategory(ctx context.Context, in CategorySaveInput) (int64, error) {
	if err := s.ensureCategoryCodeAvailable(ctx, in.Code, 0); err != nil {
		return 0, err
	}
	if err := s.ensureCategoryParentAvailable(ctx, 0, in.ParentId); err != nil {
		return 0, err
	}
	userID := s.currentUserID(ctx)
	return dao.CmsCategory.Ctx(ctx).Data(do.CmsCategory{
		ParentId:        in.ParentId,
		Code:            in.Code,
		Name:            in.Name,
		Type:            in.Type,
		Path:            in.Path,
		ListTemplate:    in.ListTemplate,
		ContentTemplate: in.ContentTemplate,
		Cover:           in.Cover,
		Outlink:         in.Outlink,
		Title:           in.Title,
		Keywords:        in.Keywords,
		Description:     in.Description,
		Sort:            in.Sort,
		Status:          in.Status,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	}).InsertAndGetId()
}

// UpdateCategory updates a CMS category.
func (s *serviceImpl) UpdateCategory(ctx context.Context, in CategorySaveInput) error {
	columns := dao.CmsCategory.Columns()
	if err := s.ensureCategoryExists(ctx, in.Id); err != nil {
		return err
	}
	if err := s.ensureCategoryCodeAvailable(ctx, in.Code, in.Id); err != nil {
		return err
	}
	if err := s.ensureCategoryParentAvailable(ctx, in.Id, in.ParentId); err != nil {
		return err
	}
	_, err := dao.CmsCategory.Ctx(ctx).
		Where(columns.Id, in.Id).
		Data(do.CmsCategory{
			ParentId:        in.ParentId,
			Code:            in.Code,
			Name:            in.Name,
			Type:            in.Type,
			Path:            in.Path,
			ListTemplate:    in.ListTemplate,
			ContentTemplate: in.ContentTemplate,
			Cover:           in.Cover,
			Outlink:         in.Outlink,
			Title:           in.Title,
			Keywords:        in.Keywords,
			Description:     in.Description,
			Sort:            in.Sort,
			Status:          in.Status,
			UpdatedBy:       s.currentUserID(ctx),
		}).
		Update()
	return err
}

// DeleteCategory deletes one CMS category when it is not referenced.
func (s *serviceImpl) DeleteCategory(ctx context.Context, id int64) error {
	columns := dao.CmsCategory.Columns()
	articleColumns := dao.CmsArticle.Columns()
	if err := s.ensureCategoryExists(ctx, id); err != nil {
		return err
	}
	children, err := dao.CmsCategory.Ctx(ctx).Where(columns.ParentId, id).Count()
	if err != nil {
		return err
	}
	if children > 0 {
		return bizerr.NewCode(CodeCategoryHasChildren)
	}
	articles, err := dao.CmsArticle.Ctx(ctx).Where(articleColumns.CategoryId, id).Count()
	if err != nil {
		return err
	}
	if articles > 0 {
		return bizerr.NewCode(CodeCategoryHasArticles)
	}
	_, err = dao.CmsCategory.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ListArticles returns paged CMS articles for management.
func (s *serviceImpl) ListArticles(ctx context.Context, in ArticleListInput) (*ArticleListOutput, error) {
	model, err := s.applyArticleManagementFilters(ctx, dao.CmsArticle.Ctx(ctx), in)
	if err != nil {
		return nil, err
	}
	return s.scanArticlePage(ctx, model, in.PageNum, in.PageSize)
}

// GetArticle retrieves one CMS article for management.
func (s *serviceImpl) GetArticle(ctx context.Context, id int64) (*ArticleItem, error) {
	var article *entitymodel.CmsArticle
	if err := dao.CmsArticle.Ctx(ctx).Where(dao.CmsArticle.Columns().Id, id).Scan(&article); err != nil {
		return nil, err
	}
	if article == nil {
		return nil, bizerr.NewCode(CodeArticleNotFound)
	}
	return s.wrapArticleItem(ctx, article)
}

// CreateArticle creates a CMS article.
func (s *serviceImpl) CreateArticle(ctx context.Context, in ArticleSaveInput) (int64, error) {
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return 0, err
	}
	if err := s.ensureArticleSlugAvailable(ctx, in.Slug, 0); err != nil {
		return 0, err
	}
	userID := s.currentUserID(ctx)
	data := do.CmsArticle{
		CategoryId:  in.CategoryId,
		Title:       in.Title,
		Subtitle:    in.Subtitle,
		Slug:        in.Slug,
		Summary:     in.Summary,
		Cover:       in.Cover,
		Author:      in.Author,
		Source:      in.Source,
		Content:     in.Content,
		Tags:        in.Tags,
		Keywords:    in.Keywords,
		Description: in.Description,
		Sort:        in.Sort,
		Status:      in.Status,
		IsTop:       in.IsTop,
		IsRecommend: in.IsRecommend,
		PublishedAt: publishedAtForStatus(in.Status, nil),
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	return dao.CmsArticle.Ctx(ctx).Data(data).InsertAndGetId()
}

// UpdateArticle updates a CMS article.
func (s *serviceImpl) UpdateArticle(ctx context.Context, in ArticleSaveInput) error {
	columns := dao.CmsArticle.Columns()
	var oldArticle *entitymodel.CmsArticle
	if err := dao.CmsArticle.Ctx(ctx).Where(columns.Id, in.Id).Scan(&oldArticle); err != nil {
		return err
	}
	if oldArticle == nil {
		return bizerr.NewCode(CodeArticleNotFound)
	}
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return err
	}
	if err := s.ensureArticleSlugAvailable(ctx, in.Slug, in.Id); err != nil {
		return err
	}

	_, err := dao.CmsArticle.Ctx(ctx).
		Where(columns.Id, in.Id).
		Data(do.CmsArticle{
			CategoryId:  in.CategoryId,
			Title:       in.Title,
			Subtitle:    in.Subtitle,
			Slug:        in.Slug,
			Summary:     in.Summary,
			Cover:       in.Cover,
			Author:      in.Author,
			Source:      in.Source,
			Content:     in.Content,
			Tags:        in.Tags,
			Keywords:    in.Keywords,
			Description: in.Description,
			Sort:        in.Sort,
			Status:      in.Status,
			IsTop:       in.IsTop,
			IsRecommend: in.IsRecommend,
			PublishedAt: publishedAtForStatus(in.Status, oldArticle.PublishedAt),
			UpdatedBy:   s.currentUserID(ctx),
		}).
		Update()
	return err
}

// DeleteArticle deletes one CMS article.
func (s *serviceImpl) DeleteArticle(ctx context.Context, id int64) error {
	columns := dao.CmsArticle.Columns()
	if _, err := s.GetArticle(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsArticle.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ListMessages returns paged visitor messages.
func (s *serviceImpl) ListMessages(ctx context.Context, in MessageListInput) (*MessageListOutput, error) {
	columns := dao.CmsMessage.Columns()
	model := dao.CmsMessage.Ctx(ctx)
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if in.Keyword != "" {
		keyword := "%" + in.Keyword + "%"
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ?)",
				columns.Name,
				columns.Email,
				columns.Mobile,
				columns.Content,
			),
			keyword,
			keyword,
			keyword,
			keyword,
		)
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	list := make([]*MessageItem, 0)
	if err = model.Page(normalizePageNum(in.PageNum), normalizePageSize(in.PageSize)).
		OrderDesc(columns.Id).
		Scan(&list); err != nil {
		return nil, err
	}
	return &MessageListOutput{List: list, Total: total}, nil
}

// ListPublicMessages returns approved visitor messages when site settings allow
// public message display.
func (s *serviceImpl) ListPublicMessages(ctx context.Context, in PublicMessageListInput) (*MessageListOutput, error) {
	site, err := s.GetSite(ctx, true)
	if err != nil {
		return nil, err
	}
	if site.ShowMessages != StatusEnabled {
		return &MessageListOutput{List: []*MessageItem{}, Total: 0}, nil
	}

	approvedStatus := MessageStatusApproved
	return s.ListMessages(ctx, MessageListInput{
		PageNum:  in.PageNum,
		PageSize: in.PageSize,
		Status:   &approvedStatus,
	})
}

// UpdateMessage updates visitor message moderation data.
func (s *serviceImpl) UpdateMessage(ctx context.Context, in MessageUpdateInput) error {
	columns := dao.CmsMessage.Columns()
	if err := s.ensureMessageExists(ctx, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsMessage.Ctx(ctx).
		Where(columns.Id, in.Id).
		Data(do.CmsMessage{
			Status:    in.Status,
			Reply:     in.Reply,
			UpdatedBy: s.currentUserID(ctx),
		}).
		Update()
	return err
}

// DeleteMessage deletes one visitor message.
func (s *serviceImpl) DeleteMessage(ctx context.Context, id int64) error {
	columns := dao.CmsMessage.Columns()
	if err := s.ensureMessageExists(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsMessage.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ListLinks returns paged friendly links for management.
func (s *serviceImpl) ListLinks(ctx context.Context, in LinkListInput) (*LinkListOutput, error) {
	columns := dao.CmsLink.Columns()
	model := dao.CmsLink.Ctx(ctx)
	if in.GroupCode != "" {
		model = model.Where(columns.GroupCode, strings.TrimSpace(in.GroupCode))
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if keywordText := strings.TrimSpace(in.Keyword); keywordText != "" {
		keyword := "%" + keywordText + "%"
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", columns.Name, columns.Url),
			keyword,
			keyword,
		)
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	list := make([]*LinkItem, 0)
	if err = model.
		Page(normalizePageNum(in.PageNum), normalizePageSize(in.PageSize)).
		OrderAsc(columns.Sort).
		OrderAsc(columns.Id).
		Scan(&list); err != nil {
		return nil, err
	}
	return &LinkListOutput{List: list, Total: total}, nil
}

// CreateLink creates a friendly link.
func (s *serviceImpl) CreateLink(ctx context.Context, in LinkSaveInput) (int64, error) {
	userID := s.currentUserID(ctx)
	return dao.CmsLink.Ctx(ctx).Data(do.CmsLink{
		GroupCode: strings.TrimSpace(in.GroupCode),
		Name:      strings.TrimSpace(in.Name),
		Url:       strings.TrimSpace(in.Url),
		Logo:      strings.TrimSpace(in.Logo),
		Sort:      in.Sort,
		Status:    in.Status,
		CreatedBy: userID,
		UpdatedBy: userID,
	}).InsertAndGetId()
}

// UpdateLink updates a friendly link.
func (s *serviceImpl) UpdateLink(ctx context.Context, in LinkSaveInput) error {
	columns := dao.CmsLink.Columns()
	if err := s.ensureLinkExists(ctx, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsLink.Ctx(ctx).
		Where(columns.Id, in.Id).
		Data(do.CmsLink{
			GroupCode: strings.TrimSpace(in.GroupCode),
			Name:      strings.TrimSpace(in.Name),
			Url:       strings.TrimSpace(in.Url),
			Logo:      strings.TrimSpace(in.Logo),
			Sort:      in.Sort,
			Status:    in.Status,
			UpdatedBy: s.currentUserID(ctx),
		}).
		Update()
	return err
}

// DeleteLink deletes one friendly link.
func (s *serviceImpl) DeleteLink(ctx context.Context, id int64) error {
	columns := dao.CmsLink.Columns()
	if err := s.ensureLinkExists(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsLink.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ListSlides returns paged carousel slides for management.
func (s *serviceImpl) ListSlides(ctx context.Context, in SlideListInput) (*SlideListOutput, error) {
	columns := dao.CmsSlide.Columns()
	model := dao.CmsSlide.Ctx(ctx)
	if in.GroupCode != "" {
		model = model.Where(columns.GroupCode, strings.TrimSpace(in.GroupCode))
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if keywordText := strings.TrimSpace(in.Keyword); keywordText != "" {
		keyword := "%" + keywordText + "%"
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", columns.Title, columns.Subtitle),
			keyword,
			keyword,
		)
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	list := make([]*SlideItem, 0)
	if err = model.
		Page(normalizePageNum(in.PageNum), normalizePageSize(in.PageSize)).
		OrderAsc(columns.Sort).
		OrderAsc(columns.Id).
		Scan(&list); err != nil {
		return nil, err
	}
	return &SlideListOutput{List: list, Total: total}, nil
}

// CreateSlide creates a carousel slide.
func (s *serviceImpl) CreateSlide(ctx context.Context, in SlideSaveInput) (int64, error) {
	userID := s.currentUserID(ctx)
	return dao.CmsSlide.Ctx(ctx).Data(do.CmsSlide{
		GroupCode: strings.TrimSpace(in.GroupCode),
		Title:     strings.TrimSpace(in.Title),
		Subtitle:  strings.TrimSpace(in.Subtitle),
		Image:     strings.TrimSpace(in.Image),
		Link:      strings.TrimSpace(in.Link),
		Sort:      in.Sort,
		Status:    in.Status,
		CreatedBy: userID,
		UpdatedBy: userID,
	}).InsertAndGetId()
}

// UpdateSlide updates a carousel slide.
func (s *serviceImpl) UpdateSlide(ctx context.Context, in SlideSaveInput) error {
	columns := dao.CmsSlide.Columns()
	if err := s.ensureSlideExists(ctx, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsSlide.Ctx(ctx).
		Where(columns.Id, in.Id).
		Data(do.CmsSlide{
			GroupCode: strings.TrimSpace(in.GroupCode),
			Title:     strings.TrimSpace(in.Title),
			Subtitle:  strings.TrimSpace(in.Subtitle),
			Image:     strings.TrimSpace(in.Image),
			Link:      strings.TrimSpace(in.Link),
			Sort:      in.Sort,
			Status:    in.Status,
			UpdatedBy: s.currentUserID(ctx),
		}).
		Update()
	return err
}

// DeleteSlide deletes one carousel slide.
func (s *serviceImpl) DeleteSlide(ctx context.Context, id int64) error {
	columns := dao.CmsSlide.Columns()
	if err := s.ensureSlideExists(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsSlide.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ListPublicArticles returns published articles with enabled categories only.
func (s *serviceImpl) ListPublicArticles(ctx context.Context, in PublicArticleListInput) (*ArticleListOutput, error) {
	columns := dao.CmsArticle.Columns()
	model := s.applyPublicArticleVisibility(ctx, dao.CmsArticle.Ctx(ctx), in.IncludeHiddenCategories)
	if in.CategoryId > 0 {
		model = model.Where(columns.CategoryId, in.CategoryId)
	}
	if keywordText := strings.TrimSpace(in.Keyword); keywordText != "" {
		keyword := "%" + keywordText + "%"
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ?)",
				columns.Title,
				columns.Subtitle,
				columns.Summary,
				columns.Content,
				columns.Tags,
				columns.Keywords,
				columns.Description,
			),
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
		)
	}
	model = s.applyPublicArticleOrder(model, in.Order)
	return s.scanArticlePage(ctx, model, in.PageNum, in.PageSize)
}

// GetPublicArticleBySlug retrieves one published article by slug.
func (s *serviceImpl) GetPublicArticleBySlug(ctx context.Context, slug string) (*ArticleItem, error) {
	columns := dao.CmsArticle.Columns()
	var article *entitymodel.CmsArticle
	err := s.applyPublicArticleVisibility(ctx, dao.CmsArticle.Ctx(ctx), true).
		Where(columns.Slug, strings.TrimSpace(slug)).
		Scan(&article)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, bizerr.NewCode(CodePublicContentNotFound)
	}
	if _, err = dao.CmsArticle.Ctx(ctx).
		Where(columns.Id, article.Id).
		Data(do.CmsArticle{Views: gdb.Raw(columns.Views + " + 1")}).
		Update(); err != nil {
		return nil, err
	}
	article.Views++
	return s.wrapArticleItem(ctx, article)
}

// ListPublicLinks returns enabled public friendly links.
func (s *serviceImpl) ListPublicLinks(ctx context.Context) ([]*LinkItem, error) {
	columns := dao.CmsLink.Columns()
	list := make([]*LinkItem, 0)
	err := dao.CmsLink.Ctx(ctx).
		Where(columns.Status, StatusEnabled).
		OrderAsc(columns.Sort).
		OrderAsc(columns.Id).
		Scan(&list)
	return list, err
}

// ListPublicSlides returns enabled public slides.
func (s *serviceImpl) ListPublicSlides(ctx context.Context) ([]*SlideItem, error) {
	columns := dao.CmsSlide.Columns()
	list := make([]*SlideItem, 0)
	err := dao.CmsSlide.Ctx(ctx).
		Where(columns.Status, StatusEnabled).
		OrderAsc(columns.Sort).
		OrderAsc(columns.Id).
		Scan(&list)
	return list, err
}

// CreatePublicMessage creates a public visitor message.
func (s *serviceImpl) CreatePublicMessage(ctx context.Context, in PublicMessageCreateInput) (int64, error) {
	return dao.CmsMessage.Ctx(ctx).Data(do.CmsMessage{
		Name:      in.Name,
		Mobile:    in.Mobile,
		Email:     in.Email,
		Content:   in.Content,
		Status:    MessageStatusPending,
		UserIp:    in.UserIp,
		UserAgent: in.UserAgent,
	}).InsertAndGetId()
}

// PurgeStorageData clears plugin-owned tables during purge uninstall.
func PurgeStorageData(ctx context.Context) error {
	tables := append(cmsContentTables(), dao.CmsSite.Table())
	for _, table := range tables {
		if _, err := dao.CmsSite.DB().Exec(ctx, "DELETE FROM "+table); err != nil {
			return err
		}
	}
	return nil
}

// PurgeStorageData delegates service cleanup to the dependency-free purge entry.
func (s *serviceImpl) PurgeStorageData(ctx context.Context) error {
	return PurgeStorageData(ctx)
}

// clearCMSData removes all CMS content tables and site settings inside caller transaction.
func clearCMSData(ctx context.Context, tx gdb.TX) error {
	for _, table := range cmsContentTables() {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM "+dao.CmsSite.Table()); err != nil {
		return err
	}
	return nil
}

// executeStarterContentSQL executes the embedded starter content SQL inside caller transaction.
func executeStarterContentSQL(ctx context.Context, tx gdb.TX) error {
	content, err := cmsplugin.EmbeddedFiles.ReadFile(cmsStarterContentSQLPath)
	if err != nil {
		return bizerr.WrapCode(err, CodeSampleDataLoadFailed)
	}
	for _, statement := range dialect.SplitSQLStatements(string(content)) {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			return bizerr.WrapCode(err, CodeSampleDataLoadFailed)
		}
	}
	return nil
}

// markSampleDataMaintainer records the operator that triggered runtime sample data loading.
func markSampleDataMaintainer(ctx context.Context, tx gdb.TX, userID int64) error {
	if userID <= 0 {
		return nil
	}
	updates := []struct {
		table string
		data  any
	}{
		{table: dao.CmsMessage.Table(), data: do.CmsMessage{CreatedBy: userID, UpdatedBy: userID}},
		{table: dao.CmsSlide.Table(), data: do.CmsSlide{CreatedBy: userID, UpdatedBy: userID}},
		{table: dao.CmsLink.Table(), data: do.CmsLink{CreatedBy: userID, UpdatedBy: userID}},
		{table: dao.CmsArticleTag.Table(), data: do.CmsArticleTag{CreatedBy: userID, UpdatedBy: userID}},
		{table: dao.CmsArticle.Table(), data: do.CmsArticle{CreatedBy: userID, UpdatedBy: userID}},
		{table: dao.CmsCategory.Table(), data: do.CmsCategory{CreatedBy: userID, UpdatedBy: userID}},
		{table: dao.CmsSite.Table(), data: do.CmsSite{CreatedBy: userID, UpdatedBy: userID}},
	}
	for _, update := range updates {
		if _, err := tx.Model(update.table).Ctx(ctx).Data(update.data).Update(); err != nil {
			return err
		}
	}
	return nil
}

// cmsContentTables returns content tables in an order that clears dependents before parents.
func cmsContentTables() []string {
	return []string{
		dao.CmsMessage.Table(),
		dao.CmsSlide.Table(),
		dao.CmsLink.Table(),
		dao.CmsArticleTag.Table(),
		dao.CmsArticle.Table(),
		dao.CmsCategory.Table(),
	}
}

// applyArticleManagementFilters applies management article list filters.
func (s *serviceImpl) applyArticleManagementFilters(
	ctx context.Context,
	model *gdb.Model,
	in ArticleListInput,
) (*gdb.Model, error) {
	columns := dao.CmsArticle.Columns()
	if in.CategoryId > 0 {
		ids, err := s.categoryIDsForArticleFilter(ctx, in.CategoryId, in.IncludeChildren)
		if err != nil {
			return nil, err
		}
		model = model.WhereIn(columns.CategoryId, ids)
	} else if in.CategoryType > 0 {
		categoryColumns := dao.CmsCategory.Columns()
		subQuery := dao.CmsCategory.Ctx(ctx).
			Fields(categoryColumns.Id).
			Where(categoryColumns.Type, in.CategoryType)
		model = model.Where(columns.CategoryId+" IN (?)", subQuery)
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if in.Title != "" {
		model = model.WhereLike(columns.Title, "%"+in.Title+"%")
	}
	return model.OrderDesc(columns.IsTop).OrderDesc(columns.PublishedAt).OrderDesc(columns.Id), nil
}

// categoryIDsForArticleFilter resolves the category IDs used by management
// article list filtering.
func (s *serviceImpl) categoryIDsForArticleFilter(
	ctx context.Context,
	categoryID int64,
	includeChildren bool,
) ([]int64, error) {
	if categoryID <= 0 {
		return nil, nil
	}
	if !includeChildren {
		return []int64{categoryID}, nil
	}
	categories := make([]*entitymodel.CmsCategory, 0)
	columns := dao.CmsCategory.Columns()
	if err := dao.CmsCategory.Ctx(ctx).
		Fields(columns.Id, columns.ParentId).
		Scan(&categories); err != nil {
		return nil, err
	}
	childrenByParent := make(map[int64][]int64)
	for _, category := range categories {
		childrenByParent[category.ParentId] = append(childrenByParent[category.ParentId], category.Id)
	}
	ids := []int64{categoryID}
	queue := []int64{categoryID}
	visited := map[int64]struct{}{categoryID: {}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range childrenByParent[current] {
			if _, ok := visited[childID]; ok {
				continue
			}
			visited[childID] = struct{}{}
			ids = append(ids, childID)
			queue = append(queue, childID)
		}
	}
	return ids, nil
}

// applyPublicArticleVisibility restricts a query to published articles whose
// category is enabled.
func (s *serviceImpl) applyPublicArticleVisibility(
	ctx context.Context,
	model *gdb.Model,
	includeHiddenCategories bool,
) *gdb.Model {
	articleColumns := dao.CmsArticle.Columns()
	model = model.Where(articleColumns.Status, ArticleStatusPublished)
	if includeHiddenCategories {
		return model
	}
	categoryColumns := dao.CmsCategory.Columns()
	enabledCategorySubQuery := dao.CmsCategory.Ctx(ctx).
		Fields(categoryColumns.Id).
		Where(categoryColumns.Status, StatusEnabled)
	return model.Where(articleColumns.CategoryId+" IN (?)", enabledCategorySubQuery)
}

// applyPublicArticleOrder applies a CMS-template-compatible article order.
func (s *serviceImpl) applyPublicArticleOrder(model *gdb.Model, order PublicArticleOrder) *gdb.Model {
	columns := dao.CmsArticle.Columns()
	switch NormalizePublicArticleOrder(string(order)) {
	case PublicArticleOrderID:
		return model.
			OrderDesc(columns.Id).
			OrderDesc(columns.IsTop).
			OrderDesc(columns.IsRecommend).
			OrderAsc(columns.Sort).
			OrderDesc(columns.PublishedAt)
	case PublicArticleOrderDate:
		return model.
			OrderDesc(columns.IsTop).
			OrderDesc(columns.PublishedAt).
			OrderDesc(columns.IsRecommend).
			OrderAsc(columns.Sort).
			OrderDesc(columns.Id)
	case PublicArticleOrderManual:
		return model.
			OrderAsc(columns.Sort).
			OrderDesc(columns.IsTop).
			OrderDesc(columns.IsRecommend).
			OrderDesc(columns.PublishedAt).
			OrderDesc(columns.Id)
	case PublicArticleOrderViews:
		return model.
			OrderDesc(columns.Views).
			OrderDesc(columns.IsTop).
			OrderDesc(columns.IsRecommend).
			OrderAsc(columns.Sort).
			OrderDesc(columns.PublishedAt).
			OrderDesc(columns.Id)
	default:
		return model.
			OrderDesc(columns.IsTop).
			OrderDesc(columns.PublishedAt).
			OrderDesc(columns.Id)
	}
}

// scanArticlePage scans one article model as a paged response.
func (s *serviceImpl) scanArticlePage(
	ctx context.Context,
	model *gdb.Model,
	pageNum int,
	pageSize int,
) (*ArticleListOutput, error) {
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*entitymodel.CmsArticle, 0)
	if err = model.Page(normalizePageNum(pageNum), normalizePageSize(pageSize)).Scan(&list); err != nil {
		return nil, err
	}
	items, err := s.wrapArticleItems(ctx, list)
	if err != nil {
		return nil, err
	}
	return &ArticleListOutput{List: items, Total: total}, nil
}

// wrapArticleItems adds category names to a list of articles.
func (s *serviceImpl) wrapArticleItems(
	ctx context.Context,
	list []*entitymodel.CmsArticle,
) ([]*ArticleItem, error) {
	categoryNames, err := s.categoryNameMap(ctx, list)
	if err != nil {
		return nil, err
	}
	items := make([]*ArticleItem, 0, len(list))
	for _, article := range list {
		article.Content = normalizeImportedArticleContent(article.Content)
		items = append(items, &ArticleItem{
			CmsArticle:   article,
			CategoryName: categoryNames[article.CategoryId],
		})
	}
	return items, nil
}

// wrapArticleItem adds category names to one article.
func (s *serviceImpl) wrapArticleItem(ctx context.Context, article *entitymodel.CmsArticle) (*ArticleItem, error) {
	items, err := s.wrapArticleItems(ctx, []*entitymodel.CmsArticle{article})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, bizerr.NewCode(CodeArticleNotFound)
	}
	return items[0], nil
}

// categoryNameMap resolves category names for article projections.
func (s *serviceImpl) categoryNameMap(
	ctx context.Context,
	list []*entitymodel.CmsArticle,
) (map[int64]string, error) {
	ids := make([]int64, 0, len(list))
	seen := make(map[int64]bool)
	for _, article := range list {
		if article == nil || article.CategoryId <= 0 || seen[article.CategoryId] {
			continue
		}
		ids = append(ids, article.CategoryId)
		seen[article.CategoryId] = true
	}
	result := make(map[int64]string)
	if len(ids) == 0 {
		return result, nil
	}
	columns := dao.CmsCategory.Columns()
	categories := make([]*entitymodel.CmsCategory, 0, len(ids))
	if err := dao.CmsCategory.Ctx(ctx).
		Fields(columns.Id, columns.Name).
		WhereIn(columns.Id, ids).
		Scan(&categories); err != nil {
		return nil, err
	}
	for _, category := range categories {
		result[category.Id] = category.Name
	}
	return result, nil
}

// normalizeImportedArticleContent restores entity-encoded legacy content so
// both the admin editor and public templates receive real HTML.
func normalizeImportedArticleContent(content string) string {
	return html.UnescapeString(content)
}

// ensureCategoryExists verifies a category exists.
func (s *serviceImpl) ensureCategoryExists(ctx context.Context, id int64) error {
	columns := dao.CmsCategory.Columns()
	count, err := dao.CmsCategory.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeCategoryNotFound)
	}
	return nil
}

// ensureCategoryCodeAvailable verifies a category code is unused by other rows.
func (s *serviceImpl) ensureCategoryCodeAvailable(ctx context.Context, code string, currentID int64) error {
	columns := dao.CmsCategory.Columns()
	model := dao.CmsCategory.Ctx(ctx).Where(columns.Code, strings.TrimSpace(code))
	if currentID > 0 {
		model = model.WhereNot(columns.Id, currentID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeCategoryCodeExists)
	}
	return nil
}

// ensureCategoryParentAvailable verifies a category parent exists and cannot
// point to itself or any existing descendant.
func (s *serviceImpl) ensureCategoryParentAvailable(ctx context.Context, categoryID int64, parentID int64) error {
	if parentID <= 0 {
		return nil
	}
	if categoryID > 0 && parentID == categoryID {
		return bizerr.NewCode(CodeCategoryParentInvalid)
	}
	if err := s.ensureCategoryExists(ctx, parentID); err != nil {
		return err
	}
	if categoryID <= 0 {
		return nil
	}

	columns := dao.CmsCategory.Columns()
	categories := make([]*entitymodel.CmsCategory, 0)
	if err := dao.CmsCategory.Ctx(ctx).
		Fields(columns.Id, columns.ParentId).
		Scan(&categories); err != nil {
		return err
	}
	parentByID := make(map[int64]int64, len(categories))
	for _, category := range categories {
		if category == nil {
			continue
		}
		parentByID[category.Id] = category.ParentId
	}

	visited := make(map[int64]struct{}, len(categories))
	for currentID := parentID; currentID > 0; currentID = parentByID[currentID] {
		if currentID == categoryID {
			return bizerr.NewCode(CodeCategoryParentInvalid)
		}
		if _, ok := visited[currentID]; ok {
			return bizerr.NewCode(CodeCategoryParentInvalid)
		}
		visited[currentID] = struct{}{}
		if _, ok := parentByID[currentID]; !ok {
			return nil
		}
	}
	return nil
}

// ensureArticleSlugAvailable verifies an article slug is unused by other rows.
func (s *serviceImpl) ensureArticleSlugAvailable(ctx context.Context, slug string, currentID int64) error {
	columns := dao.CmsArticle.Columns()
	model := dao.CmsArticle.Ctx(ctx).Where(columns.Slug, strings.TrimSpace(slug))
	if currentID > 0 {
		model = model.WhereNot(columns.Id, currentID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeArticleSlugExists)
	}
	return nil
}

// ensureMessageExists verifies a visitor message exists.
func (s *serviceImpl) ensureMessageExists(ctx context.Context, id int64) error {
	columns := dao.CmsMessage.Columns()
	count, err := dao.CmsMessage.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeMessageNotFound)
	}
	return nil
}

// ensureLinkExists verifies a friendly link exists.
func (s *serviceImpl) ensureLinkExists(ctx context.Context, id int64) error {
	columns := dao.CmsLink.Columns()
	count, err := dao.CmsLink.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeLinkNotFound)
	}
	return nil
}

// ensureSlideExists verifies a carousel slide exists.
func (s *serviceImpl) ensureSlideExists(ctx context.Context, id int64) error {
	columns := dao.CmsSlide.Columns()
	count, err := dao.CmsSlide.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeSlideNotFound)
	}
	return nil
}

// buildCategoryTree converts a flat category list into a parent-child tree.
func buildCategoryTree(list []*entitymodel.CmsCategory) []*CategoryItem {
	nodes := make(map[int64]*CategoryItem, len(list))
	roots := make([]*CategoryItem, 0)
	for _, category := range list {
		if category == nil {
			continue
		}
		nodes[category.Id] = &CategoryItem{CmsCategory: category}
	}
	for _, category := range list {
		if category == nil {
			continue
		}
		node := nodes[category.Id]
		if category.ParentId <= 0 {
			roots = append(roots, node)
			continue
		}
		parent := nodes[category.ParentId]
		if parent == nil {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	return roots
}

// publishedAtForStatus resolves publication timestamp for an article status.
func publishedAtForStatus(status int, oldPublishedAt *gtime.Time) *gtime.Time {
	if status != ArticleStatusPublished {
		return nil
	}
	if oldPublishedAt != nil {
		return oldPublishedAt
	}
	return gtime.Now()
}

// normalizePageNum normalizes page numbers for callers that bypass DTO defaults.
func normalizePageNum(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

// normalizePageSize normalizes page size for callers that bypass DTO validation.
func normalizePageSize(value int) int {
	if value < 1 {
		return 10
	}
	if value > 100 {
		return 100
	}
	return value
}

// currentUserID returns the current authenticated user ID as int64.
func (s *serviceImpl) currentUserID(ctx context.Context) int64 {
	return int64(s.bizCtxSvc.Current(ctx).UserID)
}
