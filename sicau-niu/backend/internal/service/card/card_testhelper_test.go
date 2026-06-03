// card_testhelper_test.go provides the database-gated test harness shared by the
// card and quote integration tests. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All card DB tests share one temporary PostgreSQL database provisioned once per
// package run. A single shared database is used deliberately: GoFrame caches its
// default-group connection and per-table field metadata process-wide, so dropping
// and recreating a database per test would leave the cached default instance
// pointing at a dropped database and break the next test. Instead, each test
// stays self-contained and order independent by truncating the plugin tables
// before it runs via setupPostgreSQLCardDB, so it never depends on rows left by
// another test. The shared database and the original GoFrame config are released
// once at the end of the package run.
//
// The card capability validates owning-cattle existence through the cattle
// service (which reads the niu table) and stores cards/quotes, so the harness
// applies both the 001 identity/college DDL and the 002 catalog DDL to ensure
// every consumed table exists.

package card

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
)

// sicauNiuTables lists the plugin tables truncated before each DB-gated test so
// every test starts from a clean, deterministic state.
var sicauNiuTables = []string{
	"plugin_sicau_niu_quote",
	"plugin_sicau_niu_card",
	"plugin_sicau_niu_iron",
	"plugin_sicau_niu_niu",
	"plugin_sicau_niu_college",
}

// cardSchemaFiles lists the plugin install DDL files applied to the shared test
// database in order. 001 creates the user/college tables; 002 creates the
// cattle/iron/card/quote catalog tables consumed by the card capability.
var cardSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
}

// cardDBHarness holds the lazily-provisioned shared test database state.
var (
	cardDBOnce           sync.Once
	cardDBLink           string
	cardDBPrepErr        error
	cardDBOriginalConfig gdb.Config
)

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

// setupPostgreSQLCardDB ensures the shared card test database exists and
// truncates the plugin tables so the calling test starts clean. It skips the test
// when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLCardDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL card integration tests")
	}

	cardDBOnce.Do(func() {
		cardDBPrepErr = provisionSharedCardDB(ctx, baseLink)
	})
	if cardDBPrepErr != nil {
		t.Fatalf("provision shared card database failed: %v", cardDBPrepErr)
	}

	db := g.DB()
	for _, table := range sicauNiuTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedCardDB creates the shared temporary database, points the default
// GoFrame group at it, applies the plugin install DDL files in order and records
// cleanup state for the package-level teardown.
func provisionSharedCardDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresCardTestDatabaseLink(baseLink)
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
	for _, name := range cardSchemaFiles {
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

	cardDBLink = dbLink
	cardDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedCardDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database. It is a no-op when the shared
// database was never provisioned (for example when the DB-gated tests skipped).
func teardownSharedCardDB() {
	if cardDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared card database failed: %v\n", err)
	}
	if err := gdb.SetConfig(cardDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLCardTestDatabase(ctx, cardDBLink); err != nil {
		fmt.Printf("drop shared card database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared card database
// afterwards so the temporary database and GoFrame config never leak.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedCardDB()
	os.Exit(code)
}

// postgresCardTestDatabaseLink returns a unique database link for the shared card
// test database so concurrent or repeated package runs never collide.
func postgresCardTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL card base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_card_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		cardNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLCardTestDatabase removes the shared temporary database.
func dropPostgreSQLCardTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresCardSystemLink(*targetConfig)})
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

// postgresCardSystemLink returns a PostgreSQL maintenance database link using the
// same host, credentials and extra parameters as the target link.
func postgresCardSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		cardNormalizeExtra(config.Extra),
	)
}

// cardNormalizeExtra ensures the link extra parameters are prefixed with a single
// leading question mark when present.
func cardNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
