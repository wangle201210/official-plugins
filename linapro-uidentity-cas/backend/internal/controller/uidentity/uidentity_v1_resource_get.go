package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ResourceGet returns one UIdentity resource record.
func (c *ControllerV1) ResourceGet(ctx context.Context, req *v1.ResourceGetReq) (res *v1.ResourceGetRes, err error) {
	record, err := c.uidentitySvc.GetResource(ctx, req.Resource, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.ResourceGetRes{Data: v1.ResourceRecord(record)}, nil
}
