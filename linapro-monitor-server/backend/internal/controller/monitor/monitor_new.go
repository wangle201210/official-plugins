// Package monitor implements the linapro-monitor-server plugin HTTP controllers.
package monitor

import (
	monitorapi "lina-plugin-linapro-monitor-server/backend/api/monitor"
	monitorsvc "lina-plugin-linapro-monitor-server/backend/internal/service/monitor"
)

// ControllerV1 is the server-monitor controller.
type ControllerV1 struct {
	monitorSvc monitorsvc.Service // server-monitor service
}

// NewV1 creates and returns a new linapro-monitor-server controller instance.
func NewV1(monitorSvc monitorsvc.Service) monitorapi.IMonitorV1 {
	return &ControllerV1{monitorSvc: monitorSvc}
}
