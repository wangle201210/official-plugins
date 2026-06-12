// player_login.go defines the request and response DTOs for sicau-niu player
// WeChat login. This is the only public player endpoint; it carries no auth and
// returns the issued player session token.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginReq is the request for WeChat mini-program player login.
type LoginReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/login" method:"post" tags:"寻牛小程序" summary:"微信小程序玩家登录" dc:"Exchange a WeChat mini-program login code for a player session token. Provisions the player account on first login. This endpoint is public and requires no authentication."`
	Code   string `json:"code" v:"required" dc:"WeChat mini-program login code returned by wx.login, exchanged server-side for the player's openid" eg:"081xACFa1bZ2..."`
}

// LoginRes is the response for WeChat mini-program player login.
type LoginRes struct {
	Token     string `json:"token" dc:"Issued player session token; send it as 'Authorization: Bearer <token>' on player endpoints" eg:"eyJhbGciOiJIUzI1NiIsInR5cCI6IlNJQ0FVLU5JVS1QTEFZRVIifQ.eyJwaWQiOjF9.signature"`
	PlayerId  int64  `json:"playerId" dc:"Authenticated player ID" eg:"1"`
	IsNewUser bool   `json:"isNewUser" dc:"Whether the account was newly provisioned during this login: true=new account, false=existing account" eg:"true"`
}
