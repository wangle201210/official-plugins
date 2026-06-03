// admin_new.go defines the operator-facing controller and its constructor. The
// controller holds the college and identity services as fields injected at route
// assembly time; it never constructs services on the request path. These
// endpoints are governed by the host Auth+Tenancy+Permission chain, with the
// concrete permission declared on each API DTO's g.Meta tag.

package admin

import (
	"lina-plugin-sicau-niu/backend/api/admin"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// ControllerV1 is the sicau-niu operator-facing controller.
type ControllerV1 struct {
	collegeSvc  collegesvc.Service  // collegeSvc handles college dictionary CRUD.
	identitySvc identitysvc.Service // identitySvc provides the read-only player query.
}

// NewV1 creates the operator controller with explicit service dependencies.
func NewV1(collegeSvc collegesvc.Service, identitySvc identitysvc.Service) admin.IAdminV1 {
	return &ControllerV1{
		collegeSvc:  collegeSvc,
		identitySvc: identitySvc,
	}
}
