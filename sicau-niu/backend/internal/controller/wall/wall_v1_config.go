// wall_v1_config.go implements the public memorial-wall configuration handler.

package wall

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/wall/v1"
)

// Config returns the public memorial-wall configuration (the mini-program back-link
// URL the H5 wall uses).
func (c *ControllerV1) Config(ctx context.Context, req *v1.ConfigReq) (res *v1.ConfigRes, err error) {
	config, err := c.wallSvc.WallConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ConfigRes{MiniappURL: config.MiniappURL}, nil
}
