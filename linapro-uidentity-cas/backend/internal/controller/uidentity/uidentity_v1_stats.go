package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// Stats returns aggregate UIdentity CAS statistics.
func (c *ControllerV1) Stats(ctx context.Context, req *v1.StatsReq) (res *v1.StatsRes, err error) {
	out, err := c.uidentitySvc.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.StatsRes{StatsPayload: v1.StatsPayload{
		AccountCount:     out.AccountCount,
		AuthCount:        out.AuthCount,
		AppCount:         out.AppCount,
		UserByContainer:  toAPIStatItems(out.UserByContainer),
		AppByType:        toAPIStatItems(out.AppByType),
		AuthByType:       toAPIStatItems(out.AuthByType),
		CasByAccountType: toAPIStatItems(out.CasByAccountType),
		PassLevel:        toAPIStatItems(out.PassLevel),
		LoginType:        toAPIStatItems(out.LoginType),
		LoginApp:         toAPIStatItems(out.LoginApp),
	}}, nil
}
