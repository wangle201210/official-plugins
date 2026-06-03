// player_v1_login.go implements the public WeChat player login handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// Login exchanges a WeChat login code for a player session token.
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	out, err := c.identitySvc.Login(ctx, &identitysvc.LoginInput{Code: req.Code})
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{
		Token:     out.Token,
		PlayerId:  out.PlayerID,
		IsNewUser: out.IsNewUser,
	}, nil
}
