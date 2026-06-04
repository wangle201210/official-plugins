// This file constructs the media plugin controller.

package media

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-plugin-media/backend/api/media"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ControllerV1 is the media plugin controller.
type ControllerV1 struct {
	mediaSvc mediasvc.Service // mediaSvc handles plugin business operations.
}

// NewV1 creates and returns a new media controller instance.
func NewV1(mediaSvc mediasvc.Service) (media.IMediaV1, error) {
	if mediaSvc == nil {
		return nil, gerror.New("media controller requires media service")
	}
	return &ControllerV1{mediaSvc: mediaSvc}, nil
}
