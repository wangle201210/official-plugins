// This file records old uidentity/admin management resource API contracts that
// are served by the plugin legacy HTTP layer.

package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountDetailsListReq declares GET /api/v1/account-details.
type AccountDetailsListReq struct {
	g.Meta `path:"/api/v1/account-details" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List account details" dc:"Match the old uidentity/admin account-details list contract." permission:"uidentity:cas:read"`
	LegacyAccountDetailsList
}

// AccountDetailsGetReq declares GET /api/v1/account-details/{id}.
type AccountDetailsGetReq struct {
	g.Meta    `path:"/api/v1/account-details/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get account detail" dc:"Match the old uidentity/admin account-details detail contract." permission:"uidentity:cas:read"`
	AccountId int64 `json:"accountId" dc:"Legacy accountId path binding" eg:"1"`
}

// AccountDetailsCreateReq declares POST /api/v1/account-details.
type AccountDetailsCreateReq struct {
	g.Meta `path:"/api/v1/account-details" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create account detail" dc:"Match the old uidentity/admin account-details create contract." permission:"uidentity:cas:write"`
	LegacyAccountDetailsCreateBody
}

// AccountDetailsUpdateReq declares PUT /api/v1/account-details/{id}.
type AccountDetailsUpdateReq struct {
	g.Meta    `path:"/api/v1/account-details/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update account detail" dc:"Match the old uidentity/admin account-details update contract." permission:"uidentity:cas:write"`
	AccountId int64 `json:"accountId" dc:"Legacy accountId path binding" eg:"1"`
	LegacyAccountDetailsUpdateBody
}

// AccountDetailsDeleteReq declares DELETE /api/v1/account-details.
type AccountDetailsDeleteReq struct {
	g.Meta     `path:"/api/v1/account-details" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete account details" dc:"Match the old uidentity/admin account-details delete contract." permission:"uidentity:cas:delete"`
	AccountIds []int64 `json:"account_ids" dc:"Legacy account_ids delete body" eg:"[1,2]"`
}

// AccountUnitListReq declares GET /api/v1/account-unit.
type AccountUnitListReq struct {
	g.Meta `path:"/api/v1/account-unit" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List account units" dc:"Match the old uidentity/admin account-unit list contract." permission:"uidentity:cas:read"`
	LegacyAccountUnitList
}

// AccountUnitGetReq declares GET /api/v1/account-unit/{id}.
type AccountUnitGetReq struct {
	g.Meta `path:"/api/v1/account-unit/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get account unit" dc:"Match the old uidentity/admin account-unit detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// AccountUnitCreateReq declares POST /api/v1/account-unit.
type AccountUnitCreateReq struct {
	g.Meta `path:"/api/v1/account-unit" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create account unit" dc:"Match the old uidentity/admin account-unit create contract." permission:"uidentity:cas:write"`
	LegacyAccountUnitBody
}

// AccountUnitUpdateReq declares PUT /api/v1/account-unit/{id}.
type AccountUnitUpdateReq struct {
	g.Meta `path:"/api/v1/account-unit/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update account unit" dc:"Match the old uidentity/admin account-unit update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyAccountUnitBody
}

// AccountUnitDeleteReq declares DELETE /api/v1/account-unit.
type AccountUnitDeleteReq struct {
	g.Meta `path:"/api/v1/account-unit" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete account units" dc:"Match the old uidentity/admin account-unit delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// AccountAppRoleListReq declares GET /api/v1/account-app-role.
type AccountAppRoleListReq struct {
	g.Meta `path:"/api/v1/account-app-role" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List account app roles" dc:"Match the old uidentity/admin account-app-role list contract." permission:"uidentity:cas:read"`
	LegacyAccountAppRoleList
}

// AccountAppRoleGetReq declares GET /api/v1/account-app-role/{id}.
type AccountAppRoleGetReq struct {
	g.Meta `path:"/api/v1/account-app-role/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get account app role" dc:"Match the old uidentity/admin account-app-role detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// AccountAppRoleCreateReq declares POST /api/v1/account-app-role.
type AccountAppRoleCreateReq struct {
	g.Meta `path:"/api/v1/account-app-role" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create account app role" dc:"Match the old uidentity/admin account-app-role create contract." permission:"uidentity:cas:write"`
	LegacyAccountAppRoleBody
}

// AccountAppRoleUpdateReq declares PUT /api/v1/account-app-role/{id}.
type AccountAppRoleUpdateReq struct {
	g.Meta `path:"/api/v1/account-app-role/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update account app role" dc:"Match the old uidentity/admin account-app-role update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyAccountAppRoleBody
}

// AccountAppRoleDeleteReq declares DELETE /api/v1/account-app-role.
type AccountAppRoleDeleteReq struct {
	g.Meta `path:"/api/v1/account-app-role" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete account app roles" dc:"Match the old uidentity/admin account-app-role delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// AccountAppBlacklistListReq declares GET /api/v1/account-app-blacklist.
type AccountAppBlacklistListReq struct {
	g.Meta `path:"/api/v1/account-app-blacklist" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List account app blacklist" dc:"Match the old uidentity/admin account-app-blacklist list contract." permission:"uidentity:cas:read"`
	LegacyAccountAppBlacklistList
}

// AccountAppBlacklistGetReq declares GET /api/v1/account-app-blacklist/{id}.
type AccountAppBlacklistGetReq struct {
	g.Meta `path:"/api/v1/account-app-blacklist/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get account app blacklist" dc:"Match the old uidentity/admin account-app-blacklist detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// AccountAppBlacklistCreateReq declares POST /api/v1/account-app-blacklist.
type AccountAppBlacklistCreateReq struct {
	g.Meta `path:"/api/v1/account-app-blacklist" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create account app blacklist" dc:"Match the old uidentity/admin account-app-blacklist create contract." permission:"uidentity:cas:write"`
	LegacyAccountAppBlacklistCreateBody
}

// AccountAppBlacklistUpdateReq declares PUT /api/v1/account-app-blacklist/{id}.
type AccountAppBlacklistUpdateReq struct {
	g.Meta `path:"/api/v1/account-app-blacklist/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update account app blacklist" dc:"Match the old uidentity/admin account-app-blacklist update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyAccountAppBlacklistUpdateBody
}

// AccountAppBlacklistDeleteReq declares DELETE /api/v1/account-app-blacklist.
type AccountAppBlacklistDeleteReq struct {
	g.Meta `path:"/api/v1/account-app-blacklist" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete account app blacklist" dc:"Match the old uidentity/admin account-app-blacklist delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// AccountChangeLogListReq declares GET /api/v1/account-change-log.
type AccountChangeLogListReq struct {
	g.Meta `path:"/api/v1/account-change-log" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List account change logs" dc:"Match the old uidentity/admin account-change-log list contract." permission:"uidentity:cas:read"`
	LegacyAccountChangeLogList
}

// AccountChangeLogGetReq declares GET /api/v1/account-change-log/{id}.
type AccountChangeLogGetReq struct {
	g.Meta `path:"/api/v1/account-change-log/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get account change log" dc:"Match the old uidentity/admin account-change-log detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// AccountChangeLogCreateReq declares POST /api/v1/account-change-log.
type AccountChangeLogCreateReq struct {
	g.Meta `path:"/api/v1/account-change-log" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create account change log" dc:"Match the old uidentity/admin account-change-log create contract." permission:"uidentity:cas:write"`
	LegacyAccountChangeLogBody
}

// AccountChangeLogUpdateReq declares PUT /api/v1/account-change-log/{id}.
type AccountChangeLogUpdateReq struct {
	g.Meta `path:"/api/v1/account-change-log/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update account change log" dc:"Match the old uidentity/admin account-change-log update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyAccountChangeLogBody
}

// AccountChangeLogDeleteReq declares DELETE /api/v1/account-change-log.
type AccountChangeLogDeleteReq struct {
	g.Meta `path:"/api/v1/account-change-log" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete account change logs" dc:"Match the old uidentity/admin account-change-log delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// UnitsListReq declares GET /api/v1/units.
type UnitsListReq struct {
	g.Meta `path:"/api/v1/units" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List units" dc:"Match the old uidentity/admin units list contract." permission:"uidentity:cas:read"`
	LegacyUnitsList
}

// UnitsGetReq declares GET /api/v1/units/{id}.
type UnitsGetReq struct {
	g.Meta `path:"/api/v1/units/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get unit" dc:"Match the old uidentity/admin units detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// UnitsCreateReq declares POST /api/v1/units.
type UnitsCreateReq struct {
	g.Meta `path:"/api/v1/units" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create unit" dc:"Match the old uidentity/admin units create contract." permission:"uidentity:cas:write"`
	LegacyUnitsBody
}

// UnitsUpdateReq declares PUT /api/v1/units/{id}.
type UnitsUpdateReq struct {
	g.Meta `path:"/api/v1/units/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update unit" dc:"Match the old uidentity/admin units update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyUnitsBody
}

// UnitsDeleteReq declares DELETE /api/v1/units.
type UnitsDeleteReq struct {
	g.Meta `path:"/api/v1/units" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete units" dc:"Match the old uidentity/admin units delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// GroupsListReq declares GET /api/v1/groups.
type GroupsListReq struct {
	g.Meta `path:"/api/v1/groups" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List groups" dc:"Match the old uidentity/admin groups list contract." permission:"uidentity:cas:read"`
	LegacyGroupsList
}

// GroupsGetReq declares GET /api/v1/groups/{id}.
type GroupsGetReq struct {
	g.Meta `path:"/api/v1/groups/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get group" dc:"Match the old uidentity/admin groups detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// GroupsCreateReq declares POST /api/v1/groups.
type GroupsCreateReq struct {
	g.Meta `path:"/api/v1/groups" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create group" dc:"Match the old uidentity/admin groups create contract." permission:"uidentity:cas:write"`
	LegacyGroupsBody
}

// GroupsUpdateReq declares PUT /api/v1/groups/{id}.
type GroupsUpdateReq struct {
	g.Meta `path:"/api/v1/groups/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update group" dc:"Match the old uidentity/admin groups update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyGroupsBody
}

// GroupsDeleteReq declares DELETE /api/v1/groups.
type GroupsDeleteReq struct {
	g.Meta `path:"/api/v1/groups" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete groups" dc:"Match the old uidentity/admin groups delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// ContainersListReq declares GET /api/v1/containers.
type ContainersListReq struct {
	g.Meta `path:"/api/v1/containers" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List containers" dc:"Match the old uidentity/admin containers list contract." permission:"uidentity:cas:read"`
	LegacyContainersList
}

// ContainersGetReq declares GET /api/v1/containers/{id}.
type ContainersGetReq struct {
	g.Meta `path:"/api/v1/containers/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get container" dc:"Match the old uidentity/admin containers detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// ContainersCreateReq declares POST /api/v1/containers.
type ContainersCreateReq struct {
	g.Meta `path:"/api/v1/containers" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create container" dc:"Match the old uidentity/admin containers create contract." permission:"uidentity:cas:write"`
	LegacyContainersBody
}

// ContainersUpdateReq declares PUT /api/v1/containers/{id}.
type ContainersUpdateReq struct {
	g.Meta `path:"/api/v1/containers/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update container" dc:"Match the old uidentity/admin containers update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyContainersBody
}

// ContainersDeleteReq declares DELETE /api/v1/containers.
type ContainersDeleteReq struct {
	g.Meta `path:"/api/v1/containers" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete containers" dc:"Match the old uidentity/admin containers delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// ApplicationsListReq declares GET /api/v1/applications.
type ApplicationsListReq struct {
	g.Meta `path:"/api/v1/applications" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List applications" dc:"Match the old uidentity/admin applications list contract." permission:"uidentity:cas:read"`
	LegacyApplicationsList
}

// ApplicationsGetReq declares GET /api/v1/applications/{id}.
type ApplicationsGetReq struct {
	g.Meta `path:"/api/v1/applications/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get application" dc:"Match the old uidentity/admin applications detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// ApplicationsCreateReq declares POST /api/v1/applications.
type ApplicationsCreateReq struct {
	g.Meta `path:"/api/v1/applications" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create application" dc:"Match the old uidentity/admin applications create contract." permission:"uidentity:cas:write"`
	LegacyApplicationsBody
}

// ApplicationsUpdateReq declares PUT /api/v1/applications/{id}.
type ApplicationsUpdateReq struct {
	g.Meta `path:"/api/v1/applications/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update application" dc:"Match the old uidentity/admin applications update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyApplicationsBody
}

// ApplicationsDeleteReq declares DELETE /api/v1/applications.
type ApplicationsDeleteReq struct {
	g.Meta `path:"/api/v1/applications" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete applications" dc:"Match the old uidentity/admin applications delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// GroupAppBlacklistListReq declares GET /api/v1/group-app-blacklist.
type GroupAppBlacklistListReq struct {
	g.Meta `path:"/api/v1/group-app-blacklist" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List group app blacklist" dc:"Match the old uidentity/admin group-app-blacklist list contract." permission:"uidentity:cas:read"`
	LegacyGroupAppBlacklistList
}

// GroupAppBlacklistGetReq declares GET /api/v1/group-app-blacklist/{id}.
type GroupAppBlacklistGetReq struct {
	g.Meta `path:"/api/v1/group-app-blacklist/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get group app blacklist" dc:"Match the old uidentity/admin group-app-blacklist detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// GroupAppBlacklistCreateReq declares POST /api/v1/group-app-blacklist.
type GroupAppBlacklistCreateReq struct {
	g.Meta `path:"/api/v1/group-app-blacklist" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create group app blacklist" dc:"Match the old uidentity/admin group-app-blacklist create contract." permission:"uidentity:cas:write"`
	LegacyGroupAppBlacklistBody
}

// GroupAppBlacklistUpdateReq declares PUT /api/v1/group-app-blacklist/{id}.
type GroupAppBlacklistUpdateReq struct {
	g.Meta `path:"/api/v1/group-app-blacklist/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update group app blacklist" dc:"Match the old uidentity/admin group-app-blacklist update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyGroupAppBlacklistBody
}

// GroupAppBlacklistDeleteReq declares DELETE /api/v1/group-app-blacklist.
type GroupAppBlacklistDeleteReq struct {
	g.Meta `path:"/api/v1/group-app-blacklist" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete group app blacklist" dc:"Match the old uidentity/admin group-app-blacklist delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// PassRulerListReq declares GET /api/v1/pass-ruler.
type PassRulerListReq struct {
	g.Meta `path:"/api/v1/pass-ruler" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List password rules" dc:"Match the old uidentity/admin pass-ruler list contract." permission:"uidentity:cas:read"`
	LegacyPassRulerList
}

// PassRulerGetReq declares GET /api/v1/pass-ruler/{id}.
type PassRulerGetReq struct {
	g.Meta `path:"/api/v1/pass-ruler/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get password rule" dc:"Match the old uidentity/admin pass-ruler detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// PassRulerCreateReq declares POST /api/v1/pass-ruler.
type PassRulerCreateReq struct {
	g.Meta `path:"/api/v1/pass-ruler" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create password rule" dc:"Match the old uidentity/admin pass-ruler create contract." permission:"uidentity:cas:write"`
	LegacyPassRulerBody
}

// PassRulerUpdateReq declares PUT /api/v1/pass-ruler/{id}.
type PassRulerUpdateReq struct {
	g.Meta `path:"/api/v1/pass-ruler/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update password rule" dc:"Match the old uidentity/admin pass-ruler update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyPassRulerBody
}

// PassRulerDeleteReq declares DELETE /api/v1/pass-ruler.
type PassRulerDeleteReq struct {
	g.Meta `path:"/api/v1/pass-ruler" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete password rules" dc:"Match the old uidentity/admin pass-ruler delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// SmsListReq declares GET /api/v1/sms.
type SmsListReq struct {
	g.Meta `path:"/api/v1/sms" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List SMS records" dc:"Match the old uidentity/admin generated sms list contract." permission:"uidentity:cas:read"`
	LegacySmsList
}

// SmsGetReq declares GET /api/v1/sms/{id}.
type SmsGetReq struct {
	g.Meta `path:"/api/v1/sms/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get SMS record" dc:"Match the old uidentity/admin generated sms detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// SmsCreateReq declares POST /api/v1/sms.
type SmsCreateReq struct {
	g.Meta `path:"/api/v1/sms" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create SMS record" dc:"Match the old uidentity/admin generated sms create contract." permission:"uidentity:cas:write"`
	LegacySmsBody
}

// SmsUpdateReq declares PUT /api/v1/sms/{id}.
type SmsUpdateReq struct {
	g.Meta `path:"/api/v1/sms/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update SMS record" dc:"Match the old uidentity/admin generated sms update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacySmsBody
}

// SmsDeleteReq declares DELETE /api/v1/sms.
type SmsDeleteReq struct {
	g.Meta `path:"/api/v1/sms" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete SMS records" dc:"Match the old uidentity/admin generated sms delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// CasLoginLogListReq declares GET /api/v1/cas-login-log.
type CasLoginLogListReq struct {
	g.Meta `path:"/api/v1/cas-login-log" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List CAS login logs" dc:"Match the old uidentity/admin cas-login-log list contract." permission:"uidentity:cas:read"`
	LegacyCasLoginLogList
}

// CasLoginLogGetReq declares GET /api/v1/cas-login-log/{id}.
type CasLoginLogGetReq struct {
	g.Meta `path:"/api/v1/cas-login-log/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get CAS login log" dc:"Match the old uidentity/admin cas-login-log detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// CasLoginLogCreateReq declares POST /api/v1/cas-login-log.
type CasLoginLogCreateReq struct {
	g.Meta `path:"/api/v1/cas-login-log" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create CAS login log" dc:"Match the old uidentity/admin cas-login-log create contract." permission:"uidentity:cas:write"`
	LegacyCasLoginLogBody
}

// CasLoginLogUpdateReq declares PUT /api/v1/cas-login-log/{id}.
type CasLoginLogUpdateReq struct {
	g.Meta `path:"/api/v1/cas-login-log/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update CAS login log" dc:"Match the old uidentity/admin cas-login-log update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyCasLoginLogBody
}

// CasLoginLogDeleteReq declares DELETE /api/v1/cas-login-log.
type CasLoginLogDeleteReq struct {
	g.Meta `path:"/api/v1/cas-login-log" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete CAS login logs" dc:"Match the old uidentity/admin cas-login-log delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// OAuthLogListReq declares GET /api/v1/oauth-log.
type OAuthLogListReq struct {
	g.Meta `path:"/api/v1/oauth-log" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List OAuth logs" dc:"Match the old uidentity/admin oauth-log list contract." permission:"uidentity:cas:read"`
	LegacyOAuthLogList
}

// OAuthLogGetReq declares GET /api/v1/oauth-log/{id}.
type OAuthLogGetReq struct {
	g.Meta `path:"/api/v1/oauth-log/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get OAuth log" dc:"Match the old uidentity/admin oauth-log detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// OAuthLogCreateReq declares POST /api/v1/oauth-log.
type OAuthLogCreateReq struct {
	g.Meta `path:"/api/v1/oauth-log" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create OAuth log" dc:"Match the old uidentity/admin oauth-log create contract." permission:"uidentity:cas:write"`
	LegacyOAuthLogBody
}

// OAuthLogUpdateReq declares PUT /api/v1/oauth-log/{id}.
type OAuthLogUpdateReq struct {
	g.Meta `path:"/api/v1/oauth-log/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update OAuth log" dc:"Match the old uidentity/admin oauth-log update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyOAuthLogBody
}

// OAuthLogDeleteReq declares DELETE /api/v1/oauth-log.
type OAuthLogDeleteReq struct {
	g.Meta `path:"/api/v1/oauth-log" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete OAuth logs" dc:"Match the old uidentity/admin oauth-log delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// OAuthTokenListReq declares GET /api/v1/oauth-token.
type OAuthTokenListReq struct {
	g.Meta `path:"/api/v1/oauth-token" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"List OAuth tokens" dc:"Match the old uidentity/admin oauth-token list contract." permission:"uidentity:cas:read"`
	LegacyOAuthTokenList
}

// OAuthTokenGetReq declares GET /api/v1/oauth-token/{id}.
type OAuthTokenGetReq struct {
	g.Meta `path:"/api/v1/oauth-token/{id}" method:"get" tags:"UIdentity CAS Legacy Resources" summary:"Get OAuth token" dc:"Match the old uidentity/admin oauth-token detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
}

// OAuthTokenCreateReq declares POST /api/v1/oauth-token.
type OAuthTokenCreateReq struct {
	g.Meta `path:"/api/v1/oauth-token" method:"post" tags:"UIdentity CAS Legacy Resources" summary:"Create OAuth token" dc:"Match the old uidentity/admin oauth-token create contract." permission:"uidentity:cas:write"`
	LegacyOAuthTokenBody
}

// OAuthTokenUpdateReq declares PUT /api/v1/oauth-token/{id}.
type OAuthTokenUpdateReq struct {
	g.Meta `path:"/api/v1/oauth-token/{id}" method:"put" tags:"UIdentity CAS Legacy Resources" summary:"Update OAuth token" dc:"Match the old uidentity/admin oauth-token update contract." permission:"uidentity:cas:write"`
	Id     int64 `json:"id" dc:"Legacy path ID" eg:"1"`
	LegacyOAuthTokenBody
}

// OAuthTokenDeleteReq declares DELETE /api/v1/oauth-token.
type OAuthTokenDeleteReq struct {
	g.Meta `path:"/api/v1/oauth-token" method:"delete" tags:"UIdentity CAS Legacy Resources" summary:"Delete OAuth tokens" dc:"Match the old uidentity/admin oauth-token delete contract." permission:"uidentity:cas:delete"`
	LegacyIDsBody
}

// LegacyPageReq contains the old go-admin pagination query parameters.
type LegacyPageReq struct {
	PageNum  int `json:"pageIndex" dc:"Legacy page number" eg:"1"`
	PageSize int `json:"pageSize" dc:"Legacy page size" eg:"10"`
}

// LegacyIDsBody contains the old generated delete payload.
type LegacyIDsBody struct {
	Ids []int64 `json:"ids" dc:"Legacy ids delete body" eg:"[1,2]"`
}

// LegacyAccountDetailsList contains old account detail list query parameters.
type LegacyAccountDetailsList struct {
	LegacyPageReq
	AccountId int64     `json:"accountId" dc:"Legacy accountId filter" eg:"1"`
	Birthday  time.Time `json:"birthday" dc:"Birthday filter" eg:"2000-01-01T00:00:00+08:00"`
	Email     string    `json:"email" dc:"Email filter" eg:"user@example.com"`
	Gender    int64     `json:"gender" dc:"Gender filter" eg:"1"`
	Qq        string    `json:"qq" dc:"QQ filter" eg:"10001"`
	Wechat    string    `json:"wechat" dc:"Wechat filter" eg:"unionid_001"`
	Idcard    string    `json:"idcard" dc:"Identity card filter" eg:"510000200001010000"`
	Avatar    string    `json:"avatar" dc:"Avatar filter" eg:"https://example.com/avatar.png"`
	Source    string    `json:"source" dc:"Account source filter" eg:"manual"`
	Nj        int64     `json:"nj" dc:"Grade year filter" eg:"2024"`
	Xymc      string    `json:"xymc" dc:"College name filter" eg:"College"`
	Xydm      string    `json:"xydm" dc:"College code filter" eg:"C001"`
	Xq        string    `json:"xq" dc:"Campus filter" eg:"North"`
	Xz        int64     `json:"xz" dc:"School system filter" eg:"4"`
	Yjbysj    int64     `json:"yjbysj" dc:"Expected graduation year filter" eg:"2028"`
	Zymc      string    `json:"zymc" dc:"Major name filter" eg:"Computer Science"`
	Bjmc      string    `json:"bjmc" dc:"Class name filter" eg:"Class 1"`
	AccountDetailsOrder
}

// AccountDetailsOrder contains old account detail order query parameters.
type AccountDetailsOrder struct {
	AccountIdOrder string `json:"accountIdOrder" dc:"Legacy accountId sort direction" eg:"asc"`
	BirthdayOrder  string `json:"birthdayOrder" dc:"Legacy birthday sort direction" eg:"asc"`
	EmailOrder     string `json:"emailOrder" dc:"Legacy email sort direction" eg:"asc"`
	GenderOrder    string `json:"genderOrder" dc:"Legacy gender sort direction" eg:"asc"`
	QqOrder        string `json:"qqOrder" dc:"Legacy qq sort direction" eg:"asc"`
	WechatOrder    string `json:"wechatOrder" dc:"Legacy wechat sort direction" eg:"asc"`
	IdcardOrder    string `json:"idcardOrder" dc:"Legacy idcard sort direction" eg:"asc"`
	AvatarOrder    string `json:"avatarOrder" dc:"Legacy avatar sort direction" eg:"asc"`
	SourceOrder    string `json:"sourceOrder" dc:"Legacy source sort direction" eg:"asc"`
	CreatedAtOrder string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder  string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder  string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
	NjOrder        string `json:"njOrder" dc:"Legacy nj sort direction" eg:"asc"`
	XymcOrder      string `json:"xymcOrder" dc:"Legacy xymc sort direction" eg:"asc"`
	XydmOrder      string `json:"xydmOrder" dc:"Legacy xydm sort direction" eg:"asc"`
	XqOrder        string `json:"xqOrder" dc:"Legacy xq sort direction" eg:"asc"`
	XzOrder        string `json:"xzOrder" dc:"Legacy xz sort direction" eg:"asc"`
	YjbysjOrder    string `json:"yjbysjOrder" dc:"Legacy yjbysj sort direction" eg:"asc"`
	ZymcOrder      string `json:"zymcOrder" dc:"Legacy zymc sort direction" eg:"asc"`
	BjmcOrder      string `json:"bjmcOrder" dc:"Legacy bjmc sort direction" eg:"asc"`
}

// LegacyAccountDetailsCreateBody contains old account detail create JSON fields.
type LegacyAccountDetailsCreateBody struct {
	Birthday time.Time `json:"birthday" dc:"Birthday" eg:"2000-01-01T00:00:00+08:00"`
	Email    string    `json:"email" dc:"Email address" eg:"user@example.com"`
	Gender   int64     `json:"gender" dc:"Gender code" eg:"1"`
	Qq       string    `json:"qq" dc:"QQ number" eg:"10001"`
	Wechat   string    `json:"wechat" dc:"Wechat union ID" eg:"unionid_001"`
	Idcard   string    `json:"idcard" dc:"Identity card" eg:"510000200001010000"`
	Avatar   string    `json:"avatar" dc:"Avatar URL" eg:"https://example.com/avatar.png"`
	Source   string    `json:"source" dc:"Account source" eg:"manual"`
	Nj       int64     `json:"nj" dc:"Grade year" eg:"2024"`
	Xymc     string    `json:"xymc" dc:"College name" eg:"College"`
	Xydm     string    `json:"xydm" dc:"College code" eg:"C001"`
	Xq       string    `json:"xq" dc:"Campus" eg:"North"`
	Xz       int64     `json:"xz" dc:"School system" eg:"4"`
	Yjbysj   int64     `json:"yjbysj" dc:"Expected graduation year" eg:"2028"`
	Zymc     string    `json:"zymc" dc:"Major name" eg:"Computer Science"`
	Bjmc     string    `json:"bjmc" dc:"Class name" eg:"Class 1"`
}

// LegacyAccountDetailsUpdateBody contains old account detail update JSON fields.
type LegacyAccountDetailsUpdateBody struct {
	LegacyAccountDetailsCreateBody
	Face int64 `json:"face" dc:"Face verification marker" eg:"1"`
}

// LegacyAccountUnitList contains old account-unit list query parameters.
type LegacyAccountUnitList struct {
	LegacyPageReq
	AccountUnitOrder
}

// AccountUnitOrder contains old account-unit order query parameters.
type AccountUnitOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	AccountIdOrder string `json:"accountIdOrder" dc:"Legacy accountId sort direction" eg:"asc"`
	UnitIdOrder    string `json:"unitIdOrder" dc:"Legacy unitId sort direction" eg:"asc"`
}

// LegacyAccountUnitBody contains old account-unit JSON fields.
type LegacyAccountUnitBody struct {
	AccountId string `json:"accountId" dc:"Legacy account_id value" eg:"1"`
	UnitId    string `json:"unitId" dc:"Legacy unit_id value" eg:"1"`
}

// LegacyAccountAppRoleList contains old account-app-role list query fields.
type LegacyAccountAppRoleList struct {
	LegacyPageReq
	GiveAccountId      int64     `json:"giveAccountId" dc:"Granting account ID" eg:"1"`
	EmpoweredAccountId int64     `json:"empoweredAccountId" dc:"Empowered account ID" eg:"2"`
	AppId              int64     `json:"appId" dc:"Application ID" eg:"1"`
	ExpireAt           time.Time `json:"expireAt" dc:"Expiration time" eg:"2027-06-02T00:00:00+08:00"`
	AccountAppRoleOrder
}

// AccountAppRoleOrder contains old account-app-role order query fields.
type AccountAppRoleOrder struct {
	IdOrder                 string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	GiveAccountIdOrder      string `json:"giveAccountIdOrder" dc:"Legacy giveAccountId sort direction" eg:"asc"`
	EmpoweredAccountIdOrder string `json:"empoweredAccountIdOrder" dc:"Legacy empoweredAccountId sort direction" eg:"asc"`
	AppIdOrder              string `json:"appIdOrder" dc:"Legacy appId sort direction" eg:"asc"`
	ExpireAtOrder           string `json:"expireAtOrder" dc:"Legacy expireAt sort direction" eg:"desc"`
	CreatedAtOrder          string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder          string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder          string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder           string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder           string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyAccountAppRoleBody contains old account-app-role JSON fields.
type LegacyAccountAppRoleBody struct {
	GiveAccountId      int64     `json:"giveAccountId" dc:"Granting account ID" eg:"1"`
	EmpoweredAccountId int64     `json:"empoweredAccountId" dc:"Empowered account ID" eg:"2"`
	AppId              int64     `json:"appId" dc:"Application ID" eg:"1"`
	ExpireAt           time.Time `json:"expireAt" dc:"Expiration time" eg:"2027-06-02T00:00:00+08:00"`
}

// LegacyAccountAppBlacklistList contains old account-app-blacklist list fields.
type LegacyAccountAppBlacklistList struct {
	LegacyPageReq
	Name      string    `json:"name" dc:"Rule name" eg:"deny"`
	AppId     int64     `json:"appId" dc:"Application ID" eg:"1"`
	AccountId int64     `json:"accountId" dc:"Account ID" eg:"1"`
	EffectAt  time.Time `json:"effectAt" dc:"Effective time" eg:"2026-06-02T00:00:00+08:00"`
	ExpireAt  time.Time `json:"expireAt" dc:"Expiration time" eg:"2027-06-02T00:00:00+08:00"`
	AccountAppBlacklistOrder
}

// AccountAppBlacklistOrder contains old account-app-blacklist order fields.
type AccountAppBlacklistOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	NameOrder      string `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	AppIdOrder     string `json:"appIdOrder" dc:"Legacy appId sort direction" eg:"asc"`
	AccountIdOrder string `json:"accountIdOrder" dc:"Legacy accountId sort direction" eg:"asc"`
	EffectAtOrder  string `json:"effectAtOrder" dc:"Legacy effectAt sort direction" eg:"desc"`
	ExpireAtOrder  string `json:"expireAtOrder" dc:"Legacy expireAt sort direction" eg:"desc"`
	CreatedAtOrder string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder  string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder  string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyAccountAppBlacklistUpdateBody contains old account-app-blacklist update JSON fields.
type LegacyAccountAppBlacklistUpdateBody struct {
	Name      string    `json:"name" dc:"Rule name" eg:"deny"`
	AppId     int64     `json:"appId" dc:"Application ID" eg:"1"`
	AccountId int64     `json:"accountId" dc:"Account ID" eg:"1"`
	EffectAt  time.Time `json:"effectAt" dc:"Effective time" eg:"2026-06-02T00:00:00+08:00"`
	ExpireAt  time.Time `json:"expireAt" dc:"Expiration time" eg:"2027-06-02T00:00:00+08:00"`
}

// LegacyAccountAppBlacklistCreateBody contains old account-app-blacklist create JSON fields.
type LegacyAccountAppBlacklistCreateBody struct {
	LegacyAccountAppBlacklistUpdateBody
	Number string `json:"number" dc:"Legacy account number helper field on create" eg:"A001"`
}

// LegacyAccountChangeLogList contains old account-change-log list fields.
type LegacyAccountChangeLogList struct {
	LegacyPageReq
	AccountId int64  `json:"accountId" dc:"Account ID" eg:"1"`
	TableName string `json:"tableName" dc:"Table name" eg:"account"`
	Action    string `json:"action" dc:"Action" eg:"update"`
	AccountChangeLogOrder
}

// AccountChangeLogOrder contains old account-change-log order fields.
type AccountChangeLogOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	AccountIdOrder string `json:"accountIdOrder" dc:"Legacy accountId sort direction" eg:"asc"`
	TableNameOrder string `json:"tableNameOrder" dc:"Legacy tableName sort direction" eg:"asc"`
	ActionOrder    string `json:"actionOrder" dc:"Legacy action sort direction" eg:"asc"`
	DataOldOrder   string `json:"dataOldOrder" dc:"Legacy dataOld sort direction" eg:"asc"`
	DataNewOrder   string `json:"dataNewOrder" dc:"Legacy dataNew sort direction" eg:"asc"`
	CreatedAtOrder string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder  string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder  string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyAccountChangeLogBody contains old account-change-log JSON fields.
type LegacyAccountChangeLogBody struct {
	AccountId int64  `json:"accountId" dc:"Account ID" eg:"1"`
	TableName string `json:"tableName" dc:"Table name" eg:"account"`
	Action    string `json:"action" dc:"Action" eg:"update"`
	DataOld   string `json:"dataOld" dc:"Old data" eg:"{}"`
	DataNew   string `json:"dataNew" dc:"New data" eg:"{}"`
}

// LegacyUnitsList contains old units list fields.
type LegacyUnitsList struct {
	LegacyPageReq
	Name  string `json:"name" dc:"Unit name" eg:"College"`
	Alias string `json:"alias" dc:"Unit alias" eg:"C"`
	Code  int64  `json:"code" dc:"Unit code" eg:"1001"`
	UnitsOrder
}

// UnitsOrder contains old units order fields.
type UnitsOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	NameOrder      string `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	AliasOrder     string `json:"aliasOrder" dc:"Legacy alias sort direction" eg:"asc"`
	CreatedAtOrder string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder  string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder  string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
	CodeOrder      string `json:"codeOrder" dc:"Legacy code sort direction" eg:"asc"`
}

// LegacyUnitsBody contains old units JSON fields.
type LegacyUnitsBody struct {
	Name  string `json:"name" dc:"Unit name" eg:"College"`
	Alias string `json:"alias" dc:"Unit alias" eg:"C"`
	Code  int64  `json:"code" dc:"Unit code" eg:"1001"`
}

// LegacyGroupsList contains old groups list fields.
type LegacyGroupsList struct {
	LegacyPageReq
	Name  string `json:"name" dc:"Group name" eg:"Teachers"`
	Alias string `json:"alias" dc:"Group alias" eg:"teacher"`
	GroupsOrder
}

// GroupsOrder contains old groups order fields.
type GroupsOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	NameOrder      string `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	AliasOrder     string `json:"aliasOrder" dc:"Legacy alias sort direction" eg:"asc"`
	CreatedAtOrder string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder  string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder  string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyGroupsBody contains old groups JSON fields.
type LegacyGroupsBody struct {
	Name  string `json:"name" dc:"Group name" eg:"Teachers"`
	Alias string `json:"alias" dc:"Group alias" eg:"teacher"`
}

// LegacyContainersList contains old containers list fields.
type LegacyContainersList struct {
	LegacyPageReq
	Name         string `json:"name" dc:"Container name" eg:"students"`
	Alias        string `json:"alias" dc:"Container alias" eg:"Students"`
	AccountCount int64  `json:"accountCount" dc:"Account count" eg:"100"`
	ContainersOrder
}

// ContainersOrder contains old containers order fields.
type ContainersOrder struct {
	IdOrder           string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	NameOrder         string `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	AliasOrder        string `json:"aliasOrder" dc:"Legacy alias sort direction" eg:"asc"`
	AccountCountOrder string `json:"accountCountOrder" dc:"Legacy accountCount sort direction" eg:"asc"`
	AdminCountOrder   string `json:"adminCountOrder" dc:"Legacy adminCount sort direction" eg:"asc"`
	CreatedAtOrder    string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder    string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder    string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder     string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder     string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyContainersBody contains old containers JSON fields.
type LegacyContainersBody struct {
	Name         string `json:"name" dc:"Container name" eg:"students"`
	Alias        string `json:"alias" dc:"Container alias" eg:"Students"`
	AccountCount int64  `json:"accountCount" dc:"Account count" eg:"100"`
	AdminCount   int64  `json:"adminCount" dc:"Admin count" eg:"1"`
}

// LegacyApplicationsList contains old applications list fields.
type LegacyApplicationsList struct {
	LegacyPageReq
	Name        string `json:"name" dc:"Application name" eg:"Portal"`
	Alias       string `json:"alias" dc:"Application alias" eg:"portal"`
	ClientId    string `json:"clientId" dc:"Client ID" eg:"client001"`
	SecretKey   string `json:"secretKey" dc:"Secret key" eg:"secret"`
	AccessModel string `json:"accessModel" dc:"Access model" eg:"cas"`
	Status      int64  `json:"status" dc:"Status" eg:"1"`
	CallbackUrl string `json:"callbackUrl" dc:"Callback URL" eg:"https://example.com/callback"`
	Whitelist   string `json:"whitelist" dc:"IP whitelist" eg:"127.0.0.1"`
	ApplicationsOrder
}

// ApplicationsOrder contains old applications order fields.
type ApplicationsOrder struct {
	IdOrder          string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	NameOrder        string `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	AliasOrder       string `json:"aliasOrder" dc:"Legacy alias sort direction" eg:"asc"`
	ClientIdOrder    string `json:"clientIdOrder" dc:"Legacy clientId sort direction" eg:"asc"`
	SecretKeyOrder   string `json:"secretKeyOrder" dc:"Legacy secretKey sort direction" eg:"asc"`
	AccessModelOrder string `json:"accessModelOrder" dc:"Legacy accessModel sort direction" eg:"asc"`
	StatusOrder      string `json:"statusOrder" dc:"Legacy status sort direction" eg:"asc"`
	CallbackUrlOrder string `json:"callbackUrlOrder" dc:"Legacy callbackUrl sort direction" eg:"asc"`
	WhitelistOrder   string `json:"whitelistOrder" dc:"Legacy whitelist sort direction" eg:"asc"`
	CreatedAtOrder   string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder   string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder   string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder    string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder    string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyApplicationsBody contains old applications JSON fields.
type LegacyApplicationsBody struct {
	Name        string `json:"name" dc:"Application name" eg:"Portal"`
	Alias       string `json:"alias" dc:"Application alias" eg:"portal"`
	ClientId    string `json:"clientId" dc:"Client ID" eg:"client001"`
	SecretKey   string `json:"secretKey" dc:"Secret key" eg:"secret"`
	AccessModel string `json:"accessModel" dc:"Access model" eg:"cas"`
	Status      int64  `json:"status" dc:"Status" eg:"1"`
	CallbackUrl string `json:"callbackUrl" dc:"Callback URL" eg:"https://example.com/callback"`
	Whitelist   string `json:"whitelist" dc:"IP whitelist" eg:"127.0.0.1"`
}

// LegacyGroupAppBlacklistList contains old group-app-blacklist list fields.
type LegacyGroupAppBlacklistList struct {
	LegacyPageReq
	Name     string    `json:"name" dc:"Rule name" eg:"deny"`
	AppId    int64     `json:"appId" dc:"Application ID" eg:"1"`
	GroupId  int64     `json:"groupId" dc:"Group ID" eg:"1"`
	EffectAt time.Time `json:"effectAt" dc:"Effective time" eg:"2026-06-02T00:00:00+08:00"`
	ExpireAt time.Time `json:"expireAt" dc:"Expiration time" eg:"2027-06-02T00:00:00+08:00"`
	GroupAppBlacklistOrder
}

// GroupAppBlacklistOrder contains old group-app-blacklist order fields.
type GroupAppBlacklistOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	NameOrder      string `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	AppIdOrder     string `json:"appIdOrder" dc:"Legacy appId sort direction" eg:"asc"`
	GroupIdOrder   string `json:"groupIdOrder" dc:"Legacy groupId sort direction" eg:"asc"`
	EffectAtOrder  string `json:"effectAtOrder" dc:"Legacy effectAt sort direction" eg:"desc"`
	ExpireAtOrder  string `json:"expireAtOrder" dc:"Legacy expireAt sort direction" eg:"desc"`
	CreatedAtOrder string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder  string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder  string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyGroupAppBlacklistBody contains old group-app-blacklist JSON fields.
type LegacyGroupAppBlacklistBody struct {
	Name     string    `json:"name" dc:"Rule name" eg:"deny"`
	AppId    int64     `json:"appId" dc:"Application ID" eg:"1"`
	GroupId  int64     `json:"groupId" dc:"Group ID" eg:"1"`
	EffectAt time.Time `json:"effectAt" dc:"Effective time" eg:"2026-06-02T00:00:00+08:00"`
	ExpireAt time.Time `json:"expireAt" dc:"Expiration time" eg:"2027-06-02T00:00:00+08:00"`
}

// LegacyPassRulerList contains old pass-ruler list fields.
type LegacyPassRulerList struct {
	LegacyPageReq
	Name string `json:"name" dc:"Rule name" eg:"default"`
	PassRulerOrder
}

// PassRulerOrder contains old pass-ruler order fields.
type PassRulerOrder struct {
	IdOrder             string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	NameOrder           string `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	CapitalOrder        string `json:"capitalOrder" dc:"Legacy capital sort direction" eg:"asc"`
	LowerOrder          string `json:"lowerOrder" dc:"Legacy lower sort direction" eg:"asc"`
	NumberOrder         string `json:"numberOrder" dc:"Legacy number sort direction" eg:"asc"`
	SymbolOrder         string `json:"symbolOrder" dc:"Legacy symbol sort direction" eg:"asc"`
	LengthOrder         string `json:"lengthOrder" dc:"Legacy length sort direction" eg:"asc"`
	IntervalOrder       string `json:"intervalOrder" dc:"Legacy interval sort direction" eg:"asc"`
	IntervalStatusOrder string `json:"intervalStatusOrder" dc:"Legacy intervalStatus sort direction" eg:"asc"`
	StatusOrder         string `json:"statusOrder" dc:"Legacy status sort direction" eg:"asc"`
	CreatedAtOrder      string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder      string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder      string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder       string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder       string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyPassRulerBody contains old pass-ruler JSON fields.
type LegacyPassRulerBody struct {
	Name           string `json:"name" dc:"Rule name" eg:"default"`
	Capital        int64  `json:"capital" dc:"Uppercase flag" eg:"1"`
	Lower          int64  `json:"lower" dc:"Lowercase flag" eg:"1"`
	Number         int64  `json:"number" dc:"Number flag" eg:"1"`
	Symbol         int64  `json:"symbol" dc:"Symbol flag" eg:"1"`
	Length         int64  `json:"length" dc:"Minimum length" eg:"8"`
	Interval       int64  `json:"interval" dc:"Password interval" eg:"90"`
	IntervalStatus int64  `json:"intervalStatus" dc:"Interval status" eg:"1"`
	Status         int64  `json:"status" dc:"Status" eg:"1"`
}

// LegacySmsList contains old generated sms list fields.
type LegacySmsList struct {
	LegacyPageReq
	Phone   string `json:"phone" dc:"Phone" eg:"13800000000"`
	Type    string `json:"type" dc:"SMS type" eg:"cas_login"`
	Content string `json:"content" dc:"Content" eg:"123456"`
	Status  int64  `json:"status" dc:"Status" eg:"1"`
	RespMsg string `json:"respMsg" dc:"Response message" eg:"ok"`
	SmsOrder
}

// SmsOrder contains generated sms order fields.
type SmsOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	PhoneOrder     string `json:"phoneOrder" dc:"Legacy phone sort direction" eg:"asc"`
	TypeOrder      string `json:"typeOrder" dc:"Legacy type sort direction" eg:"asc"`
	ContentOrder   string `json:"contentOrder" dc:"Legacy content sort direction" eg:"asc"`
	StatusOrder    string `json:"statusOrder" dc:"Legacy status sort direction" eg:"asc"`
	RespMsgOrder   string `json:"respMsgOrder" dc:"Legacy respMsg sort direction" eg:"asc"`
	CreatedAtOrder string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder  string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder  string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacySmsBody contains generated sms JSON fields.
type LegacySmsBody struct {
	Phone   string `json:"phone" dc:"Phone" eg:"13800000000"`
	Type    string `json:"type" dc:"SMS type" eg:"cas_login"`
	Content string `json:"content" dc:"Content" eg:"123456"`
	Status  int64  `json:"status" dc:"Status" eg:"1"`
	RespMsg string `json:"respMsg" dc:"Response message" eg:"ok"`
}

// LegacyCasLoginLogList contains old cas-login-log list fields.
type LegacyCasLoginLogList struct {
	LegacyPageReq
	AccountId       int64     `json:"accountId" dc:"Login account ID" eg:"1"`
	ChoiceAccountId int64     `json:"choiceAccountId" dc:"Chosen account ID" eg:"1"`
	AppId           int64     `json:"appId" dc:"Application ID" eg:"1"`
	Ipaddr          string    `json:"ipaddr" dc:"IP address" eg:"127.0.0.1"`
	LoginLocation   string    `json:"loginLocation" dc:"Login location" eg:"Local"`
	Browser         string    `json:"browser" dc:"Browser" eg:"Chrome"`
	Os              string    `json:"os" dc:"Operating system" eg:"macOS"`
	Platform        string    `json:"platform" dc:"Platform" eg:"web"`
	LoginTime       time.Time `json:"loginTime" dc:"Login time" eg:"2026-06-02T00:00:00+08:00"`
	Remark          string    `json:"remark" dc:"Remark" eg:"ok"`
	Msg             string    `json:"msg" dc:"Message" eg:"success"`
	LoginType       string    `json:"loginType" dc:"Login type" eg:"password"`
	CasLoginLogOrder
}

// CasLoginLogOrder contains old cas-login-log order fields.
type CasLoginLogOrder struct {
	IdOrder              string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	AccountIdOrder       string `json:"accountIdOrder" dc:"Legacy accountId sort direction" eg:"asc"`
	ChoiceAccountIdOrder string `json:"choiceAccountIdOrder" dc:"Legacy choiceAccountId sort direction" eg:"asc"`
	AppIdOrder           string `json:"appIdOrder" dc:"Legacy appId sort direction" eg:"asc"`
	IpaddrOrder          string `json:"ipaddrOrder" dc:"Legacy ipaddr sort direction" eg:"asc"`
	LoginLocationOrder   string `json:"loginLocationOrder" dc:"Legacy loginLocation sort direction" eg:"asc"`
	BrowserOrder         string `json:"browserOrder" dc:"Legacy browser sort direction" eg:"asc"`
	OsOrder              string `json:"osOrder" dc:"Legacy os sort direction" eg:"asc"`
	PlatformOrder        string `json:"platformOrder" dc:"Legacy platform sort direction" eg:"asc"`
	LoginTimeOrder       string `json:"loginTimeOrder" dc:"Legacy loginTime sort direction" eg:"desc"`
	RemarkOrder          string `json:"remarkOrder" dc:"Legacy remark sort direction" eg:"asc"`
	MsgOrder             string `json:"msgOrder" dc:"Legacy msg sort direction" eg:"asc"`
	LoginTypeOrder       string `json:"loginTypeOrder" dc:"Legacy loginType sort direction" eg:"asc"`
	CreatedAtOrder       string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder       string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	CreateByOrder        string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder        string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyCasLoginLogBody contains old cas-login-log JSON fields.
type LegacyCasLoginLogBody struct {
	AccountId       int64     `json:"accountId" dc:"Login account ID" eg:"1"`
	ChoiceAccountId int64     `json:"choiceAccountId" dc:"Chosen account ID" eg:"1"`
	AppId           int64     `json:"appId" dc:"Application ID" eg:"1"`
	Ipaddr          string    `json:"ipaddr" dc:"IP address" eg:"127.0.0.1"`
	LoginLocation   string    `json:"loginLocation" dc:"Login location" eg:"Local"`
	Browser         string    `json:"browser" dc:"Browser" eg:"Chrome"`
	Os              string    `json:"os" dc:"Operating system" eg:"macOS"`
	Platform        string    `json:"platform" dc:"Platform" eg:"web"`
	LoginTime       time.Time `json:"loginTime" dc:"Login time" eg:"2026-06-02T00:00:00+08:00"`
	Remark          string    `json:"remark" dc:"Remark" eg:"ok"`
	Msg             string    `json:"msg" dc:"Message" eg:"success"`
}

// LegacyOAuthLogList contains old oauth-log list fields.
type LegacyOAuthLogList struct {
	LegacyPageReq
	UserId      int64  `json:"userId" dc:"Account ID" eg:"1"`
	AppId       int64  `json:"appId" dc:"Application ID" eg:"1"`
	RedirectUri string `json:"redirectUri" dc:"Redirect URI" eg:"https://example.com/callback"`
	Scope       string `json:"scope" dc:"OAuth scope" eg:"openid"`
	OAuthLogOrder
}

// OAuthLogOrder contains old oauth-log order fields.
type OAuthLogOrder struct {
	IdOrder          string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	UserIdOrder      string `json:"userIdOrder" dc:"Legacy userId sort direction" eg:"asc"`
	AppIdOrder       string `json:"appIdOrder" dc:"Legacy appId sort direction" eg:"asc"`
	RedirectUriOrder string `json:"redirectUriOrder" dc:"Legacy redirectUri sort direction" eg:"asc"`
	ScopeOrder       string `json:"scopeOrder" dc:"Legacy scope sort direction" eg:"asc"`
	CreatedAtOrder   string `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder   string `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder   string `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder    string `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder    string `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}

// LegacyOAuthLogBody contains old oauth-log JSON fields.
type LegacyOAuthLogBody struct {
	UserId      int64  `json:"userId" dc:"Account ID" eg:"1"`
	AppId       int64  `json:"appId" dc:"Application ID" eg:"1"`
	RedirectUri string `json:"redirectUri" dc:"Redirect URI" eg:"https://example.com/callback"`
	Scope       string `json:"scope" dc:"OAuth scope" eg:"openid"`
}

// LegacyOAuthTokenList contains old oauth-token list fields.
type LegacyOAuthTokenList struct {
	LegacyPageReq
	ExpiredAt int64  `json:"expiredAt" dc:"Expiration unix time" eg:"1780406400"`
	Code      string `json:"code" dc:"Authorization code" eg:"code001"`
	Access    string `json:"access" dc:"Access token" eg:"access001"`
	Refresh   string `json:"refresh" dc:"Refresh token" eg:"refresh001"`
	Data      string `json:"data" dc:"Token data" eg:"{}"`
	OAuthTokenOrder
}

// OAuthTokenOrder contains old oauth-token order fields.
type OAuthTokenOrder struct {
	IdOrder        string `json:"idOrder" dc:"Legacy id sort direction" eg:"asc"`
	ExpiredAtOrder string `json:"expiredAtOrder" dc:"Legacy expiredAt sort direction" eg:"desc"`
	CodeOrder      string `json:"codeOrder" dc:"Legacy code sort direction" eg:"asc"`
	AccessOrder    string `json:"accessOrder" dc:"Legacy access sort direction" eg:"asc"`
	RefreshOrder   string `json:"refreshOrder" dc:"Legacy refresh sort direction" eg:"asc"`
	DataOrder      string `json:"dataOrder" dc:"Legacy data sort direction" eg:"asc"`
}

// LegacyOAuthTokenBody contains old oauth-token JSON fields.
type LegacyOAuthTokenBody struct {
	ExpiredAt int64  `json:"expiredAt" dc:"Expiration unix time" eg:"1780406400"`
	Code      string `json:"code" dc:"Authorization code" eg:"code001"`
	Access    string `json:"access" dc:"Access token" eg:"access001"`
	Refresh   string `json:"refresh" dc:"Refresh token" eg:"refresh001"`
	Data      string `json:"data" dc:"Token data" eg:"{}"`
}
