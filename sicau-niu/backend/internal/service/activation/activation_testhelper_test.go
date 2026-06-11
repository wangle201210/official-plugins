// activation_testhelper_test.go provides the database-gated test harness shared
// by the activation, collection and poster integration tests. Database-backed
// assertions are skipped unless LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All activation DB tests share one temporary PostgreSQL database provisioned
// once per package run. A single shared database is used deliberately: GoFrame
// caches its default-group connection and per-table field metadata process-wide,
// so dropping and recreating a database per test would leave the cached default
// instance pointing at a dropped database and break the next test. Instead, each
// test stays self-contained and order independent by truncating the plugin tables
// before it runs via setupPostgreSQLActivationDB, so it never depends on rows left
// by another test. The shared database and the original GoFrame config are
// released once at the end of the package run.
//
// The activation capability reads the niu/card/quote catalog (C2) and writes the
// activation table (C3), and the poster path reads the player identity (C1), so
// the harness applies the 001 identity, 002 catalog and 003 activation DDL to
// ensure every consumed table exists.

package activation

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
)

// sicauNiuTables lists the plugin tables truncated before each DB-gated test so
// every test starts from a clean, deterministic state.
var sicauNiuTables = []string{
	"plugin_sicau_niu_activation_attempt",
	"plugin_sicau_niu_activation",
	"plugin_sicau_niu_quote",
	"plugin_sicau_niu_card",
	"plugin_sicau_niu_iron",
	"plugin_sicau_niu_niu",
	"plugin_sicau_niu_user",
	"plugin_sicau_niu_college",
}

// activationSchemaFiles lists the plugin install DDL files applied to the shared
// test database in order: 001 user/college, 002 catalog, 003 activation.
var activationSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"003-sicau-niu-activation.sql",
	"007-sicau-niu-rule-config.sql",
}

// activationDBHarness holds the lazily-provisioned shared test database state.
var (
	activationDBOnce           sync.Once
	activationDBLink           string
	activationDBPrepErr        error
	activationDBOriginalConfig gdb.Config
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

// insertNiuRow inserts one cattle row directly for test setup and returns its ID.
// Tests use it to stage activatable cattle with explicit anchors and visibility.
func insertNiuRow(t *testing.T, ctx context.Context, row do.Niu) int64 {
	t.Helper()
	id, err := dao.Niu.Ctx(ctx).Data(row).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert niu row failed: %v", err)
	}
	return id
}

// insertCardRow inserts one card row directly for test setup and returns its ID.
func insertCardRow(t *testing.T, ctx context.Context, row do.Card) int64 {
	t.Helper()
	id, err := dao.Card.Ctx(ctx).Data(row).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert card row failed: %v", err)
	}
	return id
}

// insertQuoteRow inserts one quote row directly for test setup and returns its ID.
func insertQuoteRow(t *testing.T, ctx context.Context, row do.Quote) int64 {
	t.Helper()
	id, err := dao.Quote.Ctx(ctx).Data(row).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert quote row failed: %v", err)
	}
	return id
}

// insertUserRow inserts one player row directly for test setup and returns its ID.
func insertUserRow(t *testing.T, ctx context.Context, row do.User) int64 {
	t.Helper()
	id, err := dao.User.Ctx(ctx).Data(row).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user row failed: %v", err)
	}
	return id
}

// daoInsertActivation inserts an activation row directly for test setup, used to
// stage prior activations on a chosen activity date without going through the
// daily-limit-guarded Activate path. It returns the inserted ID.
func daoInsertActivation(ctx context.Context, userID, niuID int64, activityDate string, isFirst, orderNo int) (int64, error) {
	return dao.Activation.Ctx(ctx).Data(do.Activation{
		UserId:       userID,
		NiuId:        niuID,
		ActivityDate: activityDate,
		IsFirst:      isFirst,
		OrderNo:      orderNo,
	}).InsertAndGetId()
}

// setupPostgreSQLActivationDB ensures the shared activation test database exists
// and truncates the plugin tables so the calling test starts clean. It skips the
// test when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLActivationDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL activation integration tests")
	}

	activationDBOnce.Do(func() {
		activationDBPrepErr = provisionSharedActivationDB(ctx, baseLink)
	})
	if activationDBPrepErr != nil {
		t.Fatalf("provision shared activation database failed: %v", activationDBPrepErr)
	}

	db := g.DB()
	for _, table := range sicauNiuTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedActivationDB creates the shared temporary database, points the
// default GoFrame group at it, applies the plugin install DDL files in order and
// records cleanup state for the package-level teardown.
func provisionSharedActivationDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresActivationTestDatabaseLink(baseLink)
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
	for _, name := range activationSchemaFiles {
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

	activationDBLink = dbLink
	activationDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedActivationDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database. It is a no-op when the shared
// database was never provisioned (for example when the DB-gated tests skipped).
func teardownSharedActivationDB() {
	if activationDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared activation database failed: %v\n", err)
	}
	if err := gdb.SetConfig(activationDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLActivationTestDatabase(ctx, activationDBLink); err != nil {
		fmt.Printf("drop shared activation database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared activation database
// afterwards so the temporary database and GoFrame config never leak.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedActivationDB()
	os.Exit(code)
}

// postgresActivationTestDatabaseLink returns a unique database link for the shared
// activation test database so concurrent or repeated package runs never collide.
func postgresActivationTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL activation base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_activation_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		activationNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLActivationTestDatabase removes the shared temporary database.
func dropPostgreSQLActivationTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresActivationSystemLink(*targetConfig)})
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

// postgresActivationSystemLink returns a PostgreSQL maintenance database link
// using the same host, credentials and extra parameters as the target link.
func postgresActivationSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		activationNormalizeExtra(config.Extra),
	)
}

// activationNormalizeExtra ensures the link extra parameters are prefixed with a
// single leading question mark when present.
func activationNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
