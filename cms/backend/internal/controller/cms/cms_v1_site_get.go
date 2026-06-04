// This file implements the CMS site settings detail controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// SiteGet returns CMS site settings.
func (c *ControllerV1) SiteGet(ctx context.Context, _ *v1.SiteGetReq) (res *v1.SiteGetRes, err error) {
	site, err := c.cmsSvc.GetSite(ctx, false)
	if err != nil {
		return nil, err
	}
	return &v1.SiteGetRes{SiteItem: toAPISite(site)}, nil
}
