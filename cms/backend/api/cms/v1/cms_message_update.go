// This file declares the CMS message moderation API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MessageUpdateReq defines the request for moderating a visitor message.
type MessageUpdateReq struct {
	g.Meta `path:"/cms/messages/{id}" method:"put" tags:"CMS Messages" summary:"Update visitor message moderation" dc:"Update visitor message status and reply." permission:"cms:message:edit"`
	Id     int64  `json:"id" v:"required|min:1#gf.gvalid.rule.required|gf.gvalid.rule.min" dc:"Message ID" eg:"1"`
	Status int    `json:"status" v:"required|in:0,1,2#gf.gvalid.rule.required|gf.gvalid.rule.in" dc:"Status: 0=pending, 1=approved, 2=rejected" eg:"1"`
	Reply  string `json:"reply" dc:"Reply content" eg:"Thanks"`
}

// MessageUpdateRes defines the response for moderating a visitor message.
type MessageUpdateRes struct{}
