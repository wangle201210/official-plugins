// Package settlement implements the sicau-niu C7 operator settlement HTTP
// controllers: the operations dashboard, the player roster export, the batch
// certificate issuance, the shared-device risk view and the settlement archive
// (create and list). These endpoints are operator-facing and bound under the host
// unified Auth+Tenancy+Permission chain; per-route permission is declared on each
// DTO. The controller holds the settlement service as a field injected at route
// assembly time and never constructs services on the request path.
package settlement

import (
	settlementapi "lina-plugin-sicau-niu/backend/api/settlement"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
)

// ControllerV1 is the sicau-niu operator settlement controller.
type ControllerV1 struct {
	settlementSvc settlementsvc.Service // settlementSvc serves the dashboard, export, issuance, risk view and archive.
}

// NewV1 creates the settlement controller with its explicit service dependency.
func NewV1(settlementSvc settlementsvc.Service) settlementapi.ISettlementV1 {
	return &ControllerV1{settlementSvc: settlementSvc}
}
