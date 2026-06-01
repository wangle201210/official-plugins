package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ResourceCreate creates one UIdentity resource record.
func (c *ControllerV1) ResourceCreate(ctx context.Context, req *v1.ResourceCreateReq) (res *v1.ResourceCreateRes, err error) {
	id, err := c.uidentitySvc.CreateResource(ctx, req.Resource, req.Body)
	if err != nil {
		return nil, err
	}
	return &v1.ResourceCreateRes{Id: id}, nil
}
