// Package record implements the sicau-niu activity-record query HTTP controllers:
// the feeding, steal, gift, check-in, activation and grass-ledger read-only paged
// queries. These endpoints are operator-facing and bound under the host unified
// Auth+Tenancy+Permission chain; per-route permission is declared on each DTO. The
// controller holds the record service as a field injected at route assembly time
// and never constructs services on the request path.
package record

import (
	recordapi "lina-plugin-sicau-niu/backend/api/record"
	recordsvc "lina-plugin-sicau-niu/backend/internal/service/record"
)

// ControllerV1 is the sicau-niu activity-record query controller.
type ControllerV1 struct {
	recordSvc recordsvc.Service // recordSvc serves the six read-only record queries.
}

// NewV1 creates the record controller with its explicit service dependency.
func NewV1(recordSvc recordsvc.Service) recordapi.IRecordV1 {
	return &ControllerV1{recordSvc: recordSvc}
}
