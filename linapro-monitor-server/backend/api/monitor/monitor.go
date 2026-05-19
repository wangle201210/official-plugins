// Package monitor declares the HTTP controller contract exposed by the
// linapro-monitor-server source plugin.
package monitor

import (
	"context"

	"lina-plugin-linapro-monitor-server/backend/api/monitor/v1"
)

// IMonitorV1 defines the linapro-monitor-server HTTP handlers.
type IMonitorV1 interface {
	// ServerMonitor returns the current server-monitor projection.
	ServerMonitor(ctx context.Context, req *v1.ServerMonitorReq) (res *v1.ServerMonitorRes, err error)
}
