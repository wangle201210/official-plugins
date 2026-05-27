// Cron registration controller.

package dynamic

import (
	"context"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
)

// RegisterCrons publishes the dynamic sample plugin's built-in cron
// declarations for host-side discovery.
func (c *Controller) RegisterCrons(
	_ context.Context,
	_ *v1.RegisterCronsReq,
) (*v1.RegisterCronsRes, error) {
	if err := c.dynamicSvc.RegisterCrons(); err != nil {
		return nil, err
	}
	return &v1.RegisterCronsRes{}, nil
}
