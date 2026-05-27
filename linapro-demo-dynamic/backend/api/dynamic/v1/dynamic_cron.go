// This file defines typed cron callback DTOs for the dynamic plugin sample.

package v1

// RegisterCronsReq is the typed request for built-in cron declaration discovery.
type RegisterCronsReq struct{}

// RegisterCronsRes is the response placeholder for cron declaration discovery.
type RegisterCronsRes struct{}

// CronHeartbeatReq is the typed request for executing the sample cron heartbeat.
type CronHeartbeatReq struct{}

// CronHeartbeatRes summarizes one sample cron heartbeat execution.
type CronHeartbeatRes struct {
	Count   int    `json:"count"`
	Message string `json:"message"`
}
