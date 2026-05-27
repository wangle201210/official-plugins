// This file verifies impersonation token metadata and error contracts.

package impersonate

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	pluginmodeldo "lina-plugin-linapro-tenant-core/backend/internal/model/do"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
	tenantsvc "lina-plugin-linapro-tenant-core/backend/internal/service/tenant"
)

// TestImpersonationBusinessErrorMetadata verifies impersonation errors expose stable metadata.
func TestImpersonationBusinessErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		code        *bizerr.Code
		runtimeCode string
		messageKey  string
	}{
		{name: "permission denied", code: CodeImpersonationPermissionDenied, runtimeCode: "MULTI_TENANT_IMPERSONATION_PERMISSION_DENIED", messageKey: "error.multi.tenant.impersonation.permission.denied"},
		{name: "tenant unavailable", code: CodeImpersonationTenantUnavailable, runtimeCode: "MULTI_TENANT_IMPERSONATION_TENANT_UNAVAILABLE", messageKey: "error.multi.tenant.impersonation.tenant.unavailable"},
		{name: "token invalid", code: CodeImpersonationTokenInvalid, runtimeCode: "MULTI_TENANT_IMPERSONATION_TOKEN_INVALID", messageKey: "error.multi.tenant.impersonation.token.invalid"},
		{name: "token unavailable", code: CodeImpersonationTokenUnavailable, runtimeCode: "MULTI_TENANT_IMPERSONATION_TOKEN_UNAVAILABLE", messageKey: "error.multi.tenant.impersonation.token.unavailable"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			err := bizerr.NewCode(testCase.code)
			messageErr, ok := bizerr.As(err)
			if !ok {
				t.Fatalf("expected structured business error, got %T", err)
			}
			if messageErr.RuntimeCode() != testCase.runtimeCode {
				t.Fatalf("expected runtime code %q, got %q", testCase.runtimeCode, messageErr.RuntimeCode())
			}
			if messageErr.MessageKey() != testCase.messageKey {
				t.Fatalf("expected message key %q, got %q", testCase.messageKey, messageErr.MessageKey())
			}
		})
	}
}

// TestStartDelegatesTokenIssuanceToHostAuth verifies plugin impersonation does
// not read host JWT config and instead delegates token/session ownership to the
// host auth service.
func TestStartDelegatesTokenIssuanceToHostAuth(t *testing.T) {
	ctx := context.Background()
	configureImpersonationTestDB(t, ctx)

	username := fmt.Sprintf("impersonation-auth-delegate-%d", time.Now().UnixNano())
	userID := insertImpersonationTestPlatformAdmin(t, ctx, username)
	authSvc := &fakeImpersonationAuthService{}
	svc := &serviceImpl{
		authSvc:   authSvc,
		bizCtxSvc: impersonateGuardBizCtx{current: plugincontract.CurrentContext{UserID: int(userID), PlatformBypass: true}},
		tenantSvc: fakeImpersonationTenantService{tenant: &tenantsvc.Entity{Id: 42, Status: string(shared.TenantStatusActive)}},
	}

	out, err := svc.Start(ctx, StartInput{TenantID: 42, Reason: "unit test"})
	if err != nil {
		t.Fatalf("start impersonation: %v", err)
	}
	if out.Token != "host-impersonation-token" || out.TenantID != 42 || out.ActingUserID != userID || !out.IsImpersonated {
		t.Fatalf("unexpected impersonation output: %#v", out)
	}
	if authSvc.issuedActingUserID != int(userID) || authSvc.issuedTenantID != 42 {
		t.Fatalf("expected host auth issue call, got %#v", authSvc)
	}
}

// TestStopDelegatesTokenRevocationToHostAuth verifies impersonation stop keeps
// token parsing and revocation inside the host auth service.
func TestStopDelegatesTokenRevocationToHostAuth(t *testing.T) {
	authSvc := &fakeImpersonationAuthService{}
	svc := &serviceImpl{authSvc: authSvc}

	if err := svc.Stop(context.Background(), StopInput{TenantID: 42, Token: "Bearer host-impersonation-token"}); err != nil {
		t.Fatalf("stop impersonation: %v", err)
	}
	if authSvc.revokedBearer != "host-impersonation-token" || authSvc.revokedTenantID != 42 {
		t.Fatalf("expected host auth revoke call, got %#v", authSvc)
	}
}

// fakeImpersonationAuthService records host auth calls made by impersonation tests.
type fakeImpersonationAuthService struct {
	issuedActingUserID int
	issuedTenantID     int
	revokedBearer      string
	revokedTenantID    int
}

// SelectTenant is unused by impersonation tests.
func (s *fakeImpersonationAuthService) SelectTenant(
	context.Context,
	plugincontract.SelectTenantInput,
) (*plugincontract.TenantTokenOutput, error) {
	return &plugincontract.TenantTokenOutput{}, nil
}

// SwitchTenant is unused by impersonation tests.
func (s *fakeImpersonationAuthService) SwitchTenant(
	context.Context,
	plugincontract.SwitchTenantInput,
) (*plugincontract.TenantTokenOutput, error) {
	return &plugincontract.TenantTokenOutput{}, nil
}

// IssueImpersonationToken records the host auth impersonation request.
func (s *fakeImpersonationAuthService) IssueImpersonationToken(
	_ context.Context,
	in plugincontract.ImpersonationTokenIssueInput,
) (*plugincontract.ImpersonationTokenOutput, error) {
	s.issuedActingUserID = in.ActingUserID
	s.issuedTenantID = in.TenantID
	return &plugincontract.ImpersonationTokenOutput{
		AccessToken:  "host-impersonation-token",
		TokenID:      "host-token-id",
		TenantID:     in.TenantID,
		ActingUserID: in.ActingUserID,
	}, nil
}

// RevokeImpersonationToken records the host auth impersonation revoke request.
func (s *fakeImpersonationAuthService) RevokeImpersonationToken(
	_ context.Context,
	in plugincontract.ImpersonationTokenRevokeInput,
) error {
	s.revokedBearer = in.BearerToken
	s.revokedTenantID = in.TenantID
	return nil
}

// fakeImpersonationTenantService returns one configured tenant.
type fakeImpersonationTenantService struct {
	tenant *tenantsvc.Entity
}

// List is unused by impersonation tests.
func (s fakeImpersonationTenantService) List(context.Context, tenantsvc.ListInput) (*tenantsvc.ListOutput, error) {
	return &tenantsvc.ListOutput{}, nil
}

// Get returns the configured tenant.
func (s fakeImpersonationTenantService) Get(context.Context, int64) (*tenantsvc.Entity, error) {
	return s.tenant, nil
}

// Create is unused by impersonation tests.
func (s fakeImpersonationTenantService) Create(context.Context, tenantsvc.CreateInput) (int64, error) {
	return 0, nil
}

// Update is unused by impersonation tests.
func (s fakeImpersonationTenantService) Update(context.Context, tenantsvc.UpdateInput) error {
	return nil
}

// ChangeStatus is unused by impersonation tests.
func (s fakeImpersonationTenantService) ChangeStatus(context.Context, int64, shared.TenantStatus) error {
	return nil
}

// Delete is unused by impersonation tests.
func (s fakeImpersonationTenantService) Delete(context.Context, int64) error {
	return nil
}

// insertImpersonationTestPlatformAdmin creates a platform all-data role binding.
func insertImpersonationTestPlatformAdmin(t *testing.T, ctx context.Context, username string) int64 {
	t.Helper()

	userID, err := shared.Model(ctx, shared.TableSysUser).Data(pluginmodeldo.SysUser{
		Username: username,
		Password: "unused",
		Nickname: username,
		Status:   1,
		TenantId: shared.PlatformTenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert impersonation test user: %v", err)
	}

	roleKey := fmt.Sprintf("impersonation_admin_%d", time.Now().UnixNano())
	roleID, err := g.DB().Model("sys_role").Safe().Ctx(ctx).Data(platformRoleData{
		TenantID:  shared.PlatformTenantID,
		Name:      roleKey,
		Key:       roleKey,
		Sort:      1,
		DataScope: 1,
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert impersonation test role: %v", err)
	}
	if _, err = g.DB().Model("sys_user_role").Safe().Ctx(ctx).Data(platformUserRoleData{
		TenantID: shared.PlatformTenantID,
		UserID:   userID,
		RoleID:   roleID,
	}).Insert(); err != nil {
		t.Fatalf("insert impersonation test user role: %v", err)
	}

	t.Cleanup(func() {
		if _, err := g.DB().Model("sys_user_role").Safe().Ctx(ctx).Where(platformUserRoleData{
			TenantID: shared.PlatformTenantID,
			UserID:   userID,
			RoleID:   roleID,
		}).Delete(); err != nil {
			t.Errorf("cleanup impersonation user role: %v", err)
		}
		if _, err := g.DB().Model("sys_role").Safe().Ctx(ctx).Unscoped().Where("id", roleID).Delete(); err != nil {
			t.Errorf("cleanup impersonation role: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where(pluginmodeldo.SysUser{Id: userID}).Delete(); err != nil {
			t.Errorf("cleanup impersonation user: %v", err)
		}
	})
	return userID
}

// configureImpersonationTestDB points impersonation tests at the local
// PostgreSQL database and creates the minimal host tables they need.
func configureImpersonationTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure impersonation test database failed: %v", err)
	}
	db := g.DB()
	ensureImpersonationTestTables(t, ctx)
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close impersonation test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore impersonation test database config failed: %v", err)
		}
	})
}

// ensureImpersonationTestTables creates minimal host auth and role tables for
// impersonation service tests when the repository DB has not been initialized.
func ensureImpersonationTestTables(t *testing.T, ctx context.Context) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS sys_user (
			"id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			"tenant_id" BIGINT NOT NULL DEFAULT 0,
			"username" VARCHAR(64) NOT NULL,
			"password" VARCHAR(256) NOT NULL,
			"nickname" VARCHAR(64) NOT NULL DEFAULT '',
			"status" SMALLINT NOT NULL DEFAULT 1,
			"deleted_at" TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sys_role (
			"id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			"tenant_id" BIGINT NOT NULL DEFAULT 0,
			"name" VARCHAR(64) NOT NULL DEFAULT '',
			"key" VARCHAR(64) NOT NULL DEFAULT '',
			"sort" INT NOT NULL DEFAULT 0,
			"data_scope" SMALLINT NOT NULL DEFAULT 2,
			"status" SMALLINT NOT NULL DEFAULT 1,
			"deleted_at" TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sys_user_role (
			"tenant_id" BIGINT NOT NULL DEFAULT 0,
			"user_id" BIGINT NOT NULL,
			"role_id" BIGINT NOT NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("ensure impersonation test table failed: %v", err)
		}
	}
}
