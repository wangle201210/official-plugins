package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ResourceUpdate updates one UIdentity resource record.
func (c *ControllerV1) ResourceUpdate(ctx context.Context, req *v1.ResourceUpdateReq) (res *v1.ResourceUpdateRes, err error) {
	body := map[string]any{}
	if req.Number != "" {
		body["number"] = req.Number
	}
	if req.Name != "" {
		body["name"] = req.Name
	}
	if req.Phone != "" {
		body["phone"] = req.Phone
	}
	if value := zeroTimeToNil(req.EffectAt); value != nil {
		body["effectAt"] = value
	}
	if value := zeroTimeToNil(req.ExpireAt); value != nil {
		body["expireAt"] = value
	}
	if len(req.GroupIds) > 0 {
		body["groupIds"] = req.GroupIds
	}
	if req.PassLevel != 0 {
		body["passLevel"] = req.PassLevel
	}
	if req.ContainerId != 0 {
		body["containerId"] = req.ContainerId
	}
	if req.UnitId != 0 {
		body["unitId"] = req.UnitId
	}
	if req.Status != 0 {
		body["status"] = req.Status
	}
	if err := c.uidentitySvc.UpdateResource(ctx, "accounts", req.Id, body); err != nil {
		return nil, err
	}
	return &v1.ResourceUpdateRes{}, nil
}
