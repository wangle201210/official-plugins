// niu_ping.go declares the public ping request/response DTOs for the sicau-niu
// sample plugin API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PingReq is the request for querying the sicau-niu public ping.
type PingReq struct {
	g.Meta `path:"/plugins/sicau-niu/ping" method:"get" tags:"Sicau Niu Demo" summary:"Query sicau-niu example public ping" dc:"Return public ping information for sicau-niu, verifying that one source plugin can register both public and authenticated routes within the same API module."`
}

// PingRes is the response for querying the sicau-niu public ping.
type PingRes struct {
	Message string `json:"message" dc:"Fixed message returned by the public ping endpoint" eg:"pong"`
}
