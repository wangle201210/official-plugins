// impersonate_impl.go implements tenant impersonation token issuance, parsing,
// and stop flows for the linapro-tenant-core plugin. It validates host user context,
// tenant membership, and token signer behavior before returning tenant-bound
// credentials or stable bizerr failures.

package impersonate

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/mssola/useragent"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
)

// Start validates an impersonation request and returns token metadata.
func (s *serviceImpl) Start(ctx context.Context, in StartInput) (*StartOutput, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	if !bizCtx.PlatformBypass {
		return nil, bizerr.NewCode(CodeImpersonationPermissionDenied)
	}
	actingUserID := int64(bizCtx.UserID)
	if actingUserID <= 0 {
		return nil, bizerr.NewCode(CodeImpersonationPermissionDenied)
	}
	platformAdmin, err := s.isPlatformAdmin(ctx, actingUserID)
	if err != nil {
		return nil, err
	}
	if !platformAdmin {
		return nil, bizerr.NewCode(CodeImpersonationPermissionDenied)
	}
	tenant, err := s.tenantSvc.Get(ctx, in.TenantID)
	if err != nil {
		return nil, err
	}
	if tenant.Status != string(shared.TenantStatusActive) {
		return nil, bizerr.NewCode(CodeImpersonationTenantUnavailable)
	}
	user, err := s.currentUser(ctx, actingUserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerr.NewCode(CodeImpersonationPermissionDenied)
	}
	if s.authSvc == nil {
		return nil, bizerr.NewCode(CodeImpersonationTokenUnavailable)
	}
	tokenOut, err := s.authSvc.IssueImpersonationToken(ctx, plugincontract.ImpersonationTokenIssueInput{
		ActingUserID: int(actingUserID),
		TenantID:     int(in.TenantID),
	})
	if err != nil {
		return nil, err
	}
	client := clientInfoFromCtx(ctx)
	if err = s.writeAuditLogs(ctx, auditInput{
		TenantID:     in.TenantID,
		ActingUserID: actingUserID,
		Username:     user.Username,
		Reason:       in.Reason,
		Client:       client,
	}); err != nil {
		return nil, err
	}
	return &StartOutput{
		Token:          tokenOut.AccessToken,
		TenantID:       int64(tokenOut.TenantID),
		ActingUserID:   int64(tokenOut.ActingUserID),
		IsImpersonated: true,
	}, nil
}

// Stop revokes one current impersonation token.
func (s *serviceImpl) Stop(ctx context.Context, in StopInput) error {
	tokenString := strings.TrimSpace(strings.TrimPrefix(in.Token, "Bearer "))
	if tokenString == "" {
		return bizerr.NewCode(CodeImpersonationTokenInvalid)
	}
	if s.authSvc == nil {
		return bizerr.NewCode(CodeImpersonationTokenUnavailable)
	}
	return s.authSvc.RevokeImpersonationToken(ctx, plugincontract.ImpersonationTokenRevokeInput{
		BearerToken: tokenString,
		TenantID:    int(in.TenantID),
	})
}

// currentUser returns the current platform user projection.
func (s *serviceImpl) currentUser(ctx context.Context, userID int64) (*userRow, error) {
	var user *userRow
	err := shared.Model(ctx, shared.TableSysUser).
		Fields("id", "username", "status").
		Where("id", userID).
		Scan(&user)
	return user, err
}

// isPlatformAdmin reports whether userID is bound to an all-data role in platform context.
func (s *serviceImpl) isPlatformAdmin(ctx context.Context, userID int64) (bool, error) {
	count, err := shared.Model(ctx, "sys_user_role").As("ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Where("ur.user_id", userID).
		Where("ur.tenant_id", shared.PlatformTenantID).
		Where("r.data_scope", 1).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// writeAuditLogs writes optional login and operation log rows when monitor tables exist.
func (s *serviceImpl) writeAuditLogs(ctx context.Context, in auditInput) error {
	tables, err := g.DB().Tables(ctx)
	if err != nil {
		return err
	}
	exists := make(map[string]struct{}, len(tables))
	for _, table := range tables {
		exists[table] = struct{}{}
	}
	if _, ok := exists["plugin_linapro_monitor_loginlog"]; ok {
		if _, err = shared.Model(ctx, "plugin_linapro_monitor_loginlog").Data(loginLogData{
			TenantID:           in.TenantID,
			ActingUserID:       in.ActingUserID,
			OnBehalfOfTenantID: in.TenantID,
			IsImpersonation:    true,
			UserName:           in.Username,
			Status:             0,
			IP:                 in.Client.IP,
			Browser:            in.Client.Browser,
			OS:                 in.Client.OS,
			Msg:                "Impersonation started",
		}).Insert(); err != nil {
			return err
		}
	}
	if _, ok := exists["plugin_linapro_monitor_operlog"]; ok {
		if _, err = shared.Model(ctx, "plugin_linapro_monitor_operlog").Data(operLogData{
			TenantID:           in.TenantID,
			ActingUserID:       in.ActingUserID,
			OnBehalfOfTenantID: in.TenantID,
			IsImpersonation:    true,
			Title:              "Tenant Impersonation",
			OperSummary:        in.Reason,
			RouteOwner:         "linapro-tenant-core",
			RouteMethod:        "POST",
			RoutePath:          "/platform/tenants/{id}/impersonate",
			RouteDocKey:        "platform.tenant.impersonate",
			OperType:           "other",
			Method:             "impersonate.Start",
			RequestMethod:      "POST",
			OperName:           in.Username,
			OperURL:            in.Client.URL,
			OperIP:             in.Client.IP,
			OperParam:          "{}",
			JsonResult:         "{}",
			Status:             0,
			ErrorMsg:           "",
			CostTime:           0,
		}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

// clientInfoFromCtx extracts browser, OS, IP, and URL metadata from the request.
func clientInfoFromCtx(ctx context.Context) clientInfo {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return clientInfo{}
	}
	browser, osName := parseUserAgent(request)
	return clientInfo{
		IP:      request.GetClientIp(),
		Browser: browser,
		OS:      osName,
		URL:     request.URL.String(),
	}
}

// parseUserAgent parses browser and OS names from a request.
func parseUserAgent(request *ghttp.Request) (string, string) {
	if request == nil {
		return "", ""
	}
	ua := useragent.New(request.GetHeader("User-Agent"))
	browserName, browserVersion := ua.Browser()
	return strings.TrimSpace(browserName + " " + browserVersion), ua.OS()
}
