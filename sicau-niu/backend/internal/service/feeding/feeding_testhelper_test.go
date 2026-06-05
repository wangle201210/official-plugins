// feeding_testhelper_test.go provides the database-gated test harness shared by
// the feeding integration tests. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All feeding DB tests share one temporary PostgreSQL database provisioned once
// per package run. A single shared database is used deliberately: GoFrame caches
// its default-group connection and per-table field metadata process-wide, so
// dropping and recreating a database per test would leave the cached default
// instance pointing at a dropped database and break the next test. Instead, each
// test stays self-contained and order independent by truncating the plugin tables
// before it runs. The shared database and the original GoFrame config are released
// once at the end of the package run.
//
// The feeding capability reads the niu/quote catalog (C2) and the iron table (C2),
// writes the feeding table and debits the grass ledger (C4), so the harness
// applies the 001 identity, 002 catalog and 004 grass DDL.

package feeding

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
	"lina-plugin-sicau-niu/backend/internal/service/feeding/internal/ironlocation"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

// feedingTables lists the plugin tables truncated before each DB-gated test.
var feedingTables = []string{
	"plugin_sicau_niu_feeding",
	"plugin_sicau_niu_grass_txn",
	"plugin_sicau_niu_grass_account",
	"plugin_sicau_niu_quote",
	"plugin_sicau_niu_iron",
	"plugin_sicau_niu_niu",
	"plugin_sicau_niu_user",
}

// feedingSchemaFiles lists the plugin install DDL files applied to the shared test
// database in order: 001 user/college, 002 catalog, 004 grass.
var feedingSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"004-sicau-niu-grass.sql",
}

// feedingDBHarness holds the lazily-provisioned shared test database state.
var (
	feedingDBOnce           sync.Once
	feedingDBLink           string
	feedingDBPrepErr        error
	feedingDBOriginalConfig gdb.Config
)

// fakeIronLocation is a controllable iron-location gateway returning fixed
// positions so the proximity-bonus decision is deterministic in tests.
type fakeIronLocation struct {
	positions []*ironlocation.IronPosition
}

// Positions returns the configured fixed iron positions.
func (f *fakeIronLocation) Positions(ctx context.Context) ([]*ironlocation.IronPosition, error) {
	return f.positions, nil
}

// newFeedingServiceForTest builds a feeding service with a 12m bonus threshold and
// the provided iron gateway, backed by a real grass ledger service.
func newFeedingServiceForTest(iron ironlocation.Gateway) Service {
	grassService := grasssvc.New(nil, grasssvc.Config{CheckinMinAmount: 30, CheckinMaxAmount: 30})
	return New(grassService, iron, nil, Config{IronBonusThresholdMeters: 12})
}

// creditGrass seeds a player's grass balance through the ledger so feeding has
// something to deduct.
func creditGrass(t *testing.T, ctx context.Context, userID int64, amount int64) {
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

// insertNiuRow inserts one cattle row directly for test setup and returns its ID.
func insertNiuRow(t *testing.T, ctx context.Context, row do.Niu) int64 {
	t.Helper()
	id, err := dao.Niu.Ctx(ctx).Data(row).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert niu row failed: %v", err)
	}
	return id
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

// insertQuoteRow inserts one quote row directly for test setup.
func insertQuoteRow(t *testing.T, ctx context.Context, content string, enabled int) {
	t.Helper()
	if _, err := dao.Quote.Ctx(ctx).Data(do.Quote{Content: content, Enabled: enabled}).Insert(); err != nil {
		t.Fatalf("insert quote row failed: %v", err)
	}
}

// setupPostgreSQLFeedingDB ensures the shared feeding test database exists and
// truncates the plugin tables so the calling test starts clean.
func setupPostgreSQLFeedingDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL feeding integration tests")
	}

	feedingDBOnce.Do(func() {
		feedingDBPrepErr = provisionSharedFeedingDB(ctx, baseLink)
	})
	if feedingDBPrepErr != nil {
		t.Fatalf("provision shared feeding database failed: %v", feedingDBPrepErr)
	}

	db := g.DB()
	for _, table := range feedingTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedFeedingDB creates the shared temporary database, points the
// default GoFrame group at it, applies the plugin install DDL files in order and
// records cleanup state for the package-level teardown.
func provisionSharedFeedingDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresFeedingTestDatabaseLink(baseLink)
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
	for _, name := range feedingSchemaFiles {
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

	feedingDBLink = dbLink
	feedingDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedFeedingDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database.
func teardownSharedFeedingDB() {
	if feedingDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared feeding database failed: %v\n", err)
	}
	if err := gdb.SetConfig(feedingDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLFeedingTestDatabase(ctx, feedingDBLink); err != nil {
		fmt.Printf("drop shared feeding database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared feeding database.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedFeedingDB()
	os.Exit(code)
}

// postgresFeedingTestDatabaseLink returns a unique database link for the shared
// feeding test database so concurrent or repeated package runs never collide.
func postgresFeedingTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL feeding base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_feeding_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		feedingNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLFeedingTestDatabase removes the shared temporary database.
func dropPostgreSQLFeedingTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresFeedingSystemLink(*targetConfig)})
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

// postgresFeedingSystemLink returns a PostgreSQL maintenance database link using
// the same host, credentials and extra parameters as the target link.
func postgresFeedingSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		feedingNormalizeExtra(config.Extra),
	)
}

// feedingNormalizeExtra ensures the link extra parameters are prefixed with a
// single leading question mark when present.
func feedingNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
