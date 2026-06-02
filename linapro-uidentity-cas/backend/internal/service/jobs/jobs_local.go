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
	mover, cleanup := s.changeContainerMover(ctx)
	defer cleanup()
	stats, err := executeChangeContainerJob(ctx, tenantID, time.Now().Year(), mover)
	if err != nil {
		return err
	}
	logger.Infof(ctx, "uidentity graduation container changed tenant=%d updated=%d failed=%d", tenantID, stats.updateNum, stats.errNum)
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

type changeContainerMover func(ctx context.Context, account *entity.Account, target *entity.Container) error

func (s *serviceImpl) changeContainerMover(ctx context.Context) (changeContainerMover, func()) {
	cfg, cfgErr := s.ldapJobConfig(ctx)
	if cfgErr != nil {
		return func(context.Context, *entity.Account, *entity.Container) error {
			return cfgErr
		}, func() {}
	}
	conn, connErr := openLDAPConn(cfg)
	if connErr != nil {
		return func(context.Context, *entity.Account, *entity.Container) error {
			return connErr
		}, func() {}
	}
	return func(ctx context.Context, account *entity.Account, target *entity.Container) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return moveLDAPAccountToContainer(conn, cfg, account, target)
		}, func() {
			if err := conn.Close(); err != nil {
				logger.Warningf(ctx, "uidentity ldap connection close failed err=%v", err)
			}
		}
}

func executeChangeContainerJob(ctx context.Context, tenantID int, graduationYear int, mover changeContainerMover) (jobRunStats, error) {
	if mover == nil {
		return jobRunStats{}, bizerr.NewCode(CodeJobInvalid)
	}
	target, err := legacyContainerByName(ctx, tenantID, legacyContainerAlumni)
	if err != nil {
		return jobRunStats{}, err
	}
	stats := jobRunStats{}
	for page := 0; ; page++ {
		accounts, err := graduatingAccountPage(ctx, tenantID, graduationYear, page, defaultPageSize)
		if err != nil {
			return jobRunStats{}, err
		}
		if len(accounts) == 0 {
			break
		}
		accountIDs, failed := movableChangeContainerAccountIDs(ctx, accounts, target, mover)
		stats.errNum += failed
		if len(accountIDs) > 0 {
			updated, err := updateAccountsContainerByIDs(ctx, tenantID, accountIDs, target.Id)
			if err != nil {
				return jobRunStats{}, err
			}
			stats.updateNum += updated
		}
		if len(accounts) < defaultPageSize {
			break
		}
	}
	return stats, nil
}

func legacyContainerIDByName(ctx context.Context, tenantID int, name string) (int64, error) {
	container, err := legacyContainerByName(ctx, tenantID, name)
	if err != nil {
		return 0, err
	}
	return container.Id, nil
}

func legacyContainerByName(ctx context.Context, tenantID int, name string) (*entity.Container, error) {
	var container *entity.Container
	cols := dao.Container.Columns()
	err := dao.Container.Ctx(ctx).
		Fields(cols.Id, cols.Name).
		Where(cols.TenantId, tenantID).
		Where(cols.Name, strings.TrimSpace(name)).
		Scan(&container)
	if err != nil {
		return nil, err
	}
	if container == nil || container.Id <= 0 {
		return nil, bizerr.NewCode(CodeJobExecutorUnsupported)
	}
	return container, nil
}

func graduatingAccountPage(ctx context.Context, tenantID int, graduationYear int, page int, pageSize int) ([]*entity.Account, error) {
	detailCols := dao.AccountDetail.Columns()
	yearText := fmt.Sprintf("%d", graduationYear)
	var details []*entity.AccountDetail
	err := dao.AccountDetail.Ctx(ctx).
		Fields(detailCols.AccountId).
		Where(detailCols.TenantId, tenantID).
		Where(detailCols.GraduatedAt, yearText).
		OrderAsc(detailCols.AccountId).
		Offset(page * pageSize).
		Limit(pageSize).
		Scan(&details)
	if err != nil {
		return nil, err
	}
	accountIDs := make(map[int64]struct{}, len(details))
	for _, detail := range details {
		if detail != nil && detail.AccountId > 0 {
			accountIDs[detail.AccountId] = struct{}{}
		}
	}
	if len(accountIDs) == 0 {
		return nil, nil
	}
	accountsByID, err := legacyAccountsByIDs(ctx, tenantID, int64sFromSet(accountIDs))
	if err != nil {
		return nil, err
	}
	accounts := make([]*entity.Account, 0, len(details))
	for _, detail := range details {
		if detail == nil || detail.AccountId <= 0 {
			continue
		}
		accounts = append(accounts, accountsByID[detail.AccountId])
	}
	return accounts, nil
}

func legacyAccountsByIDs(ctx context.Context, tenantID int, accountIDs []int64) (map[int64]*entity.Account, error) {
	result := make(map[int64]*entity.Account, len(accountIDs))
	if len(accountIDs) == 0 {
		return result, nil
	}
	var accounts []*entity.Account
	err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().TenantId, tenantID).
		WhereIn(dao.Account.Columns().Id, accountIDs).
		Scan(&accounts)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account != nil {
			result[account.Id] = account
		}
	}
	return result, nil
}

func movableChangeContainerAccountIDs(ctx context.Context, accounts []*entity.Account, target *entity.Container, mover changeContainerMover) ([]int64, int64) {
	accountIDs := make([]int64, 0, len(accounts))
	var failed int64
	for _, account := range accounts {
		if account == nil || account.Id <= 0 {
			failed++
			continue
		}
		if err := mover(ctx, account, target); err != nil {
			failed++
			logger.Warningf(ctx, "uidentity change container ldap move failed number=%s err=%v", account.Number, err)
			continue
		}
		accountIDs = append(accountIDs, account.Id)
	}
	return accountIDs, failed
}

func updateAccountsContainerByIDs(ctx context.Context, tenantID int, accountIDs []int64, containerID int64) (int64, error) {
	if tenantID < 0 || len(accountIDs) == 0 || containerID <= 0 {
		return 0, nil
	}
	result, err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().TenantId, tenantID).
		WhereIn(dao.Account.Columns().Id, accountIDs).
		Data(do.Account{ContainerId: containerID}).
		Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
