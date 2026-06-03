// player_v1_bind_phone.go implements the player phone-binding handler. It binds
// the phone to the authenticated player only, never to an arbitrary ID.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// BindPhone binds a WeChat-authorized phone number to the current player.
func (c *ControllerV1) BindPhone(ctx context.Context, req *v1.BindPhoneReq) (res *v1.BindPhoneRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	err = c.identitySvc.BindPhone(ctx, playerID, &identitysvc.BindPhoneInput{
		Code:              req.Code,
		EncryptedData:     req.EncryptedData,
		IV:                req.Iv,
		PhoneOverride:     req.Phone,
		DeviceFingerprint: req.DeviceFingerprint,
	})
	if err != nil {
		return nil, err
	}
	return &v1.BindPhoneRes{}, nil
}
