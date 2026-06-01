package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// AccountPassword resets one account password by administrator action.
func (c *ControllerV1) AccountPassword(ctx context.Context, req *v1.AccountPasswordReq) (res *v1.AccountPasswordRes, err error) {
	if err := c.uidentitySvc.ResetAccountPassword(ctx, req.Id, req.NewPassword); err != nil {
		return nil, err
	}
	return &v1.AccountPasswordRes{}, nil
}

// AccountPasswordUnlock clears password failure counters by administrator action.
func (c *ControllerV1) AccountPasswordUnlock(ctx context.Context, req *v1.AccountPasswordUnlockReq) (res *v1.AccountPasswordUnlockRes, err error) {
	unlocked, err := c.uidentitySvc.UnlockPasswordFailures(ctx, req.UserNumberList)
	if err != nil {
		return nil, err
	}
	return &v1.AccountPasswordUnlockRes{UserNumberList: unlocked}, nil
}
