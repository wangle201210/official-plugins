// This file implements the CMS visitor message moderation controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// MessageUpdate updates CMS visitor message moderation state.
func (c *ControllerV1) MessageUpdate(ctx context.Context, req *v1.MessageUpdateReq) (res *v1.MessageUpdateRes, err error) {
	err = c.cmsSvc.UpdateMessage(ctx, cmssvc.MessageUpdateInput{
		Id:     req.Id,
		Status: req.Status,
		Reply:  req.Reply,
	})
	if err != nil {
		return nil, err
	}
	return &v1.MessageUpdateRes{}, nil
}
