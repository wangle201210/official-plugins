// This file verifies CMS article scheduled publishing rules and batch
// operations against an isolated PostgreSQL schema.

package cms

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
)

// cmsWallClock formats a time as the wall-clock string used for assertions;
// the timestamp column has no timezone, so instants cannot round-trip and
// tests compare wall-clock values instead.
func cmsWallClock(value *gtime.Time) string {
	if value == nil {
		return ""
	}
	return value.Layout("2006-01-02 15:04:05")
}

// cmsWallClockOfMillis formats a Unix millisecond timestamp as the local
// wall-clock string written to the database.
func cmsWallClockOfMillis(millis int64) string {
	return cmsWallClock(gtime.New(time.UnixMilli(millis)))
}

// loadCMSArticleRow loads one article row including soft-deleted filtering.
func loadCMSArticleRow(t *testing.T, ctx context.Context, id int64) *entitymodel.CmsArticle {
	t.Helper()

	var article *entitymodel.CmsArticle
	if err := dao.CmsArticle.Ctx(ctx).Where(dao.CmsArticle.Columns().Id, id).Scan(&article); err != nil {
		t.Fatalf("load CMS article %d: %v", id, err)
	}
	return article
}

// TestCreateArticleWithFuturePublishedAtSchedulesIt verifies an explicit
// future publication time is stored and hides the article from every public
// read path until the time is reached.
func TestCreateArticleWithFuturePublishedAtSchedulesIt(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "schedule", StatusEnabled)
	futureMillis := time.Now().Add(time.Hour).UnixMilli()
	svc := newTestCMSService()
	id, err := svc.CreateArticle(ctx, ArticleSaveInput{
		CategoryId:  categoryID,
		Title:       "Scheduled",
		Slug:        "scheduled-article",
		Content:     "<p>Scheduled</p>",
		Status:      ArticleStatusPublished,
		PublishedAt: &futureMillis,
	})
	if err != nil {
		t.Fatalf("create scheduled CMS article: %v", err)
	}

	row := loadCMSArticleRow(t, ctx, id)
	if row == nil || row.PublishedAt == nil {
		t.Fatalf("expected stored publication time for scheduled article")
	}
	if got := cmsWallClock(row.PublishedAt); got != cmsWallClockOfMillis(futureMillis) {
		t.Fatalf("expected published_at %s, got %s", cmsWallClockOfMillis(futureMillis), got)
	}

	publicOut, err := svc.ListPublicArticles(ctx, PublicArticleListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list public CMS articles: %v", err)
	}
	if publicOut.Total != 0 {
		t.Fatalf("expected scheduled article hidden from public list, got total=%d", publicOut.Total)
	}
	if _, err = svc.GetPublicArticleBySlug(ctx, "scheduled-article"); !bizerr.Is(err, CodePublicContentNotFound) {
		t.Fatalf("expected public content not found for scheduled article, got %v", err)
	}

	managementOut, err := svc.ListArticles(ctx, ArticleListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list management CMS articles: %v", err)
	}
	if managementOut.Total != 1 {
		t.Fatalf("expected scheduled article in management list, got total=%d", managementOut.Total)
	}
}

// TestScheduledArticleBecomesVisibleAfterPublishedAt verifies an article whose
// publication time already passed is served by public reads.
func TestScheduledArticleBecomesVisibleAfterPublishedAt(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "released", StatusEnabled)
	pastMillis := time.Now().Add(-time.Hour).UnixMilli()
	svc := newTestCMSService()
	if _, err := svc.CreateArticle(ctx, ArticleSaveInput{
		CategoryId:  categoryID,
		Title:       "Released",
		Slug:        "released-article",
		Content:     "<p>Released</p>",
		Status:      ArticleStatusPublished,
		PublishedAt: &pastMillis,
	}); err != nil {
		t.Fatalf("create released CMS article: %v", err)
	}

	item, err := svc.GetPublicArticleBySlug(ctx, "released-article")
	if err != nil {
		t.Fatalf("expected released article public detail, got %v", err)
	}
	if item.Slug != "released-article" {
		t.Fatalf("expected released article slug, got %q", item.Slug)
	}
}

// TestUpdateArticlePublishedAtDefaultRules verifies first publishes stamp the
// current time, republish saves keep it, and explicit values override it.
func TestUpdateArticlePublishedAtDefaultRules(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "rules", StatusEnabled)
	svc := newTestCMSService()
	id, err := svc.CreateArticle(ctx, ArticleSaveInput{
		CategoryId: categoryID,
		Title:      "Rules",
		Slug:       "publish-rules",
		Content:    "<p>Rules</p>",
		Status:     ArticleStatusDraft,
	})
	if err != nil {
		t.Fatalf("create draft CMS article: %v", err)
	}
	if row := loadCMSArticleRow(t, ctx, id); row.PublishedAt != nil {
		t.Fatalf("expected draft without publication time, got %v", row.PublishedAt)
	}

	saveInput := ArticleSaveInput{
		Id:         id,
		CategoryId: categoryID,
		Title:      "Rules",
		Slug:       "publish-rules",
		Content:    "<p>Rules</p>",
		Status:     ArticleStatusPublished,
	}
	if err = svc.UpdateArticle(ctx, saveInput); err != nil {
		t.Fatalf("publish CMS article: %v", err)
	}
	firstPublish := loadCMSArticleRow(t, ctx, id).PublishedAt
	if firstPublish == nil {
		t.Fatalf("expected first publish to stamp the current time")
	}

	if err = svc.UpdateArticle(ctx, saveInput); err != nil {
		t.Fatalf("republish CMS article: %v", err)
	}
	if kept := loadCMSArticleRow(t, ctx, id).PublishedAt; kept == nil || !kept.Equal(firstPublish) {
		t.Fatalf("expected republish to keep publication time %v, got %v", firstPublish, kept)
	}

	explicitMillis := time.Now().Add(-30 * time.Minute).Round(time.Second).UnixMilli()
	saveInput.PublishedAt = &explicitMillis
	if err = svc.UpdateArticle(ctx, saveInput); err != nil {
		t.Fatalf("update CMS article publication time: %v", err)
	}
	if got := cmsWallClock(loadCMSArticleRow(t, ctx, id).PublishedAt); got != cmsWallClockOfMillis(explicitMillis) {
		t.Fatalf("expected explicit published_at %s, got %s", cmsWallClockOfMillis(explicitMillis), got)
	}
}

// TestBatchUpdateArticleStatusPublishesAndKeepsTimes verifies batch publishing
// fills empty publication times and keeps existing ones.
func TestBatchUpdateArticleStatusPublishesAndKeepsTimes(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "batch", StatusEnabled)
	draftID := insertCMSArticle(t, ctx, categoryID, "batch-draft", ArticleStatusDraft)
	keptTime := gtime.New(time.Now().Add(-2 * time.Hour).Round(time.Second))
	timedID := insertCMSArticleWithOptions(t, ctx, cmsArticleOptions{
		categoryID:  categoryID,
		slug:        "batch-timed",
		status:      ArticleStatusDraft,
		publishedAt: keptTime,
	})
	if _, err := dao.CmsArticle.Ctx(ctx).Where(dao.CmsArticle.Columns().Id, timedID).Data(do.CmsArticle{PublishedAt: keptTime}).Update(); err != nil {
		t.Fatalf("seed draft publication time: %v", err)
	}

	if err := newTestCMSService().BatchUpdateArticleStatus(ctx, ArticleBatchStatusInput{Ids: []int64{draftID, timedID}, Status: ArticleStatusPublished}); err != nil {
		t.Fatalf("batch publish CMS articles: %v", err)
	}

	draftRow := loadCMSArticleRow(t, ctx, draftID)
	if draftRow.Status != ArticleStatusPublished || draftRow.PublishedAt == nil {
		t.Fatalf("expected published draft with stamped time, got status=%d publishedAt=%v", draftRow.Status, draftRow.PublishedAt)
	}
	timedRow := loadCMSArticleRow(t, ctx, timedID)
	if timedRow.Status != ArticleStatusPublished || cmsWallClock(timedRow.PublishedAt) != cmsWallClock(keptTime) {
		t.Fatalf("expected kept publication time %v, got status=%d publishedAt=%v", keptTime, timedRow.Status, timedRow.PublishedAt)
	}
}

// TestBatchUpdateArticleStatusUnpublishKeepsPublishedAt verifies batch
// unpublishing turns articles into drafts without clearing publication times.
func TestBatchUpdateArticleStatusUnpublishKeepsPublishedAt(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "unpublish", StatusEnabled)
	id := insertCMSArticle(t, ctx, categoryID, "unpublish-target", ArticleStatusPublished)
	before := loadCMSArticleRow(t, ctx, id).PublishedAt
	if before == nil {
		t.Fatalf("expected published article to carry a publication time")
	}

	if err := newTestCMSService().BatchUpdateArticleStatus(ctx, ArticleBatchStatusInput{Ids: []int64{id}, Status: ArticleStatusDraft}); err != nil {
		t.Fatalf("batch unpublish CMS articles: %v", err)
	}
	row := loadCMSArticleRow(t, ctx, id)
	if row.Status != ArticleStatusDraft {
		t.Fatalf("expected draft status after unpublish, got %d", row.Status)
	}
	if row.PublishedAt == nil || !row.PublishedAt.Equal(before) {
		t.Fatalf("expected kept publication time %v, got %v", before, row.PublishedAt)
	}
}

// TestBatchUpdateArticleStatusRejectsMissingTargets verifies a batch with any
// unknown ID fails as a whole and leaves all targets untouched.
func TestBatchUpdateArticleStatusRejectsMissingTargets(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "reject", StatusEnabled)
	id := insertCMSArticle(t, ctx, categoryID, "reject-target", ArticleStatusDraft)

	err := newTestCMSService().BatchUpdateArticleStatus(ctx, ArticleBatchStatusInput{Ids: []int64{id, id + 999}, Status: ArticleStatusPublished})
	if !bizerr.Is(err, CodeArticleNotFound) {
		t.Fatalf("expected article not found error, got %v", err)
	}
	if row := loadCMSArticleRow(t, ctx, id); row.Status != ArticleStatusDraft {
		t.Fatalf("expected untouched draft after rejected batch, got status=%d", row.Status)
	}
}

// TestBatchDeleteArticlesRemovesTargets verifies batch deletes soft delete all
// targets and reject batches containing unknown IDs.
func TestBatchDeleteArticlesRemovesTargets(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "remove", StatusEnabled)
	firstID := insertCMSArticle(t, ctx, categoryID, "remove-first", ArticleStatusPublished)
	secondID := insertCMSArticle(t, ctx, categoryID, "remove-second", ArticleStatusDraft)
	svc := newTestCMSService()

	if err := svc.BatchDeleteArticles(ctx, []int64{firstID, firstID + 999}); !bizerr.Is(err, CodeArticleNotFound) {
		t.Fatalf("expected article not found error for unknown target, got %v", err)
	}

	if err := svc.BatchDeleteArticles(ctx, []int64{firstID, secondID}); err != nil {
		t.Fatalf("batch delete CMS articles: %v", err)
	}
	out, err := svc.ListArticles(ctx, ArticleListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list management CMS articles: %v", err)
	}
	if out.Total != 0 {
		t.Fatalf("expected empty management list after batch delete, got total=%d", out.Total)
	}
}
