// This file maps the legacy server monitor service projection into the public
// controller DTO without collecting host metrics in the controller layer.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// LegacyServerMonitor returns old server-monitor compatible data.
func (c *ControllerV1) LegacyServerMonitor(ctx context.Context, req *v1.LegacyServerMonitorReq) (res *v1.LegacyServerMonitorRes, err error) {
	out, err := c.uidentitySvc.ServerMonitor(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.LegacyServerMonitorRes{
		Code:     out.Code,
		OS:       out.OS,
		Mem:      out.Mem,
		CPU:      out.CPU,
		Disk:     out.Disk,
		Net:      out.Net,
		Swap:     out.Swap,
		Location: out.Location,
		BootTime: out.BootTime,
	}, nil
}
