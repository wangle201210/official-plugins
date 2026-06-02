// This file implements UIdentity scheduled jobs that only touch plugin-owned
// local tables and do not require external Oracle or LDAP integrations.

package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

// wannaT preserves the lightweight logging behavior of the old WannaT job.
func (s *serviceImpl) wannaT(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		logger.Infof(ctx, "uidentity WannaT tick=%d time=%s", i+1, time.Now().Format(time.DateTime))
		if !sleepWithContext(ctx, time.Second) {
			return ctx.Err()
		}
	}
	logger.Infof(ctx, "uidentity WannaT end")
	return nil
}

// newContainerAccount refreshes account counters for every container.
func (s *serviceImpl) newContainerAccount(ctx context.Context) error {
	tenantID := s.tenantID(ctx)
	stats, err := executeContainerAccountJob(ctx, tenantID)
	if err != nil {
		return err
	}
	logger.Infof(ctx, "uidentity container account counts refreshed tenant=%d updated=%d", tenantID, stats.updateNum)
	return nil
}

// changeContainer moves accounts whose graduated_at equals the current year
// into the tenant-local xy container.
func (s *serviceImpl) changeContainer(ctx context.Context) error {
	tenantID := s.tenantID(ctx)
	stats, err := executeChangeContainerJob(ctx, tenantID, time.Now().Year())
	if err != nil {
		return err
	}
	logger.Infof(ctx, "uidentity graduation container changed tenant=%d updated=%d", tenantID, stats.updateNum)
	return nil
}

type jobRunStats struct {
	createNum                int64
	updateNum                int64
	deleteNum                int64
	updateAccountCount       int64
	updateAccountDetailCount int64
	errNum                   int64
}

func executeContainerAccountJob(ctx context.Context, tenantID int) (jobRunStats, error) {
	accountCols := dao.Account.Columns()
	containerCols := dao.Container.Columns()
	tableAccount := dao.Account.Table()
	tableContainer := dao.Container.Table()
	accountCountSubquery := fmt.Sprintf(
		`(SELECT COUNT(*) FROM %s WHERE %s.%s = %d AND %s.%s IS NULL AND %s.%s = %s.%s)`,
		tableAccount,
		tableAccount,
		accountCols.TenantId,
		tenantID,
		tableAccount,
		accountCols.DeletedAt,
		tableAccount,
		accountCols.ContainerId,
		tableContainer,
		containerCols.Id,
	)
	result, err := dao.Container.Ctx(ctx).
		Where(containerCols.TenantId, tenantID).
		Data(do.Container{AccountCount: gdb.Raw(accountCountSubquery)}).
		Update()
	if err != nil {
		return jobRunStats{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return jobRunStats{}, err
	}
	return jobRunStats{updateNum: affected}, nil
}

func executeChangeContainerJob(ctx context.Context, tenantID int, graduationYear int) (jobRunStats, error) {
	containerID, err := legacyContainerIDByName(ctx, tenantID, legacyContainerAlumni)
	if err != nil {
		return jobRunStats{}, err
	}
	return updateGraduatingAccountContainer(ctx, tenantID, containerID, graduationYear)
}

func legacyContainerIDByName(ctx context.Context, tenantID int, name string) (int64, error) {
	var container *entity.Container
	cols := dao.Container.Columns()
	err := dao.Container.Ctx(ctx).
		Fields(cols.Id).
		Where(cols.TenantId, tenantID).
		Where(cols.Name, strings.TrimSpace(name)).
		Scan(&container)
	if err != nil {
		return 0, err
	}
	if container == nil || container.Id <= 0 {
		return 0, bizerr.NewCode(CodeJobExecutorUnsupported)
	}
	return container.Id, nil
}

func updateGraduatingAccountContainer(ctx context.Context, tenantID int, containerID int64, graduationYear int) (jobRunStats, error) {
	if tenantID < 0 || containerID <= 0 || graduationYear <= 0 {
		return jobRunStats{}, bizerr.NewCode(CodeJobInvalid)
	}
	accountCols := dao.Account.Columns()
	detailCols := dao.AccountDetail.Columns()
	yearText := fmt.Sprintf("%d", graduationYear)
	subquery := dao.AccountDetail.Ctx(ctx).
		Fields(detailCols.AccountId).
		Where(detailCols.TenantId, tenantID).
		Where(detailCols.GraduatedAt, yearText)
	result, err := dao.Account.Ctx(ctx).
		Where(accountCols.TenantId, tenantID).
		WhereIn(accountCols.Id, subquery).
		Data(do.Account{ContainerId: containerID}).
		Update()
	if err != nil {
		return jobRunStats{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return jobRunStats{}, err
	}
	return jobRunStats{updateNum: affected}, nil
}
