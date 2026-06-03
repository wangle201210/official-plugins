// honor_testhelper_test.go provides the database-gated test harness shared by the
// honor integration tests. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All honor DB tests share one temporary PostgreSQL database provisioned once per
// package run. A single shared database is used deliberately: GoFrame caches its
// default-group connection and per-table field metadata process-wide, so dropping
// and recreating a database per test would leave the cached default instance
// pointing at a dropped database and break the next test. Instead, each test stays
// self-contained and order independent by truncating the plugin tables before it
// runs. The shared database and the original GoFrame config are released once at
// the end of the package run.
//
// The honor capability owns the honor_def table and reads the C3/C4 feeding,
// activation and card tables for the player unlock computation, so the harness
// applies the 001 identity, 002 catalog, 003 activation, 004 grass and 005 honor
// DDL.

package honor

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

// honorTables lists the plugin tables truncated before each DB-gated test.
var honorTables = []string{
	"plugin_sicau_niu_honor_def",
	"plugin_sicau_niu_feeding",
	"plugin_sicau_niu_activation",
	"plugin_sicau_niu_card",
	"plugin_sicau_niu_user",
}

// honorSchemaFiles lists the plugin install DDL files applied to the shared test
// database in order: 001 user/college, 002 catalog (niu/card), 003 activation,
// 004 grass (feeding), 005 honor.
var honorSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"003-sicau-niu-activation.sql",
	"004-sicau-niu-grass.sql",
	"005-sicau-niu-honor.sql",
}

// honorDBHarness holds the lazily-provisioned shared test database state.
var (
	honorDBOnce           sync.Once
	honorDBLink           string
	honorDBPrepErr        error
	honorDBOriginalConfig gdb.Config
)

// insertUserRow inserts one player row directly for test setup and returns its ID.
func insertUserRow(t *testing.T, ctx context.Context, nickname string) int64 {
	t.Helper()
	id, err := dao.User.Ctx(ctx).Data(do.User{Nickname: nickname}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user row failed: %v", err)
	}
	return id
}

// insertFeedingRow inserts one feeding record for a player so feed_count rules can
// be exercised.
func insertFeedingRow(t *testing.T, ctx context.Context, userID int64) {
	t.Helper()
	_, err := dao.Feeding.Ctx(ctx).Data(do.Feeding{
		UserId:           userID,
		NiuId:            1,
		BaseAmount:       10,
		CoefficientBasis: 100,
		EffectAmount:     10,
		IsIronBonus:      0,
	}).Insert()
	if err != nil {
		t.Fatalf("insert feeding row failed: %v", err)
	}
}

// insertActivationRow inserts one activation record of a cattle for a player so
// activation_count and collection rules can be exercised.
func insertActivationRow(t *testing.T, ctx context.Context, userID, niuID int64) {
	t.Helper()
	// Each activation is on a distinct natural day (one activation per player per
	// day), derived from niuID so the per-day unique constraint is respected.
	activityDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, int(niuID)).
		Format("2006-01-02")
	_, err := dao.Activation.Ctx(ctx).Data(do.Activation{
		UserId:       userID,
		NiuId:        niuID,
		ActivityDate: activityDate,
		IsFirst:      0,
		OrderNo:      1,
	}).Insert()
	if err != nil {
		t.Fatalf("insert activation row failed: %v", err)
	}
}

// insertCardRow inserts one main card of a cattle in a category and returns the
// cattle ID it binds, so collection completion rules can be exercised.
func insertCardRow(t *testing.T, ctx context.Context, niuID int64, category string) {
	t.Helper()
	_, err := dao.Card.Ctx(ctx).Data(do.Card{
		NiuId:    niuID,
		Category: category,
		Title:    "card",
	}).Insert()
	if err != nil {
		t.Fatalf("insert card row failed: %v", err)
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

// setupPostgreSQLHonorDB ensures the shared honor test database exists and
// truncates the plugin tables so the calling test starts clean. It skips the test
// when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLHonorDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL honor integration tests")
	}

	honorDBOnce.Do(func() {
		honorDBPrepErr = provisionSharedHonorDB(ctx, baseLink)
	})
	if honorDBPrepErr != nil {
		t.Fatalf("provision shared honor database failed: %v", honorDBPrepErr)
	}

	db := g.DB()
	for _, table := range honorTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedHonorDB creates the shared temporary database, points the default
// GoFrame group at it, applies the plugin install DDL files in order and records
// cleanup state for the package-level teardown.
func provisionSharedHonorDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresHonorTestDatabaseLink(baseLink)
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
	for _, name := range honorSchemaFiles {
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

	honorDBLink = dbLink
	honorDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedHonorDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database.
func teardownSharedHonorDB() {
	if honorDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared honor database failed: %v\n", err)
	}
	if err := gdb.SetConfig(honorDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLHonorTestDatabase(ctx, honorDBLink); err != nil {
		fmt.Printf("drop shared honor database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared honor database.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedHonorDB()
	os.Exit(code)
}

// postgresHonorTestDatabaseLink returns a unique database link for the shared
// honor test database so concurrent or repeated package runs never collide.
func postgresHonorTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL honor base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_honor_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		honorNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLHonorTestDatabase removes the shared temporary database.
func dropPostgreSQLHonorTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresHonorSystemLink(*targetConfig)})
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

// postgresHonorSystemLink returns a PostgreSQL maintenance database link using the
// same host, credentials and extra parameters as the target link.
func postgresHonorSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		honorNormalizeExtra(config.Extra),
	)
}

// honorNormalizeExtra ensures the link extra parameters are prefixed with a single
// leading question mark when present.
func honorNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
