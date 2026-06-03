// wall_testhelper_test.go provides the database-gated test harness shared by the
// public memorial-wall integration tests. Database-backed assertions are skipped
// unless LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All wall DB tests share one temporary PostgreSQL database provisioned once per
// package run. A single shared database is used deliberately: GoFrame caches its
// default-group connection and per-table field metadata process-wide, so dropping
// and recreating a database per test would leave the cached default instance
// pointing at a dropped database and break the next test. Instead, each test stays
// self-contained and order independent by truncating the plugin tables before it
// runs. The shared database and the original GoFrame config are released once at
// the end of the package run.
//
// The wall capability owns no table; it reads the C1 user, C2 cattle/card/quote
// and C3 activation tables, so the harness applies the 001 identity, 002 catalog
// and 003 activation DDL.

package wall

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

	_ "lina-core/pkg/dbdriver" // registers the supported PostgreSQL GoFrame driver.
	"lina-core/pkg/dialect"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

// wallTables lists the plugin tables truncated before each DB-gated test.
var wallTables = []string{
	"plugin_sicau_niu_activation",
	"plugin_sicau_niu_card",
	"plugin_sicau_niu_quote",
	"plugin_sicau_niu_niu",
	"plugin_sicau_niu_user",
}

// wallSchemaFiles lists the plugin install DDL files applied to the shared test
// database in order: 001 user/college, 002 catalog (niu/card/quote), 003
// activation.
var wallSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"003-sicau-niu-activation.sql",
}

// wallDBHarness holds the lazily-provisioned shared test database state.
var (
	wallDBOnce           sync.Once
	wallDBLink           string
	wallDBPrepErr        error
	wallDBOriginalConfig gdb.Config
)

// userOpenidSeq provides process-unique openids for seeded test players.
var userOpenidSeq int64

// niuCodeSeq provides process-unique codes for seeded test cattle.
var niuCodeSeq int64

// insertUserRow inserts one player row with a unique openid and a phone, and
// returns its ID. Seeding a phone lets the privacy assertions verify the public
// wall never surfaces it.
func insertUserRow(t *testing.T, ctx context.Context, nickname, identityType string) int64 {
	t.Helper()
	seq := atomic.AddInt64(&userOpenidSeq, 1)
	id, err := dao.User.Ctx(ctx).Data(do.User{
		Openid:       fmt.Sprintf("openid-%d", seq),
		Phone:        fmt.Sprintf("1380000%04d", seq),
		Nickname:     nickname,
		IdentityType: identityType,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user row failed: %v", err)
	}
	return id
}

// insertNiuRow inserts one cattle row with a unique code and the given name and
// status, and returns its ID.
func insertNiuRow(t *testing.T, ctx context.Context, name, status string) int64 {
	t.Helper()
	seq := atomic.AddInt64(&niuCodeSeq, 1)
	id, err := dao.Niu.Ctx(ctx).Data(do.Niu{
		Code:    fmt.Sprintf("NIU-%04d", seq),
		NiuType: "common",
		Name:    name,
		Status:  status,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert niu row failed: %v", err)
	}
	return id
}

// insertActivationRow inserts one activation for a player on a cattle at a
// distinct natural day and activation instant derived from seq, so the per-day
// unique constraint is respected and first-activator ordering is deterministic.
func insertActivationRow(t *testing.T, ctx context.Context, userID, niuID int64, isFirst, seq int) {
	t.Helper()
	day := time.Date(2026, time.January, 1, 8, 0, 0, 0, time.UTC).AddDate(0, 0, seq)
	_, err := dao.Activation.Ctx(ctx).Data(do.Activation{
		UserId:       userID,
		NiuId:        niuID,
		ActivityDate: day.Format("2006-01-02"),
		ActivatedAt:  &day,
		IsFirst:      isFirst,
		OrderNo:      1,
	}).Insert()
	if err != nil {
		t.Fatalf("insert activation row failed: %v", err)
	}
}

// insertCardRow inserts one campus-history card bound to a cattle and returns its
// ID.
func insertCardRow(t *testing.T, ctx context.Context, niuID int64, category, title string) int64 {
	t.Helper()
	id, err := dao.Card.Ctx(ctx).Data(do.Card{
		NiuId:    niuID,
		Category: category,
		Title:    title,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert card row failed: %v", err)
	}
	return id
}

// insertQuoteRow inserts one campus-history quote with the given enabled flag and
// returns its ID.
func insertQuoteRow(t *testing.T, ctx context.Context, content string, enabled int) int64 {
	t.Helper()
	id, err := dao.Quote.Ctx(ctx).Data(do.Quote{
		Content: content,
		Enabled: enabled,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert quote row failed: %v", err)
	}
	return id
}

// setupPostgreSQLWallDB ensures the shared wall test database exists and
// truncates the plugin tables so the calling test starts clean. It skips the test
// when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLWallDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL wall integration tests")
	}

	wallDBOnce.Do(func() {
		wallDBPrepErr = provisionSharedWallDB(ctx, baseLink)
	})
	if wallDBPrepErr != nil {
		t.Fatalf("provision shared wall database failed: %v", wallDBPrepErr)
	}

	db := g.DB()
	for _, table := range wallTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedWallDB creates the shared temporary database, points the default
// GoFrame group at it, applies the plugin install DDL files in order and records
// cleanup state for the package-level teardown.
func provisionSharedWallDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresWallTestDatabaseLink(baseLink)
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
	for _, name := range wallSchemaFiles {
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

	wallDBLink = dbLink
	wallDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedWallDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database.
func teardownSharedWallDB() {
	if wallDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared wall database failed: %v\n", err)
	}
	if err := gdb.SetConfig(wallDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLWallTestDatabase(ctx, wallDBLink); err != nil {
		fmt.Printf("drop shared wall database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared wall database.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedWallDB()
	os.Exit(code)
}

// postgresWallTestDatabaseLink returns a unique database link for the shared wall
// test database so concurrent or repeated package runs never collide.
func postgresWallTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL wall base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_wall_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		wallNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLWallTestDatabase removes the shared temporary database.
func dropPostgreSQLWallTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresWallSystemLink(*targetConfig)})
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

// postgresWallSystemLink returns a PostgreSQL maintenance database link using the
// same host, credentials and extra parameters as the target link.
func postgresWallSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		wallNormalizeExtra(config.Extra),
	)
}

// wallNormalizeExtra ensures the link extra parameters are prefixed with a single
// leading question mark when present.
func wallNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
