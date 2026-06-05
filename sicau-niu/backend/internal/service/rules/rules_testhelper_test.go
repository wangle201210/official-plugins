// rules_testhelper_test.go provides the database-gated test harness shared by the
// runtime-rule service integration tests. Database-backed assertions are skipped
// unless LINA_TEST_PGSQL_LINK is explicitly provided.

package rules

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

	_ "lina-core/pkg/dbdriver" // registers the supported PostgreSQL GoFrame driver.
	"lina-core/pkg/dialect"
)

var (
	rulesDBOnce           sync.Once
	rulesDBLink           string
	rulesDBPrepErr        error
	rulesDBOriginalConfig gdb.Config
)

// setupPostgreSQLRulesDB ensures the shared rules test database exists and
// truncates the rule_config table so the calling test starts clean.
func setupPostgreSQLRulesDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL rules integration tests")
	}

	rulesDBOnce.Do(func() {
		rulesDBPrepErr = provisionSharedRulesDB(ctx, baseLink)
	})
	if rulesDBPrepErr != nil {
		t.Fatalf("provision shared rules database failed: %v", rulesDBPrepErr)
	}

	if _, err := g.DB().Exec(ctx, `TRUNCATE TABLE plugin_sicau_niu_rule_config RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate plugin_sicau_niu_rule_config failed: %v", err)
	}
}

func provisionSharedRulesDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresRulesTestDatabaseLink(baseLink)
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
	sqlPath := filepath.Join("..", "..", "..", "..", "manifest", "sql", "007-sicau-niu-rule-config.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	translated, err := dbDialect.TranslateDDL(ctx, sqlPath, string(content))
	if err != nil {
		return err
	}
	for _, statement := range dialect.SplitSQLStatements(translated) {
		if _, execErr := db.Exec(ctx, statement); execErr != nil {
			return fmt.Errorf("execute schema SQL failed: %w\nSQL:\n%s", execErr, statement)
		}
	}

	rulesDBLink = dbLink
	rulesDBOriginalConfig = originalConfig
	return nil
}

func teardownSharedRulesDB() {
	if rulesDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared rules database failed: %v\n", err)
	}
	if err := gdb.SetConfig(rulesDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLRulesTestDatabase(ctx, rulesDBLink); err != nil {
		fmt.Printf("drop shared rules database failed: %v\n", err)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedRulesDB()
	os.Exit(code)
}

func postgresRulesTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL rules base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_rules_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		rulesNormalizeExtra(config.Extra),
	), nil
}

func dropPostgreSQLRulesTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresRulesSystemLink(*targetConfig)})
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

func postgresRulesSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		rulesNormalizeExtra(config.Extra),
	)
}

func rulesNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
