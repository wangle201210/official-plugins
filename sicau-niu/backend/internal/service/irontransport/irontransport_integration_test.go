// irontransport_integration_test.go verifies cloud-moving persistence and invariants on PostgreSQL.
package irontransport

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/dialect"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

var (
	integrationDBOnce           sync.Once
	integrationDBLink           string
	integrationDBPrepErr        error
	integrationDBOriginalConfig gdb.Config
)

var integrationSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"003-sicau-niu-activation.sql",
	"004-sicau-niu-grass.sql",
	"007-sicau-niu-rule-config.sql",
	"010-sicau-niu-miniapp-interfaces.sql",
}

var integrationTables = []string{
	"plugin_sicau_niu_transport_report",
	"plugin_sicau_niu_transport_member",
	"plugin_sicau_niu_transport_team",
	"plugin_sicau_niu_user",
}

func setupIntegrationDB(t *testing.T, ctx context.Context) {
	t.Helper()
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run cloud-moving integration tests")
	}
	integrationDBOnce.Do(func() { integrationDBPrepErr = provisionIntegrationDB(ctx, baseLink) })
	if integrationDBPrepErr != nil {
		t.Fatalf("provision cloud-moving database failed: %v", integrationDBPrepErr)
	}
	for _, table := range integrationTables {
		if _, err := g.DB().Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

func provisionIntegrationDB(ctx context.Context, baseLink string) error {
	link, err := uniqueIntegrationDBLink(baseLink)
	if err != nil {
		return err
	}
	dbDialect, err := dialect.From(link)
	if err != nil {
		return err
	}
	if err = dbDialect.PrepareDatabase(ctx, link, true); err != nil {
		return err
	}
	integrationDBOriginalConfig = gdb.GetAllConfig()
	if err = gdb.SetConfig(gdb.Config{gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}}}); err != nil {
		return err
	}
	coreDictionarySchema := filepath.Join("..", "..", "..", "..", "..", "..", "lina-core", "manifest", "sql", "002-dictionary-management.sql")
	if err = executeIntegrationSQL(ctx, dbDialect, coreDictionarySchema); err != nil {
		return err
	}
	for _, name := range integrationSchemaFiles {
		path := filepath.Join("..", "..", "..", "..", "manifest", "sql", name)
		if err = executeIntegrationSQL(ctx, dbDialect, path); err != nil {
			return err
		}
	}
	if err = seedLegacyTransportModel(ctx); err != nil {
		return err
	}
	cloudMovingMigration := filepath.Join("..", "..", "..", "..", "manifest", "sql", "011-sicau-niu-cloud-moving.sql")
	if err = executeIntegrationSQL(ctx, dbDialect, cloudMovingMigration); err != nil {
		return err
	}
	if err = assertLegacyTransportMigration(ctx); err != nil {
		return err
	}
	// The cloud-moving migration is explicitly idempotent and must tolerate a
	// second run against the already-upgraded schema.
	if err = executeIntegrationSQL(ctx, dbDialect, cloudMovingMigration); err != nil {
		return fmt.Errorf("repeat cloud-moving schema failed: %w", err)
	}
	integrationDBLink = link
	return nil
}

func seedLegacyTransportModel(ctx context.Context) error {
	_, err := g.DB().Exec(ctx, `
WITH legacy_user AS (
    INSERT INTO plugin_sicau_niu_user (openid, nickname)
    VALUES ('transport-migration-legacy-user', '旧搬运玩家')
    RETURNING id
), legacy_team AS (
    INSERT INTO plugin_sicau_niu_transport_team (
        code, name, campus_id, leader_user_id, create_request_id, status, visible
    )
    SELECT 'LEGACY01', '旧共享会话团', 'cd', id, 'legacy-create', 'active', 1
    FROM legacy_user
    RETURNING id, leader_user_id
), legacy_member AS (
    INSERT INTO plugin_sicau_niu_transport_member (
        team_id, user_id, join_request_id, role, last_heartbeat_at
    )
    SELECT id, leader_user_id, 'legacy-join', 'leader', CURRENT_TIMESTAMP
    FROM legacy_team
    RETURNING id, team_id, user_id
), legacy_session AS (
    INSERT INTO plugin_sicau_niu_transport_session (
        team_id, started_by_user_id, start_request_id, status, last_lat, last_lng
    )
    SELECT team_id, user_id, 'legacy-start', 'active', 30.7058, 103.8318
    FROM legacy_member
    RETURNING id
)
INSERT INTO plugin_sicau_niu_transport_track (
    session_id, user_id, request_id, lat, lng, distance_meters
)
SELECT legacy_session.id, legacy_member.user_id, 'legacy-track', 30.7068, 103.8318, 111
FROM legacy_session, legacy_member`)
	if err != nil {
		return fmt.Errorf("seed legacy transport model failed: %w", err)
	}
	return nil
}

func assertLegacyTransportMigration(ctx context.Context) error {
	var team struct {
		Status        string     `json:"status"`
		Visible       int        `json:"visible"`
		MemberCount   int        `json:"memberCount"`
		InvalidatedAt *time.Time `json:"invalidatedAt"`
		InvalidReason string     `json:"invalidReason"`
	}
	if err := dao.TransportTeam.Ctx(ctx).
		Fields(
			dao.TransportTeam.Columns().Status,
			dao.TransportTeam.Columns().Visible,
			dao.TransportTeam.Columns().MemberCount,
			dao.TransportTeam.Columns().InvalidatedAt,
			dao.TransportTeam.Columns().InvalidReason,
		).
		Where(do.TransportTeam{CreateRequestId: "legacy-create"}).Scan(&team); err != nil {
		return fmt.Errorf("query migrated legacy team failed: %w", err)
	}
	if team.Status != string(teamStatusInvalid) || team.Visible != 0 || team.MemberCount != 0 || team.InvalidatedAt == nil || team.InvalidReason != "legacy_model" {
		return fmt.Errorf("legacy team was not safely invalidated: %+v", team)
	}
	var member struct {
		LeftAt *time.Time `json:"leftAt"`
	}
	if err := dao.TransportMember.Ctx(ctx).Fields(dao.TransportMember.Columns().LeftAt).
		Where(do.TransportMember{JoinRequestId: "legacy-join"}).Scan(&member); err != nil {
		return fmt.Errorf("query migrated legacy member failed: %w", err)
	}
	if member.LeftAt == nil {
		return errors.New("legacy active membership was not closed")
	}
	for _, table := range []string{"plugin_sicau_niu_transport_session", "plugin_sicau_niu_transport_track"} {
		value, queryErr := g.DB().GetValue(ctx, "SELECT to_regclass(?)", "public."+table)
		if queryErr != nil {
			return fmt.Errorf("query dropped legacy table %s failed: %w", table, queryErr)
		}
		if !value.IsNil() {
			return fmt.Errorf("legacy table %s still exists after migration", table)
		}
	}
	dictCount, err := g.DB().GetValue(ctx, `
SELECT COUNT(*)
FROM sys_dict_data
WHERE tenant_id = 0
  AND dict_type IN (
    'sicau_niu_transport_team_status',
    'sicau_niu_transport_member_role',
    'sicau_niu_transport_invalid_reason'
  )`)
	if err != nil {
		return fmt.Errorf("query cloud-moving dictionary data failed: %w", err)
	}
	if dictCount.Int() != 6 {
		return fmt.Errorf("expected six cloud-moving dictionary values, got %d", dictCount.Int())
	}
	return nil
}

func executeIntegrationSQL(ctx context.Context, dbDialect dialect.Dialect, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	translated, err := dbDialect.TranslateDDL(ctx, path, string(content))
	if err != nil {
		return err
	}
	for _, statement := range dialect.SplitSQLStatements(translated) {
		if _, err = g.DB().Exec(ctx, statement); err != nil {
			return fmt.Errorf("execute schema SQL failed: %w\nSQL:\n%s", err, statement)
		}
	}
	return nil
}

func uniqueIntegrationDBLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", errors.New("PostgreSQL base link configuration is empty")
	}
	if err = db.Close(context.Background()); err != nil {
		return "", err
	}
	return fmt.Sprintf("pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_transport_%d%s", config.User, config.Pass, config.Protocol, config.Host, config.Port, time.Now().UnixNano(), normalizeIntegrationExtra(config.Extra)), nil
}

func normalizeIntegrationExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}

func TestMain(m *testing.M) {
	code := m.Run()
	teardownIntegrationDB()
	os.Exit(code)
}

func teardownIntegrationDB() {
	if integrationDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close cloud-moving database failed: %v\n", err)
	}
	if err := gdb.SetConfig(integrationDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropIntegrationDB(ctx, integrationDBLink); err != nil {
		fmt.Printf("drop cloud-moving database failed: %v\n", err)
	}
}

func dropIntegrationDB(ctx context.Context, targetLink string) (err error) {
	target, err := gdb.New(gdb.ConfigNode{Link: targetLink})
	if err != nil {
		return err
	}
	config := target.GetConfig()
	if config == nil {
		return target.Close(ctx)
	}
	name := strings.TrimSpace(config.Name)
	if err = target.Close(ctx); err != nil || name == "" {
		return err
	}
	system, err := gdb.New(gdb.ConfigNode{Link: fmt.Sprintf("pgsql:%s:%s@%s(%s:%s)/postgres%s", config.User, config.Pass, config.Protocol, config.Host, config.Port, normalizeIntegrationExtra(config.Extra))})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := system.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = system.Exec(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()", name); err != nil {
		return err
	}
	quoted := `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	_, err = system.Exec(ctx, "DROP DATABASE IF EXISTS "+quoted)
	return err
}

func TestEmptyState(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	playerID := insertIntegrationUser(t, ctx, "openid-no-transport-team")
	state, err := New(Config{}).State(ctx, playerID)
	if err != nil {
		t.Fatalf("read empty transport state failed: %v", err)
	}
	if state.MaxEffectiveTeams != 120 || state.DailyReportLimit != 12 || state.TodayReportCount != 0 || state.MyTeam != nil || len(state.Teams) != 0 {
		t.Fatalf("unexpected empty cloud-moving state: %+v", state)
	}
}

func TestCreateJoinAndContributionFacts(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	creator := insertIntegrationUser(t, ctx, "openid-transport-creator")
	member := insertIntegrationUser(t, ctx, "openid-transport-member")
	outsider := insertIntegrationUser(t, ctx, "openid-transport-outsider")
	svc := New(Config{})
	createInput := &CreateTeamInput{RequestID: "create-cloud-team", Name: "云搬牛测试团"}
	state, err := svc.CreateTeam(ctx, creator, createInput)
	if err != nil {
		t.Fatalf("create cloud-moving team failed: %v", err)
	}
	if state.MyTeam == nil || state.MyTeam.MemberCount != 1 || state.MyTeam.Name != createInput.Name {
		t.Fatalf("unexpected created team state: %+v", state)
	}
	teamID := state.MyTeam.ID
	replay, err := svc.CreateTeam(ctx, creator, createInput)
	if err != nil || replay.MyTeam == nil || replay.MyTeam.ID != teamID {
		t.Fatalf("create replay was not idempotent: state=%+v err=%v", replay, err)
	}
	if _, err = svc.JoinTeam(ctx, member, teamID, "join-cloud-team"); err != nil {
		t.Fatalf("join team failed: %v", err)
	}
	if _, err = svc.JoinTeam(ctx, member, teamID, "join-cloud-team"); err != nil {
		t.Fatalf("join replay failed: %v", err)
	}
	_, err = svc.CreateTeam(ctx, member, &CreateTeamInput{RequestID: "create-while-joined", Name: "不应创建的团"})
	assertIntegrationCode(t, err, CodeAlreadyInTeam.RuntimeCode())
	members, err := svc.ListMembers(ctx, member, teamID, &PageInput{PageNum: 1, PageSize: 20})
	if err != nil || members.Total != 2 || len(members.List) != 2 || members.List[0].Role != string(roleCreator) {
		t.Fatalf("unexpected member projection: out=%+v err=%v", members, err)
	}
	memberPage, err := svc.ListMembers(ctx, member, teamID, &PageInput{PageNum: 2, PageSize: 1})
	if err != nil || memberPage.Total != 2 || len(memberPage.List) != 1 {
		t.Fatalf("member pagination was not bounded: out=%+v err=%v", memberPage, err)
	}
	_, err = svc.ListMembers(ctx, outsider, teamID, &PageInput{PageNum: 1, PageSize: 20})
	assertIntegrationCode(t, err, CodeTeamNotFound.RuntimeCode())
	first, err := svc.Report(ctx, member, &ReportInput{RequestID: "report-cloud-1", TeamID: teamID, Lat: 30.7058, Lng: 103.8318, SampledAt: time.Now()})
	if err != nil || first.ContributionMeters != 0 || first.TodayReportCount != 1 || first.StartLat != nil {
		t.Fatalf("unexpected first report: out=%+v err=%v", first, err)
	}
	second, err := svc.Report(ctx, member, &ReportInput{RequestID: "report-cloud-2", TeamID: teamID, Lat: 30.7068, Lng: 103.8318, SampledAt: time.Now()})
	if err != nil || second.ContributionMeters < 110 || second.ContributionMeters > 112 || second.UserTotalMeters != second.ContributionMeters || second.TeamTotalMeters != second.ContributionMeters {
		t.Fatalf("unexpected distance report: out=%+v err=%v", second, err)
	}
	secondReplay, err := svc.Report(ctx, member, &ReportInput{RequestID: "report-cloud-2", TeamID: teamID, Lat: 31, Lng: 104, SampledAt: time.Now()})
	if err != nil || secondReplay.ID != second.ID || secondReplay.EndLat != second.EndLat || secondReplay.TodayReportCount != second.TodayReportCount {
		t.Fatalf("report replay was not stable: out=%+v err=%v", secondReplay, err)
	}
	zeroDistance, err := svc.Report(ctx, member, &ReportInput{RequestID: "report-cloud-zero-distance", TeamID: teamID, Lat: 30.7068, Lng: 103.8318, SampledAt: time.Now()})
	if err != nil || zeroDistance.ContributionMeters != 0 || zeroDistance.TodayReportCount != 3 || zeroDistance.UserTotalMeters != second.UserTotalMeters {
		t.Fatalf("zero-distance report was not retained as a successful fact: out=%+v err=%v", zeroDistance, err)
	}
	memberState, err := svc.State(ctx, member)
	if err != nil || memberState.MyTeam == nil || !memberState.MyTeam.HasReportBaseline || memberState.MyTeam.MyContributionMeters != second.UserTotalMeters {
		t.Fatalf("current player contribution projection is incomplete: state=%+v err=%v", memberState, err)
	}
	reports, err := svc.ListMyReports(ctx, member, &PageInput{PageNum: 1, PageSize: 20})
	if err != nil || reports.Total != 3 || len(reports.List) != 3 {
		t.Fatalf("successful report facts missing: out=%+v err=%v", reports, err)
	}
	persistedCount, err := dao.TransportReport.Ctx(ctx).Where(do.TransportReport{UserId: member, TeamId: teamID}).Count()
	if err != nil || persistedCount != 3 {
		t.Fatalf("expected three persisted immutable reports, count=%d err=%v", persistedCount, err)
	}
	stats, err := svc.AdminStats(ctx, "")
	if err != nil || stats.ReportCount != 3 || stats.ContributionMeters != second.ContributionMeters || stats.ActiveMemberCount != 2 {
		t.Fatalf("unexpected admin stats: out=%+v err=%v", stats, err)
	}
	lastActiveBeforeRename := teamLastActiveAt(t, ctx, teamID)
	if err = svc.RenameTeam(ctx, teamID, "管理员改名团"); err != nil {
		t.Fatalf("admin rename failed: %v", err)
	}
	if lastActiveAfterRename := teamLastActiveAt(t, ctx, teamID); !lastActiveAfterRename.Equal(lastActiveBeforeRename) {
		t.Fatalf("admin rename unexpectedly refreshed team activity: before=%v after=%v", lastActiveBeforeRename, lastActiveAfterRename)
	}
	team, err := svc.GetTeam(ctx, member, teamID)
	if err != nil || team.Name != "管理员改名团" {
		t.Fatalf("renamed team was not visible: out=%+v err=%v", team, err)
	}
	adminTeams, err := svc.ListAdminTeams(ctx, &AdminTeamListInput{PageNum: 1, PageSize: 1, Status: string(teamStatusEffective)})
	if err != nil || adminTeams.Total != 1 || len(adminTeams.List) != 1 || adminTeams.List[0].Name != team.Name {
		t.Fatalf("unexpected admin team list: out=%+v err=%v", adminTeams, err)
	}
	adminReports, err := svc.ListAdminReports(ctx, &AdminReportListInput{PageNum: 1, PageSize: 1, TeamID: teamID, UserID: member})
	if err != nil || adminReports.Total != 3 || len(adminReports.List) != 1 || adminReports.List[0].EndLat == 0 {
		t.Fatalf("unexpected admin report audit: out=%+v err=%v", adminReports, err)
	}
}

func TestEffectiveTeamNamesAreUniqueAndReleasedAfterInvalidation(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	firstCreator := insertIntegrationUser(t, ctx, "openid-team-name-first")
	secondCreator := insertIntegrationUser(t, ctx, "openid-team-name-second")
	thirdCreator := insertIntegrationUser(t, ctx, "openid-team-name-third")
	svc := New(Config{InactiveAfter: time.Hour})

	first, err := svc.CreateTeam(ctx, firstCreator, &CreateTeamInput{RequestID: "create-team-name-first", Name: "有效团唯一名称"})
	if err != nil {
		t.Fatalf("create first named team failed: %v", err)
	}
	now := time.Now()
	_, directInsertErr := dao.TransportTeam.Ctx(ctx).Data(do.TransportTeam{
		Name: "有效团唯一名称", LeaderUserId: secondCreator, CreateRequestId: "direct-duplicate-team-name",
		Status: teamStatusEffective, Visible: 1, MemberCount: 1, LastActiveAt: &now,
	}).Insert()
	if !dialect.IsUniqueConstraintViolation(directInsertErr) {
		t.Fatalf("effective team-name index did not reject a direct duplicate insert: %v", directInsertErr)
	}
	_, err = svc.CreateTeam(ctx, secondCreator, &CreateTeamInput{RequestID: "create-team-name-duplicate", Name: "  有效团唯一名称  "})
	assertIntegrationCode(t, err, CodeTeamNameTaken.RuntimeCode())
	if member, memberErr := activeMembership(ctx, secondCreator, false); memberErr != nil || member != nil {
		t.Fatalf("duplicate team name unexpectedly created membership: member=%+v err=%v", member, memberErr)
	}

	second, err := svc.CreateTeam(ctx, secondCreator, &CreateTeamInput{RequestID: "create-team-name-second", Name: "后台改名目标团"})
	if err != nil {
		t.Fatalf("create second named team failed: %v", err)
	}
	err = svc.RenameTeam(ctx, second.MyTeam.ID, "有效团唯一名称")
	assertIntegrationCode(t, err, CodeTeamNameTaken.RuntimeCode())
	secondTeam, err := svc.GetTeam(ctx, secondCreator, second.MyTeam.ID)
	if err != nil || secondTeam.Name != "后台改名目标团" {
		t.Fatalf("rejected duplicate rename changed team: team=%+v err=%v", secondTeam, err)
	}

	reference := time.Now().Truncate(time.Microsecond)
	stale := reference.Add(-time.Hour)
	if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: first.MyTeam.ID}).Data(do.TransportTeam{LastActiveAt: &stale}).Update(); err != nil {
		t.Fatalf("age first named team failed: %v", err)
	}
	invalidated, err := svc.ExpireInactive(ctx, reference)
	if err != nil || invalidated != 1 {
		t.Fatalf("expire first named team failed: count=%d err=%v", invalidated, err)
	}
	reused, err := svc.CreateTeam(ctx, thirdCreator, &CreateTeamInput{RequestID: "create-team-name-reused", Name: "有效团唯一名称"})
	if err != nil || reused.MyTeam == nil || reused.MyTeam.Name != "有效团唯一名称" {
		t.Fatalf("invalidated team name was not released: state=%+v err=%v", reused, err)
	}
	effectiveCount, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Name: "有效团唯一名称", Status: teamStatusEffective}).Count()
	if err != nil || effectiveCount != 1 {
		t.Fatalf("expected exactly one effective team with reused name: count=%d err=%v", effectiveCount, err)
	}
	historyCount, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Name: "有效团唯一名称"}).Count()
	if err != nil || historyCount != 2 {
		t.Fatalf("expected invalid history and new effective team to share released name: count=%d err=%v", historyCount, err)
	}
}

func TestConcurrentCreatesPreserveEffectiveTeamNameUniqueness(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	firstPlayer := insertIntegrationUser(t, ctx, "openid-team-name-race-first")
	secondPlayer := insertIntegrationUser(t, ctx, "openid-team-name-race-second")

	results := make(chan error, 2)
	start := make(chan struct{})
	for index, playerID := range []int64{firstPlayer, secondPlayer} {
		go func(index int, playerID int64) {
			<-start
			_, createErr := New(Config{}).CreateTeam(ctx, playerID, &CreateTeamInput{
				RequestID: fmt.Sprintf("team-name-race-%d", index), Name: "并发唯一团名",
			})
			results <- createErr
		}(index, playerID)
	}
	close(start)
	successes, nameTaken := 0, 0
	for index := 0; index < 2; index++ {
		err := <-results
		if err == nil {
			successes++
			continue
		}
		parsed, ok := bizerr.As(err)
		if ok && parsed.RuntimeCode() == CodeTeamNameTaken.RuntimeCode() {
			nameTaken++
			continue
		}
		t.Fatalf("unexpected concurrent same-name create error: %v", err)
	}
	if successes != 1 || nameTaken != 1 {
		t.Fatalf("expected one same-name create and one rejection, successes=%d rejected=%d", successes, nameTaken)
	}
	count, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Name: "并发唯一团名", Status: teamStatusEffective}).Count()
	if err != nil || count != 1 {
		t.Fatalf("effective team-name uniqueness was not preserved: count=%d err=%v", count, err)
	}
}

func TestDailyReportLimit(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	playerID := insertIntegrationUser(t, ctx, "openid-report-limit")
	svc := New(Config{})
	state, err := svc.CreateTeam(ctx, playerID, &CreateTeamInput{RequestID: "create-limit-team", Name: "日报限额团"})
	if err != nil {
		t.Fatalf("create report-limit team failed: %v", err)
	}
	for i := 1; i <= 11; i++ {
		_, err = svc.Report(ctx, playerID, &ReportInput{RequestID: fmt.Sprintf("daily-report-%d", i), TeamID: state.MyTeam.ID, Lat: float64(i) / 10000, Lng: 103.8, SampledAt: time.Now()})
		if err != nil {
			t.Fatalf("report %d failed: %v", i, err)
		}
	}
	results := make(chan error, 2)
	start := make(chan struct{})
	for index := 12; index <= 13; index++ {
		go func(index int) {
			<-start
			_, reportErr := svc.Report(ctx, playerID, &ReportInput{RequestID: fmt.Sprintf("daily-report-%d", index), TeamID: state.MyTeam.ID, Lat: float64(index) / 10000, Lng: 103.8, SampledAt: time.Now()})
			results <- reportErr
		}(index)
	}
	close(start)
	successes, limited := 0, 0
	for index := 0; index < 2; index++ {
		reportErr := <-results
		if reportErr == nil {
			successes++
			continue
		}
		parsed, ok := bizerr.As(reportErr)
		if ok && parsed.RuntimeCode() == CodeDailyLimit.RuntimeCode() {
			limited++
			continue
		}
		t.Fatalf("unexpected concurrent daily-limit error: %v", reportErr)
	}
	if successes != 1 || limited != 1 {
		t.Fatalf("expected one twelfth report and one daily-limit rejection, successes=%d limited=%d", successes, limited)
	}
	zeroPlayer := insertIntegrationUser(t, ctx, "openid-zero-coordinate")
	zeroState, err := svc.CreateTeam(ctx, zeroPlayer, &CreateTeamInput{RequestID: "create-zero-coordinate", Name: "零纬度合法坐标团"})
	if err != nil {
		t.Fatalf("create zero-coordinate team failed: %v", err)
	}
	if _, err = svc.Report(ctx, zeroPlayer, &ReportInput{RequestID: "valid-zero-latitude", TeamID: zeroState.MyTeam.ID, Lat: 0, Lng: 103.8, SampledAt: time.Now()}); err != nil {
		t.Fatalf("valid zero latitude should be accepted: %v", err)
	}
	_, err = svc.Report(ctx, playerID, &ReportInput{RequestID: "daily-report-after-limit", TeamID: state.MyTeam.ID, Lat: 0, Lng: 103.8, SampledAt: time.Now()})
	assertIntegrationCode(t, err, CodeDailyLimit.RuntimeCode())
	count, countErr := dao.TransportReport.Ctx(ctx).Where(do.TransportReport{UserId: playerID}).Count()
	if countErr != nil || count != 12 {
		t.Fatalf("rejected thirteenth report was persisted: count=%d err=%v", count, countErr)
	}
}

func TestJoinAndReportRejectExpiredTeamAndPersistInvalidation(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	creator := insertIntegrationUser(t, ctx, "openid-expiry-boundary-creator")
	joiner := insertIntegrationUser(t, ctx, "openid-expiry-boundary-joiner")
	svc := New(Config{})
	state, err := svc.CreateTeam(ctx, creator, &CreateTeamInput{RequestID: "create-expiry-boundary", Name: "锁内失效团"})
	if err != nil {
		t.Fatalf("create boundary team failed: %v", err)
	}
	teamID := state.MyTeam.ID
	stale := time.Now().Add(-73 * time.Hour)
	if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).Data(do.TransportTeam{LastActiveAt: &stale}).Update(); err != nil {
		t.Fatalf("age boundary team failed: %v", err)
	}
	_, err = svc.JoinTeam(ctx, joiner, teamID, "join-expired-boundary")
	assertIntegrationCode(t, err, CodeTeamNotFound.RuntimeCode())
	assertIntegrationTeamInvalid(t, ctx, teamID)

	reportState, err := svc.CreateTeam(ctx, joiner, &CreateTeamInput{RequestID: "create-report-boundary", Name: "上报边界团"})
	if err != nil {
		t.Fatalf("create report boundary team failed: %v", err)
	}
	reportTeamID := reportState.MyTeam.ID
	if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: reportTeamID}).Data(do.TransportTeam{LastActiveAt: &stale}).Update(); err != nil {
		t.Fatalf("age report boundary team failed: %v", err)
	}
	_, err = svc.Report(ctx, joiner, &ReportInput{RequestID: "report-expired-boundary", TeamID: reportTeamID, Lat: 30.7, Lng: 103.8, SampledAt: time.Now()})
	assertIntegrationCode(t, err, CodeTeamNotFound.RuntimeCode())
	assertIntegrationTeamInvalid(t, ctx, reportTeamID)
	reportCount, countErr := dao.TransportReport.Ctx(ctx).Where(do.TransportReport{TeamId: reportTeamID}).Count()
	if countErr != nil || reportCount != 0 {
		t.Fatalf("expired team report must not persist: count=%d err=%v", reportCount, countErr)
	}
}

func TestInvalidatesInactiveTeamAndReleasesMembership(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	playerID := insertIntegrationUser(t, ctx, "openid-expired-team")
	memberID := insertIntegrationUser(t, ctx, "openid-expired-team-member")
	svc := New(Config{InactiveAfter: time.Hour})
	state, err := svc.CreateTeam(ctx, playerID, &CreateTeamInput{RequestID: "create-expired-team", Name: "待失效团"})
	if err != nil {
		t.Fatalf("create expiring team failed: %v", err)
	}
	teamID := state.MyTeam.ID
	if _, err = svc.JoinTeam(ctx, memberID, teamID, "join-expired-team"); err != nil {
		t.Fatalf("join expiring team failed: %v", err)
	}
	if _, err = svc.Report(ctx, playerID, &ReportInput{RequestID: "history-before-expiry", TeamID: teamID, Lat: 30.7, Lng: 103.8, SampledAt: time.Now()}); err != nil {
		t.Fatalf("seed report history failed: %v", err)
	}
	reference := time.Now().Truncate(time.Microsecond)
	stale := reference.Add(-time.Hour)
	if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).Data(do.TransportTeam{LastActiveAt: &stale}).Update(); err != nil {
		t.Fatalf("age team failed: %v", err)
	}
	invalidated, err := svc.ExpireInactive(ctx, reference)
	if err != nil || invalidated != 1 {
		t.Fatalf("expire inactive team failed: count=%d err=%v", invalidated, err)
	}
	state, err = svc.State(ctx, playerID)
	if err != nil || state.MyTeam != nil || len(state.Teams) != 0 {
		t.Fatalf("expired team remained player-visible: state=%+v err=%v", state, err)
	}
	var teamHistory struct {
		Status      string `json:"status"`
		MemberCount int    `json:"memberCount"`
	}
	if err = dao.TransportTeam.Ctx(ctx).Fields(dao.TransportTeam.Columns().Status, dao.TransportTeam.Columns().MemberCount).Where(do.TransportTeam{Id: teamID}).Scan(&teamHistory); err != nil || teamHistory.Status != string(teamStatusInvalid) || teamHistory.MemberCount != 0 {
		t.Fatalf("team history was not invalidated: out=%+v err=%v", teamHistory, err)
	}
	historyCount, err := dao.TransportReport.Ctx(ctx).Where(do.TransportReport{TeamId: teamID}).Count()
	if err != nil || historyCount != 1 {
		t.Fatalf("expired team report history was not retained: count=%d err=%v", historyCount, err)
	}
	activeCount, err := dao.TransportMember.Ctx(ctx).
		Where(do.TransportMember{TeamId: teamID}).
		Where(dao.TransportMember.Columns().LeftAt + " IS NULL").Count()
	if err != nil || activeCount != 0 {
		t.Fatalf("expired team retained active members: count=%d err=%v", activeCount, err)
	}
	newState, err := svc.CreateTeam(ctx, playerID, &CreateTeamInput{RequestID: "create-after-expiry", Name: "待失效团"})
	if err != nil {
		t.Fatalf("expired membership was not released: %v", err)
	}
	newReport, err := svc.Report(ctx, playerID, &ReportInput{RequestID: "report-after-expiry", TeamID: newState.MyTeam.ID, Lat: 31, Lng: 104, SampledAt: time.Now()})
	if err != nil || newReport.ContributionMeters != 0 || newReport.TodayReportCount != 2 {
		t.Fatalf("new-team baseline or cross-team daily quota was reset incorrectly: out=%+v err=%v", newReport, err)
	}
}

func TestEffectiveTeamLimitAndExpiredSlotRelease(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	playerID := insertIntegrationUser(t, ctx, "openid-team-limit")
	now := time.Now()
	for i := 0; i < maxEffectiveTeams; i++ {
		if _, err := dao.TransportTeam.Ctx(ctx).Data(do.TransportTeam{
			Name: fmt.Sprintf("有效团-%03d", i), LeaderUserId: playerID + int64(i) + 1,
			CreateRequestId: fmt.Sprintf("seed-team-%03d", i), Status: teamStatusEffective,
			Visible: 1, MemberCount: 1, LastActiveAt: &now,
		}).Insert(); err != nil {
			t.Fatalf("seed effective team %d failed: %v", i, err)
		}
	}
	svc := New(Config{})
	_, err := svc.CreateTeam(ctx, playerID, &CreateTeamInput{RequestID: "create-over-limit", Name: "超过上限的团"})
	assertIntegrationCode(t, err, CodeTeamLimit.RuntimeCode())
	stale := now.Add(-73 * time.Hour)
	if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{CreateRequestId: "seed-team-000"}).Data(do.TransportTeam{LastActiveAt: &stale}).Update(); err != nil {
		t.Fatalf("age one effective team failed: %v", err)
	}
	state, err := svc.CreateTeam(ctx, playerID, &CreateTeamInput{RequestID: "create-after-slot-release", Name: "释放名额后的团"})
	if err != nil || state.MyTeam == nil || state.MyTeam.Name != "释放名额后的团" || len(state.Teams) != maxEffectiveTeams {
		t.Fatalf("expired slot was not released: state=%+v err=%v", state, err)
	}
}

func TestConcurrentCreatesCannotExceedEffectiveTeamLimit(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	firstPlayer := insertIntegrationUser(t, ctx, "openid-capacity-race-first")
	secondPlayer := insertIntegrationUser(t, ctx, "openid-capacity-race-second")
	now := time.Now()
	for index := 0; index < maxEffectiveTeams-1; index++ {
		if _, err := dao.TransportTeam.Ctx(ctx).Data(do.TransportTeam{
			Name: fmt.Sprintf("容量预置团-%03d", index), LeaderUserId: 900000 + index,
			CreateRequestId: fmt.Sprintf("capacity-seed-%03d", index), Status: teamStatusEffective,
			Visible: 1, MemberCount: 1, LastActiveAt: &now,
		}).Insert(); err != nil {
			t.Fatalf("seed capacity team %d failed: %v", index, err)
		}
	}

	type createResult struct {
		err error
	}
	results := make(chan createResult, 2)
	start := make(chan struct{})
	for index, playerID := range []int64{firstPlayer, secondPlayer} {
		go func(index int, playerID int64) {
			<-start
			_, err := New(Config{}).CreateTeam(ctx, playerID, &CreateTeamInput{
				RequestID: fmt.Sprintf("capacity-race-%d", index), Name: fmt.Sprintf("容量竞争团-%d", index),
			})
			results <- createResult{err: err}
		}(index, playerID)
	}
	close(start)
	successes, limited := 0, 0
	for index := 0; index < 2; index++ {
		result := <-results
		if result.err == nil {
			successes++
			continue
		}
		parsed, ok := bizerr.As(result.err)
		if ok && parsed.RuntimeCode() == CodeTeamLimit.RuntimeCode() {
			limited++
			continue
		}
		t.Fatalf("unexpected concurrent create error: %v", result.err)
	}
	if successes != 1 || limited != 1 {
		t.Fatalf("expected one create and one capacity rejection, successes=%d limited=%d", successes, limited)
	}
	count, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Status: teamStatusEffective, Visible: 1}).Count()
	if err != nil || count != maxEffectiveTeams {
		t.Fatalf("effective team cap was not preserved: count=%d err=%v", count, err)
	}
}

func TestConcurrentJoinsPreserveSingleActiveMembership(t *testing.T) {
	ctx := context.Background()
	setupIntegrationDB(t, ctx)
	firstCreator := insertIntegrationUser(t, ctx, "openid-join-race-first-creator")
	secondCreator := insertIntegrationUser(t, ctx, "openid-join-race-second-creator")
	joiner := insertIntegrationUser(t, ctx, "openid-join-race-player")
	svc := New(Config{})
	first, err := svc.CreateTeam(ctx, firstCreator, &CreateTeamInput{RequestID: "join-race-first-team", Name: "并发加入一团"})
	if err != nil {
		t.Fatalf("create first join target failed: %v", err)
	}
	second, err := svc.CreateTeam(ctx, secondCreator, &CreateTeamInput{RequestID: "join-race-second-team", Name: "并发加入二团"})
	if err != nil {
		t.Fatalf("create second join target failed: %v", err)
	}

	results := make(chan error, 2)
	start := make(chan struct{})
	for index, teamID := range []int64{first.MyTeam.ID, second.MyTeam.ID} {
		go func(index int, teamID int64) {
			<-start
			_, err := svc.JoinTeam(ctx, joiner, teamID, fmt.Sprintf("join-race-%d", index))
			results <- err
		}(index, teamID)
	}
	close(start)
	successes, alreadyJoined := 0, 0
	for index := 0; index < 2; index++ {
		err = <-results
		if err == nil {
			successes++
			continue
		}
		parsed, ok := bizerr.As(err)
		if ok && parsed.RuntimeCode() == CodeAlreadyInTeam.RuntimeCode() {
			alreadyJoined++
			continue
		}
		t.Fatalf("unexpected concurrent join error: %v", err)
	}
	if successes != 1 || alreadyJoined != 1 {
		t.Fatalf("expected one join and one single-membership rejection, successes=%d rejected=%d", successes, alreadyJoined)
	}
	count, err := activeMemberModel(ctx, joiner).Count()
	if err != nil || count != 1 {
		t.Fatalf("single active membership was not preserved: count=%d err=%v", count, err)
	}
}

func insertIntegrationUser(t *testing.T, ctx context.Context, openid string) int64 {
	t.Helper()
	id, err := dao.User.Ctx(ctx).Data(do.User{Openid: openid, Nickname: openid}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user failed: %v", err)
	}
	return id
}

func teamLastActiveAt(t *testing.T, ctx context.Context, teamID int64) time.Time {
	t.Helper()
	var row struct {
		LastActiveAt *time.Time `json:"lastActiveAt"`
	}
	if err := dao.TransportTeam.Ctx(ctx).
		Fields(dao.TransportTeam.Columns().LastActiveAt).
		Where(do.TransportTeam{Id: teamID}).Scan(&row); err != nil || row.LastActiveAt == nil {
		t.Fatalf("query team last-active time failed: teamID=%d out=%+v err=%v", teamID, row, err)
	}
	return *row.LastActiveAt
}

func assertIntegrationTeamInvalid(t *testing.T, ctx context.Context, teamID int64) {
	t.Helper()
	var row struct {
		Status        string     `json:"status"`
		MemberCount   int        `json:"memberCount"`
		InvalidatedAt *time.Time `json:"invalidatedAt"`
	}
	if err := dao.TransportTeam.Ctx(ctx).
		Fields(dao.TransportTeam.Columns().Status, dao.TransportTeam.Columns().MemberCount, dao.TransportTeam.Columns().InvalidatedAt).
		Where(do.TransportTeam{Id: teamID}).Scan(&row); err != nil {
		t.Fatalf("query invalid team failed: %v", err)
	}
	if row.Status != string(teamStatusInvalid) || row.MemberCount != 0 || row.InvalidatedAt == nil {
		t.Fatalf("team was not persistently invalidated: %+v", row)
	}
}

func assertIntegrationCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected business error %s", want)
	}
	parsed, ok := bizerr.As(err)
	if !ok || parsed.RuntimeCode() != want {
		t.Fatalf("expected business error %s, got %v", want, err)
	}
}
