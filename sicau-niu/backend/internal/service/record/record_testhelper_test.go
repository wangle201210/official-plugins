// record_testhelper_test.go provides the database-gated test harness shared by the
// activity-record query tests. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All record DB tests share one temporary PostgreSQL database provisioned once per
// package run; each test stays self-contained and order independent by truncating
// the plugin tables before it runs. The record capability reads the C1 player, C2
// cattle, C3 activation and C4 grass tables, so the harness applies the 001 identity,
// 002 catalog, 003 activation and 004 grass DDL.

package record

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

// recordTables lists the plugin tables truncated before each DB-gated test.
var recordTables = []string{
	"plugin_sicau_niu_feeding",
	"plugin_sicau_niu_steal",
	"plugin_sicau_niu_gift",
	"plugin_sicau_niu_checkin",
	"plugin_sicau_niu_grass_txn",
	"plugin_sicau_niu_activation",
	"plugin_sicau_niu_niu",
	"plugin_sicau_niu_user",
}

// recordSchemaFiles lists the plugin install DDL applied in order.
var recordSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"003-sicau-niu-activation.sql",
	"004-sicau-niu-grass.sql",
}

var (
	recordDBOnce           sync.Once
	recordDBLink           string
	recordDBPrepErr        error
	recordDBOriginalConfig gdb.Config
	openidSeq              int64
	niuCodeSeq             int64
)

// insertUserRow inserts one player with a unique openid and returns its ID.
func insertUserRow(t *testing.T, ctx context.Context, nickname string) int64 {
	t.Helper()
	id, err := dao.User.Ctx(ctx).Data(do.User{
		Openid:   fmt.Sprintf("openid-%d", atomic.AddInt64(&openidSeq, 1)),
		Nickname: nickname,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user failed: %v", err)
	}
	return id
}

// insertNiuRow inserts one cattle with a unique code and returns its ID.
func insertNiuRow(t *testing.T, ctx context.Context, name string) int64 {
	t.Helper()
	id, err := dao.Niu.Ctx(ctx).Data(do.Niu{
		Code: fmt.Sprintf("NIU-%04d", atomic.AddInt64(&niuCodeSeq, 1)), NiuType: "common", Name: name,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert niu failed: %v", err)
	}
	return id
}

// insertFeedingRow inserts one feeding record.
func insertFeedingRow(t *testing.T, ctx context.Context, userID, niuID int64, base, effect int) {
	t.Helper()
	_, err := dao.Feeding.Ctx(ctx).Data(do.Feeding{
		UserId: userID, NiuId: niuID, BaseAmount: base, CoefficientBasis: 100, EffectAmount: effect, IsIronBonus: 0,
	}).Insert()
	if err != nil {
		t.Fatalf("insert feeding failed: %v", err)
	}
}

// insertActivationRow inserts one activation record with optional uploaded photo
// evidence path.
func insertActivationRow(t *testing.T, ctx context.Context, userID, niuID int64, photoPath string) {
	t.Helper()
	now := time.Now().UTC()
	_, err := dao.Activation.Ctx(ctx).Data(do.Activation{
		UserId:       userID,
		NiuId:        niuID,
		ActivityDate: now.Format("2006-01-02"),
		ActivatedAt:  &now,
		IsFirst:      1,
		OrderNo:      1,
		PhotoPath:    photoPath,
	}).Insert()
	if err != nil {
		t.Fatalf("insert activation failed: %v", err)
	}
}

// insertStealRow inserts one steal record on a distinct day.
func insertStealRow(t *testing.T, ctx context.Context, actorID, targetID int64, seq int) {
	t.Helper()
	day := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, seq).Format("2006-01-02")
	_, err := dao.Steal.Ctx(ctx).Data(do.Steal{
		ActorUserId: actorID, TargetUserId: targetID, Amount: 5, StealDate: day,
	}).Insert()
	if err != nil {
		t.Fatalf("insert steal failed: %v", err)
	}
}

// insertGrassTxnRow inserts one grass ledger entry.
func insertGrassTxnRow(t *testing.T, ctx context.Context, userID, delta int64, txnType string) {
	t.Helper()
	_, err := dao.GrassTxn.Ctx(ctx).Data(do.GrassTxn{
		UserId: userID, Delta: delta, TxnType: txnType, RefId: 0,
	}).Insert()
	if err != nil {
		t.Fatalf("insert grass txn failed: %v", err)
	}
}

// setupPostgreSQLRecordDB ensures the shared record test database exists and
// truncates the plugin tables. It skips when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLRecordDB(t *testing.T, ctx context.Context) {
	t.Helper()
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL record integration tests")
	}
	recordDBOnce.Do(func() { recordDBPrepErr = provisionSharedRecordDB(ctx, baseLink) })
	if recordDBPrepErr != nil {
		t.Fatalf("provision shared record database failed: %v", recordDBPrepErr)
	}
	db := g.DB()
	for _, table := range recordTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedRecordDB creates the shared temporary database, applies the DDL and
// records cleanup state for teardown.
func provisionSharedRecordDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresRecordTestDatabaseLink(baseLink)
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
	if err = gdb.SetConfig(gdb.Config{gdb.DefaultGroupName: gdb.ConfigGroup{{Link: dbLink}}}); err != nil {
		return err
	}
	db := g.DB()
	for _, name := range recordSchemaFiles {
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
	recordDBLink = dbLink
	recordDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedRecordDB releases the shared database and restores config.
func teardownSharedRecordDB() {
	if recordDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared record database failed: %v\n", err)
	}
	if err := gdb.SetConfig(recordDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLRecordTestDatabase(ctx, recordDBLink); err != nil {
		fmt.Printf("drop shared record database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared record database.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedRecordDB()
	os.Exit(code)
}

// postgresRecordTestDatabaseLink returns a unique database link.
func postgresRecordTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL record base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_record_%d%s",
		config.User, config.Pass, config.Protocol, config.Host, config.Port,
		time.Now().UnixNano(), recordNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLRecordTestDatabase removes the shared temporary database.
func dropPostgreSQLRecordTestDatabase(ctx context.Context, targetLink string) (err error) {
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
	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresRecordSystemLink(*targetConfig)})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := systemDB.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = systemDB.Exec(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()", targetName); err != nil {
		return err
	}
	quotedName := `"` + strings.ReplaceAll(targetName, `"`, `""`) + `"`
	if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		return err
	}
	return nil
}

// postgresRecordSystemLink returns a maintenance database link.
func postgresRecordSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf("pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User, config.Pass, config.Protocol, config.Host, config.Port, recordNormalizeExtra(config.Extra))
}

// recordNormalizeExtra prefixes link extra parameters with a single question mark.
func recordNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
