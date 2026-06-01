// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

type IUidentityV1 interface {
	AccountImportCheck(ctx context.Context, req *v1.AccountImportCheckReq) (res *v1.AccountImportCheckRes, err error)
	AccountImport(ctx context.Context, req *v1.AccountImportReq) (res *v1.AccountImportRes, err error)
	AccountPassword(ctx context.Context, req *v1.AccountPasswordReq) (res *v1.AccountPasswordRes, err error)
	AccountPasswordChallenge(ctx context.Context, req *v1.AccountPasswordChallengeReq) (res *v1.AccountPasswordChallengeRes, err error)
	AccountPasswordPhoneVerify(ctx context.Context, req *v1.AccountPasswordPhoneVerifyReq) (res *v1.AccountPasswordPhoneVerifyRes, err error)
	AccountPasswordSelfReset(ctx context.Context, req *v1.AccountPasswordSelfResetReq) (res *v1.AccountPasswordSelfResetRes, err error)
	ActivationStart(ctx context.Context, req *v1.ActivationStartReq) (res *v1.ActivationStartRes, err error)
	ActivationFace(ctx context.Context, req *v1.ActivationFaceReq) (res *v1.ActivationFaceRes, err error)
	ActivationPassword(ctx context.Context, req *v1.ActivationPasswordReq) (res *v1.ActivationPasswordRes, err error)
	ActivationPhone(ctx context.Context, req *v1.ActivationPhoneReq) (res *v1.ActivationPhoneRes, err error)
	ActivationWechat(ctx context.Context, req *v1.ActivationWechatReq) (res *v1.ActivationWechatRes, err error)
	ActivationState(ctx context.Context, req *v1.ActivationStateReq) (res *v1.ActivationStateRes, err error)
	CasLogin(ctx context.Context, req *v1.CasLoginReq) (res *v1.CasLoginRes, err error)
	CasPasswordLogin(ctx context.Context, req *v1.CasPasswordLoginReq) (res *v1.CasPasswordLoginRes, err error)
	CasPhoneLogin(ctx context.Context, req *v1.CasPhoneLoginReq) (res *v1.CasPhoneLoginRes, err error)
	CasUnionIDLogin(ctx context.Context, req *v1.CasUnionIDLoginReq) (res *v1.CasUnionIDLoginRes, err error)
	CasServiceTicket(ctx context.Context, req *v1.CasServiceTicketReq) (res *v1.CasServiceTicketRes, err error)
	CasServiceValidate(ctx context.Context, req *v1.CasServiceValidateReq) (res *v1.CasServiceValidateRes, err error)
	CasTicketLogout(ctx context.Context, req *v1.CasTicketLogoutReq) (res *v1.CasTicketLogoutRes, err error)
	LegacyCASServiceValidateXML(ctx context.Context, req *v1.LegacyCASServiceValidateXMLReq) (res *v1.LegacyCASServiceValidateXMLRes, err error)
	LegacyCASConfig(ctx context.Context, req *v1.LegacyCASConfigReq) (res *v1.LegacyCASConfigRes, err error)
	LegacyLDAPConfig(ctx context.Context, req *v1.LegacyLDAPConfigReq) (res *v1.LegacyLDAPConfigRes, err error)
	LegacyOAuthConfig(ctx context.Context, req *v1.LegacyOAuthConfigReq) (res *v1.LegacyOAuthConfigRes, err error)
	LegacyTokenConfig(ctx context.Context, req *v1.LegacyTokenConfigReq) (res *v1.LegacyTokenConfigRes, err error)
	LegacyUpload(ctx context.Context, req *v1.LegacyUploadReq) (res *v1.LegacyUploadRes, err error)
	LegacyHealth(ctx context.Context, req *v1.LegacyHealthReq) (res *v1.LegacyHealthRes, err error)
	LegacyServerMonitor(ctx context.Context, req *v1.LegacyServerMonitorReq) (res *v1.LegacyServerMonitorRes, err error)
	LegacyLogSnapshot(ctx context.Context, req *v1.LegacyLogSnapshotReq) (res *v1.LegacyLogSnapshotRes, err error)
	LegacyExternalAction(ctx context.Context, req *v1.LegacyExternalActionReq) (res *v1.LegacyExternalActionRes, err error)
	OAuthIssue(ctx context.Context, req *v1.OAuthIssueReq) (res *v1.OAuthIssueRes, err error)
	OAuthAuthorizationCode(ctx context.Context, req *v1.OAuthAuthorizationCodeReq) (res *v1.OAuthAuthorizationCodeRes, err error)
	OAuthAccessToken(ctx context.Context, req *v1.OAuthAccessTokenReq) (res *v1.OAuthAccessTokenRes, err error)
	OAuthAccessTokenInfo(ctx context.Context, req *v1.OAuthAccessTokenInfoReq) (res *v1.OAuthAccessTokenInfoRes, err error)
	ResourceCreate(ctx context.Context, req *v1.ResourceCreateReq) (res *v1.ResourceCreateRes, err error)
	ResourceDelete(ctx context.Context, req *v1.ResourceDeleteReq) (res *v1.ResourceDeleteRes, err error)
	ResourceGet(ctx context.Context, req *v1.ResourceGetReq) (res *v1.ResourceGetRes, err error)
	ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error)
	ResourceUpdate(ctx context.Context, req *v1.ResourceUpdateReq) (res *v1.ResourceUpdateRes, err error)
	SmsSend(ctx context.Context, req *v1.SmsSendReq) (res *v1.SmsSendRes, err error)
	Stats(ctx context.Context, req *v1.StatsReq) (res *v1.StatsRes, err error)
	RuntimeTokenIssue(ctx context.Context, req *v1.RuntimeTokenIssueReq) (res *v1.RuntimeTokenIssueRes, err error)
	RuntimeTokenInfo(ctx context.Context, req *v1.RuntimeTokenInfoReq) (res *v1.RuntimeTokenInfoRes, err error)
	UserUnionIDLookup(ctx context.Context, req *v1.UserUnionIDLookupReq) (res *v1.UserUnionIDLookupRes, err error)
	UserUnionIDBind(ctx context.Context, req *v1.UserUnionIDBindReq) (res *v1.UserUnionIDBindRes, err error)
	UserPasswordChange(ctx context.Context, req *v1.UserPasswordChangeReq) (res *v1.UserPasswordChangeRes, err error)
	UserPhoneChange(ctx context.Context, req *v1.UserPhoneChangeReq) (res *v1.UserPhoneChangeRes, err error)
	UserEmailChange(ctx context.Context, req *v1.UserEmailChangeReq) (res *v1.UserEmailChangeRes, err error)
	UserQQChange(ctx context.Context, req *v1.UserQQChangeReq) (res *v1.UserQQChangeRes, err error)
	UserWechatUnbind(ctx context.Context, req *v1.UserWechatUnbindReq) (res *v1.UserWechatUnbindRes, err error)
	UserInfo(ctx context.Context, req *v1.UserInfoReq) (res *v1.UserInfoRes, err error)
	UserLoginLogs(ctx context.Context, req *v1.UserLoginLogsReq) (res *v1.UserLoginLogsRes, err error)
	UserApplications(ctx context.Context, req *v1.UserApplicationsReq) (res *v1.UserApplicationsRes, err error)
	UserAppRoles(ctx context.Context, req *v1.UserAppRolesReq) (res *v1.UserAppRolesRes, err error)
	UserAppRoleCreate(ctx context.Context, req *v1.UserAppRoleCreateReq) (res *v1.UserAppRoleCreateRes, err error)
	UserAppRoleUpdate(ctx context.Context, req *v1.UserAppRoleUpdateReq) (res *v1.UserAppRoleUpdateRes, err error)
}
