// admin_new.go defines the operator-facing controller and its constructor. The
// controller holds the college, identity, cattle and card services as fields
// injected at route assembly time; it never constructs services on the request
// path. These endpoints are governed by the host Auth+Tenancy+Permission chain,
// with the concrete permission declared on each API DTO's g.Meta tag.

package admin

import (
	"lina-plugin-sicau-niu/backend/api/admin"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// ControllerV1 is the sicau-niu operator-facing controller.
type ControllerV1 struct {
	collegeSvc  collegesvc.Service  // collegeSvc handles college dictionary CRUD.
	identitySvc identitysvc.Service // identitySvc provides the read-only player query.
	cattleSvc   cattlesvc.Service   // cattleSvc handles cattle and iron-cow CRUD.
	cardSvc     cardsvc.Service     // cardSvc handles card and quote CRUD.
}

// NewV1 creates the operator controller with explicit service dependencies.
func NewV1(
	collegeSvc collegesvc.Service,
	identitySvc identitysvc.Service,
	cattleSvc cattlesvc.Service,
	cardSvc cardsvc.Service,
) admin.IAdminV1 {
	return &ControllerV1{
		collegeSvc:  collegeSvc,
		identitySvc: identitySvc,
		cattleSvc:   cattleSvc,
		cardSvc:     cardSvc,
	}
}
