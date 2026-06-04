// Package wall implements the sicau-niu C6 public memorial-wall HTTP controllers:
// the first-activator wall, the campus-history highlights and the public activity
// stats. These endpoints are public (no authentication, no player token); they
// expose only nicknames and activity information and never return phone numbers,
// openids or device fingerprints. The controller holds the wall service as a field
// injected at route assembly time and never constructs services on the request
// path.
package wall

import (
	wallapi "lina-plugin-sicau-niu/backend/api/wall"
	wallsvc "lina-plugin-sicau-niu/backend/internal/service/wall"
)

// ControllerV1 is the sicau-niu public memorial-wall controller.
type ControllerV1 struct {
	wallSvc wallsvc.Service // wallSvc serves the first-activator wall, highlights and stats.
}

// NewV1 creates the wall controller with its explicit service dependency.
func NewV1(wallSvc wallsvc.Service) wallapi.IWallV1 {
	return &ControllerV1{wallSvc: wallSvc}
}
