// admin_new.go defines the operator-facing controller and its constructor. The
// controller holds the college, identity, cattle, IOT reporting-cycle, card and
// honor services as fields injected at route assembly time; it never constructs
// services on the request path. These endpoints are governed by the host
// Auth+Tenancy+Permission chain, with the concrete permission declared on each
// API DTO's g.Meta tag.

package admin

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-plugin-sicau-niu/backend/api/admin"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
	activationphotosvc "lina-plugin-sicau-niu/backend/internal/service/activationphoto"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
	miniappconfigsvc "lina-plugin-sicau-niu/backend/internal/service/miniappconfig"
)

// ControllerV1 is the sicau-niu operator-facing controller.
type ControllerV1 struct {
	collegeSvc            collegesvc.Service                   // collegeSvc handles college dictionary CRUD.
	identitySvc           identitysvc.Service                  // identitySvc provides the read-only player query.
	cattleSvc             cattlesvc.Service                    // cattleSvc handles cattle and iron-cow CRUD.
	reportingCycleUpdater feedingsvc.IronReportingCycleUpdater // reportingCycleUpdater sends IOT locator commands.
	cardSvc               cardsvc.Service                      // cardSvc handles card and quote CRUD.
	honorSvc              honorsvc.Service                     // honorSvc handles honor-definition CRUD.
	miniappConfigSvc      miniappconfigsvc.Service             // miniappConfigSvc maintains public runtime config.
	photoSvc              activationphotosvc.Service           // photoSvc serves protected evidence audit.
	activationSvc         activationsvc.Service                // activationSvc repairs erroneous activations.
}

// NewV1 creates the operator controller with explicit service dependencies.
func NewV1(
	collegeSvc collegesvc.Service,
	identitySvc identitysvc.Service,
	cattleSvc cattlesvc.Service,
	reportingCycleUpdater feedingsvc.IronReportingCycleUpdater,
	cardSvc cardsvc.Service,
	honorSvc honorsvc.Service,
	miniappConfigSvc miniappconfigsvc.Service,
	photoSvc activationphotosvc.Service,
	activationSvc activationsvc.Service,
) (admin.IAdminV1, error) {
	switch {
	case collegeSvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires college service")
	case identitySvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires identity service")
	case cattleSvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires cattle service")
	case reportingCycleUpdater == nil:
		return nil, gerror.New("sicau-niu admin controller requires reporting-cycle updater")
	case cardSvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires card service")
	case honorSvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires honor service")
	case miniappConfigSvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires miniapp config service")
	case photoSvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires activation photo service")
	case activationSvc == nil:
		return nil, gerror.New("sicau-niu admin controller requires activation service")
	}
	return &ControllerV1{
		collegeSvc:            collegeSvc,
		identitySvc:           identitySvc,
		cattleSvc:             cattleSvc,
		reportingCycleUpdater: reportingCycleUpdater,
		cardSvc:               cardSvc,
		honorSvc:              honorSvc,
		miniappConfigSvc:      miniappConfigSvc,
		photoSvc:              photoSvc,
		activationSvc:         activationSvc,
	}, nil
}
