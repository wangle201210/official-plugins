// settlement_testhelper_test.go provides the database-gated test harness shared by
// the operator settlement integration tests. Database-backed assertions are skipped
// unless LINA_TEST_PGSQL_LINK is explicitly provided.
//
// All settlement DB tests share one temporary PostgreSQL database provisioned once
// per package run. A single shared database is used deliberately: GoFrame caches its
// default-group connection and per-table field metadata process-wide, so dropping
// and recreating a database per test would leave the cached default instance
// pointing at a dropped database and break the next test. Instead, each test stays
// self-contained and order independent by truncating the plugin tables before it
// runs. The shared database and the original GoFrame config are released once at the
// end of the package run.
//
// The settlement capability owns the settlement table and reads the C1-C5 player,
// cattle, activation, feeding, steal, gift, check-in, honor and grant tables, so the
// harness applies the 001-006 install DDL.

package settlement

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

// settlementTables lists the plugin tables truncated before each DB-gated test.
var settlementTables = []string{
	"plugin_sicau_niu_settlement",
	"plugin_sicau_niu_user_honor",
	"plugin_sicau_niu_honor_def",
	"plugin_sicau_niu_checkin",
	"plugin_sicau_niu_gift",
	"plugin_sicau_niu_steal",
	"plugin_sicau_niu_feeding",
	"plugin_sicau_niu_activation",
	"plugin_sicau_niu_niu",
	"plugin_sicau_niu_user",
}

// settlementSchemaFiles lists the plugin install DDL files applied to the shared
// test database in order: 001 user/college, 002 catalog, 003 activation, 004 grass,
// 005 honor, 006 settlement.
var settlementSchemaFiles = []string{
	"001-sicau-niu-identity.sql",
	"002-sicau-niu-catalog.sql",
	"003-sicau-niu-activation.sql",
	"004-sicau-niu-grass.sql",
	"005-sicau-niu-honor.sql",
	"006-sicau-niu-settlement.sql",
}

// settlementDBHarness holds the lazily-provisioned shared test database state.
var (
	settlementDBOnce           sync.Once
	settlementDBLink           string
	settlementDBPrepErr        error
	settlementDBOriginalConfig gdb.Config
)

// seq counters provide process-unique values for seeded rows so the active-unique
// indexes (openid, cattle code, honor code) are respected.
var (
	openidSeq    int64
	niuCodeSeq   int64
	honorCodeSeq int64
)

// insertUserRow inserts one player row with a unique openid, the given identity and
// device fingerprint, and returns its ID.
func insertUserRow(t *testing.T, ctx context.Context, nickname, identityType, deviceFingerprint string) int64 {
	t.Helper()
	seq := atomic.AddInt64(&openidSeq, 1)
	id, err := dao.User.Ctx(ctx).Data(do.User{
		Openid:            fmt.Sprintf("openid-%d", seq),
		Nickname:          nickname,
		IdentityType:      identityType,
		DeviceFingerprint: deviceFingerprint,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert user row failed: %v", err)
	}
	return id
}

// insertNiuRow inserts one cattle row with a unique code, the given name and status,
// and returns its ID.
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

// insertActivationRow inserts one activation for a player on a cattle at a distinct
// natural day derived from seq so the per-day unique constraint is respected.
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

// insertStealRow inserts one steal record between two players on a distinct day.
func insertStealRow(t *testing.T, ctx context.Context, actorID, targetID int64, seq int) {
	t.Helper()
	day := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, seq)
	_, err := dao.Steal.Ctx(ctx).Data(do.Steal{
		ActorUserId:  actorID,
		TargetUserId: targetID,
		Amount:       5,
		StealDate:    day.Format("2006-01-02"),
	}).Insert()
	if err != nil {
		t.Fatalf("insert steal row failed: %v", err)
	}
}

// insertGiftRow inserts one gift record between two players on a distinct day.
func insertGiftRow(t *testing.T, ctx context.Context, fromID, toID int64, seq int) {
	t.Helper()
	day := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, seq)
	_, err := dao.Gift.Ctx(ctx).Data(do.Gift{
		FromUserId: fromID,
		ToUserId:   toID,
		Amount:     12,
		GiftDate:   day.Format("2006-01-02"),
	}).Insert()
	if err != nil {
		t.Fatalf("insert gift row failed: %v", err)
	}
}

// insertCheckinRow inserts one check-in for a player on a distinct day derived from
// seq so the per-day unique constraint is respected.
func insertCheckinRow(t *testing.T, ctx context.Context, userID int64, seq int) {
	t.Helper()
	day := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, seq)
	_, err := dao.Checkin.Ctx(ctx).Data(do.Checkin{
		UserId:       userID,
		CheckinDate:  day.Format("2006-01-02"),
		Amount:       20,
	}).Insert()
	if err != nil {
		t.Fatalf("insert checkin row failed: %v", err)
	}
}

// insertHonorDefRow inserts one honor definition with a unique code and returns its
// ID.
func insertHonorDefRow(t *testing.T, ctx context.Context, honorType, unlockType string, threshold int) int64 {
	t.Helper()
	seq := atomic.AddInt64(&honorCodeSeq, 1)
	id, err := dao.HonorDef.Ctx(ctx).Data(do.HonorDef{
		HonorType:  honorType,
		Code:       fmt.Sprintf("honor-%d", seq),
		Name:       "荣誉",
		UnlockType: unlockType,
		Threshold:  threshold,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert honor def row failed: %v", err)
	}
	return id
}

// countUserHonors returns the number of active grant rows for a honor.
func countUserHonors(t *testing.T, ctx context.Context, honorID int64) int {
	t.Helper()
	count, err := dao.UserHonor.Ctx(ctx).Where(dao.UserHonor.Columns().HonorId, honorID).Count()
	if err != nil {
		t.Fatalf("count user honors failed: %v", err)
	}
	return count
}

// setupPostgreSQLSettlementDB ensures the shared settlement test database exists and
// truncates the plugin tables so the calling test starts clean. It skips the test
// when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLSettlementDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL settlement integration tests")
	}

	settlementDBOnce.Do(func() {
		settlementDBPrepErr = provisionSharedSettlementDB(ctx, baseLink)
	})
	if settlementDBPrepErr != nil {
		t.Fatalf("provision shared settlement database failed: %v", settlementDBPrepErr)
	}

	db := g.DB()
	for _, table := range settlementTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedSettlementDB creates the shared temporary database, points the
// default GoFrame group at it, applies the plugin install DDL files in order and
// records cleanup state for the package-level teardown.
func provisionSharedSettlementDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresSettlementTestDatabaseLink(baseLink)
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
	for _, name := range settlementSchemaFiles {
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

	settlementDBLink = dbLink
	settlementDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedSettlementDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database.
func teardownSharedSettlementDB() {
	if settlementDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared settlement database failed: %v\n", err)
	}
	if err := gdb.SetConfig(settlementDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLSettlementTestDatabase(ctx, settlementDBLink); err != nil {
		fmt.Printf("drop shared settlement database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared settlement database.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedSettlementDB()
	os.Exit(code)
}

// postgresSettlementTestDatabaseLink returns a unique database link for the shared
// settlement test database so concurrent or repeated package runs never collide.
func postgresSettlementTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL settlement base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_settlement_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		settlementNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLSettlementTestDatabase removes the shared temporary database.
func dropPostgreSQLSettlementTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresSettlementSystemLink(*targetConfig)})
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

// postgresSettlementSystemLink returns a PostgreSQL maintenance database link using
// the same host, credentials and extra parameters as the target link.
func postgresSettlementSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		settlementNormalizeExtra(config.Extra),
	)
}

// settlementNormalizeExtra ensures the link extra parameters are prefixed with a
// single leading question mark when present.
func settlementNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
