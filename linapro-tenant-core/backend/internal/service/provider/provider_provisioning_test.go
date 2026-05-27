// This file verifies startup tenant-plugin provisioning exposed through the
// host-facing linapro-tenant-core provider.

package provider

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-tenant-core/backend/internal/dao"
	"lina-plugin-linapro-tenant-core/backend/internal/model/do"
	"lina-plugin-linapro-tenant-core/backend/internal/service/membership"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolver"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
	"lina-plugin-linapro-tenant-core/backend/internal/service/tenantplugin"
)

// TestProvisionAutoEnabledTenantPluginsAppliesExistingActiveTenants verifies
// startup reconciliation provisions tenant-scoped default plugins for active
// tenants that existed before the host image started.
func TestProvisionAutoEnabledTenantPluginsAppliesExistingActiveTenants(t *testing.T) {
	ctx := context.Background()
	configureProviderProvisioningTestDB(t, ctx)

	const (
		pluginID        = "provider-existing-tenant-default-plugin"
		activeTenant    = "provider-active-tenant"
		disabledTenant  = "provider-disabled-tenant"
		suspendedTenant = "provider-suspended-tenant"
	)
	activeTenantID := insertProviderProvisioningTenant(t, ctx, activeTenant, shared.TenantStatusActive)
	disabledTenantID := insertProviderProvisioningTenant(t, ctx, disabledTenant, shared.TenantStatusActive)
	suspendedTenantID := insertProviderProvisioningTenant(t, ctx, suspendedTenant, shared.TenantStatusSuspended)
	insertProviderProvisioningPlugin(t, ctx, pluginID, true)
	setProviderProvisioningPluginState(t, ctx, pluginID, disabledTenantID, false)
	t.Cleanup(func() {
		cleanupProviderProvisioningRows(t, ctx, pluginID, activeTenantID, disabledTenantID, suspendedTenantID)
	})

	bizCtxSvc := providerProvisioningBizCtxService{}
	membershipSvc := membership.New(bizCtxSvc)
	resolverConfigSvc := resolverconfig.New()
	tenantPluginSvc := tenantplugin.New(bizCtxSvc, nil)
	resolverSvc := resolver.New(bizCtxSvc, membershipSvc)
	providerSvc, err := New(membershipSvc, resolverSvc, resolverConfigSvc, tenantPluginSvc)
	if err != nil {
		t.Fatalf("create provider service failed: %v", err)
	}
	if err := providerSvc.ProvisionAutoEnabledTenantPlugins(ctx); err != nil {
		t.Fatalf("provision existing tenants failed: %v", err)
	}

	if enabled := providerProvisioningPluginEnabled(t, ctx, pluginID, activeTenantID); !enabled {
		t.Fatalf("expected active tenant %d to receive plugin %s enablement", activeTenantID, pluginID)
	}
	if enabled := providerProvisioningPluginEnabled(t, ctx, pluginID, disabledTenantID); enabled {
		t.Fatalf("expected explicit disabled tenant %d to preserve plugin %s disablement", disabledTenantID, pluginID)
	}
	if enabled := providerProvisioningPluginEnabled(t, ctx, pluginID, suspendedTenantID); enabled {
		t.Fatalf("expected suspended tenant %d to skip plugin %s enablement", suspendedTenantID, pluginID)
	}
}

// providerProvisioningBizCtxService is unused by startup provisioning but
// satisfies the tenantplugin constructor contract.
type providerProvisioningBizCtxService struct{}

// Current returns an empty plugin-visible business context.
func (providerProvisioningBizCtxService) Current(context.Context) plugincontract.CurrentContext {
	return plugincontract.CurrentContext{}
}

// configureProviderProvisioningTestDB points the provider package test at local PostgreSQL.
func configureProviderProvisioningTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure provider provisioning test database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close provider provisioning test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore provider provisioning test database config failed: %v", err)
		}
	})
}

// insertProviderProvisioningTenant creates one tenant row for provisioning tests.
func insertProviderProvisioningTenant(
	t *testing.T,
	ctx context.Context,
	code string,
	status shared.TenantStatus,
) int64 {
	t.Helper()
	if _, err := dao.Tenant.Ctx(ctx).Unscoped().Where(do.Tenant{Code: code}).Delete(); err != nil {
		t.Fatalf("cleanup stale tenant %s failed: %v", code, err)
	}
	id, err := dao.Tenant.Ctx(ctx).Data(do.Tenant{
		Code:   code,
		Name:   code,
		Status: string(status),
		Remark: "",
	}).OmitEmptyData().InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant %s failed: %v", code, err)
	}
	return id
}

// insertProviderProvisioningPlugin creates one tenant-scoped plugin registry row.
func insertProviderProvisioningPlugin(
	t *testing.T,
	ctx context.Context,
	pluginID string,
	autoEnableForNewTenants bool,
) {
	t.Helper()
	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().Where(do.SysPlugin{PluginId: pluginID}).Delete(); err != nil {
		t.Fatalf("cleanup stale plugin registry %s failed: %v", pluginID, err)
	}
	_, err := dao.SysPlugin.Ctx(ctx).OmitEmptyData().Data(do.SysPlugin{
		PluginId:                pluginID,
		Name:                    pluginID,
		Version:                 "v0.1.0",
		Type:                    "source",
		Installed:               1,
		Status:                  1,
		DesiredState:            "enabled",
		CurrentState:            "enabled",
		ScopeNature:             "tenant_aware",
		InstallMode:             "tenant_scoped",
		AutoEnableForNewTenants: autoEnableForNewTenants,
	}).InsertIgnore()
	if err != nil {
		t.Fatalf("insert plugin registry %s failed: %v", pluginID, err)
	}
}

// providerProvisioningPluginEnabled reads one tenant enablement row.
func providerProvisioningPluginEnabled(t *testing.T, ctx context.Context, pluginID string, tenantID int64) bool {
	t.Helper()
	value, err := dao.SysPluginState.Ctx(ctx).
		Where(do.SysPluginState{
			PluginId: pluginID,
			TenantId: tenantID,
			StateKey: "__tenant_enabled__",
		}).
		Value(dao.SysPluginState.Columns().Enabled)
	if err != nil {
		t.Fatalf("read tenant plugin state failed: %v", err)
	}
	return value != nil && !value.IsNil() && value.Bool()
}

// setProviderProvisioningPluginState writes one explicit tenant plugin state.
func setProviderProvisioningPluginState(
	t *testing.T,
	ctx context.Context,
	pluginID string,
	tenantID int64,
	enabled bool,
) {
	t.Helper()
	if _, err := dao.SysPluginState.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginState{
			PluginId: pluginID,
			TenantId: tenantID,
			StateKey: "__tenant_enabled__",
		}).
		Delete(); err != nil {
		t.Fatalf("cleanup explicit tenant plugin state failed: %v", err)
	}
	stateValue := "disabled"
	if enabled {
		stateValue = "enabled"
	}
	_, err := dao.SysPluginState.Ctx(ctx).OmitEmptyData().Data(do.SysPluginState{
		PluginId:   pluginID,
		TenantId:   tenantID,
		StateKey:   "__tenant_enabled__",
		StateValue: stateValue,
		Enabled:    enabled,
	}).Insert()
	if err != nil {
		t.Fatalf("insert explicit tenant plugin state failed: %v", err)
	}
}

// cleanupProviderProvisioningRows removes rows created by provider provisioning tests.
func cleanupProviderProvisioningRows(t *testing.T, ctx context.Context, pluginID string, tenantIDs ...int64) {
	t.Helper()
	if _, err := dao.SysPluginState.Ctx(ctx).Unscoped().Where(do.SysPluginState{PluginId: pluginID}).Delete(); err != nil {
		t.Errorf("cleanup tenant plugin state failed: %v", err)
	}
	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().Where(do.SysPlugin{PluginId: pluginID}).Delete(); err != nil {
		t.Errorf("cleanup plugin registry failed: %v", err)
	}
	if len(tenantIDs) > 0 {
		if _, err := dao.Tenant.Ctx(ctx).Unscoped().WhereIn(dao.Tenant.Columns().Id, tenantIDs).Delete(); err != nil {
			t.Errorf("cleanup tenant rows failed: %v", err)
		}
	}
}
