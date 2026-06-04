// This file implements the CMS site data clearing controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// SiteClearData clears CMS business content and resets default site settings.
func (c *ControllerV1) SiteClearData(ctx context.Context, _ *v1.SiteClearDataReq) (res *v1.SiteClearDataRes, err error) {
	if err := c.cmsSvc.ClearSiteData(ctx); err != nil {
		return nil, err
	}
	return &v1.SiteClearDataRes{}, nil
}
