// This file declares the generic resource list endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceListReq defines the request for listing a UIdentity resource.
type ResourceListReq struct {
	g.Meta      `path:"/uidentity/{resource}" method:"get" tags:"UIdentity CAS" summary:"List UIdentity resource records" dc:"Query one plugin-owned UIdentity resource by page. Supported resources are accounts, account-details, groups, units, containers, applications, account-groups, account-units, account-app-roles, account-app-blacklists, group-app-blacklists, pass-rules, sms-records, cas-login-logs, oauth-logs, oauth-tokens, account-change-logs, sys-jobs, and job-logs." permission:"uidentity:cas:read"`
	Resource    string  `json:"resource" v:"required" dc:"Resource name" eg:"accounts"`
	PageNum     int     `json:"pageNum" d:"1" v:"min:1" dc:"Page number, starting from 1" eg:"1"`
	PageSize    int     `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page; max 100" eg:"10"`
	Keyword     string  `json:"keyword" dc:"Optional fuzzy keyword. Accounts match number, name, and phone; other resources match their configured name-like fields." eg:"alice"`
	AccountId   int64   `json:"accountId" dc:"Optional account ID filter when the resource supports account ownership" eg:"1"`
	AppId       int64   `json:"appId" dc:"Optional application ID filter when the resource supports application ownership" eg:"1"`
	GroupId     int64   `json:"groupId" dc:"Optional group ID filter when the resource supports group ownership" eg:"1"`
	ContainerId int64   `json:"containerId" dc:"Optional account container filter" eg:"1"`
	UnitId      int64   `json:"unitId" dc:"Optional account unit filter" eg:"1"`
	Status      *int    `json:"status" dc:"Optional status filter. Resource-specific enum values are documented by the table comments and plugin spec." eg:"1"`
	PassLevels  []int64 `json:"passLevels" dc:"Optional account password-level filter" eg:"[3,4]"`
	GroupIds    []int64 `json:"groupIds" dc:"Optional account group membership filter" eg:"[1,2]"`
	OrderBy     string  `json:"orderBy" dc:"Optional order field. Only resource allowlisted API field names are accepted; defaults to id." eg:"createdAt"`
	Order       string  `json:"order" d:"desc" dc:"Order direction: asc or desc" eg:"desc"`
}
