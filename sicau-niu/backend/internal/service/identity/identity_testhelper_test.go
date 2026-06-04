// identity_testhelper_test.go provides the database-gated test harness shared by
// the identity integration tests. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All identity DB tests share one temporary PostgreSQL database provisioned once
// per package run. A single shared database is used deliberately: GoFrame caches
// its default-group connection and per-table field metadata process-wide, so
// dropping and recreating a database per test would leave the cached default
// instance pointing at a dropped database and break the next test. Instead, each
// test stays self-contained and order independent by truncating the player and
// college tables before it runs via setupPostgreSQLIdentityDB, so it never
// depends on rows left by another test. The shared database and the original
// GoFrame config are released once at the end of the package run.

package identity

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
var sicauNiuTables = []string{"plugin_sicau_niu_user", "plugin_sicau_niu_college"}

// identityDBHarness holds the lazily-provisioned shared test database state.
var (
	identityDBOnce           sync.Once
	identityDBLink           string
	identityDBPrepErr        error
	identityDBOriginalConfig gdb.Config
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

// setupPostgreSQLIdentityDB ensures the shared identity test database exists and
// truncates the plugin tables so the calling test starts clean. It skips the
// test when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLIdentityDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL identity integration tests")
	}

	identityDBOnce.Do(func() {
		identityDBPrepErr = provisionSharedIdentityDB(ctx, baseLink)
	})
	if identityDBPrepErr != nil {
		t.Fatalf("provision shared identity database failed: %v", identityDBPrepErr)
	}

	truncateSicauNiuTables(t, ctx)
}

// truncateSicauNiuTables clears the plugin tables so each test runs against a
// clean dataset regardless of execution order.
func truncateSicauNiuTables(t *testing.T, ctx context.Context) {
	t.Helper()
	db := g.DB()
	for _, table := range sicauNiuTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedIdentityDB creates the shared temporary database, points the
// default GoFrame group at it, applies the plugin install DDL, and registers a
// process-exit cleanup that drops the database and restores the original config.
func provisionSharedIdentityDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresIdentityTestDatabaseLink(baseLink)
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
	sqlPath := filepath.Join("..", "..", "..", "..", "manifest", "sql", "001-sicau-niu-identity.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	translated, err := dbDialect.TranslateDDL(ctx, sqlPath, string(content))
	if err != nil {
		return err
	}
	for _, statement := range dialect.SplitSQLStatements(translated) {
		if _, err = db.Exec(ctx, statement); err != nil {
			return fmt.Errorf("execute schema SQL failed: %w\nSQL:\n%s", err, statement)
		}
	}

	identityDBLink = dbLink
	identityDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedIdentityDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database. It is a no-op when the shared
// database was never provisioned (for example when the DB-gated tests skipped).
func teardownSharedIdentityDB() {
	if identityDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared identity database failed: %v\n", err)
	}
	if err := gdb.SetConfig(identityDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLIdentityTestDatabase(ctx, identityDBLink); err != nil {
		fmt.Printf("drop shared identity database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared identity database
// afterwards so the temporary database and GoFrame config never leak.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedIdentityDB()
	os.Exit(code)
}

// postgresIdentityTestDatabaseLink returns a unique database link for the shared
// identity test database so concurrent or repeated package runs never collide.
func postgresIdentityTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL identity base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		normalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLIdentityTestDatabase removes the shared temporary database.
func dropPostgreSQLIdentityTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresIdentitySystemLink(*targetConfig)})
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

// postgresIdentitySystemLink returns a PostgreSQL maintenance database link using
// the same host, credentials and extra parameters as the target link.
func postgresIdentitySystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		normalizeExtra(config.Extra),
	)
}

// normalizeExtra ensures the link extra parameters are prefixed with a single
// leading question mark when present.
func normalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
