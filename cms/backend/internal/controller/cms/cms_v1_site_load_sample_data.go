// This file implements the CMS sample data loading controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// SiteLoadSampleData clears CMS content and reloads packaged starter data.
func (c *ControllerV1) SiteLoadSampleData(ctx context.Context, _ *v1.SiteLoadSampleDataReq) (res *v1.SiteLoadSampleDataRes, err error) {
	if err := c.cmsSvc.LoadSampleData(ctx); err != nil {
		return nil, err
	}
	return &v1.SiteLoadSampleDataRes{}, nil
}
