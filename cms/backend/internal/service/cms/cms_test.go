// This file verifies CMS service publication boundaries and business errors.

package cms

import (
	"context"
	"html"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/dialect"
	"lina-core/pkg/plugin/capability/bizctx"
	plugincontract "lina-core/pkg/plugin/capability/contract"
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

// newTestCMSServiceForUser creates a CMS service that sees a fixed user ID in
// plugin business context, matching authenticated management requests.
func newTestCMSServiceForUser(userID int) Service {
	svc, err := New(bizctx.New(cmsTestBizCtx{userID: userID}))
	if err != nil {
		panic(err)
	}
	return svc
}

// cmsTestBizCtx provides a minimal plugin business context for service tests.
type cmsTestBizCtx struct {
	userID int
}

// Current returns the static business context configured for the test case.
func (c cmsTestBizCtx) Current(context.Context) plugincontract.CurrentContext {
	return plugincontract.CurrentContext{UserID: c.userID}
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

// TestListPublicMessagesRequiresApprovalAndHonorsDisabledSwitch verifies public
// message listing is enabled by default, honors explicit disable, and exposes
// only approved rows.
func TestListPublicMessagesRequiresApprovalAndHonorsDisabledSwitch(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)
	insertCMSMessage(t, ctx, "approved-before-switch", MessageStatusApproved, "before")

	svc := newTestCMSService()
	openByDefault, err := svc.ListPublicMessages(ctx, PublicMessageListInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list public CMS messages while switch is enabled by default: %v", err)
	}
	if openByDefault.Total != 1 || len(openByDefault.List) != 1 {
		t.Fatalf("expected one approved message by default, got total=%d len=%d", openByDefault.Total, len(openByDefault.List))
	}

	columns := dao.CmsSite.Columns()
	if _, err = dao.CmsSite.Ctx(ctx).
		Where(columns.SiteKey, "default").
		Data(do.CmsSite{ShowMessages: StatusDisabled}).
		Update(); err != nil {
		t.Fatalf("disable public CMS message display: %v", err)
	}
	closed, err := svc.ListPublicMessages(ctx, PublicMessageListInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list public CMS messages after disabling switch: %v", err)
	}
	if closed.Total != 0 || len(closed.List) != 0 {
		t.Fatalf("expected disabled public message switch to hide messages, got total=%d len=%d", closed.Total, len(closed.List))
	}

	if _, err = dao.CmsSite.Ctx(ctx).
		Where(columns.SiteKey, "default").
		Data(do.CmsSite{ShowMessages: StatusEnabled}).
		Update(); err != nil {
		t.Fatalf("enable public CMS message display: %v", err)
	}
	insertCMSMessage(t, ctx, "pending-message", MessageStatusPending, "pending reply")
	insertCMSMessage(t, ctx, "approved-after-switch", MessageStatusApproved, "approved reply")
	insertCMSMessage(t, ctx, "rejected-message", MessageStatusRejected, "rejected reply")

	open, err := svc.ListPublicMessages(ctx, PublicMessageListInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list public CMS messages while switch is enabled: %v", err)
	}
	if open.Total != 2 || len(open.List) != 2 {
		t.Fatalf("expected two approved messages, got total=%d len=%d", open.Total, len(open.List))
	}
	for _, item := range open.List {
		if item.Status != MessageStatusApproved {
			t.Fatalf("expected only approved public messages, got status=%d", item.Status)
		}
		if item.Reply == "" {
			t.Fatalf("expected public message reply to be preserved")
		}
	}
}

// TestCMSMockDataDefaultShowsApprovedMessages verifies CMS reference-site mock
// data ships with visible public messages and the expected moderation mix.
func TestCMSMockDataDefaultShowsApprovedMessages(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)
	loadCMSMockData(t, ctx)

	assertCMSSeedMessages(t, ctx, "mock")
}

// TestCMSInstallStarterContentIsDelivered verifies normal plugin installation
// SQL ships a usable starter site without requiring optional mock data.
func TestCMSInstallStarterContentIsDelivered(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)
	loadCMSStarterContent(t, ctx)

	assertCMSSeedMessages(t, ctx, "starter")
	assertCMSArticleBodiesAreRich(t, ctx, "starter")
	assertCMSExpertSummariesAreShort(t, ctx, "starter")

	articleCount, err := dao.CmsArticle.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count CMS starter articles: %v", err)
	}
	messageCount, err := dao.CmsMessage.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count CMS starter messages: %v", err)
	}
	loadCMSStarterContent(t, ctx)
	articleCountAfter, err := dao.CmsArticle.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count CMS starter articles after reload: %v", err)
	}
	messageCountAfter, err := dao.CmsMessage.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count CMS starter messages after reload: %v", err)
	}
	if articleCountAfter != articleCount || messageCountAfter != messageCount {
		t.Fatalf("expected starter SQL to be idempotent, articles %d->%d messages %d->%d", articleCount, articleCountAfter, messageCount, messageCountAfter)
	}
}

// TestClearSiteDataRemovesCMSBusinessContent verifies the management clear
// action gives users a fresh CMS state after reviewing starter content.
func TestClearSiteDataRemovesCMSBusinessContent(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)
	loadCMSStarterContent(t, ctx)
	insertCMSTag(t, ctx, "starter-tag")

	if err := newTestCMSService().ClearSiteData(ctx); err != nil {
		t.Fatalf("clear CMS site data: %v", err)
	}

	for _, table := range cmsContentTables() {
		count, err := dao.CmsSite.DB().Model(table).Ctx(ctx).Count()
		if err != nil {
			t.Fatalf("count cleared CMS table %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected CMS table %s to be empty, got %d", table, count)
		}
	}

	site, err := newTestCMSService().GetSite(ctx, false)
	if err != nil {
		t.Fatalf("get reset CMS site: %v", err)
	}
	if site.SiteKey != "default" || site.Name != "LinaPro CMS" {
		t.Fatalf("expected blank default site, got key=%q name=%q", site.SiteKey, site.Name)
	}
	if site.Status != StatusEnabled || site.ShowMessages != StatusEnabled {
		t.Fatalf("expected reset site enabled with messages visible, got status=%d showMessages=%d", site.Status, site.ShowMessages)
	}
	if site.Logo != "" || site.Slogan != "" || site.Description != "" {
		t.Fatalf("expected reset site to drop starter profile fields, got logo=%q slogan=%q description=%q", site.Logo, site.Slogan, site.Description)
	}
}

// TestLoadSampleDataReplacesCMSContent verifies the management sample loading
// action clears custom CMS content and restores the delivered starter site.
func TestLoadSampleDataReplacesCMSContent(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	customCategoryID := insertCMSCategory(t, ctx, "custom-only", StatusEnabled)
	insertCMSArticle(t, ctx, customCategoryID, "custom-only-article", ArticleStatusPublished)
	insertCMSMessage(t, ctx, "Custom message", MessageStatusApproved, "")
	insertCMSTag(t, ctx, "custom-only-tag")

	const sampleUserID = 88
	if err := newTestCMSServiceForUser(sampleUserID).LoadSampleData(ctx); err != nil {
		t.Fatalf("load CMS sample data: %v", err)
	}

	assertCMSSeedMessages(t, ctx, "starter")
	assertCMSArticleBodiesAreRich(t, ctx, "starter")
	assertCMSExpertSummariesAreShort(t, ctx, "starter")

	for table, slug := range map[string]string{
		dao.CmsCategory.Table():   "custom-only",
		dao.CmsArticle.Table():    "custom-only-article",
		dao.CmsArticleTag.Table(): "custom-only-tag",
	} {
		count, err := dao.CmsSite.DB().Model(table).Ctx(ctx).Where("slug", slug).Count()
		if table == dao.CmsCategory.Table() {
			count, err = dao.CmsSite.DB().Model(table).Ctx(ctx).Where("code", slug).Count()
		}
		if err != nil {
			t.Fatalf("count custom CMS row in %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected custom CMS row %q in %s to be cleared, got %d", slug, table, count)
		}
	}

	site, err := newTestCMSService().GetSite(ctx, false)
	if err != nil {
		t.Fatalf("get sample CMS site: %v", err)
	}
	if site.Name != "启明先进材料产业研究院" || site.Logo != "/static/logo.svg" {
		t.Fatalf("expected starter site profile, got name=%q logo=%q", site.Name, site.Logo)
	}
	if site.CreatedBy != sampleUserID || site.UpdatedBy != sampleUserID {
		t.Fatalf("expected starter site maintainer %d, got createdBy=%d updatedBy=%d", sampleUserID, site.CreatedBy, site.UpdatedBy)
	}

	articleColumns := dao.CmsArticle.Columns()
	maintainedArticles, err := dao.CmsArticle.Ctx(ctx).
		Where(articleColumns.CreatedBy, sampleUserID).
		Where(articleColumns.UpdatedBy, sampleUserID).
		Count()
	if err != nil {
		t.Fatalf("count sample CMS article maintainers: %v", err)
	}
	if maintainedArticles == 0 {
		t.Fatalf("expected sample CMS articles to be stamped with user %d", sampleUserID)
	}
}

// assertCMSSeedMessages verifies starter/mock message seed visibility and
// moderation mix through both direct DB state and the public service boundary.
func assertCMSSeedMessages(t *testing.T, ctx context.Context, seedName string) {
	t.Helper()

	siteColumns := dao.CmsSite.Columns()
	showMessages, err := dao.CmsSite.Ctx(ctx).
		Where(siteColumns.SiteKey, "default").
		Value(siteColumns.ShowMessages)
	if err != nil {
		t.Fatalf("read CMS %s site message switch: %v", seedName, err)
	}
	if showMessages.Int() != StatusEnabled {
		t.Fatalf("expected CMS %s site message switch enabled, got %d", seedName, showMessages.Int())
	}

	messageColumns := dao.CmsMessage.Columns()
	total, err := dao.CmsMessage.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count CMS %s messages: %v", seedName, err)
	}
	approved, err := dao.CmsMessage.Ctx(ctx).
		Where(messageColumns.Status, MessageStatusApproved).
		Count()
	if err != nil {
		t.Fatalf("count approved CMS %s messages: %v", seedName, err)
	}
	if total != 5 || approved != 3 {
		t.Fatalf("expected five CMS %s messages with three approved, got total=%d approved=%d", seedName, total, approved)
	}

	out, err := newTestCMSService().ListPublicMessages(ctx, PublicMessageListInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list public CMS %s messages: %v", seedName, err)
	}
	if out.Total != 3 || len(out.List) != 3 {
		t.Fatalf("expected three public approved %s messages, got total=%d len=%d", seedName, out.Total, len(out.List))
	}
}

// TestCMSMockDataArticleBodiesAreRich verifies every reference-site article has
// enough body text for a realistic CMS demonstration.
func TestCMSMockDataArticleBodiesAreRich(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)
	loadCMSMockData(t, ctx)

	assertCMSArticleBodiesAreRich(t, ctx, "mock")
	assertCMSExpertSummariesAreShort(t, ctx, "mock")
}

// assertCMSArticleBodiesAreRich verifies every seeded article has enough body
// text for a realistic CMS demonstration.
func assertCMSArticleBodiesAreRich(t *testing.T, ctx context.Context, seedName string) {
	t.Helper()

	columns := dao.CmsArticle.Columns()
	var articles []mockArticleContent
	if err := dao.CmsArticle.Ctx(ctx).
		Fields(columns.Slug, columns.Title, columns.Summary, columns.Cover, columns.Content).
		OrderAsc(columns.Slug).
		Scan(&articles); err != nil {
		t.Fatalf("scan CMS %s articles: %v", seedName, err)
	}
	if len(articles) == 0 {
		t.Fatalf("expected CMS %s articles", seedName)
	}
	for _, article := range articles {
		if length := cmsPlainTextLength(article.Content); length < 300 {
			t.Fatalf("expected article %s body text to be at least 300 chars, got %d", article.Slug, length)
		}
		if strings.TrimSpace(article.Summary) == "" {
			t.Fatalf("expected article %s summary to be populated", article.Slug)
		}
		if strings.TrimSpace(article.Cover) == "" {
			t.Fatalf("expected article %s cover to be populated", article.Slug)
		}
		content := html.UnescapeString(article.Content)
		if strings.Contains(content, "习近平") || strings.Contains(content, "123") || strings.Contains(content, "<br/>") {
			t.Fatalf("article %s still contains placeholder or unrelated content", article.Slug)
		}
	}
}

// assertCMSExpertSummariesAreShort verifies the public expert cards keep
// compact summaries that fit the CMS reference site layout.
func assertCMSExpertSummariesAreShort(t *testing.T, ctx context.Context, seedName string) {
	t.Helper()

	var experts []mockArticleContent
	if err := g.DB().GetScan(ctx, &experts, `
		SELECT article."slug", article."title", article."summary"
		FROM plugin_cms_article AS article
		INNER JOIN plugin_cms_category AS category ON category."id" = article."category_id"
		WHERE category."code" = '46'
		ORDER BY article."slug"
	`); err != nil {
		t.Fatalf("scan CMS %s expert summaries: %v", seedName, err)
	}
	if len(experts) == 0 {
		t.Fatalf("expected CMS %s expert articles", seedName)
	}
	expectedSummaries := map[string]string{
		"cms-72":  "教授",
		"cms-73":  "博士",
		"cms-80":  "研究员",
		"cms-102": "副教授",
		"cms-103": "高工",
		"cms-106": "超纤专家",
	}
	for _, expert := range experts {
		if length := len([]rune(strings.TrimSpace(expert.Summary))); length > 6 {
			t.Fatalf("expected CMS %s expert %s summary to be at most 6 chars, got %d: %q", seedName, expert.Slug, length, expert.Summary)
		}
		if want := expectedSummaries[expert.Slug]; want != "" && expert.Summary != want {
			t.Fatalf("expected CMS %s expert %s summary %q, got %q", seedName, expert.Slug, want, expert.Summary)
		}
	}
}

// mockArticleContent captures the fields needed for mock data quality checks.
type mockArticleContent struct {
	Slug    string `orm:"slug"`
	Title   string `orm:"title"`
	Summary string `orm:"summary"`
	Cover   string `orm:"cover"`
	Content string `orm:"content"`
}

// TestNormalizeImportedArticleContentIsIdempotent verifies already-decoded
// admin-authored HTML is preserved.
func TestNormalizeImportedArticleContentIsIdempotent(t *testing.T) {
	const content = `<p><strong>Already HTML</strong></p>`
	if got := normalizeImportedArticleContent(content); got != content {
		t.Fatalf("expected normalized HTML %q, got %q", content, got)
	}
}

// setupSQLiteCMSDB points generated CMS DAOs at an isolated PostgreSQL schema
// and executes the schema SQL needed by service tests. Starter content stays
// opt-in so tests that need an empty CMS database can keep that boundary.
func setupSQLiteCMSDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		baseLink = "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
	}
	schemaName := "cms_test_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: baseLink}},
	}); err != nil {
		t.Fatalf("configure PostgreSQL CMS database failed: %v", err)
	}
	setupDB := g.DB()
	if _, err := setupDB.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create isolated CMS test schema failed: %v", err)
	}
	if err := setupDB.Close(ctx); err != nil {
		t.Fatalf("close setup PostgreSQL CMS database failed: %v", err)
	}
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: cmsTestLinkWithSearchPath(baseLink, schemaName)}},
	}); err != nil {
		t.Fatalf("configure PostgreSQL CMS schema database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if _, err := db.Exec(ctx, "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE"); err != nil {
			t.Errorf("drop isolated CMS test schema failed: %v", err)
		}
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Errorf("close PostgreSQL CMS database failed: %v", closeErr)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore GoFrame database config failed: %v", err)
		}
	})

	createCMSHostDictTables(t, ctx, db)
	for _, sqlName := range []string{"001-cms-schema.sql", "002-cms-message-visibility.sql"} {
		executeCMSManifestSQLFile(t, ctx, "sql", sqlName)
	}
}

// loadCMSStarterContent executes the optional starter sample data.
func loadCMSStarterContent(t *testing.T, ctx context.Context) {
	t.Helper()

	executeCMSManifestSQLFile(t, ctx, "sql", "mock-data", "002-cms-starter-content.sql")
}

// loadCMSMockData executes the reference-site mock data against the isolated
// CMS test schema so tests can assert delivered demo content quality.
func loadCMSMockData(t *testing.T, ctx context.Context) {
	t.Helper()

	executeCMSManifestSQLFile(t, ctx, "sql", "mock-data", "001-cms-mock-data.sql")
}

// executeCMSManifestSQLFile runs one CMS manifest SQL file against the current
// isolated test schema.
func executeCMSManifestSQLFile(t *testing.T, ctx context.Context, pathParts ...string) {
	t.Helper()

	sqlPathParts := append([]string{"..", "..", "..", "..", "manifest"}, pathParts...)
	sqlPath := filepath.Join(sqlPathParts...)
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read CMS manifest SQL %s failed: %v", filepath.Join(pathParts...), err)
	}
	db := g.DB()
	for _, statement := range dialect.SplitSQLStatements(string(content)) {
		if _, err = db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute CMS manifest SQL %s failed: %v\nSQL:\n%s", filepath.Join(pathParts...), err, statement)
		}
	}
}

// cmsTestLinkWithSearchPath appends an isolated PostgreSQL schema to a GoFrame
// pgsql link without changing the caller-provided database endpoint.
func cmsTestLinkWithSearchPath(link string, schemaName string) string {
	separator := "?"
	if strings.Contains(link, "?") {
		separator = "&"
	}
	return link + separator + "search_path=" + schemaName
}

// cmsPlainTextLength returns visible text length after unescaping imported HTML
// entities and removing markup tags.
func cmsPlainTextLength(content string) int {
	decoded := html.UnescapeString(content)
	var text strings.Builder
	inTag := false
	for _, r := range decoded {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				text.WriteRune(r)
			}
		}
	}
	return len([]rune(strings.Join(strings.Fields(text.String()), "")))
}

// createCMSHostDictTables creates the minimal host dictionary tables required
// by the CMS plugin installation seed data.
func createCMSHostDictTables(t *testing.T, ctx context.Context, db gdb.DB) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS sys_dict_type (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(128) NOT NULL DEFAULT '',
			type VARCHAR(128) NOT NULL DEFAULT '',
			status SMALLINT NOT NULL DEFAULT 1,
			is_builtin SMALLINT NOT NULL DEFAULT 0,
			remark VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
			UNIQUE (type)
		)`,
		`CREATE TABLE IF NOT EXISTS sys_dict_data (
			id BIGSERIAL PRIMARY KEY,
			dict_type VARCHAR(128) NOT NULL DEFAULT '',
			label VARCHAR(128) NOT NULL DEFAULT '',
			value VARCHAR(128) NOT NULL DEFAULT '',
			sort INTEGER NOT NULL DEFAULT 0,
			tag_style VARCHAR(32) NOT NULL DEFAULT '',
			css_class VARCHAR(64) NOT NULL DEFAULT '',
			status SMALLINT NOT NULL DEFAULT 1,
			is_builtin SMALLINT NOT NULL DEFAULT 0,
			remark VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
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

// insertCMSMessage inserts one visitor message with an explicit moderation status.
func insertCMSMessage(t *testing.T, ctx context.Context, content string, status int, reply string) int64 {
	t.Helper()

	id, err := dao.CmsMessage.Ctx(ctx).Data(do.CmsMessage{
		Name:    "Visitor",
		Content: content,
		Reply:   reply,
		Status:  status,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert CMS message %s: %v", content, err)
	}
	return id
}

// insertCMSTag inserts one CMS tag and returns its ID.
func insertCMSTag(t *testing.T, ctx context.Context, slug string) int64 {
	t.Helper()

	id, err := dao.CmsArticleTag.Ctx(ctx).Data(do.CmsArticleTag{
		Name:   slug,
		Slug:   slug,
		Status: StatusEnabled,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert CMS tag %s: %v", slug, err)
	}
	return id
}
