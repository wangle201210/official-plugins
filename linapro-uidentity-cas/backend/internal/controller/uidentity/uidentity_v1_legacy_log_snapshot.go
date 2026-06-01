// This file adapts bounded legacy log snapshot requests to the UIdentity
// service while keeping file-system access out of the controller layer.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// LegacyLogSnapshot returns a bounded tail snapshot from plugin log files.
func (c *ControllerV1) LegacyLogSnapshot(ctx context.Context, req *v1.LegacyLogSnapshotReq) (res *v1.LegacyLogSnapshotRes, err error) {
	out, err := c.uidentitySvc.LogSnapshot(ctx, uidentitysvc.LegacyLogSnapshotInput{
		Date:  req.Date,
		Lines: req.Lines,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LegacyLogSnapshotRes{
		Date:      out.Date,
		Path:      out.Path,
		Lines:     out.Lines,
		Exists:    out.Exists,
		Truncated: out.Truncated,
	}, nil
}
