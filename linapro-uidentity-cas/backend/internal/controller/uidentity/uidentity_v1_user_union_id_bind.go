package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserUnionIDBind binds a union ID challenge to one account.
func (c *ControllerV1) UserUnionIDBind(ctx context.Context, req *v1.UserUnionIDBindReq) (res *v1.UserUnionIDBindRes, err error) {
	out, err := c.uidentitySvc.BindUnionID(ctx, uidentitysvc.UnionIDBindInput{
		ChallengeID: req.ChallengeId,
		BindType:    req.BindType,
		Phone:       req.Phone,
		Code:        req.Code,
		Number:      req.Number,
		Password:    req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserUnionIDBindRes{Number: out.Number}, nil
}
