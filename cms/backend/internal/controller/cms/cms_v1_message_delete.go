// This file implements the CMS visitor message deletion controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// MessageDelete deletes one CMS visitor message.
func (c *ControllerV1) MessageDelete(ctx context.Context, req *v1.MessageDeleteReq) (res *v1.MessageDeleteRes, err error) {
	if err = c.cmsSvc.DeleteMessage(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.MessageDeleteRes{}, nil
}
