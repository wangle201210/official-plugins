package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ResourceDelete deletes one or more UIdentity resource records.
func (c *ControllerV1) ResourceDelete(ctx context.Context, req *v1.ResourceDeleteReq) (res *v1.ResourceDeleteRes, err error) {
	if err := c.uidentitySvc.DeleteResource(ctx, req.Resource, req.Ids); err != nil {
		return nil, err
	}
	return &v1.ResourceDeleteRes{}, nil
}
