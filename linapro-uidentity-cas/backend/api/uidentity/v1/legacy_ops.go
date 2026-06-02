// This file declares legacy operations, upload, monitor, and external-boundary DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LegacyUploadReq defines the legacy-compatible file upload request.
type LegacyUploadReq struct {
	g.Meta `path:"/uidentity/legacy/uploads" method:"post" mime:"multipart/form-data" tags:"UIdentity Legacy Operations" summary:"Upload legacy runtime file" dc:"Upload one file or base64 image into plugin-owned local storage and return fields compatible with the old public uploadFile backend response." permission:"uidentity:cas:write"`
	Type   string `json:"type" dc:"Legacy upload type: 1=single file, 2=multiple files, 3=base64 image; empty defaults to single file" eg:"1"`
	Source string `json:"source" dc:"Legacy third-party source selector; only local plugin storage is used by default" eg:"1"`
	File   string `json:"file" dc:"Base64 image payload when type=3; multipart uploads use the file form field instead" eg:"data:image/png;base64,iVBORw0KGgo="`
}

// LegacyHealthReq defines the health check request.
type LegacyHealthReq struct {
	g.Meta `path:"/uidentity/legacy/health" method:"get" tags:"UIdentity Legacy Operations" summary:"Check legacy backend health" dc:"Return a lightweight health status for old health-check clients without requiring external monitor services."`
}

// LegacyServerMonitorReq defines the server monitor request.
type LegacyServerMonitorReq struct {
	g.Meta `path:"/uidentity/legacy/server-monitor" method:"get" tags:"UIdentity Legacy Operations" summary:"Get legacy server monitor data" dc:"Return runtime, memory, disk, and host data compatible with the old server-monitor backend using local process and operating-system information." permission:"uidentity:cas:read"`
}

// LegacyLogSnapshotReq defines a bounded log snapshot request.
type LegacyLogSnapshotReq struct {
	g.Meta `path:"/uidentity/legacy/log-snapshots" method:"get" tags:"UIdentity Legacy Operations" summary:"Get legacy log snapshot" dc:"Read a bounded tail snapshot from a configured plugin log directory instead of opening an unbounded SSE file watcher." permission:"uidentity:cas:read"`
	Date   string `json:"date" dc:"Log date in YYYY-MM-DD format; defaults to today's date" eg:"2026-06-01"`
	Lines  int    `json:"lines" d:"200" v:"min:1|max:1000" dc:"Maximum number of trailing lines to return, from 1 to 1000" eg:"200"`
}

// LegacyExternalActionReq defines legacy external action boundary requests.
type LegacyExternalActionReq struct {
	g.Meta `path:"/uidentity/legacy/external-actions" method:"post" tags:"UIdentity Legacy Operations" summary:"Run legacy external action" dc:"Expose a stable plugin boundary for old LDAP, external file storage, and monitor actions. Scheduled jobs are provided through LinaPro task management; unconfigured external executors return a structured unsupported-flow error." permission:"uidentity:cas:write"`
	Type   string `json:"type" v:"required" dc:"External action type: ldap_sync, ldap_password_sync, file_external_upload, monitor_external" eg:"ldap_sync"`
	Target string `json:"target" dc:"Optional target identifier such as account number, application client ID, or external resource key" eg:"A001"`
}

// LegacyUploadFile carries one uploaded file response.
type LegacyUploadFile struct {
	Size     int64  `json:"size" dc:"File size in bytes" eg:"102400"`
	Path     string `json:"path" dc:"Plugin-local public path" eg:"/uidentity/uploads/2026/06/file.png"`
	FullPath string `json:"fullPath" dc:"Full public path. The default local implementation returns the same value as path." eg:"/uidentity/uploads/2026/06/file.png"`
	Name     string `json:"name" dc:"Original file name when available" eg:"avatar.png"`
	Type     string `json:"type" dc:"Detected or declared MIME type" eg:"image/png"`
}

// LegacyUploadRes returns upload metadata.
type LegacyUploadRes struct {
	Files []*LegacyUploadFile `json:"files" dc:"Uploaded files. Single-file legacy clients should use the first element." eg:"[]"`
}

// LegacyHealthRes returns health status.
type LegacyHealthRes struct {
	Status string `json:"status" dc:"Health status" eg:"ok"`
}

// LegacyServerMonitorRes returns server monitor data.
type LegacyServerMonitorRes struct {
	Code     int            `json:"code" dc:"Legacy monitor response code" eg:"200"`
	OS       map[string]any `json:"os" dc:"Operating-system and Go runtime projection" eg:"{}"`
	Mem      map[string]any `json:"mem" dc:"Memory usage projection" eg:"{}"`
	CPU      map[string]any `json:"cpu" dc:"CPU usage projection" eg:"{}"`
	Disk     map[string]any `json:"disk" dc:"Root disk usage projection" eg:"{}"`
	Net      map[string]any `json:"net" dc:"Network speed projection. Defaults to zero when no external sampler is configured." eg:"{}"`
	Swap     map[string]any `json:"swap" dc:"Swap usage projection" eg:"{}"`
	Location string         `json:"location" dc:"Configured or local monitor location label" eg:"local"`
	BootTime int64          `json:"bootTime" dc:"Host uptime in hours" eg:"72"`
}

// LegacyLogSnapshotRes returns log tail content.
type LegacyLogSnapshotRes struct {
	Date      string   `json:"date" dc:"Resolved log date in YYYY-MM-DD format" eg:"2026-06-01"`
	Path      string   `json:"path" dc:"Resolved log path" eg:"temp/logs/2026-06-01.log"`
	Lines     []string `json:"lines" dc:"Bounded log tail lines" eg:"[]"`
	Exists    bool     `json:"exists" dc:"Whether the log file existed" eg:"true"`
	Truncated bool     `json:"truncated" dc:"Whether older lines were omitted because of the line limit" eg:"false"`
}

// LegacyExternalActionRes returns external action status when configured.
type LegacyExternalActionRes struct {
	Type    string `json:"type" dc:"External action type" eg:"ldap_sync"`
	Target  string `json:"target" dc:"Requested target identifier" eg:"A001"`
	Success bool   `json:"success" dc:"Whether the action ran successfully" eg:"false"`
}
