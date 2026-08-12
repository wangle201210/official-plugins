// irontransport_state.go assembles bounded player and operator projections.
package irontransport

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

type teamProjection struct {
	ID                      int64      `json:"id"`
	Name                    string     `json:"name"`
	CreatorID               int64      `json:"creatorId"`
	CreatorName             string     `json:"creatorName"`
	MemberCount             int        `json:"memberCount"`
	TotalContributionMeters int64      `json:"totalContributionMeters"`
	LastActiveAt            *time.Time `json:"lastActiveAt"`
	CreatedAt               *time.Time `json:"createdAt"`
	Status                  string     `json:"status"`
	InvalidatedAt           *time.Time `json:"invalidatedAt"`
	InvalidReason           string     `json:"invalidReason"`
}

type memberProjection struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"userId"`
	Name               string     `json:"name"`
	Avatar             string     `json:"avatar"`
	Role               string     `json:"role"`
	ContributionMeters int64      `json:"contributionMeters"`
	JoinedAt           *time.Time `json:"joinedAt"`
}

type reportRecord struct {
	ID                 int64      `json:"id"`
	TeamID             int64      `json:"teamId"`
	MemberID           int64      `json:"memberId"`
	UserID             int64      `json:"userId"`
	TeamName           string     `json:"teamName"`
	UserName           string     `json:"userName"`
	ActivityDate       string     `json:"activityDate"`
	StartLat           *float64   `json:"startLat"`
	StartLng           *float64   `json:"startLng"`
	EndLat             float64    `json:"endLat"`
	EndLng             float64    `json:"endLng"`
	SampledAt          *time.Time `json:"sampledAt"`
	AcceptedAt         *time.Time `json:"acceptedAt"`
	ContributionMeters int64      `json:"contributionMeters"`
	UserTotalMeters    int64      `json:"userTotalMeters"`
	TeamTotalMeters    int64      `json:"teamTotalMeters"`
	DailyReportCount   int        `json:"dailyReportCount"`
}

func (s *serviceImpl) State(ctx context.Context, playerID int64) (*State, error) {
	now := time.Now()
	if _, err := s.ExpireInactive(ctx, now); err != nil {
		return nil, err
	}
	rows := make([]*teamProjection, 0, maxEffectiveTeams)
	if err := baseTeamProjection(ctx).
		Where("t."+dao.TransportTeam.Columns().Status, teamStatusEffective).
		Where("t."+dao.TransportTeam.Columns().Visible, 1).
		Where("t." + dao.TransportTeam.Columns().DeletedAt + " IS NULL").
		OrderDesc("t." + dao.TransportTeam.Columns().LastActiveAt).
		OrderDesc("t." + dao.TransportTeam.Columns().Id).
		Limit(maxEffectiveTeams).Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	state := &State{
		Explanation:       "每人只能加入一个团，每天最多上报12次位置；每次贡献为上次成功上报点到当前点的直线距离，连续72小时无成功操作的团自动失效。",
		MaxEffectiveTeams: maxEffectiveTeams, DailyReportLimit: dailyReportLimit,
		Teams: make([]*Team, 0, len(rows)),
	}
	for _, row := range rows {
		team := teamFromProjection(row, row.ID > 0 && playerID > 0)
		team.Mine = false
		state.Teams = append(state.Teams, team)
	}
	member, err := activeMembership(ctx, playerID, false)
	if err != nil {
		return nil, err
	}
	if member != nil {
		for _, team := range state.Teams {
			if team.ID == member.TeamId {
				team.Mine = true
				team.MyContributionMeters = member.TotalContributionMeters
				team.HasReportBaseline = member.LastReportAt != nil
				state.MyTeam = team
				break
			}
		}
	}
	state.TodayReportCount, err = dao.TransportReport.Ctx(ctx).
		Where(do.TransportReport{UserId: playerID, ActivityDate: activityday.Date(now)}).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	state.TodayReportRemaining = dailyReportLimit - state.TodayReportCount
	if state.TodayReportRemaining < 0 {
		state.TodayReportRemaining = 0
	}
	return state, nil
}

func (s *serviceImpl) GetTeam(ctx context.Context, playerID, teamID int64) (*Team, error) {
	if teamID <= 0 {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	if _, err := s.ExpireInactive(ctx, time.Now()); err != nil {
		return nil, err
	}
	var row *teamProjection
	if err := baseTeamProjection(ctx).
		Where("t."+dao.TransportTeam.Columns().Id, teamID).
		Where("t."+dao.TransportTeam.Columns().Status, teamStatusEffective).
		Where("t."+dao.TransportTeam.Columns().Visible, 1).
		Where("t." + dao.TransportTeam.Columns().DeletedAt + " IS NULL").Limit(1).Scan(&row); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeTeamNotFound)
	}
	member, err := activeMembership(ctx, playerID, false)
	if err != nil {
		return nil, err
	}
	team := teamFromProjection(row, member != nil && member.TeamId == teamID)
	if team.Mine {
		team.MyContributionMeters = member.TotalContributionMeters
		team.HasReportBaseline = member.LastReportAt != nil
	}
	return team, nil
}

func (s *serviceImpl) ListMembers(ctx context.Context, playerID, teamID int64, in *PageInput) (*MemberList, error) {
	team, err := s.GetTeam(ctx, playerID, teamID)
	if err != nil {
		return nil, err
	}
	if !team.Mine {
		return nil, bizerr.NewCode(CodeTeamNotFound)
	}
	pageNum, pageSize := normalizePagination(in)
	total, err := dao.TransportMember.Ctx(ctx).
		Where(dao.TransportMember.Columns().TeamId, teamID).
		Where(dao.TransportMember.Columns().LeftAt + " IS NULL").Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	model := dao.TransportMember.Ctx(ctx).
		Unscoped().
		As("m").
		LeftJoin(dao.User.Table()+" u", "u."+dao.User.Columns().Id+"=m."+dao.TransportMember.Columns().UserId+" AND u."+dao.User.Columns().DeletedAt+" IS NULL").
		Fields(
			"m."+dao.TransportMember.Columns().Id+" AS id",
			"m."+dao.TransportMember.Columns().UserId+" AS user_id",
			"COALESCE(u."+dao.User.Columns().Nickname+", '') AS name",
			"COALESCE(u."+dao.User.Columns().Avatar+", '') AS avatar",
			"m."+dao.TransportMember.Columns().Role+" AS role",
			"m."+dao.TransportMember.Columns().TotalContributionMeters+" AS contribution_meters",
			"m."+dao.TransportMember.Columns().JoinedAt+" AS joined_at",
		).
		Where("m."+dao.TransportMember.Columns().TeamId, teamID).
		Where("m." + dao.TransportMember.Columns().LeftAt + " IS NULL")
	rows := make([]*memberProjection, 0, pageSize)
	if err = model.OrderAsc("m."+dao.TransportMember.Columns().JoinedAt).
		OrderAsc("m."+dao.TransportMember.Columns().Id).
		Page(pageNum, pageSize).Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	list := make([]*Member, 0, len(rows))
	for _, row := range rows {
		list = append(list, &Member{
			ID: row.ID, UserID: row.UserID, Name: row.Name, Avatar: row.Avatar,
			Role: row.Role, ContributionMeters: row.ContributionMeters, JoinedAt: row.JoinedAt,
		})
	}
	return &MemberList{List: list, Total: total}, nil
}

func (s *serviceImpl) ListMyReports(ctx context.Context, playerID int64, in *PageInput) (*ReportList, error) {
	pageNum, pageSize := normalizePagination(in)
	total, err := dao.TransportReport.Ctx(ctx).Where(dao.TransportReport.Columns().UserId, playerID).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	model := baseReportProjection(ctx).Where("r."+dao.TransportReport.Columns().UserId, playerID)
	rows := make([]*reportRecord, 0, pageSize)
	if err = model.OrderDesc("r."+dao.TransportReport.Columns().AcceptedAt).
		OrderDesc("r."+dao.TransportReport.Columns().Id).Page(pageNum, pageSize).Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	list := make([]*Report, 0, len(rows))
	for _, row := range rows {
		list = append(list, reportFromRecord(row))
	}
	return &ReportList{List: list, Total: total}, nil
}

func (s *serviceImpl) ListAdminTeams(ctx context.Context, in *AdminTeamListInput) (*AdminTeamList, error) {
	if _, err := s.ExpireInactive(ctx, time.Now()); err != nil {
		return nil, err
	}
	if in == nil {
		in = &AdminTeamListInput{}
	}
	if in.Status != "" && in.Status != string(teamStatusEffective) && in.Status != string(teamStatusInvalid) {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	pageNum, pageSize := normalizePagination(&PageInput{PageNum: in.PageNum, PageSize: in.PageSize})
	countModel := dao.TransportTeam.Ctx(ctx)
	model := baseTeamProjection(ctx).Where("t." + dao.TransportTeam.Columns().DeletedAt + " IS NULL")
	if name := strings.TrimSpace(in.Name); name != "" {
		countModel = countModel.WhereLike(dao.TransportTeam.Columns().Name, "%"+name+"%")
		model = model.WhereLike("t."+dao.TransportTeam.Columns().Name, "%"+name+"%")
	}
	if in.Status != "" {
		countModel = countModel.Where(dao.TransportTeam.Columns().Status, in.Status)
		model = model.Where("t."+dao.TransportTeam.Columns().Status, in.Status)
	}
	total, err := countModel.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	rows := make([]*teamProjection, 0, pageSize)
	if err = model.OrderDesc("t."+dao.TransportTeam.Columns().CreatedAt).
		OrderDesc("t."+dao.TransportTeam.Columns().Id).Page(pageNum, pageSize).Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	list := make([]*AdminTeam, 0, len(rows))
	for _, row := range rows {
		list = append(list, &AdminTeam{
			Team: teamFromProjection(row, false), Status: row.Status,
			InvalidatedAt: row.InvalidatedAt, InvalidReason: row.InvalidReason,
		})
	}
	return &AdminTeamList{List: list, Total: total}, nil
}

func (s *serviceImpl) AdminStats(ctx context.Context, activityDate string) (*Stats, error) {
	if _, err := s.ExpireInactive(ctx, time.Now()); err != nil {
		return nil, err
	}
	activityDate = strings.TrimSpace(activityDate)
	if !validActivityDate(activityDate) {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	effective, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Status: teamStatusEffective}).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	invalid, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Status: teamStatusInvalid}).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	members, err := dao.TransportMember.Ctx(ctx).Where(dao.TransportMember.Columns().LeftAt + " IS NULL").Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	reportModel := dao.TransportReport.Ctx(ctx)
	if activityDate != "" {
		reportModel = reportModel.Where(dao.TransportReport.Columns().ActivityDate, activityDate)
	}
	reportCount, err := reportModel.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	var aggregate struct {
		Meters int64 `json:"meters"`
	}
	if err = reportModel.Fields("COALESCE(SUM(" + dao.TransportReport.Columns().ContributionMeters + "), 0) AS meters").Scan(&aggregate); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return &Stats{EffectiveTeamCount: effective, InvalidTeamCount: invalid, ActiveMemberCount: members, ReportCount: reportCount, ContributionMeters: aggregate.Meters}, nil
}

func (s *serviceImpl) ListAdminReports(ctx context.Context, in *AdminReportListInput) (*AdminReportList, error) {
	if in == nil {
		in = &AdminReportListInput{}
	}
	if !validActivityDate(in.ActivityDate) {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	pageNum, pageSize := normalizePagination(&PageInput{PageNum: in.PageNum, PageSize: in.PageSize})
	countModel := dao.TransportReport.Ctx(ctx)
	model := baseReportProjection(ctx)
	if in.TeamID > 0 {
		countModel = countModel.Where(dao.TransportReport.Columns().TeamId, in.TeamID)
		model = model.Where("r."+dao.TransportReport.Columns().TeamId, in.TeamID)
	}
	if in.UserID > 0 {
		countModel = countModel.Where(dao.TransportReport.Columns().UserId, in.UserID)
		model = model.Where("r."+dao.TransportReport.Columns().UserId, in.UserID)
	}
	if value := strings.TrimSpace(in.ActivityDate); value != "" {
		countModel = countModel.Where(dao.TransportReport.Columns().ActivityDate, value)
		model = model.Where("r."+dao.TransportReport.Columns().ActivityDate, value)
	}
	total, err := countModel.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	rows := make([]*reportRecord, 0, pageSize)
	if err = model.OrderDesc("r."+dao.TransportReport.Columns().AcceptedAt).
		OrderDesc("r."+dao.TransportReport.Columns().Id).Page(pageNum, pageSize).Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	list := make([]*AdminReport, 0, len(rows))
	for _, row := range rows {
		list = append(list, &AdminReport{Report: reportFromRecord(row), TeamName: row.TeamName, MemberID: row.MemberID, UserID: row.UserID, UserName: row.UserName})
	}
	return &AdminReportList{List: list, Total: total}, nil
}

func findReportByRequest(ctx context.Context, playerID int64, requestID string) (*reportRecord, error) {
	var row *reportRecord
	if err := baseReportProjection(ctx).
		Where("r."+dao.TransportReport.Columns().UserId, playerID).
		Where("r."+dao.TransportReport.Columns().RequestId, requestID).Limit(1).Scan(&row); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return row, nil
}

func baseTeamProjection(ctx context.Context) *gdb.Model {
	return dao.TransportTeam.Ctx(ctx).
		Unscoped().
		As("t").
		LeftJoin(dao.User.Table()+" u", "u."+dao.User.Columns().Id+"=t."+dao.TransportTeam.Columns().LeaderUserId+" AND u."+dao.User.Columns().DeletedAt+" IS NULL").
		Fields(
			"t."+dao.TransportTeam.Columns().Id+" AS id",
			"t."+dao.TransportTeam.Columns().Name+" AS name",
			"t."+dao.TransportTeam.Columns().LeaderUserId+" AS creator_id",
			"COALESCE(u."+dao.User.Columns().Nickname+", '') AS creator_name",
			"t."+dao.TransportTeam.Columns().MemberCount+" AS member_count",
			"t."+dao.TransportTeam.Columns().TotalContributionMeters+" AS total_contribution_meters",
			"t."+dao.TransportTeam.Columns().LastActiveAt+" AS last_active_at",
			"t."+dao.TransportTeam.Columns().CreatedAt+" AS created_at",
			"t."+dao.TransportTeam.Columns().Status+" AS status",
			"t."+dao.TransportTeam.Columns().InvalidatedAt+" AS invalidated_at",
			"t."+dao.TransportTeam.Columns().InvalidReason+" AS invalid_reason",
		)
}

func baseReportProjection(ctx context.Context) *gdb.Model {
	return dao.TransportReport.Ctx(ctx).
		Unscoped().
		As("r").
		LeftJoin(dao.TransportTeam.Table()+" t", "t."+dao.TransportTeam.Columns().Id+"=r."+dao.TransportReport.Columns().TeamId).
		LeftJoin(dao.User.Table()+" u", "u."+dao.User.Columns().Id+"=r."+dao.TransportReport.Columns().UserId+" AND u."+dao.User.Columns().DeletedAt+" IS NULL").
		Fields(
			"r."+dao.TransportReport.Columns().Id+" AS id",
			"r."+dao.TransportReport.Columns().TeamId+" AS team_id",
			"r."+dao.TransportReport.Columns().MemberId+" AS member_id",
			"r."+dao.TransportReport.Columns().UserId+" AS user_id",
			"COALESCE(t."+dao.TransportTeam.Columns().Name+", '') AS team_name",
			"COALESCE(u."+dao.User.Columns().Nickname+", '') AS user_name",
			"r."+dao.TransportReport.Columns().ActivityDate+" AS activity_date",
			"r."+dao.TransportReport.Columns().StartLat+" AS start_lat",
			"r."+dao.TransportReport.Columns().StartLng+" AS start_lng",
			"r."+dao.TransportReport.Columns().EndLat+" AS end_lat",
			"r."+dao.TransportReport.Columns().EndLng+" AS end_lng",
			"r."+dao.TransportReport.Columns().SampledAt+" AS sampled_at",
			"r."+dao.TransportReport.Columns().AcceptedAt+" AS accepted_at",
			"r."+dao.TransportReport.Columns().ContributionMeters+" AS contribution_meters",
			"r."+dao.TransportReport.Columns().UserTotalMeters+" AS user_total_meters",
			"r."+dao.TransportReport.Columns().TeamTotalMeters+" AS team_total_meters",
			"r."+dao.TransportReport.Columns().DailyReportCount+" AS daily_report_count",
		)
}

func teamFromProjection(row *teamProjection, mine bool) *Team {
	return &Team{ID: row.ID, Name: row.Name, CreatorID: row.CreatorID, CreatorName: row.CreatorName, MemberCount: row.MemberCount, TotalContributionMeters: row.TotalContributionMeters, LastActiveAt: row.LastActiveAt, CreatedAt: row.CreatedAt, Mine: mine}
}

func reportResultFromRow(row *reportRecord) *ReportResult {
	return &ReportResult{ID: row.ID, TeamID: row.TeamID, StartLat: row.StartLat, StartLng: row.StartLng, EndLat: row.EndLat, EndLng: row.EndLng, ContributionMeters: row.ContributionMeters, UserTotalMeters: row.UserTotalMeters, TeamTotalMeters: row.TeamTotalMeters, TodayReportCount: row.DailyReportCount, TodayReportRemaining: dailyReportLimit - row.DailyReportCount, AcceptedAt: row.AcceptedAt}
}

func reportFromRecord(row *reportRecord) *Report {
	return &Report{ReportResult: reportResultFromRow(row), ActivityDate: row.ActivityDate, SampledAt: row.SampledAt}
}
