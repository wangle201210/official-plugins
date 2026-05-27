// provider_impl.go implements the host tenant-capability provider backed by
// linapro-tenant-core plugin tables. It injects tenant filters, membership checks, and
// platform fallback behavior so host services can remain decoupled from plugin
// storage details.

package provider

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-plugin-linapro-tenant-core/backend/internal/dao"
	"lina-plugin-linapro-tenant-core/backend/internal/model/do"
	"lina-plugin-linapro-tenant-core/backend/internal/model/entity"
	"lina-plugin-linapro-tenant-core/backend/internal/service/resolverconfig"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
)

// ResolveTenant resolves a tenant from request metadata.
func (p *Provider) ResolveTenant(ctx context.Context, request *ghttp.Request) (*tenantcap.ResolverResult, error) {
	config, err := p.resolverConfigSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	result, err := p.resolverSvc.Resolve(ctx, request, resolverconfig.ToResolverConfig(config))
	if err != nil {
		return nil, err
	}
	return &tenantcap.ResolverResult{
		TenantID:        tenantcap.TenantID(result.TenantID),
		Matched:         true,
		ActingAsTenant:  result.ActingAsTenant,
		IsImpersonation: result.ActingAsTenant,
	}, nil
}

// ValidateUserInTenant validates one user belongs to one tenant.
func (p *Provider) ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcap.TenantID) error {
	_, err := p.membershipSvc.GetByUserAndTenant(ctx, int64(userID), int64(tenantID))
	return err
}

// ListUserTenants returns tenant options for one user.
func (p *Provider) ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	tenants, err := p.membershipSvc.ListUserTenants(ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	result := make([]tenantcap.TenantInfo, 0, len(tenants))
	for _, item := range tenants {
		if item == nil {
			continue
		}
		result = append(result, tenantcap.TenantInfo{
			ID:     tenantcap.TenantID(item.Id),
			Code:   item.Code,
			Name:   item.Name,
			Status: item.Status,
		})
	}
	return result, nil
}

// SwitchTenant validates one user can switch to a target tenant.
func (p *Provider) SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error {
	return p.ValidateUserInTenant(ctx, userID, target)
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (p *Provider) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	return p.membershipSvc.ApplyUserTenantScope(ctx, model, userIDColumn)
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (p *Provider) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID tenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return p.membershipSvc.ApplyUserTenantFilter(ctx, model, userIDColumn, tenantID)
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (p *Provider) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*tenantcap.UserTenantProjection, error) {
	return p.membershipSvc.ListUserTenantProjections(ctx, userIDs)
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (p *Provider) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []tenantcap.TenantID,
	mode tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	return p.membershipSvc.ResolveUserTenantAssignment(ctx, requested, mode)
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (p *Provider) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *tenantcap.UserTenantAssignmentPlan,
) error {
	return p.membershipSvc.ReplaceUserTenantAssignments(ctx, userID, plan)
}

// EnsureUsersInTenant verifies every user has active membership in the tenant.
func (p *Provider) EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID tenantcap.TenantID) error {
	return p.membershipSvc.EnsureUsersInTenant(ctx, userIDs, tenantID)
}

// ValidateStartupConsistency returns user-membership startup consistency failures.
func (p *Provider) ValidateStartupConsistency(ctx context.Context) ([]string, error) {
	return p.membershipSvc.ValidateStartupConsistency(ctx)
}

// ProvisionAutoEnabledTenantPlugins provisions platform-approved tenant
// plugins for every existing active tenant. The host calls this during startup
// after plugin.autoEnable has enabled tenant-scoped plugins and after the
// linapro-tenant-core provider has registered through source-plugin route callbacks.
func (p *Provider) ProvisionAutoEnabledTenantPlugins(ctx context.Context) error {
	if p == nil || p.tenantPluginSvc == nil {
		return nil
	}
	var rows []*entity.Tenant
	err := dao.Tenant.Ctx(ctx).
		Where(do.Tenant{Status: string(shared.TenantStatusActive)}).
		OrderAsc(dao.Tenant.Columns().Id).
		Scan(&rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil || row.Id <= shared.PlatformTenantID {
			continue
		}
		if err = p.tenantPluginSvc.ProvisionForTenant(ctx, row.Id); err != nil {
			return err
		}
	}
	return nil
}
