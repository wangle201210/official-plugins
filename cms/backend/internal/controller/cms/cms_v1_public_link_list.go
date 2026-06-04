// This file implements the public CMS link list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// PublicLinkList returns enabled public links.
func (c *ControllerV1) PublicLinkList(ctx context.Context, _ *v1.PublicLinkListReq) (res *v1.PublicLinkListRes, err error) {
	list, err := c.cmsSvc.ListPublicLinks(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.PublicLinkListRes{List: toAPILinks(list)}, nil
}
