package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ResourceUpdate updates one UIdentity resource record.
func (c *ControllerV1) ResourceUpdate(ctx context.Context, req *v1.ResourceUpdateReq) (res *v1.ResourceUpdateRes, err error) {
	if err := c.uidentitySvc.UpdateResource(ctx, req.Resource, req.Id, req.Body); err != nil {
		return nil, err
	}
	return &v1.ResourceUpdateRes{}, nil
}
