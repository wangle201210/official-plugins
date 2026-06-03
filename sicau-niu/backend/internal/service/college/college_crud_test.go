// college_crud_test.go verifies the college dictionary CRUD behaviour against a
// gated PostgreSQL database: duplicate-name rejection on create, reference
// protection on delete and successful soft-delete of an unreferenced college. It
// also covers name and pagination validation as pure logic. Database-backed
// assertions are skipped unless LINA_TEST_PGSQL_LINK is set.
//
// All college DB tests share one temporary PostgreSQL database provisioned once
// per package run. A single shared database is used deliberately: GoFrame caches
// its default-group connection and per-table field metadata process-wide, so
// dropping and recreating a database per test would leave the cached default
// instance pointing at a dropped database and break the next test. Each test
// stays self-contained and order independent by truncating the plugin tables
// before it runs via setupPostgreSQLCollegeDB.

package college

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
var sicauNiuTables = []string{"plugin_sicau_niu_user", "plugin_sicau_niu_college"}

// collegeDBHarness holds the lazily-provisioned shared test database state.
var (
	collegeDBOnce           sync.Once
	collegeDBLink           string
	collegeDBPrepErr        error
	collegeDBOriginalConfig gdb.Config
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

// TestValidateNameRejectsBlank verifies blank or whitespace names are rejected.
// Pure-logic assertion that runs even when the DB harness skips.
func TestValidateNameRejectsBlank(t *testing.T) {
	for _, in := range []*MutateInput{nil, {Name: ""}, {Name: "   "}} {
		_, err := validateName(in)
		assertBizCode(t, err, CodeCollegeNameRequired.RuntimeCode())
	}

	name, err := validateName(&MutateInput{Name: "  Science  "})
	if err != nil {
		t.Fatalf("expected valid name, got %v", err)
	}
	if name != "Science" {
		t.Fatalf("expected trimmed name Science, got %q", name)
	}
}

// TestNormalizePagination verifies paging defaults and the max page-size cap.
func TestNormalizePagination(t *testing.T) {
	cases := []struct {
		in       *ListInput
		wantNum  int
		wantSize int
	}{
		{in: nil, wantNum: defaultPageNum, wantSize: defaultPageSize},
		{in: &ListInput{PageNum: 0, PageSize: 0}, wantNum: defaultPageNum, wantSize: defaultPageSize},
		{in: &ListInput{PageNum: 3, PageSize: 25}, wantNum: 3, wantSize: 25},
		{in: &ListInput{PageNum: -1, PageSize: 1000}, wantNum: defaultPageNum, wantSize: maxPageSize},
	}
	for _, tc := range cases {
		gotNum, gotSize := normalizePagination(tc.in)
		if gotNum != tc.wantNum || gotSize != tc.wantSize {
			t.Fatalf("normalizePagination(%#v) = (%d,%d), want (%d,%d)", tc.in, gotNum, gotSize, tc.wantNum, tc.wantSize)
		}
	}
}

// TestCreateRejectsDuplicateName verifies a second college with the same name is
// rejected with CodeCollegeNameExists while the first remains.
func TestCreateRejectsDuplicateName(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCollegeDB(t, ctx)

	svc := New()
	const name = "Veterinary Medicine"
	if _, err := svc.Create(ctx, &MutateInput{Name: name, Sort: 1}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err := svc.Create(ctx, &MutateInput{Name: name, Sort: 2})
	assertBizCode(t, err, CodeCollegeNameExists.RuntimeCode())

	count, err := dao.College.Ctx(ctx).Where(dao.College.Columns().Name, name).Count()
	if err != nil {
		t.Fatalf("count colleges failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one college after duplicate create, got %d", count)
	}
}

// TestDeleteReferencedCollegeRejectedAndUnreferencedSoftDeletes verifies a
// college referenced by a player cannot be deleted while an unreferenced college
// soft-deletes successfully.
func TestDeleteReferencedCollegeRejectedAndUnreferencedSoftDeletes(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCollegeDB(t, ctx)

	svc := New()
	referencedID, err := svc.Create(ctx, &MutateInput{Name: "Referenced College", Sort: 1})
	if err != nil {
		t.Fatalf("create referenced college failed: %v", err)
	}
	unreferencedID, err := svc.Create(ctx, &MutateInput{Name: "Unreferenced College", Sort: 2})
	if err != nil {
		t.Fatalf("create unreferenced college failed: %v", err)
	}

	// Attach a player to the referenced college so delete is protected.
	if _, err = dao.User.Ctx(ctx).Data(do.User{
		Openid:    "openid-college-ref",
		CollegeId: referencedID,
	}).InsertAndGetId(); err != nil {
		t.Fatalf("insert referencing player failed: %v", err)
	}

	err = svc.Delete(ctx, referencedID)
	assertBizCode(t, err, CodeCollegeReferenced.RuntimeCode())

	// The referenced college must still exist.
	exists, err := svc.Exists(ctx, referencedID)
	if err != nil {
		t.Fatalf("exists check failed: %v", err)
	}
	if !exists {
		t.Fatal("expected referenced college to remain after rejected delete")
	}

	// The unreferenced college soft-deletes.
	if err = svc.Delete(ctx, unreferencedID); err != nil {
		t.Fatalf("delete unreferenced college failed: %v", err)
	}
	exists, err = svc.Exists(ctx, unreferencedID)
	if err != nil {
		t.Fatalf("exists check after delete failed: %v", err)
	}
	if exists {
		t.Fatal("expected unreferenced college to be soft-deleted")
	}

	// The same name can be reused after the soft delete because the unique index
	// excludes soft-deleted rows.
	if _, err = svc.Create(ctx, &MutateInput{Name: "Unreferenced College", Sort: 3}); err != nil {
		t.Fatalf("recreate after soft delete failed: %v", err)
	}
}

// setupPostgreSQLCollegeDB ensures the shared college test database exists and
// truncates the plugin tables so the calling test starts clean. It skips the
// test when LINA_TEST_PGSQL_LINK is not set.
func setupPostgreSQLCollegeDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL college integration tests")
	}

	collegeDBOnce.Do(func() {
		collegeDBPrepErr = provisionSharedCollegeDB(ctx, baseLink)
	})
	if collegeDBPrepErr != nil {
		t.Fatalf("provision shared college database failed: %v", collegeDBPrepErr)
	}

	db := g.DB()
	for _, table := range sicauNiuTables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, table)); err != nil {
			t.Fatalf("truncate %s failed: %v", table, err)
		}
	}
}

// provisionSharedCollegeDB creates the shared temporary database, points the
// default GoFrame group at it and applies the plugin install DDL.
func provisionSharedCollegeDB(ctx context.Context, baseLink string) error {
	dbLink, err := postgresCollegeTestDatabaseLink(baseLink)
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

	collegeDBLink = dbLink
	collegeDBOriginalConfig = originalConfig
	return nil
}

// teardownSharedCollegeDB closes the shared connection, restores the original
// GoFrame config and drops the temporary database. It is a no-op when the shared
// database was never provisioned.
func teardownSharedCollegeDB() {
	if collegeDBLink == "" {
		return
	}
	ctx := context.Background()
	if err := g.DB().Close(ctx); err != nil {
		fmt.Printf("close shared college database failed: %v\n", err)
	}
	if err := gdb.SetConfig(collegeDBOriginalConfig); err != nil {
		fmt.Printf("restore GoFrame database config failed: %v\n", err)
	}
	if err := dropPostgreSQLCollegeTestDatabase(ctx, collegeDBLink); err != nil {
		fmt.Printf("drop shared college database failed: %v\n", err)
	}
}

// TestMain runs the package tests and tears down the shared college database
// afterwards so the temporary database and GoFrame config never leak.
func TestMain(m *testing.M) {
	code := m.Run()
	teardownSharedCollegeDB()
	os.Exit(code)
}

// postgresCollegeTestDatabaseLink returns a unique database link for the shared
// college test database.
func postgresCollegeTestDatabaseLink(baseLink string) (string, error) {
	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		return "", err
	}
	config := db.GetConfig()
	if config == nil {
		_ = db.Close(context.Background())
		return "", fmt.Errorf("PostgreSQL college base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		return "", closeErr
	}

	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sicau_niu_college_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		collegeNormalizeExtra(config.Extra),
	), nil
}

// dropPostgreSQLCollegeTestDatabase removes the shared temporary database.
func dropPostgreSQLCollegeTestDatabase(ctx context.Context, targetLink string) (err error) {
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

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresCollegeSystemLink(*targetConfig)})
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

// postgresCollegeSystemLink returns a PostgreSQL maintenance database link using
// the same host, credentials and extra parameters as the target link.
func postgresCollegeSystemLink(config gdb.ConfigNode) string {
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		collegeNormalizeExtra(config.Extra),
	)
}

// collegeNormalizeExtra ensures the link extra parameters are prefixed with a
// single leading question mark when present.
func collegeNormalizeExtra(extra string) string {
	extra = strings.TrimSpace(extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return extra
}
