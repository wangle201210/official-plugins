// player_v1_update_profile.go implements the current-player profile update handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// UpdateProfile updates the current player's own identity profile.
func (c *ControllerV1) UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	err = c.identitySvc.UpdateProfile(ctx, playerID, &identitysvc.UpdateProfileInput{
		Nickname:       req.Nickname,
		Avatar:         req.Avatar,
		IdentityType:   req.IdentityType,
		CollegeId:      req.CollegeId,
		Grade:          req.Grade,
		GraduationYear: req.GraduationYear,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateProfileRes{}, nil
}
