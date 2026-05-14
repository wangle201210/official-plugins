// This file defines the CMS controller constructor and dependencies.

package cms

import (
	"lina-plugin-cms/backend/api/cms"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ControllerV1 is the CMS plugin controller.
type ControllerV1 struct {
	cmsSvc cmssvc.Service // cmsSvc owns CMS business operations.
}

// NewV1 creates and returns the CMS API controller interface.
func NewV1() cms.ICmsV1 {
	return NewControllerV1()
}

// NewControllerV1 creates and returns the concrete CMS controller.
func NewControllerV1() *ControllerV1 {
	return newControllerV1(cmssvc.New())
}

// NewControllerV1WithService creates and returns the concrete CMS controller
// with a caller-provided service implementation.
func NewControllerV1WithService(cmsSvc cmssvc.Service) *ControllerV1 {
	return newControllerV1(cmsSvc)
}

// newControllerV1 builds the CMS controller with a valid service dependency.
func newControllerV1(cmsSvc cmssvc.Service) *ControllerV1 {
	if cmsSvc == nil {
		cmsSvc = cmssvc.New()
	}
	return &ControllerV1{cmsSvc: cmsSvc}
}
