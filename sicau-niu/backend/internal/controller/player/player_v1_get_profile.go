// player_v1_get_profile.go implements the current-player profile read handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// GetProfile returns the current player's own identity profile.
func (c *ControllerV1) GetProfile(ctx context.Context, req *v1.GetProfileReq) (res *v1.GetProfileRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.identitySvc.GetProfile(ctx, playerID)
	if err != nil {
		return nil, err
	}
	progress, err := c.grassSvc.Progress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &v1.GetProfileRes{
		PlayerId:        out.Id,
		Phone:           out.Phone,
		Nickname:        out.Nickname,
		Avatar:          out.Avatar,
		IdentityType:    out.IdentityType,
		CollegeId:       out.CollegeId,
		Grade:           out.Grade,
		GraduationYear:  out.GraduationYear,
		Level:           progress.Level,
		ExpIntoLevel:    progress.ExpIntoLevel,
		ExpForNextLevel: progress.ExpForNextLevel,
		Exp:             progress.Exp,
		CreatedAt:       out.CreatedAt,
		UpdatedAt:       out.UpdatedAt,
	}, nil
}
