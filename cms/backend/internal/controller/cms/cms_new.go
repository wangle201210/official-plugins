// This file defines the CMS controller constructor and dependencies.

package cms

import (
	"github.com/gogf/gf/v2/errors/gerror"

	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ControllerV1 is the CMS plugin controller.
type ControllerV1 struct {
	cmsSvc cmssvc.Service // cmsSvc owns CMS business operations.
}

// NewV1 creates and returns the concrete CMS API controller.
func NewV1(cmsSvc cmssvc.Service) (*ControllerV1, error) {
	if cmsSvc == nil {
		return nil, gerror.New("cms controller requires cms service")
	}
	return &ControllerV1{cmsSvc: cmsSvc}, nil
}
