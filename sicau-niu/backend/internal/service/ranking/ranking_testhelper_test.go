// ranking_testhelper_test.go provides the database-gated test harness shared by
// the ranking integration tests. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All ranking DB tests share one temporary PostgreSQL database provisioned once
// per package run. A single shared database is used deliberately: GoFrame caches
// its default-group connection and per-table field metadata process-wide, so
// dropping and recreating a database per test would leave the cached default
// instance pointing at a dropped database and break the next test. Instead, each
// test stays self-contained and order independent by truncating the plugin tables
// before it runs. The shared database and the original GoFrame config are released
// once at the end of the package run.
//
// The ranking capability aggregates the C4 feeding effect joined with the C1
// player/college tables, so the harness applies the 001 identity, 002 catalog and
// 004 grass DDL.

package ranking

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver" // registers the supported PostgreSQL GoFrame driver.
	"lina-core/pkg/dialect"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

// rankingTables lists the plugin tables truncated before each DB-gated test.
var rankingTables = []string{
	"plugin_sicau_niu_feeding",
	"plugin_sicau_niu_user",
	"plugin_sicau_niu_college",
}

// rankingSchemaFiles lists the plugin install DDL files applied to the shared test
// database in order: 001 user/college, 002 catalog, 004 grass (feeding).
var rankingSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"004-sicau-niu-grass.sql",
}

// rankingDBHarness holds the lazily-provisioned shared test database state.
var (
	rankingDBOnce           sync.Once
	rankingDBLink           string
	rankingDBPrepErr        error
	rankingDBOriginalConfig gdb.Config
)

// insertUserRow inserts one player row directly for test setup and returns its ID.
func insertUserRow(t *testing.T, ctx context.Context, nickname, identityType string, collegeID int64) int64 {
	t.Helper()
	// Every player has a distinct WeChat openid in production; seed a unique one
	// so the active-openid unique index is respected across rows.
	openid := fmt.Sprintf("openid-%d", atomic.AddInt64(&userOpenidSeq, 1))
	id, err := dao.User.Ctx(ctx).Data(do.User{
		Openid:       openid,
		Nickname:     nickname,
		IdentityType: identityType,
		CollegeId:    collegeID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user row failed: %v", err)
	}
	return id
}

// userOpenidSeq provides process-unique openids for seeded test players.
var userOpenidSeq int64

// insertCollegeRow inserts one college row directly for test setup and returns its ID.
func insertCollegeRow(t *testing.T, ctx context.Context, name string) int64 {
	t.Helper()
	id, err := dao.College.Ctx(ctx).Data(do.College{Name: name}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert college row failed: %v", err)
	}
	return id
}

// insertFeedingRow inserts one feeding record with the given effect for a player.
func insertFeedingRow(t *testing.T, ctx context.Context, userID int64, effect int) {
	t.Helper()
	_, err := dao.Feeding.Ctx(ctx).Data(do.Feeding{
		UserId:           userID,
		NiuId:            1,
		BaseAmount:       effect,
		CoefficientBasis: 100,
		EffectAmount:     effect,
		IsIronBonus:      0,
	}).Insert()
	if err != nil {
		t.Fatalf("insert feeding row failed: %v", err)
	}
}

// assertBizCode fails the test unless err is a structured business error whose
// runtime code equals wantCode.
func assertBizCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected business error with code %s, got nil", wantCode)
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error with code %s, got %T: %v", wantCode, err, err)
	}
	if bizErr.RuntimeCode() != wantCode {
		t.Fatalf("expected code %s, got %s", wantCode, bizErr.RuntimeCode())
	}
}

// setupPostgreSQLRankingDB ensures the shared ranking test database exists and
// truncates the plugin tables so the calling test starts clean. It skips the test
// when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLRankingDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL ranking integration tests")
	}

	rankingDBOnce.Do(func() {
		rankingDBPrepErr = provisionSharedRankingDB(ctx, baseLink)
	})
	if rankingDBPrepErr != nil {
		t.Fatalf("provision shared ranking database failed: %v", rankingDBPrepErr)
	}

	db := g.DB()
	for _, table := range rankingTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedRankingDB creates the shared temporary database, points the
// default GoFrame group at it, applies the plugin install DDL files in order and
// records cleanup state for the package-level teardown.
func provisionSharedRankingDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresRankingTestDatabaseLink(baseLink)
	if err != nil {
		return err
	}
	dbDialect, err := dialect.From(dbLink)
	if err != nil {
		return err
	}
	if err = dbDialect.PrepareDatabase(ctx, dbLink, true); err != nil {
		return err
	}

	originalConfig := gdb.GetAllConfig()
	if err = gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: dbLink}},
	}); err != nil {
		return err
	}

	db := g.DB()
	for _, name := range rankingSchemaFiles {
		sqlPath := filepath.Join("..", "..", "..", "..", "manifest", "sql", name)
		content, readErr := os.ReadFile(sqlPath)
		if readErr != nil {
			return readErr
		}
		translated, translateErr := dbDialect.TranslateDDL(ctx, sqlPath, string(content))
		if translateErr != nil {
			return translateErr
		}
		for _, statement := range dialect.SplitSQLStatements(translated) {
			if _, execErr := db.Exec(ctx, statement); execErr != nil {
				return fmt.Errorf("execute schema SQL failed: %w\nSQL:\n%s", execErr, statement)
			}
		}
	}

	rankingDBLink = dbLink
	rankingDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedRankingDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database.
func teardownSharedRankingDB() {
	if rankingDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared ranking database failed: %v\n", err)
	}
	if err := gdb.SetConfig(rankingDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLRankingTestDatabase(ctx, rankingDBLink); err != nil {
		fmt.Printf("drop shared ranking database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared ranking database.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedRankingDB()
	os.Exit(code)
}

// postgresRankingTestDatabaseLink returns a unique database link for the shared
// ranking test database so concurrent or repeated package runs never collide.
func postgresRankingTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL ranking base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_ranking_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		rankingNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLRankingTestDatabase removes the shared temporary database.
func dropPostgreSQLRankingTestDatabase(ctx context.Context, targetLink string) (err error) {
	targetDB, err := gdb.New(gdb.ConfigNode{Link: targetLink})
	if err != nil {
		return err
	}
	targetConfig := targetDB.GetConfig()
	if targetConfig == nil {
		if closeErr := targetDB.Close(ctx); closeErr != nil {
			return closeErr
		}
		return nil
	}
	targetName := strings.TrimSpace(targetConfig.Name)
	if closeErr := targetDB.Close(ctx); closeErr != nil {
		return closeErr
	}
	if targetName == "" {
		return nil
	}

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresRankingSystemLink(*targetConfig)})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := systemDB.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = systemDB.Exec(
		ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
		targetName,
	); err != nil {
		return err
	}
	quotedName := `"` + strings.ReplaceAll(targetName, `"`, `""`) + `"`
	if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		return err
	}
	return nil
}

// postgresRankingSystemLink returns a PostgreSQL maintenance database link using
// the same host, credentials and extra parameters as the target link.
func postgresRankingSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		rankingNormalizeExtra(config.Extra),
	)
}

// rankingNormalizeExtra ensures the link extra parameters are prefixed with a
// single leading question mark when present.
func rankingNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
