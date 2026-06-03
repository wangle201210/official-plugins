// identity_login.go implements WeChat code login: resolve openid via the gateway,
// look up or provision the player by openid, and issue a player session token.

package identity

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// LoginInput defines the WeChat login input.
type LoginInput struct {
	// Code is the WeChat mini-program login code exchanged for an openid.
	Code string
}

// LoginOutput defines the WeChat login result.
type LoginOutput struct {
	// Token is the issued player session token.
	Token string
	// PlayerID is the authenticated player ID.
	PlayerID int64
	// IsNewUser reports whether the account was provisioned during this login.
	IsNewUser bool
}

// Login exchanges a WeChat login code for an authenticated player session.
func (s *serviceImpl) Login(ctx context.Context, in *LoginInput) (*LoginOutput, error) {
	if in == nil || strings.TrimSpace(in.Code) == "" {
		return nil, bizerr.NewCode(CodeLoginCodeRequired)
	}

	openid, err := s.gateway.Code2Session(ctx, strings.TrimSpace(in.Code))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(openid) == "" {
		return nil, bizerr.NewCode(CodeLoginFailed)
	}

	playerID, isNewUser, err := s.resolvePlayerByOpenid(ctx, openid)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenSvc.Sign(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if isNewUser {
		logger.Infof(ctx, "sicau-niu provisioned new player id=%d", playerID)
	}
	return &LoginOutput{Token: token, PlayerID: playerID, IsNewUser: isNewUser}, nil
}

// resolvePlayerByOpenid returns the existing player ID for openid or provisions
// a new account. It reports whether the account was newly created.
func (s *serviceImpl) resolvePlayerByOpenid(ctx context.Context, openid string) (int64, bool, error) {
	var existing *entitymodel.User
	err := dao.User.Ctx(ctx).
		Where(do.User{Openid: openid}).
		Fields(dao.User.Columns().Id).
		Scan(&existing)
	if err != nil {
		return 0, false, bizerr.WrapCode(err, CodePlayerQueryFailed)
	}
	if existing != nil {
		return existing.Id, false, nil
	}

	playerID, err := dao.User.Ctx(ctx).Data(do.User{Openid: openid}).InsertAndGetId()
	if err != nil {
		return 0, false, bizerr.WrapCode(err, CodePlayerWriteFailed)
	}
	return playerID, true, nil
}
