// This file defines shared login-log response DTOs for the linapro-monitor-loginlog API.
package v1

// LoginLogItem exposes login-log fields visible to monitoring callers.
type LoginLogItem struct {
	Id                 int    `json:"id" dc:"Log ID" eg:"1"`
	TenantId           int    `json:"tenantId" dc:"Owning tenant ID, where 0 means platform" eg:"1001"`
	ActingUserId       int    `json:"actingUserId" dc:"Actual acting user ID for platform operations or impersonation" eg:"1"`
	OnBehalfOfTenantId int    `json:"onBehalfOfTenantId" dc:"Target tenant ID when a platform administrator acts on behalf of a tenant" eg:"1001"`
	IsImpersonation    bool   `json:"isImpersonation" dc:"Whether this log was produced during tenant impersonation" eg:"true"`
	UserName           string `json:"userName" dc:"Login account" eg:"admin"`
	Status             int    `json:"status" dc:"Login status: 0=success 1=failed" eg:"0"`
	Ip                 string `json:"ip" dc:"Login IP address" eg:"127.0.0.1"`
	Browser            string `json:"browser" dc:"Browser type" eg:"Chrome 120.0"`
	Os                 string `json:"os" dc:"Operating system" eg:"macOS"`
	Msg                string `json:"msg" dc:"Message" eg:"Login succeeded"`
	LoginTime          *int64 `json:"loginTime" dc:"Login time as Unix timestamp in milliseconds" eg:"1735689600000"`
}
