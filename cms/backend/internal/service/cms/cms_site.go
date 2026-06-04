// This file implements CMS site settings, data reset, sample loading, and purge helpers.

package cms

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/dialect"
	cmsplugin "lina-plugin-cms"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
)

// GetSite returns the default site settings and optionally enforces public visibility.
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

// UpdateSite persists management changes for the default CMS site settings.
func (s *serviceImpl) UpdateSite(ctx context.Context, in SiteUpdateInput) error {
	columns := dao.CmsSite.Columns()
	site, err := s.GetSite(ctx, false)
	if err != nil {
		return err
	}
	_, err = dao.CmsSite.Ctx(ctx).Where(columns.Id, site.Id).Data(do.CmsSite{Name: in.Name, Logo: in.Logo, Weixin: in.Weixin, Domain: in.Domain, Slogan: in.Slogan, Keywords: in.Keywords, Description: in.Description, Icp: in.Icp, Contact: in.Contact, Phone: in.Phone, Email: in.Email, Address: in.Address, Status: in.Status, ShowMessages: in.ShowMessages, UpdatedBy: s.currentUserID(ctx)}).Update()
	return err
}

// ClearSiteData removes CMS business content and recreates the default enabled site record.
func (s *serviceImpl) ClearSiteData(ctx context.Context) error {
	userID := s.currentUserID(ctx)
	return dao.CmsSite.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := clearCMSData(ctx, tx); err != nil {
			return err
		}
		_, err := tx.Model(dao.CmsSite.Table()).Ctx(ctx).Data(do.CmsSite{SiteKey: "default", Name: "LinaPro CMS", Status: StatusEnabled, ShowMessages: StatusEnabled, CreatedBy: userID, UpdatedBy: userID}).Insert()
		return err
	})
}

// LoadSampleData replaces CMS business content with the embedded optional sample dataset.
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

// PurgeStorageData removes CMS plugin storage data during an uninstall purge.
func PurgeStorageData(ctx context.Context) error {
	tables := append(cmsContentTables(), dao.CmsSite.Table())
	for _, table := range tables {
		if _, err := dao.CmsSite.DB().Exec(ctx, "DELETE FROM "+table); err != nil {
			return err
		}
	}
	return nil
}

// PurgeStorageData removes CMS plugin storage data during an uninstall purge.
func (s *serviceImpl) PurgeStorageData(ctx context.Context) error {
	return PurgeStorageData(ctx)
}

// clearCMSData deletes CMS content tables and the site record inside the supplied transaction.
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

// executeStarterContentSQL runs the embedded optional starter dataset SQL statements.
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

// markSampleDataMaintainer stamps loaded sample rows with the current management user.
func markSampleDataMaintainer(ctx context.Context, tx gdb.TX, userID int64) error {
	if userID <= 0 {
		return nil
	}
	updates := []struct {
		table string
		data  any
	}{{table: dao.CmsMessage.Table(), data: do.CmsMessage{CreatedBy: userID, UpdatedBy: userID}}, {table: dao.CmsSlide.Table(), data: do.CmsSlide{CreatedBy: userID, UpdatedBy: userID}}, {table: dao.CmsLink.Table(), data: do.CmsLink{CreatedBy: userID, UpdatedBy: userID}}, {table: dao.CmsArticleTag.Table(), data: do.CmsArticleTag{CreatedBy: userID, UpdatedBy: userID}}, {table: dao.CmsArticle.Table(), data: do.CmsArticle{CreatedBy: userID, UpdatedBy: userID}}, {table: dao.CmsCategory.Table(), data: do.CmsCategory{CreatedBy: userID, UpdatedBy: userID}}, {table: dao.CmsSite.Table(), data: do.CmsSite{CreatedBy: userID, UpdatedBy: userID}}}
	for _, update := range updates {
		if _, err := tx.Model(update.table).Ctx(ctx).Data(update.data).Update(); err != nil {
			return err
		}
	}
	return nil
}

// cmsContentTables returns CMS business tables in an order safe for full-content cleanup.
func cmsContentTables() []string {
	return []string{dao.CmsMessage.Table(), dao.CmsSlide.Table(), dao.CmsLink.Table(), dao.CmsArticleTag.Table(), dao.CmsArticle.Table(), dao.CmsCategory.Table()}
}
