// This file implements the CMS friendly link delete controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// LinkDelete deletes one CMS friendly link.
func (c *ControllerV1) LinkDelete(ctx context.Context, req *v1.LinkDeleteReq) (res *v1.LinkDeleteRes, err error) {
	if err := c.cmsSvc.DeleteLink(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.LinkDeleteRes{}, nil
}
