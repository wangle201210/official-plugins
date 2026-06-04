// This file implements the public CMS site settings controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// PublicSite returns public CMS site settings.
func (c *ControllerV1) PublicSite(ctx context.Context, _ *v1.PublicSiteReq) (res *v1.PublicSiteRes, err error) {
	site, err := c.cmsSvc.GetSite(ctx, true)
	if err != nil {
		return nil, err
	}
	return &v1.PublicSiteRes{SiteItem: toAPISite(site)}, nil
}
