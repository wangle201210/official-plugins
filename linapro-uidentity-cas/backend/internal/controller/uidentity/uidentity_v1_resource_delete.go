package uidentity

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ResourceDelete deletes one or more UIdentity resource records.
func (c *ControllerV1) ResourceDelete(ctx context.Context, req *v1.ResourceDeleteReq) (res *v1.ResourceDeleteRes, err error) {
	if err := c.uidentitySvc.DeleteResource(ctx, "accounts", legacyIDCSV(req.Ids)); err != nil {
		return nil, err
	}
	return &v1.ResourceDeleteRes{}, nil
}

func legacyIDCSV(ids []int64) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, gconv.String(id))
		}
	}
	return strings.Join(parts, ",")
}
