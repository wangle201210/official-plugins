// Package monitor implements the linapro-monitor-online plugin HTTP controllers.
package monitor

import (
	monitorapi "lina-plugin-linapro-monitor-online/backend/api/monitor"
	monitorsvc "lina-plugin-linapro-monitor-online/backend/internal/service/monitor"
)

// ControllerV1 is the linapro-monitor-online controller.
type ControllerV1 struct {
	monitorSvc monitorsvc.Service // monitor service
}

// NewV1 creates and returns a new linapro-monitor-online controller instance.
func NewV1(monitorSvc monitorsvc.Service) monitorapi.IMonitorV1 {
	return &ControllerV1{monitorSvc: monitorSvc}
}
