// Declared cron heartbeat controller.

package dynamic

import (
	"context"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
)

// CronHeartbeat executes the declared cron heartbeat task for the dynamic
// sample plugin.
func (c *Controller) CronHeartbeat(
	_ context.Context,
	_ *v1.CronHeartbeatReq,
) (*v1.CronHeartbeatRes, error) {
	payload, err := c.dynamicSvc.BuildCronHeartbeatPayload()
	if err != nil {
		return nil, err
	}
	return &v1.CronHeartbeatRes{
		Count:   payload.Count,
		Message: payload.Message,
	}, nil
}
