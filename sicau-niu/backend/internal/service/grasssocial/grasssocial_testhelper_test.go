// grasssocial_testhelper_test.go provides the database-gated test harness shared
// by the steal, gift and inbox integration tests. Database-backed assertions are
// skipped unless LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All grass-social DB tests share one temporary PostgreSQL database provisioned
// once per package run. A single shared database is used deliberately: GoFrame
// caches its default-group connection and per-table field metadata process-wide,
// so dropping and recreating a database per test would break the cached default
// instance. Instead, each test stays self-contained and order independent by
// truncating the plugin tables before it runs. The shared database and the
// original GoFrame config are released once at the end of the package run.
//
// The grass-social capability reads the user table, moves grass through the ledger
// and writes the steal/gift/inbox tables, so the harness applies the 001 identity
// and 004 grass DDL.

package grasssocial

import (
	"context"
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
	_ "lina-core/pkg/dbdriver" // registers the supported PostgreSQL GoFrame driver.
	"lina-core/pkg/dialect"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

// socialTables lists the plugin tables truncated before each DB-gated test.
var socialTables = []string{
	"plugin_sicau_niu_inbox_msg",
	"plugin_sicau_niu_steal",
	"plugin_sicau_niu_gift",
	"plugin_sicau_niu_grass_txn",
	"plugin_sicau_niu_grass_account",
	"plugin_sicau_niu_user",
}

// socialSchemaFiles lists the plugin install DDL files applied to the shared test
// database in order: 001 user/college then 004 grass.
var socialSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"004-sicau-niu-grass.sql",
}

// socialDBHarness holds the lazily-provisioned shared test database state.
var (
	socialDBOnce           sync.Once
	socialDBLink           string
	socialDBPrepErr        error
	socialDBOriginalConfig gdb.Config
)

// newSocialServiceForTest builds a grass-social service with deterministic limits
// for tests: a fixed steal amount (min==max) and small daily caps.
func newSocialServiceForTest() Service {
	grassService := grasssvc.New(nil, grasssvc.Config{CheckinMinAmount: 1, CheckinMaxAmount: 1})
	return New(grassService, nil, Config{
		StealDailyTargets: 12,
		StealDailyLimit:   2,
		StealMinAmount:    10,
		StealMaxAmount:    10,
		GiftDailyLimit:    2,
		GiftMinAmount:     12,
	})
}

// seedGrass credits a player's grass balance through the ledger so steal/gift have
// grass to move.
func seedGrass(t *testing.T, ctx context.Context, userID int64, amount int64) {
	t.Helper()
	grassService := grasssvc.New(nil, grasssvc.Config{CheckinMinAmount: 1, CheckinMaxAmount: 1})
	err := dao.GrassAccount.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, applyErr := grassService.ApplyDelta(ctx, tx, userID, amount, grasssvc.TxnTypeCheckin, 0)
		return applyErr
	})
	if err != nil {
		t.Fatalf("seed grass failed: %v", err)
	}
}

// balanceOf returns a player's current grass balance for assertions.
func balanceOf(t *testing.T, ctx context.Context, userID int64) int64 {
	t.Helper()
	value, err := dao.GrassAccount.Ctx(ctx).
		Fields(dao.GrassAccount.Columns().Balance).
		Where(dao.GrassAccount.Columns().UserId, userID).
		Value()
	if err != nil {
		t.Fatalf("read balance failed: %v", err)
	}
	return value.Int64()
}

// inboxCount returns the number of inbox messages of a type for a player.
func inboxCount(t *testing.T, ctx context.Context, userID int64, msgType MsgType) int {
	t.Helper()
	count, err := dao.InboxMsg.Ctx(ctx).
		Where(dao.InboxMsg.Columns().UserId, userID).
		Where(dao.InboxMsg.Columns().MsgType, msgType.String()).
		Count()
	if err != nil {
		t.Fatalf("count inbox failed: %v", err)
	}
	return count
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

// insertUserRow inserts one player row directly for test setup and returns its ID.
func insertUserRow(t *testing.T, ctx context.Context, openid string) int64 {
	t.Helper()
	id, err := dao.User.Ctx(ctx).Data(do.User{Openid: openid}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user row failed: %v", err)
	}
	return id
}

// setupPostgreSQLSocialDB ensures the shared grass-social test database exists and
// truncates the plugin tables so the calling test starts clean.
func setupPostgreSQLSocialDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL grass-social integration tests")
	}

	socialDBOnce.Do(func() {
		socialDBPrepErr = provisionSharedSocialDB(ctx, baseLink)
	})
	if socialDBPrepErr != nil {
		t.Fatalf("provision shared grass-social database failed: %v", socialDBPrepErr)
	}

	db := g.DB()
	for _, table := range socialTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedSocialDB creates the shared temporary database, points the
// default GoFrame group at it, applies the plugin install DDL files in order and
// records cleanup state for the package-level teardown.
func provisionSharedSocialDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresSocialTestDatabaseLink(baseLink)
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
	for _, name := range socialSchemaFiles {
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

	socialDBLink = dbLink
	socialDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedSocialDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database.
func teardownSharedSocialDB() {
	if socialDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared grass-social database failed: %v\n", err)
	}
	if err := gdb.SetConfig(socialDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLSocialTestDatabase(ctx, socialDBLink); err != nil {
		fmt.Printf("drop shared grass-social database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared grass-social database.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedSocialDB()
	os.Exit(code)
}

// postgresSocialTestDatabaseLink returns a unique database link for the shared
// grass-social test database so concurrent or repeated package runs never collide.
func postgresSocialTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL grass-social base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_social_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		socialNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLSocialTestDatabase removes the shared temporary database.
func dropPostgreSQLSocialTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresSocialSystemLink(*targetConfig)})
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

// postgresSocialSystemLink returns a PostgreSQL maintenance database link using
// the same host, credentials and extra parameters as the target link.
func postgresSocialSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		socialNormalizeExtra(config.Extra),
	)
}

// socialNormalizeExtra ensures the link extra parameters are prefixed with a
// single leading question mark when present.
func socialNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
