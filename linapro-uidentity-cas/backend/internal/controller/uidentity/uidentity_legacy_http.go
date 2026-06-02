// This file implements the old uidentity/admin HTTP contract as a thin
// compatibility layer. It translates legacy root paths, form field names, and
// response envelopes into the plugin service contract without creating a
// separate business service graph.

package uidentity

import (
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

const (
	legacyStatusOK        = 200
	legacyStatusError     = 500
	legacyTGTCookieName   = "go-admin-tgt"
	legacyTGTCookieMaxAge = 24 * 10 * time.Hour
	legacySysJobGroup     = "uidentity"
	legacySysJobType      = 1
	legacySysJobStatusOn  = 2
)

// LegacyController serves legacy uidentity/admin paths and response envelopes.
type LegacyController struct {
	uidentitySvc uidentitysvc.Service // UIdentity CAS domain service.
}

// NewLegacy creates one old-contract compatibility controller.
func NewLegacy(uidentitySvc uidentitysvc.Service) *LegacyController {
	return &LegacyController{uidentitySvc: uidentitySvc}
}

// ResourceList handles old CRUD list routes such as GET /api/v1/account.
func (c *LegacyController) ResourceList(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		pageIndex, pageSize := legacyPage(r)
		out, err := c.uidentitySvc.ListResource(r.Context(), uidentitysvc.ResourceListInput{
			Resource:    resource,
			PageNum:     pageIndex,
			PageSize:    pageSize,
			Keyword:     legacyKeyword(r),
			AccountId:   legacyInt64Param(r, "accountId", "account_id"),
			AppId:       legacyInt64Param(r, "appId", "app_id", "applicationId", "application_id"),
			GroupId:     legacyInt64Param(r, "groupId", "group_id"),
			ContainerId: legacyInt64Param(r, "containerId", "container_id"),
			UnitId:      legacyInt64Param(r, "unitId", "unit_id"),
			Status:      legacyOptionalInt(r, "status"),
			PassLevels:  legacyInt64ListParam(r, "passLevels", "pass_levels", "passLevel", "pass_level"),
			GroupIds:    legacyInt64ListParam(r, "groupIds", "group_ids"),
			OrderBy:     legacyStringParam(r, "orderBy", "order_by", "sort"),
			Order:       legacyStringParam(r, "order", "sortOrder", "sort_order"),
		})
		if err != nil {
			legacyError(r, err)
			return
		}
		legacyPageOK(r, out.List, out.Total, pageIndex, pageSize)
	}
}

// ResourceGet handles old CRUD detail routes.
func (c *LegacyController) ResourceGet(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		record, err := c.uidentitySvc.GetResource(r.Context(), resource, legacyRouterID(r))
		if err != nil {
			legacyError(r, err)
			return
		}
		legacyOK(r, record)
	}
}

// ResourceCreate handles old CRUD create routes.
func (c *LegacyController) ResourceCreate(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		id, err := c.uidentitySvc.CreateResource(r.Context(), resource, legacyRequestMap(r))
		if err != nil {
			legacyError(r, err)
			return
		}
		legacyOK(r, map[string]any{"id": id})
	}
}

// ResourceUpdate handles old CRUD update routes.
func (c *LegacyController) ResourceUpdate(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if err := c.uidentitySvc.UpdateResource(r.Context(), resource, legacyRouterID(r), legacyRequestMap(r)); err != nil {
			legacyError(r, err)
			return
		}
		legacyOK(r, nil)
	}
}

// ResourceDelete handles old CRUD delete routes.
func (c *LegacyController) ResourceDelete(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if err := c.uidentitySvc.DeleteResource(r.Context(), resource, legacyDeleteIDs(r)); err != nil {
			legacyError(r, err)
			return
		}
		legacyOK(r, nil)
	}
}

// AccountPasswordUnlock handles POST /api/v1/account/unlockPassword.
func (c *LegacyController) AccountPasswordUnlock(r *ghttp.Request) {
	numbers := legacyStringListParam(r, "numbers", "number", "usernames", "username")
	out, err := c.uidentitySvc.UnlockPasswordFailures(r.Context(), numbers)
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"numbers": out})
}

// AccountPasswordUpdate handles POST /api/v1/account/updatePassword.
func (c *LegacyController) AccountPasswordUpdate(r *ghttp.Request) {
	id := legacyInt64Param(r, "id", "accountId", "account_id")
	password := legacyStringParam(r, "password", "newPassword", "new_password")
	if err := c.uidentitySvc.ResetAccountPassword(r.Context(), id, password); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// AccountImportCheck handles POST /api/v1/account/importCheck.
func (c *LegacyController) AccountImportCheck(r *ghttp.Request) {
	out, err := c.uidentitySvc.CheckAccountImport(r.Context(), legacyAccountImportInput(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"rows": out.Rows})
}

// AccountImport handles POST /api/v1/account/import.
func (c *LegacyController) AccountImport(r *ghttp.Request) {
	out, err := c.uidentitySvc.ImportAccounts(r.Context(), legacyAccountImportInput(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"success": out.Success, "failedNumber": out.FailedNumber, "failed_number": out.FailedNumber})
}

// AccountPasswordChallenge handles POST /api/v1/account/updatePasswordGetUser.
func (c *LegacyController) AccountPasswordChallenge(r *ghttp.Request) {
	out, err := c.uidentitySvc.CreatePasswordChallenge(r.Context(), legacyStringParam(r, "number", "username"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"uuid": out.ChallengeID, "challengeId": out.ChallengeID, "status": out.Status})
}

// AccountPasswordPhoneVerify handles POST /api/v1/account/updatePasswordBySelfPhone.
func (c *LegacyController) AccountPasswordPhoneVerify(r *ghttp.Request) {
	token, err := c.uidentitySvc.VerifyPasswordChallengePhone(
		r.Context(),
		legacyChallengeID(r),
		legacyStringParam(r, "phone"),
		legacyStringParam(r, "code"),
	)
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"uuid": token, "challengeId": token})
}

// AccountPasswordSelfReset handles POST /api/v1/account/updatePasswordBySelf.
func (c *LegacyController) AccountPasswordSelfReset(r *ghttp.Request) {
	if err := c.uidentitySvc.ResetPasswordByChallenge(r.Context(), legacyChallengeID(r), legacyStringParam(r, "password", "newPassword", "new_password")); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// CasLoginByCookie handles GET /api/v1/cas/login.
func (c *LegacyController) CasLoginByCookie(r *ghttp.Request) {
	tgt := legacyStringParam(r, "tgt", "ticket")
	if tgt == "" {
		tgt = r.Cookie.Get(legacyTGTCookieName).String()
	}
	out, err := c.uidentitySvc.IssueServiceTicketFromTGT(r.Context(), uidentitysvc.ServiceTicketInput{
		ClientID:  legacyClientID(r),
		TGT:       tgt,
		AccountID: legacyInt64Param(r, "accountId", "account_id", "userId", "user_id"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"callbackUrl": out.CallbackURL, "callback_url": out.CallbackURL, "tgt": tgt, "st": out.ST})
}

// CasPasswordLogin handles POST /api/v1/cas/login.
func (c *LegacyController) CasPasswordLogin(r *ghttp.Request) {
	out, err := c.uidentitySvc.LoginByPassword(r.Context(), uidentitysvc.PasswordLoginInput{
		ClientID: legacyClientID(r),
		Number:   legacyStringParam(r, "number", "username"),
		Password: legacyStringParam(r, "password"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacySetTGTCookie(r, out)
	legacyOK(r, legacyRuntimeLoginPayload(out))
}

// CasPhoneLogin handles POST /api/v1/cas/loginByPhone.
func (c *LegacyController) CasPhoneLogin(r *ghttp.Request) {
	out, err := c.uidentitySvc.LoginByPhone(r.Context(), uidentitysvc.PhoneLoginInput{
		ClientID: legacyClientID(r),
		Phone:    legacyStringParam(r, "phone"),
		Code:     legacyStringParam(r, "code"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacySetTGTCookie(r, out)
	legacyOK(r, legacyRuntimeLoginPayload(out))
}

// CasUnionIDLogin handles GET /api/v1/cas/loginByUnionID.
func (c *LegacyController) CasUnionIDLogin(r *ghttp.Request) {
	out, err := c.uidentitySvc.LoginByUnionID(r.Context(), uidentitysvc.UnionIDLoginInput{
		ClientID: legacyClientID(r),
		UnionID:  legacyStringParam(r, "unionID", "unionId", "union_id", "uuid"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacySetTGTCookie(r, out)
	legacyOK(r, legacyRuntimeLoginPayload(out))
}

// CasTicketLogout handles DELETE /api/v1/cas/tickets/:ticket.
func (c *LegacyController) CasTicketLogout(r *ghttp.Request) {
	if err := c.uidentitySvc.DeleteTicket(r.Context(), legacyStringParam(r, "ticket")); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// CasServiceValidateXML handles old CAS XML validation paths.
func (c *LegacyController) CasServiceValidateXML(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyCASServiceXML(r.Context(), uidentitysvc.LegacyCASServiceXMLInput{
		Ticket: legacyStringParam(r, "ticket"),
		UserID: legacyInt64Param(r, "userId", "user_id"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	r.Response.Write(out.XML)
	r.Exit()
}

// CasGetLoginQR handles POST /api/v1/cas/getLoginQr.
func (c *LegacyController) CasGetLoginQR(r *ghttp.Request) {
	out, err := c.uidentitySvc.CreateWechatLoginQR(r.Context(), uidentitysvc.WechatLoginQRInput{
		ClientID: legacyClientID(r),
		Callback: legacyStringParam(r, "callback", "cascallback", "casCallback"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"uuid": out.State, "state": out.State, "url": out.URL, "expiredAt": out.ExpiredAt, "expired_at": out.ExpiredAt})
}

// CasLoginByQR handles GET /api/v1/cas/loginByQr.
func (c *LegacyController) CasLoginByQR(r *ghttp.Request) {
	out, err := c.uidentitySvc.CompleteWechatLoginQR(r.Context(), uidentitysvc.WechatLoginCallbackInput{
		State:    legacyStringParam(r, "state", "uuid"),
		ClientID: legacyClientID(r),
		Code:     legacyStringParam(r, "code"),
		UnionID:  legacyStringParam(r, "unionID", "unionId", "union_id"),
		Callback: legacyStringParam(r, "callback", "cascallback", "casCallback"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyWechatLoginResultPayload(out))
}

// CasGetLoginQRResult handles POST /api/v1/cas/getCasLoginQrRes.
func (c *LegacyController) CasGetLoginQRResult(r *ghttp.Request) {
	out, err := c.uidentitySvc.GetWechatLoginQRResult(r.Context(), legacyStringParam(r, "state", "uuid"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyWechatLoginResultPayload(out))
}

// SSOLogin handles old /sso/login paths using the same TGT-to-ST flow.
func (c *LegacyController) SSOLogin(r *ghttp.Request) {
	c.CasLoginByCookie(r)
}

// SSOLogout handles old /sso/logout paths.
func (c *LegacyController) SSOLogout(r *ghttp.Request) {
	ticket := legacyStringParam(r, "ticket", "tgt")
	if ticket == "" {
		ticket = r.Cookie.Get(legacyTGTCookieName).String()
	}
	if err := c.uidentitySvc.DeleteTicket(r.Context(), ticket); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// SSOLoginToken handles POST /api/v1/ssologin/getToken.
func (c *LegacyController) SSOLoginToken(r *ghttp.Request) {
	c.CasLoginByCookie(r)
}

// RuntimeTokenIssue handles POST /api/v1/token/get.
func (c *LegacyController) RuntimeTokenIssue(r *ghttp.Request) {
	out, err := c.uidentitySvc.IssueRuntimeToken(r.Context(), uidentitysvc.RuntimeTokenInput{
		ClientID: legacyClientID(r),
		Secret:   legacyStringParam(r, "secret", "clientSecret", "client_secret"),
		Number:   legacyStringParam(r, "number", "username"),
		Password: legacyStringParam(r, "password"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"AccessToken": out.AccessToken, "accessToken": out.AccessToken, "access_token": out.AccessToken, "expiredAt": out.ExpiredAt, "expired_at": out.ExpiredAt})
}

// RuntimeTokenInfo handles GET /api/v1/token/getUserInfoByToken.
func (c *LegacyController) RuntimeTokenInfo(r *ghttp.Request) {
	out, err := c.uidentitySvc.GetUserInfoByRuntimeToken(r.Context(), legacyAccessToken(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"user": legacyRuntimeAccountPayload(out.User), "users": legacyRuntimeAccountPayloads(out.Users), "app": legacyRuntimeApplicationPayload(out.App)})
}

// ActivationStart handles POST /api/v1/activate/baseInfo.
func (c *LegacyController) ActivationStart(r *ghttp.Request) {
	out, err := c.uidentitySvc.StartActivation(r.Context(), uidentitysvc.ActivationStartInput{
		Number: legacyStringParam(r, "number"),
		Name:   legacyStringParam(r, "name"),
		Idcard: legacyStringParam(r, "idcard", "idCard"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"uuid": out.ChallengeID, "challengeId": out.ChallengeID, "face": out.NeedFace, "status": out.Status})
}

// ActivationFace handles POST /api/v1/activate/face.
func (c *LegacyController) ActivationFace(r *ghttp.Request) {
	out, err := c.uidentitySvc.RecordActivationFace(r.Context(), uidentitysvc.ActivationFaceInput{
		ChallengeID: legacyChallengeID(r),
		FaceURL:     legacyStringParam(r, "faceUrl", "face_url", "face"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyActivationStepPayload(out))
}

// ActivationPassword handles POST /api/v1/activate/password.
func (c *LegacyController) ActivationPassword(r *ghttp.Request) {
	out, err := c.uidentitySvc.SetActivationPassword(r.Context(), uidentitysvc.ActivationPasswordInput{
		ChallengeID: legacyChallengeID(r),
		Password:    legacyStringParam(r, "password"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyActivationStepPayload(out))
}

// ActivationPhone handles POST /api/v1/activate/phone.
func (c *LegacyController) ActivationPhone(r *ghttp.Request) {
	out, err := c.uidentitySvc.SetActivationPhone(r.Context(), uidentitysvc.ActivationPhoneInput{
		ChallengeID: legacyChallengeID(r),
		Phone:       legacyStringParam(r, "phone"),
		Code:        legacyStringParam(r, "code"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyActivationStepPayload(out))
}

// ActivationWechatQR handles POST /api/v1/activate/wechatQr.
func (c *LegacyController) ActivationWechatQR(r *ghttp.Request) {
	out, err := c.uidentitySvc.CreateActivationWechatState(r.Context(), uidentitysvc.ActivationWechatStateInput{
		ChallengeID: legacyChallengeID(r),
		Callback:    legacyStringParam(r, "callback", "cascallback", "casCallback"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyActivationWechatPayload(out))
}

// ActivationWechatScan handles GET /api/v1/activate/wechatScan.
func (c *LegacyController) ActivationWechatScan(r *ghttp.Request) {
	out, err := c.uidentitySvc.CompleteActivationWechat(r.Context(), uidentitysvc.ActivationWechatCallbackInput{
		State:    legacyStringParam(r, "state", "uuid"),
		UnionID:  legacyStringParam(r, "unionID", "unionId", "union_id"),
		Code:     legacyStringParam(r, "code"),
		Callback: legacyStringParam(r, "callback", "cascallback", "casCallback"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyActivationWechatPayload(out))
}

// ActivationState handles POST /api/v1/activate/state.
func (c *LegacyController) ActivationState(r *ghttp.Request) {
	out, err := c.uidentitySvc.ActivationState(r.Context(), legacyChallengeID(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{
		"uuid": out.ChallengeID, "challengeId": out.ChallengeID, "success": out.Success, "status": out.Status,
		"stage": out.Stage, "wechatStatus": out.WechatStatus, "wechat_status": out.WechatStatus,
		"redirectUrl": out.RedirectURL, "redirect_url": out.RedirectURL, "errorCode": out.ErrorCode, "error_code": out.ErrorCode,
		"message": out.Message,
	})
}

// UserGetByUnionID handles POST /api/v1/user/getByUnionID.
func (c *LegacyController) UserGetByUnionID(r *ghttp.Request) {
	out, err := c.uidentitySvc.LookupUnionID(r.Context(), legacyStringParam(r, "union_id", "unionID", "unionId"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"number": out.Number, "uuid": out.ChallengeID, "challengeId": out.ChallengeID, "call_back_url": out.CallbackURL, "callbackUrl": out.CallbackURL})
}

// UserBindUnionIDCallback handles GET /api/v1/user/bindUnionIDCallBack.
func (c *LegacyController) UserBindUnionIDCallback(r *ghttp.Request) {
	legacyOK(r, legacyRequestMap(r))
}

// UserBindUnionID handles POST /api/v1/user/bindUnionID.
func (c *LegacyController) UserBindUnionID(r *ghttp.Request) {
	out, err := c.uidentitySvc.BindUnionID(r.Context(), uidentitysvc.UnionIDBindInput{
		ChallengeID: legacyChallengeID(r),
		BindType:    legacyIntParam(r, "bind_type", "bindType"),
		Phone:       legacyStringParam(r, "phone"),
		Code:        legacyStringParam(r, "code"),
		Number:      legacyStringParam(r, "number"),
		Password:    legacyStringParam(r, "password"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"number": out.Number})
}

// UserChangePassword handles POST /api/v1/user/changePassword.
func (c *LegacyController) UserChangePassword(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	if err := c.uidentitySvc.ChangeRuntimePassword(r.Context(), number, legacyStringParam(r, "new_password", "newPassword", "password")); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// UserChangePhone handles POST /api/v1/user/changePhone.
func (c *LegacyController) UserChangePhone(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	if err := c.uidentitySvc.ChangeRuntimePhone(r.Context(), uidentitysvc.ChangePhoneInput{
		Number: number,
		Phone:  legacyStringParam(r, "phone"),
		Code:   legacyStringParam(r, "code"),
	}); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// UserChangeEmail handles POST /api/v1/user/changeEmail.
func (c *LegacyController) UserChangeEmail(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	if err := c.uidentitySvc.ChangeRuntimeEmail(r.Context(), number, legacyStringParam(r, "email")); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// UserChangeQQ handles POST /api/v1/user/changeQQ.
func (c *LegacyController) UserChangeQQ(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	if err := c.uidentitySvc.ChangeRuntimeQQ(r.Context(), number, legacyStringParam(r, "qq")); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// UserUnbindWechat handles POST /api/v1/user/unbindWechat.
func (c *LegacyController) UserUnbindWechat(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	if err := c.uidentitySvc.UnbindRuntimeWechat(r.Context(), number); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// UserInfo handles POST /api/v1/user/getUserInfo.
func (c *LegacyController) UserInfo(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	out, err := c.uidentitySvc.GetRuntimeUserInfo(r.Context(), number)
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyRuntimeAccountPayload(out))
}

// UserLoginLogs handles GET /api/v1/user/getUserCasLoginLog.
func (c *LegacyController) UserLoginLogs(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	pageIndex, pageSize := legacyPage(r)
	out, err := c.uidentitySvc.ListRuntimeUserLoginLogs(r.Context(), uidentitysvc.UserLogListInput{Number: number, PageNum: pageIndex, PageSize: pageSize})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyPageOK(r, out.List, out.Total, pageIndex, pageSize)
}

// UserApplications handles GET /api/v1/user/accountAppList.
func (c *LegacyController) UserApplications(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	pageIndex, pageSize := legacyPage(r)
	out, err := c.uidentitySvc.ListRuntimeApplications(r.Context(), uidentitysvc.UserApplicationListInput{Number: number, PageNum: pageIndex, PageSize: pageSize})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"count": out.Total, "pageIndex": pageIndex, "pageSize": pageSize, "list": legacyRuntimeApplicationPayloads(out.List)})
}

// UserAppRoles handles GET /api/v1/user/accountAppRole.
func (c *LegacyController) UserAppRoles(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	pageIndex, pageSize := legacyPage(r)
	out, err := c.uidentitySvc.ListRuntimeAppRoles(r.Context(), uidentitysvc.UserAppRoleListInput{Number: number, PageNum: pageIndex, PageSize: pageSize})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyPageOK(r, out.List, out.Total, pageIndex, pageSize)
}

// UserAppRoleCreate handles POST /api/v1/user/accountAppRole.
func (c *LegacyController) UserAppRoleCreate(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	id, err := c.uidentitySvc.CreateRuntimeAppRole(r.Context(), uidentitysvc.UserAppRoleCreateInput{
		Number:          number,
		EmpoweredNumber: legacyStringParam(r, "empoweredNumber", "empowered_number"),
		AppID:           legacyInt64Param(r, "appId", "app_id"),
		ExpireAt:        legacyOptionalInt64(r, "expireAt", "expire_at"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"id": id})
}

// UserAppRoleUpdate handles POST /api/v1/user/accountAppRoleUpdate.
func (c *LegacyController) UserAppRoleUpdate(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	if err := c.uidentitySvc.UpdateRuntimeAppRole(r.Context(), uidentitysvc.UserAppRoleUpdateInput{
		Number:   number,
		ID:       legacyInt64Param(r, "id"),
		ExpireAt: legacyOptionalInt64(r, "expireAt", "expire_at"),
	}); err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, nil)
}

// UserChangeWechatQR handles POST /api/v1/user/changeWechatQr.
func (c *LegacyController) UserChangeWechatQR(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	out, err := c.uidentitySvc.CreateRuntimeWechatRebindState(r.Context(), uidentitysvc.WechatRebindStateInput{
		Number:   number,
		Callback: legacyStringParam(r, "callback", "cascallback", "casCallback"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyWechatRebindPayload(out))
}

// UserChangeWechatState handles POST /api/v1/user/changeWechatState.
func (c *LegacyController) UserChangeWechatState(r *ghttp.Request) {
	number, err := c.legacyRuntimeNumber(r)
	if err != nil {
		legacyError(r, err)
		return
	}
	out, err := c.uidentitySvc.GetRuntimeWechatRebindState(r.Context(), uidentitysvc.WechatRebindStateLookupInput{
		Number: number,
		State:  legacyStringParam(r, "state", "uuid"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyWechatRebindPayload(out))
}

// UserWechatRebindCallback handles POST /api/v1/user/changeWechatCallBack.
func (c *LegacyController) UserWechatRebindCallback(r *ghttp.Request) {
	out, err := c.uidentitySvc.CompleteRuntimeWechatRebind(r.Context(), uidentitysvc.WechatRebindCallbackInput{
		State:    legacyStringParam(r, "state", "uuid"),
		UnionID:  legacyStringParam(r, "unionID", "unionId", "union_id"),
		Code:     legacyStringParam(r, "code"),
		Callback: legacyStringParam(r, "callback", "cascallback", "casCallback"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, legacyWechatRebindPayload(out))
}

// OAuthLogin handles old OAuth login page and form submission paths.
func (c *LegacyController) OAuthLogin(r *ghttp.Request) {
	c.OAuthAuthorize(r)
}

// OAuthAuthorize handles /api/v1/oauth/auth and /api/v1/oauth/authorize.
func (c *LegacyController) OAuthAuthorize(r *ghttp.Request) {
	out, err := c.uidentitySvc.IssueOAuthAuthorizationCode(r.Context(), uidentitysvc.OAuthAuthorizationCodeInput{
		ClientID:    legacyClientID(r),
		RedirectURI: legacyStringParam(r, "redirect_uri", "redirectUri"),
		Scope:       legacyStringParam(r, "scope"),
		State:       legacyStringParam(r, "state"),
		Number:      legacyStringParam(r, "number", "username"),
		Password:    legacyStringParam(r, "password"),
		TtlSeconds:  legacyInt64Param(r, "ttlSeconds", "ttl_seconds"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	payload := map[string]any{"code": out.Code, "redirectUrl": out.RedirectURL, "redirect_url": out.RedirectURL, "expiredAt": out.ExpiredAt, "expired_at": out.ExpiredAt, "state": out.State}
	if legacyBoolParam(r, "json", "api") || out.RedirectURL == "" {
		legacyOK(r, payload)
		return
	}
	r.Response.RedirectTo(out.RedirectURL)
	r.Exit()
}

// OAuthToken handles POST /api/v1/oauth/token with the raw OAuth token body.
func (c *LegacyController) OAuthToken(r *ghttp.Request) {
	out, err := c.uidentitySvc.ExchangeOAuthAuthorizationCode(r.Context(), uidentitysvc.OAuthTokenExchangeInput{
		GrantType:    legacyStringParam(r, "grant_type", "grantType"),
		ClientID:     legacyClientID(r),
		ClientSecret: legacyStringParam(r, "client_secret", "clientSecret", "secret"),
		Code:         legacyStringParam(r, "code"),
		RedirectURI:  legacyStringParam(r, "redirect_uri", "redirectUri"),
		TtlSeconds:   legacyInt64Param(r, "ttlSeconds", "ttl_seconds"),
	})
	if err != nil {
		r.Response.WriteJson(map[string]any{"error": "invalid_grant", "error_description": err.Error()})
		r.Exit()
		return
	}
	r.Response.WriteJson(map[string]any{
		"access_token": out.AccessToken, "accessToken": out.AccessToken,
		"refresh_token": out.RefreshToken, "refreshToken": out.RefreshToken,
		"token_type": out.TokenType, "tokenType": out.TokenType,
		"expires_in": out.ExpiresIn, "expiresIn": out.ExpiresIn,
		"expiredAt": out.ExpiredAt, "expired_at": out.ExpiredAt, "scope": out.Scope,
	})
	r.Exit()
}

// OAuthTest handles old /api/v1/oauth/test.
func (c *LegacyController) OAuthTest(r *ghttp.Request) {
	legacyOK(r, map[string]any{"status": "ok"})
}

// SmsSend handles POST /api/v1/sms/send.
func (c *LegacyController) SmsSend(r *ghttp.Request) {
	out, err := c.uidentitySvc.SendSMSCode(r.Context(), uidentitysvc.SMSSendInput{
		Type:  legacyStringParam(r, "type", "scene"),
		Phone: legacyStringParam(r, "phone"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, map[string]any{"id": out.ID})
}

// Upload handles POST /api/v1/public/uploadFile.
func (c *LegacyController) Upload(r *ghttp.Request) {
	out, err := c.uidentitySvc.UploadLegacyFiles(r.Context(), uidentitysvc.LegacyUploadInput{
		Type:        legacyStringParam(r, "type"),
		Source:      legacyStringParam(r, "source"),
		Base64File:  legacyStringParam(r, "file", "base64File", "base64_file"),
		UploadFiles: r.GetUploadFiles("file"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, out)
}

// Health handles old /api/v1/health.
func (c *LegacyController) Health(r *ghttp.Request) {
	r.Response.WriteStatus(http.StatusOK)
	r.Exit()
}

// LegacyCASConfig handles GET /api/v1/config/cas.
func (c *LegacyController) LegacyCASConfig(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyCASConfig(r.Context())
	legacyWriteServiceOutput(r, out, err)
}

// LegacyLDAPConfig handles GET /api/v1/config/ldap.
func (c *LegacyController) LegacyLDAPConfig(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyLDAPConfig(r.Context())
	legacyWriteServiceOutput(r, out, err)
}

// LegacyOAuthConfig handles GET /api/v1/config/oauth.
func (c *LegacyController) LegacyOAuthConfig(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyOAuthConfig(r.Context())
	legacyWriteServiceOutput(r, out, err)
}

// LegacyTokenConfig handles GET /api/v1/config/token.
func (c *LegacyController) LegacyTokenConfig(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyTokenConfig(r.Context())
	legacyWriteServiceOutput(r, out, err)
}

// Stats handles GET /api/v1/stat/get.
func (c *LegacyController) Stats(r *ghttp.Request) {
	out, err := c.uidentitySvc.Stats(r.Context())
	legacyWriteServiceOutput(r, out, err)
}

// ServerMonitor handles GET /api/v1/server-monitor.
func (c *LegacyController) ServerMonitor(r *ghttp.Request) {
	out, err := c.uidentitySvc.ServerMonitor(r.Context())
	legacyWriteServiceOutput(r, out, err)
}

// LogSnapshot handles GET /api/v1/log/watch.
func (c *LegacyController) LogSnapshot(r *ghttp.Request) {
	out, err := c.uidentitySvc.LogSnapshot(r.Context(), uidentitysvc.LegacyLogSnapshotInput{
		Date:  legacyStringParam(r, "date"),
		Lines: legacyIntParam(r, "lines", "limit"),
	})
	legacyWriteServiceOutput(r, out, err)
}

// SysJobList handles old GET /api/v1/sysjob using plugin managed-job snapshots.
func (c *LegacyController) SysJobList(r *ghttp.Request) {
	pageIndex, pageSize := legacyPage(r)
	records := legacySysJobSnapshots()
	records = legacyFilterSysJobs(r, records)
	total := len(records)
	records = legacyPaginateSysJobs(records, pageIndex, pageSize)
	legacyPageOK(r, records, total, pageIndex, pageSize)
}

// SysJobGet handles old GET /api/v1/sysjob/{id}.
func (c *LegacyController) SysJobGet(r *ghttp.Request) {
	id := legacyRouterID(r)
	if id <= 0 {
		id = legacyInt64Param(r, "jobId", "id")
	}
	for _, record := range legacySysJobSnapshots() {
		if gconv.Int64(record["jobId"]) == id {
			legacyOK(r, record)
			return
		}
	}
	legacyError(r, bizerr.NewCode(uidentitysvc.CodeResourceNotFound))
}

// SysJobExternalAction reports old sysjob mutation routes as host-owned gf jobs.
func (c *LegacyController) SysJobExternalAction(actionType string) ghttp.HandlerFunc {
	return c.ExternalAction(actionType)
}

// ExternalAction handles compatibility stubs for old external execution routes.
func (c *LegacyController) ExternalAction(actionType string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		out, err := c.uidentitySvc.RunExternalAction(r.Context(), uidentitysvc.LegacyExternalActionInput{
			Type:   actionType,
			Target: legacyStringParam(r, "target", "name", "jobName", "id"),
		})
		legacyWriteServiceOutput(r, out, err)
	}
}

func (c *LegacyController) legacyRuntimeNumber(r *ghttp.Request) (string, error) {
	number := legacyStringParam(r, "number", "username", "workCode", "work_code")
	if number != "" {
		return number, nil
	}
	for _, header := range []string{"X-Uidentity-Number", "X-Account-Number", "X-User-Number"} {
		if value := strings.TrimSpace(r.GetHeader(header)); value != "" {
			return value, nil
		}
	}
	for _, key := range []string{"number", "userNumber", "runtimeNumber"} {
		if value := strings.TrimSpace(r.GetCtxVar(key).String()); value != "" {
			return value, nil
		}
	}
	accessToken := legacyAccessToken(r)
	if accessToken != "" {
		out, err := c.uidentitySvc.GetUserInfoByRuntimeToken(r.Context(), accessToken)
		if err != nil {
			return "", err
		}
		if out != nil && out.User != nil && out.User.Number != "" {
			return out.User.Number, nil
		}
	}
	return "", bizerr.NewCode(uidentitysvc.CodeInvalidCredentials)
}

func legacyWriteServiceOutput(r *ghttp.Request, out any, err error) {
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOK(r, out)
}

func legacyOK(r *ghttp.Request, data any) {
	r.Response.WriteJson(map[string]any{
		"requestId": legacyRequestID(r),
		"code":      legacyStatusOK,
		"msg":       "操作成功",
		"data":      data,
	})
	r.Exit()
}

func legacyPageOK(r *ghttp.Request, list any, total int, pageIndex int, pageSize int) {
	legacyOK(r, map[string]any{
		"count":     total,
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
		"list":      list,
	})
}

func legacyError(r *ghttp.Request, err error) {
	msg := "操作失败"
	if err != nil {
		msg = err.Error()
	}
	r.Response.WriteStatus(http.StatusOK)
	r.Response.WriteJson(map[string]any{
		"requestId": legacyRequestID(r),
		"code":      legacyStatusError,
		"msg":       msg,
		"status":    "error",
	})
	r.Exit()
}

func legacyRequestID(r *ghttp.Request) string {
	for _, key := range []string{"requestId", "RequestId", "traceID", "traceId"} {
		if value := strings.TrimSpace(r.GetCtxVar(key).String()); value != "" {
			return value
		}
	}
	return ""
}

func legacyRequestMap(r *ghttp.Request) map[string]any {
	data := r.GetRequestMap()
	if data == nil {
		return map[string]any{}
	}
	if id := legacyRouterID(r); id > 0 {
		data["id"] = id
	}
	return data
}

func legacyStringParam(r *ghttp.Request, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(r.GetRequest(name).String()); value != "" {
			return value
		}
		if value := strings.TrimSpace(r.GetRouter(name).String()); value != "" {
			return value
		}
	}
	return ""
}

func legacyIntParam(r *ghttp.Request, names ...string) int {
	return gconv.Int(legacyStringParam(r, names...))
}

func legacyInt64Param(r *ghttp.Request, names ...string) int64 {
	return gconv.Int64(legacyStringParam(r, names...))
}

func legacyOptionalInt(r *ghttp.Request, names ...string) *int {
	for _, name := range names {
		value := strings.TrimSpace(r.GetRequest(name).String())
		if value == "" {
			value = strings.TrimSpace(r.GetRouter(name).String())
		}
		if value != "" {
			parsed := gconv.Int(value)
			return &parsed
		}
	}
	return nil
}

func legacyOptionalInt64(r *ghttp.Request, names ...string) *int64 {
	for _, name := range names {
		value := strings.TrimSpace(r.GetRequest(name).String())
		if value == "" {
			value = strings.TrimSpace(r.GetRouter(name).String())
		}
		if value != "" {
			parsed := gconv.Int64(value)
			return &parsed
		}
	}
	return nil
}

func legacyBoolParam(r *ghttp.Request, names ...string) bool {
	for _, name := range names {
		if value := strings.TrimSpace(r.GetRequest(name).String()); value != "" {
			return gconv.Bool(value)
		}
	}
	return false
}

func legacyStringListParam(r *ghttp.Request, names ...string) []string {
	for _, name := range names {
		value := r.GetRequest(name)
		if value.IsEmpty() {
			continue
		}
		list := stringsFromAny(value.Val())
		if len(list) > 0 {
			return list
		}
	}
	return nil
}

func legacyInt64ListParam(r *ghttp.Request, names ...string) []int64 {
	values := legacyStringListParam(r, names...)
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if parsed := gconv.Int64(value); parsed > 0 {
			result = append(result, parsed)
		}
	}
	return result
}

func legacyDeleteIDs(r *ghttp.Request) string {
	if routerID := legacyRouterID(r); routerID > 0 {
		return gconv.String(routerID)
	}
	for _, name := range []string{"ids", "id"} {
		values := legacyStringListParam(r, name)
		if len(values) > 0 {
			return strings.Join(values, ",")
		}
	}
	return ""
}

func legacyRouterID(r *ghttp.Request) int64 {
	return gconv.Int64(r.GetRouter("id").String())
}

func legacyClientID(r *ghttp.Request) string {
	return legacyStringParam(r, "appid", "appId", "app_id", "client_id", "clientId")
}

func legacyChallengeID(r *ghttp.Request) string {
	return legacyStringParam(r, "uuid", "challengeId", "challenge_id", "state")
}

func legacyAccessToken(r *ghttp.Request) string {
	token := legacyStringParam(r, "accessToken", "access_token", "token")
	if token != "" {
		return token
	}
	auth := strings.TrimSpace(r.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return auth
}

func legacyKeyword(r *ghttp.Request) string {
	for _, name := range []string{"keyword", "search", "number", "name", "phone"} {
		if value := legacyStringParam(r, name); value != "" {
			return value
		}
	}
	return ""
}

func legacyPage(r *ghttp.Request) (int, int) {
	pageIndex := legacyIntParam(r, "pageIndex", "page_index", "pageNum", "page", "current")
	pageSize := legacyIntParam(r, "pageSize", "page_size", "limit")
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return pageIndex, pageSize
}

func legacyAccountImportInput(r *ghttp.Request) uidentitysvc.AccountImportInput {
	return uidentitysvc.AccountImportInput{
		Filepath: legacyStringParam(r, "filepath", "filePath", "path"),
		Limit:    legacyIntParam(r, "limit"),
	}
}

func legacySetTGTCookie(r *ghttp.Request, out *uidentitysvc.RuntimeLoginOutput) {
	if out == nil || strings.TrimSpace(out.TGT) == "" {
		return
	}
	r.Cookie.SetCookie(legacyTGTCookieName, out.TGT, "", "/", legacyTGTCookieMaxAge)
}

func legacyRuntimeLoginPayload(out *uidentitysvc.RuntimeLoginOutput) map[string]any {
	if out == nil {
		return map[string]any{}
	}
	return map[string]any{
		"callbackUrl":  out.CallbackURL,
		"callback_url": out.CallbackURL,
		"tgt":          out.TGT,
		"st":           out.ST,
		"user":         legacyRuntimeAccountPayload(out.User),
		"users":        legacyRuntimeAccountPayloads(out.Users),
		"app":          legacyRuntimeApplicationPayload(out.App),
	}
}

func legacyWechatLoginResultPayload(out *uidentitysvc.WechatLoginQRResultOutput) map[string]any {
	if out == nil {
		return map[string]any{}
	}
	payload := map[string]any{
		"uuid": out.State, "state": out.State, "status": out.Status, "redirectUrl": out.RedirectURL, "redirect_url": out.RedirectURL,
		"challengeId": out.ChallengeID, "challenge_id": out.ChallengeID, "callbackUrl": out.CallbackURL, "callback_url": out.CallbackURL,
		"errorCode": out.ErrorCode, "error_code": out.ErrorCode, "message": out.Message,
	}
	for key, value := range legacyRuntimeLoginPayload(out.Login) {
		payload[key] = value
	}
	return payload
}

func legacyActivationStepPayload(out *uidentitysvc.ActivationStepOutput) map[string]any {
	if out == nil {
		return map[string]any{}
	}
	return map[string]any{"uuid": out.ChallengeID, "challengeId": out.ChallengeID, "success": out.Success}
}

func legacyActivationWechatPayload(out *uidentitysvc.ActivationWechatStateOutput) map[string]any {
	if out == nil {
		return map[string]any{}
	}
	return map[string]any{
		"uuid": out.State, "state": out.State, "status": out.Status, "success": out.Success, "qrcode": out.URL, "url": out.URL,
		"redirectUrl": out.RedirectURL, "redirect_url": out.RedirectURL, "errorCode": out.ErrorCode, "error_code": out.ErrorCode,
		"message": out.Message,
	}
}

func legacyWechatRebindPayload(out *uidentitysvc.WechatRebindStateOutput) map[string]any {
	if out == nil {
		return map[string]any{}
	}
	return map[string]any{
		"uuid": out.State, "state": out.State, "status": out.Status, "success": out.Success, "qrcode": out.URL, "url": out.URL,
		"redirectUrl": out.RedirectURL, "redirect_url": out.RedirectURL, "expiredAt": out.ExpiredAt, "expired_at": out.ExpiredAt,
		"errorCode": out.ErrorCode, "error_code": out.ErrorCode, "message": out.Message,
	}
}

func legacyRuntimeAccountPayload(account *uidentitysvc.RuntimeAccount) map[string]any {
	if account == nil {
		return nil
	}
	payload := map[string]any{
		"id": account.ID, "number": account.Number, "name": account.Name, "phone": account.Phone, "status": account.Status,
		"passLevel": account.PassLevel, "pass_level": account.PassLevel, "containerId": account.ContainerID, "container_id": account.ContainerID,
		"containerName": account.ContainerName, "container_name": account.ContainerName, "unitId": account.UnitID, "unit_id": account.UnitID,
		"unitName": account.UnitName, "unit_name": account.UnitName, "unit": account.UnitName, "expireAt": account.ExpireAt,
		"expire_at": account.ExpireAt, "groups": account.Groups,
	}
	if account.Detail != nil {
		payload["detail"] = account.Detail
		payload["birthday"] = account.Detail.Birthday
		payload["email"] = account.Detail.Email
		payload["gender"] = account.Detail.Gender
		payload["qq"] = account.Detail.QQ
		payload["wechat"] = account.Detail.Wechat
		payload["idcard"] = account.Detail.Idcard
		payload["idCard"] = account.Detail.Idcard
		payload["avatar"] = account.Detail.Avatar
		payload["face"] = account.Detail.Face
	}
	return payload
}

func legacyRuntimeAccountPayloads(accounts []*uidentitysvc.RuntimeAccount) []map[string]any {
	result := make([]map[string]any, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, legacyRuntimeAccountPayload(account))
	}
	return result
}

func legacyRuntimeApplicationPayload(app *uidentitysvc.RuntimeApplication) map[string]any {
	if app == nil {
		return nil
	}
	return map[string]any{
		"id": app.ID, "name": app.Name, "alias": app.Alias, "clientId": app.ClientID, "client_id": app.ClientID,
		"accessModel": app.AccessModel, "access_model": app.AccessModel, "callbackUrl": app.CallbackURL, "callback_url": app.CallbackURL,
	}
}

func legacyRuntimeApplicationPayloads(apps []*uidentitysvc.RuntimeApplication) []map[string]any {
	result := make([]map[string]any, 0, len(apps))
	for _, app := range apps {
		result = append(result, legacyRuntimeApplicationPayload(app))
	}
	return result
}

func stringsFromAny(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case []string:
		return cleanupStrings(typed)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, splitCSV(gconv.String(item))...)
		}
		return cleanupStrings(result)
	default:
		return cleanupStrings(splitCSV(gconv.String(typed)))
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(value), ",")
}

func cleanupStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func legacyRouteKey(method string, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func legacySysJobSnapshots() []map[string]any {
	return []map[string]any{
		legacySysJobSnapshot(1, "SyncDept", "0 1 * * *"),
		legacySysJobSnapshot(2, "SyncJzg", "15 1 * * *"),
		legacySysJobSnapshot(3, "SyncStudent", "30 1 * * *"),
		legacySysJobSnapshot(4, "SyncStudentYJS", "45 1 * * *"),
		legacySysJobSnapshot(5, "SyncStudentWJ", "0 2 * * *"),
		legacySysJobSnapshot(6, "ChangeContainer", "15 2 * * *"),
		legacySysJobSnapshot(7, "NewContainerAccount", "30 2 * * *"),
		legacySysJobSnapshot(8, "SyncMysql2Ldap", "45 2 * * *"),
		legacySysJobSnapshot(9, "WannaT", "0 * * * *"),
	}
}

func legacySysJobSnapshot(id int, name string, cron string) map[string]any {
	return map[string]any{
		"jobId":          id,
		"jobName":        name,
		"jobGroup":       legacySysJobGroup,
		"jobType":        legacySysJobType,
		"cronExpression": cron,
		"invokeTarget":   name,
		"args":           "",
		"misfirePolicy":  1,
		"concurrent":     0,
		"status":         legacySysJobStatusOn,
		"entryId":        0,
		"dataScope":      "",
	}
}

func legacyFilterSysJobs(r *ghttp.Request, records []map[string]any) []map[string]any {
	jobID := legacyIntParam(r, "jobId")
	jobName := strings.ToLower(legacyStringParam(r, "jobName"))
	jobGroup := legacyStringParam(r, "jobGroup")
	status := legacyIntParam(r, "status")
	result := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if jobID > 0 && gconv.Int(record["jobId"]) != jobID {
			continue
		}
		if jobName != "" && !strings.Contains(strings.ToLower(gconv.String(record["jobName"])), jobName) {
			continue
		}
		if jobGroup != "" && gconv.String(record["jobGroup"]) != jobGroup {
			continue
		}
		if status > 0 && gconv.Int(record["status"]) != status {
			continue
		}
		result = append(result, record)
	}
	return result
}

func legacyPaginateSysJobs(records []map[string]any, pageIndex int, pageSize int) []map[string]any {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	start := (pageIndex - 1) * pageSize
	if start >= len(records) {
		return []map[string]any{}
	}
	end := start + pageSize
	if end > len(records) {
		end = len(records)
	}
	return records[start:end]
}
