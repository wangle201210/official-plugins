// This file declares the CMS message delete API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MessageDeleteReq defines the request for deleting a visitor message.
type MessageDeleteReq struct {
	g.Meta `path:"/cms/messages/{id}" method:"delete" tags:"CMS Messages" summary:"Delete visitor message" dc:"Delete one visitor message by ID." permission:"cms:message:edit"`
	Id     int64 `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Message ID" eg:"1"`
}

// MessageDeleteRes defines the response for deleting a visitor message.
type MessageDeleteRes struct{}
