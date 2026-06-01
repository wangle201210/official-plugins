// This file implements aggregate UIdentity statistics with database-side
// grouping and batch name projection.

package uidentity

import (
	"context"
	"fmt"

	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
)

type totalByContainer struct {
	ContainerId int64 `json:"containerId"`
	Total       int64 `json:"total"`
}

type totalByAccessModel struct {
	AccessModel string `json:"accessModel"`
	Total       int64  `json:"total"`
}

type totalByPassLevel struct {
	PassLevel int64 `json:"passLevel"`
	Total     int64 `json:"total"`
}

type totalByLoginType struct {
	LoginType string `json:"loginType"`
	Total     int64  `json:"total"`
}

type totalByApp struct {
	AppId int64 `json:"appId"`
	Total int64 `json:"total"`
}

// Stats returns aggregate identity and authentication statistics.
func (s *serviceImpl) Stats(ctx context.Context) (*StatsOutput, error) {
	accountCount, err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").Count()
	if err != nil {
		return nil, err
	}
	appCount, err := s.tenantFilter.Apply(ctx, dao.Application.Ctx(ctx), "").Count()
	if err != nil {
		return nil, err
	}
	casCount, err := s.tenantFilter.Apply(ctx, dao.CasLoginLog.Ctx(ctx), "").Count()
	if err != nil {
		return nil, err
	}
	oauthCount, err := s.tenantFilter.Apply(ctx, dao.OauthLog.Ctx(ctx), "").Count()
	if err != nil {
		return nil, err
	}

	userByContainer, err := s.userByContainer(ctx)
	if err != nil {
		return nil, err
	}
	casByAccountType, err := s.casByAccountType(ctx)
	if err != nil {
		return nil, err
	}
	appByType, err := s.appByType(ctx)
	if err != nil {
		return nil, err
	}
	passLevel, err := s.passLevelStats(ctx)
	if err != nil {
		return nil, err
	}
	loginType, err := s.loginTypeStats(ctx)
	if err != nil {
		return nil, err
	}
	loginApp, err := s.loginAppStats(ctx)
	if err != nil {
		return nil, err
	}

	return &StatsOutput{
		AccountCount:     int64(accountCount),
		AuthCount:        int64(casCount + oauthCount),
		AppCount:         int64(appCount),
		UserByContainer:  userByContainer,
		AppByType:        appByType,
		AuthByType:       []*StatItem{{Name: "cas", Total: int64(casCount)}, {Name: "oauth", Total: int64(oauthCount)}},
		CasByAccountType: casByAccountType,
		PassLevel:        passLevel,
		LoginType:        loginType,
		LoginApp:         loginApp,
	}, nil
}

func (s *serviceImpl) userByContainer(ctx context.Context) ([]*StatItem, error) {
	rows := make([]*totalByContainer, 0)
	accountColumns := dao.Account.Columns()
	err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Fields(accountColumns.ContainerId, "COUNT(*) AS total").
		Group(accountColumns.ContainerId).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return s.containerStats(ctx, rows)
}

func (s *serviceImpl) casByAccountType(ctx context.Context) ([]*StatItem, error) {
	accountColumns := dao.Account.Columns()
	loginColumns := dao.CasLoginLog.Columns()
	rows := make([]*totalByContainer, 0)
	err := s.tenantFilter.Apply(ctx, dao.CasLoginLog.Ctx(ctx), "").
		LeftJoin(dao.Account.Table()+" a", "a."+accountColumns.Id+"="+dao.CasLoginLog.Table()+"."+loginColumns.AccountId).
		Fields("a."+accountColumns.ContainerId+" AS container_id", "COUNT(*) AS total").
		Group("a." + accountColumns.ContainerId).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return s.containerStats(ctx, rows)
}

func (s *serviceImpl) containerStats(ctx context.Context, rows []*totalByContainer) ([]*StatItem, error) {
	containerColumns := dao.Container.Columns()
	result, err := s.tenantFilter.Apply(ctx, dao.Container.Ctx(ctx), "").
		Fields(containerColumns.Id, containerColumns.Alias).
		All()
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(result))
	for _, row := range result {
		names[row[containerColumns.Id].Int64()] = row[containerColumns.Alias].String()
	}
	totals := make(map[int64]int64, len(rows))
	var unknown int64
	for _, row := range rows {
		if _, ok := names[row.ContainerId]; !ok {
			unknown += row.Total
			continue
		}
		totals[row.ContainerId] = row.Total
	}
	items := make([]*StatItem, 0, len(names)+1)
	for id, name := range names {
		items = append(items, &StatItem{Name: name, Total: totals[id]})
	}
	if unknown > 0 {
		items = append(items, &StatItem{Name: "unknown", Total: unknown})
	}
	return items, nil
}

func (s *serviceImpl) appByType(ctx context.Context) ([]*StatItem, error) {
	rows := make([]*totalByAccessModel, 0)
	columns := dao.Application.Columns()
	err := s.tenantFilter.Apply(ctx, dao.Application.Ctx(ctx), "").
		Fields(columns.AccessModel, "COUNT(*) AS total").
		Group(columns.AccessModel).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	items := make([]*StatItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &StatItem{Name: row.AccessModel, Total: row.Total})
	}
	return items, nil
}

func (s *serviceImpl) passLevelStats(ctx context.Context) ([]*StatItem, error) {
	rows := make([]*totalByPassLevel, 0)
	columns := dao.Account.Columns()
	err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Fields(columns.PassLevel, "COUNT(*) AS total").
		Group(columns.PassLevel).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	items := make([]*StatItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &StatItem{Name: fmt.Sprintf("%d", row.PassLevel), Total: row.Total})
	}
	return items, nil
}

func (s *serviceImpl) loginTypeStats(ctx context.Context) ([]*StatItem, error) {
	rows := make([]*totalByLoginType, 0)
	columns := dao.CasLoginLog.Columns()
	err := s.tenantFilter.Apply(ctx, dao.CasLoginLog.Ctx(ctx), "").
		Fields(columns.LoginType, "COUNT(*) AS total").
		Group(columns.LoginType).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	items := make([]*StatItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &StatItem{Name: row.LoginType, Total: row.Total})
	}
	return items, nil
}

func (s *serviceImpl) loginAppStats(ctx context.Context) ([]*StatItem, error) {
	rows := make([]*totalByApp, 0)
	columns := dao.CasLoginLog.Columns()
	err := s.tenantFilter.Apply(ctx, dao.CasLoginLog.Ctx(ctx), "").
		Fields(columns.AppId, "COUNT(*) AS total").
		Group(columns.AppId).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	appIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.AppId > 0 {
			appIDs = append(appIDs, row.AppId)
		}
	}
	appColumns := dao.Application.Columns()
	names, err := s.nameMap(ctx, dao.Application.Ctx(ctx), appColumns.Id, appColumns.Alias, appIDs)
	if err != nil {
		return nil, err
	}
	items := make([]*StatItem, 0, len(rows))
	for _, row := range rows {
		name := names[row.AppId]
		if name == "" {
			name = "unknown"
		}
		items = append(items, &StatItem{Name: name, Total: row.Total})
	}
	return items, nil
}
