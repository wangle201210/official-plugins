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
	rankingsvc "lina-plugin-sicau-niu/backend/internal/service/ranking"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
)

// ControllerV1 is the sicau-niu operator settlement controller.
type ControllerV1 struct {
	settlementSvc settlementsvc.Service // settlementSvc serves the dashboard, export, issuance, risk view and archive.
	rankingSvc    rankingsvc.Service    // rankingSvc serves operator leaderboard projections.
	rulesSvc      rulessvc.Service      // rulesSvc serves operator runtime-rule configuration.
}

// NewV1 creates the settlement controller with its explicit service dependencies.
func NewV1(settlementSvc settlementsvc.Service, rankingSvc rankingsvc.Service, rulesSvc rulessvc.Service) settlementapi.ISettlementV1 {
	return &ControllerV1{settlementSvc: settlementSvc, rankingSvc: rankingSvc, rulesSvc: rulesSvc}
}
