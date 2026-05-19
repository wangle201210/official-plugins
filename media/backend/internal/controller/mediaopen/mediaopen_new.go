// This file constructs the public media token-authentication controller.

package mediaopen

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-plugin-media/backend/api/mediaopen"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ControllerV1 is the public media token-authentication controller.
type ControllerV1 struct {
	mediaSvc mediasvc.Service // mediaSvc handles plugin business operations.
}

// NewV1 creates and returns a new public media controller instance.
func NewV1(mediaSvc mediasvc.Service) (mediaopen.IMediaopenV1, error) {
	if mediaSvc == nil {
		return nil, gerror.New("mediaopen controller requires media service")
	}
	return &ControllerV1{mediaSvc: mediaSvc}, nil
}
