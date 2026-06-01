package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// UserUnionIDLookup resolves a union ID or creates a bind challenge.
func (c *ControllerV1) UserUnionIDLookup(ctx context.Context, req *v1.UserUnionIDLookupReq) (res *v1.UserUnionIDLookupRes, err error) {
	out, err := c.uidentitySvc.LookupUnionID(ctx, req.UnionId)
	if err != nil {
		return nil, err
	}
	return &v1.UserUnionIDLookupRes{
		Number:      out.Number,
		ChallengeId: out.ChallengeID,
		CallbackUrl: out.CallbackURL,
	}, nil
}
