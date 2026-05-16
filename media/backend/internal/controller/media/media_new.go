// This file constructs the media plugin controller.

package media

import (
	"lina-plugin-media/backend/api/media"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ControllerV1 is the media plugin controller.
type ControllerV1 struct {
	mediaSvc mediasvc.Service // mediaSvc handles plugin business operations.
}

// NewV1 creates and returns a new media controller instance.
func NewV1(mediaSvc mediasvc.Service) media.IMediaV1 {
	if mediaSvc == nil {
		panic("media controller requires media service")
	}
	return &ControllerV1{mediaSvc: mediaSvc}
}
