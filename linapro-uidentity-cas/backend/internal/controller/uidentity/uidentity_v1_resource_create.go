package uidentity

import (
	"context"
	"time"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ResourceCreate creates one UIdentity resource record.
func (c *ControllerV1) ResourceCreate(ctx context.Context, req *v1.ResourceCreateReq) (res *v1.ResourceCreateRes, err error) {
	id, err := c.uidentitySvc.CreateResource(ctx, "accounts", map[string]any{
		"number":      req.Number,
		"name":        req.Name,
		"phone":       req.Phone,
		"effectAt":    zeroTimeToNil(req.EffectAt),
		"expireAt":    zeroTimeToNil(req.ExpireAt),
		"unitId":      req.UnitId,
		"passLevel":   req.PassLevel,
		"containerId": req.ContainerId,
		"groupIds":    req.GroupIDs,
		"status":      req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResourceCreateRes{Id: id}, nil
}

func zeroTimeToNil(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
