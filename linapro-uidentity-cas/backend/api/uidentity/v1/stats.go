// This file declares aggregate UIdentity statistics endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// StatsReq defines the request for aggregate identity statistics.
type StatsReq struct {
	g.Meta `path:"/uidentity/stats" method:"get" tags:"UIdentity CAS" summary:"Get UIdentity aggregate statistics" dc:"Return aggregate account, application, CAS, OAuth, password-level, login-type, and login-application statistics using database-side grouping and batch projection." permission:"uidentity:cas:read"`
}

// StatsRes returns aggregate identity statistics.
type StatsRes struct {
	StatsPayload
}
