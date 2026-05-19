// This file verifies CMS service publication boundaries and business errors.

package cms

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/dialect"
	"lina-core/pkg/pluginservice/bizctx"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
)

// newTestCMSService creates a CMS service with an explicit test bizctx adapter.
func newTestCMSService() Service {
	svc, err := New(bizctx.New(nil))
	if err != nil {
		panic(err)
	}
	return svc
}

// TestPublicArticlesFilterDraftsAndDisabledCategories verifies public article
// lists expose only published content under enabled categories.
func TestPublicArticlesFilterDraftsAndDisabledCategories(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	enabledCategoryID := insertCMSCategory(t, ctx, "visible", StatusEnabled)
	disabledCategoryID := insertCMSCategory(t, ctx, "hidden", StatusDisabled)
	insertCMSArticle(t, ctx, enabledCategoryID, "visible-article", ArticleStatusPublished)
	insertCMSArticle(t, ctx, enabledCategoryID, "draft-article", ArticleStatusDraft)
	insertCMSArticle(t, ctx, disabledCategoryID, "hidden-category-article", ArticleStatusPublished)

	out, err := newTestCMSService().ListPublicArticles(ctx, PublicArticleListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list public CMS articles: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 {
		t.Fatalf("expected one public article, got total=%d len=%d", out.Total, len(out.List))
	}
	if out.List[0].Slug != "visible-article" {
		t.Fatalf("expected visible article slug, got %q", out.List[0].Slug)
	}
}

// TestPublicArticleDetailRejectsDrafts verifies public detail lookup never
// exposes draft content.
func TestPublicArticleDetailRejectsDrafts(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "drafts", StatusEnabled)
	insertCMSArticle(t, ctx, categoryID, "draft-only", ArticleStatusDraft)

	_, err := newTestCMSService().GetPublicArticleBySlug(ctx, "draft-only")
	if !bizerr.Is(err, CodePublicContentNotFound) {
		t.Fatalf("expected public content not found error, got %v", err)
	}
}

// TestCreateArticleRejectsDuplicateSlug verifies duplicate slugs are returned
// as structured business errors before the database unique key is hit.
func TestCreateArticleRejectsDuplicateSlug(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "news", StatusEnabled)
	insertCMSArticle(t, ctx, categoryID, "duplicate-slug", ArticleStatusPublished)

	_, err := newTestCMSService().CreateArticle(ctx, ArticleSaveInput{
		CategoryId: categoryID,
		Title:      "Duplicate",
		Slug:       "duplicate-slug",
		Content:    "<p>Duplicate</p>",
		Status:     ArticleStatusDraft,
	})
	if !bizerr.Is(err, CodeArticleSlugExists) {
		t.Fatalf("expected duplicate slug business error, got %v", err)
	}
}

// TestGetArticleDecodesImportedEntityHTML verifies management details return
// restored body HTML instead of entity text.
func TestGetArticleDecodesImportedEntityHTML(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "about", StatusEnabled)
	articleID := insertCMSArticleWithContent(
		t,
		ctx,
		categoryID,
		"company-profile",
		ArticleStatusPublished,
		`&lt;p&gt;&lt;span style=&quot;font-family: SimSun;&quot;&gt;公司简介&lt;/span&gt;&lt;/p&gt;`,
	)

	article, err := newTestCMSService().GetArticle(ctx, articleID)
	if err != nil {
		t.Fatalf("get CMS article detail: %v", err)
	}
	expected := `<p><span style="font-family: SimSun;">公司简介</span></p>`
	if article.Content != expected {
		t.Fatalf("expected decoded article content %q, got %q", expected, article.Content)
	}
}

// TestListArticlesDecodesImportedEntityHTML verifies management lists expose
// decoded HTML when the article content field is requested by callers.
func TestListArticlesDecodesImportedEntityHTML(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "about-list", StatusEnabled)
	insertCMSArticleWithContent(
		t,
		ctx,
		categoryID,
		"company-profile-list",
		ArticleStatusPublished,
		`&lt;p&gt;公司简介&lt;/p&gt;`,
	)

	out, err := newTestCMSService().ListArticles(ctx, ArticleListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list CMS articles: %v", err)
	}
	if len(out.List) != 1 {
		t.Fatalf("expected one article, got %d", len(out.List))
	}
	if out.List[0].Content != "<p>公司简介</p>" {
		t.Fatalf("expected decoded list content, got %q", out.List[0].Content)
	}
}

// TestListArticlesFiltersByCategoryType verifies management content modules
// can be split by CMS model semantics.
func TestListArticlesFiltersByCategoryType(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	singleID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:   "single-model",
		status: StatusEnabled,
		typeID: CategoryTypeSingle,
	})
	listID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:   "list-model",
		status: StatusEnabled,
		typeID: CategoryTypeList,
	})
	insertCMSArticle(t, ctx, singleID, "single-content", ArticleStatusPublished)
	insertCMSArticle(t, ctx, listID, "list-content", ArticleStatusPublished)

	out, err := newTestCMSService().ListArticles(ctx, ArticleListInput{
		CategoryType: CategoryTypeSingle,
		PageNum:      1,
		PageSize:     20,
	})
	if err != nil {
		t.Fatalf("list CMS single model articles: %v", err)
	}
	if out.Total != 1 || out.List[0].Slug != "single-content" {
		t.Fatalf("expected single model content only, got total=%d list=%v", out.Total, out.List)
	}
}

// TestListArticlesIncludesChildCategories verifies selecting a parent category
// in management shows content from the whole category group.
func TestListArticlesIncludesChildCategories(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	parentID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:   "news-parent",
		status: StatusEnabled,
		typeID: CategoryTypeList,
	})
	childID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:     "news-child",
		parentID: parentID,
		status:   StatusEnabled,
		typeID:   CategoryTypeList,
	})
	insertCMSArticle(t, ctx, childID, "child-news", ArticleStatusPublished)

	out, err := newTestCMSService().ListArticles(ctx, ArticleListInput{
		CategoryId:      parentID,
		IncludeChildren: true,
		PageNum:         1,
		PageSize:        20,
	})
	if err != nil {
		t.Fatalf("list CMS category group articles: %v", err)
	}
	if out.Total != 1 || out.List[0].Slug != "child-news" {
		t.Fatalf("expected child category content from parent filter, got total=%d list=%v", out.Total, out.List)
	}
}

// TestUpdateCategoryRejectsParentCycle verifies category updates cannot assign
// a descendant as the parent category.
func TestUpdateCategoryRejectsParentCycle(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	parentID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:   "cycle-parent",
		status: StatusEnabled,
		typeID: CategoryTypeList,
	})
	childID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:     "cycle-child",
		parentID: parentID,
		status:   StatusEnabled,
		typeID:   CategoryTypeList,
	})

	err := newTestCMSService().UpdateCategory(ctx, CategorySaveInput{
		Id:       parentID,
		ParentId: childID,
		Code:     "cycle-parent",
		Name:     "cycle-parent",
		Type:     CategoryTypeList,
		Status:   StatusEnabled,
	})
	if !bizerr.Is(err, CodeCategoryParentInvalid) {
		t.Fatalf("expected category parent invalid error, got %v", err)
	}
}

// TestListArticlesIncludeChildrenSkipsCategoryCycles verifies article filters
// still terminate if existing category data is already corrupted.
func TestListArticlesIncludeChildrenSkipsCategoryCycles(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	parentID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:   "corrupt-parent",
		status: StatusEnabled,
		typeID: CategoryTypeList,
	})
	childID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:     "corrupt-child",
		parentID: parentID,
		status:   StatusEnabled,
		typeID:   CategoryTypeList,
	})
	insertCMSArticle(t, ctx, childID, "corrupt-child-news", ArticleStatusPublished)

	columns := dao.CmsCategory.Columns()
	if _, err := dao.CmsCategory.Ctx(ctx).
		Where(columns.Id, parentID).
		Data(do.CmsCategory{ParentId: childID}).
		Update(); err != nil {
		t.Fatalf("create corrupt CMS category cycle: %v", err)
	}

	out, err := newTestCMSService().ListArticles(ctx, ArticleListInput{
		CategoryId:      parentID,
		IncludeChildren: true,
		PageNum:         1,
		PageSize:        20,
	})
	if err != nil {
		t.Fatalf("list CMS category group articles with corrupt cycle: %v", err)
	}
	if out.Total != 1 || out.List[0].Slug != "corrupt-child-news" {
		t.Fatalf("expected child category content from cyclic parent filter, got total=%d list=%v", out.Total, out.List)
	}
}

// TestListPublicArticlesOrderManual verifies public templates can request
// manual order before pagination is applied.
func TestListPublicArticlesOrderManual(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "manual", StatusEnabled)
	insertCMSArticleWithOptions(t, ctx, cmsArticleOptions{
		categoryID:  categoryID,
		slug:        "late",
		status:      ArticleStatusPublished,
		sort:        30,
		publishedAt: gtime.NewFromTime(time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)),
	})
	insertCMSArticleWithOptions(t, ctx, cmsArticleOptions{
		categoryID:  categoryID,
		slug:        "first",
		status:      ArticleStatusPublished,
		sort:        10,
		publishedAt: gtime.NewFromTime(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)),
	})
	insertCMSArticleWithOptions(t, ctx, cmsArticleOptions{
		categoryID:  categoryID,
		slug:        "second",
		status:      ArticleStatusPublished,
		sort:        20,
		publishedAt: gtime.NewFromTime(time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)),
	})

	out, err := newTestCMSService().ListPublicArticles(ctx, PublicArticleListInput{
		PageNum:    1,
		PageSize:   2,
		CategoryId: categoryID,
		Order:      PublicArticleOrderManual,
	})
	if err != nil {
		t.Fatalf("list public CMS articles by manual order: %v", err)
	}
	if out.Total != 3 || len(out.List) != 2 {
		t.Fatalf("expected first page of three public articles, got total=%d len=%d", out.Total, len(out.List))
	}
	if out.List[0].Slug != "first" || out.List[1].Slug != "second" {
		t.Fatalf("expected manual order before pagination, got %q then %q", out.List[0].Slug, out.List[1].Slug)
	}
}

// TestNormalizePublicArticleOrderAcceptsTemplateValues verifies documented
// template order values map directly to service order values.
func TestNormalizePublicArticleOrderAcceptsTemplateValues(t *testing.T) {
	if got := NormalizePublicArticleOrder("manual"); got != PublicArticleOrderManual {
		t.Fatalf("expected manual order value, got %q", got)
	}
	if got := NormalizePublicArticleOrder("views"); got != PublicArticleOrderViews {
		t.Fatalf("expected views order value, got %q", got)
	}
}

// TestListPublicArticlesSearchesBodyContent verifies the public search page can
// find articles whose keyword only appears in the body.
func TestListPublicArticlesSearchesBodyContent(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "search-body", StatusEnabled)
	insertCMSArticleWithOptions(t, ctx, cmsArticleOptions{
		categoryID: categoryID,
		slug:       "body-hit",
		status:     ArticleStatusPublished,
		content:    "<p>前沿算力网络专题报告</p>",
		title:      "技术观察",
	})
	insertCMSArticleWithOptions(t, ctx, cmsArticleOptions{
		categoryID: categoryID,
		slug:       "body-miss",
		status:     ArticleStatusPublished,
		content:    "<p>普通产业动态</p>",
		title:      "产业动态",
	})

	out, err := newTestCMSService().ListPublicArticles(ctx, PublicArticleListInput{
		PageNum:    1,
		PageSize:   10,
		Keyword:    "算力网络",
		CategoryId: categoryID,
	})
	if err != nil {
		t.Fatalf("search public CMS articles by body content: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 {
		t.Fatalf("expected one body search hit, got total=%d len=%d", out.Total, len(out.List))
	}
	if out.List[0].Slug != "body-hit" {
		t.Fatalf("expected body-hit article, got %q", out.List[0].Slug)
	}
}

// TestNormalizeImportedArticleContentIsIdempotent verifies already-decoded
// admin-authored HTML is preserved.
func TestNormalizeImportedArticleContentIsIdempotent(t *testing.T) {
	const content = `<p><strong>Already HTML</strong></p>`
	if got := normalizeImportedArticleContent(content); got != content {
		t.Fatalf("expected normalized HTML %q, got %q", content, got)
	}
}

// setupSQLiteCMSDB points generated CMS DAOs at a temporary SQLite database and
// executes the plugin installation SQL through the shared dialect translator.
func setupSQLiteCMSDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	dbPath := filepath.Join(t.TempDir(), "cms.db")
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "sqlite::@file(" + dbPath + ")"}},
	}); err != nil {
		t.Fatalf("configure SQLite CMS database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Errorf("close SQLite CMS database failed: %v", closeErr)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore GoFrame database config failed: %v", err)
		}
	})

	dbDialect, err := dialect.From("sqlite::@file(" + dbPath + ")")
	if err != nil {
		t.Fatalf("resolve SQLite dialect failed: %v", err)
	}
	createCMSHostDictTables(t, ctx, db)
	sqlPath := filepath.Join("..", "..", "..", "..", "manifest", "sql", "001-cms-schema.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read CMS schema SQL failed: %v", err)
	}
	translated, err := dbDialect.TranslateDDL(ctx, sqlPath, string(content))
	if err != nil {
		t.Fatalf("translate CMS schema SQL failed: %v", err)
	}
	for _, statement := range dialect.SplitSQLStatements(translated) {
		if _, err = db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute CMS schema SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
}

// createCMSHostDictTables creates the minimal host dictionary tables required
// by the CMS plugin installation seed data.
func createCMSHostDictTables(t *testing.T, ctx context.Context, db gdb.DB) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS sys_dict_type (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			type TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			is_builtin INTEGER NOT NULL DEFAULT 0,
			remark TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			UNIQUE (type)
		)`,
		`CREATE TABLE IF NOT EXISTS sys_dict_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dict_type TEXT NOT NULL DEFAULT '',
			label TEXT NOT NULL DEFAULT '',
			value TEXT NOT NULL DEFAULT '',
			sort INTEGER NOT NULL DEFAULT 0,
			tag_style TEXT NOT NULL DEFAULT '',
			css_class TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			is_builtin INTEGER NOT NULL DEFAULT 0,
			remark TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			UNIQUE (dict_type, value)
		)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			t.Fatalf("create CMS test host dict table failed: %v", err)
		}
	}
}

// insertCMSCategory inserts one category and returns its ID.
func insertCMSCategory(t *testing.T, ctx context.Context, code string, status int) int64 {
	t.Helper()
	return insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{
		code:   code,
		status: status,
		typeID: CategoryTypeList,
	})
}

// cmsCategoryOptions defines test category insert options.
type cmsCategoryOptions struct {
	code     string
	parentID int64
	status   int
	typeID   int
}

// insertCMSCategoryWithOptions inserts one category with explicit test options.
func insertCMSCategoryWithOptions(t *testing.T, ctx context.Context, opts cmsCategoryOptions) int64 {
	t.Helper()

	id, err := dao.CmsCategory.Ctx(ctx).Data(do.CmsCategory{
		Code:     opts.code,
		Name:     opts.code,
		ParentId: opts.parentID,
		Type:     opts.typeID,
		Status:   opts.status,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert CMS category %s: %v", opts.code, err)
	}
	return id
}

// insertCMSArticle inserts one article and returns its ID.
func insertCMSArticle(t *testing.T, ctx context.Context, categoryID int64, slug string, status int) int64 {
	t.Helper()
	return insertCMSArticleWithContent(t, ctx, categoryID, slug, status, "<p>"+slug+"</p>")
}

// cmsArticleOptions defines explicit test article insert options.
type cmsArticleOptions struct {
	categoryID  int64
	slug        string
	title       string
	status      int
	content     string
	sort        int
	publishedAt *gtime.Time
}

// insertCMSArticleWithContent inserts one article with explicit content and
// returns its ID.
func insertCMSArticleWithContent(
	t *testing.T,
	ctx context.Context,
	categoryID int64,
	slug string,
	status int,
	content string,
) int64 {
	t.Helper()
	return insertCMSArticleWithOptions(t, ctx, cmsArticleOptions{
		categoryID: categoryID,
		slug:       slug,
		status:     status,
		content:    content,
	})
}

// insertCMSArticleWithOptions inserts one article using explicit test options.
func insertCMSArticleWithOptions(
	t *testing.T,
	ctx context.Context,
	opts cmsArticleOptions,
) int64 {
	t.Helper()
	content := opts.content
	if content == "" {
		content = "<p>" + opts.slug + "</p>"
	}
	title := opts.title
	if title == "" {
		title = opts.slug
	}

	id, err := dao.CmsArticle.Ctx(ctx).Data(do.CmsArticle{
		CategoryId:  opts.categoryID,
		Title:       title,
		Slug:        opts.slug,
		Content:     content,
		Sort:        opts.sort,
		Status:      opts.status,
		PublishedAt: publishedAtForStatus(opts.status, opts.publishedAt),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert CMS article %s: %v", opts.slug, err)
	}
	return id
}
