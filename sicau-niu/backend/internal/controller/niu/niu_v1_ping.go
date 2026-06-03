// niu_v1_ping.go projects the public ping service output to the sicau-niu API
// response.

package niu

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/niu/v1"
)

// Ping returns one anonymous ping payload for public route verification.
func (c *ControllerV1) Ping(ctx context.Context, _ *v1.PingReq) (res *v1.PingRes, err error) {
	out, err := c.niuSvc.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.PingRes{Message: out.Message}, nil
}
