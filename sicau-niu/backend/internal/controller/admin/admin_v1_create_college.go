// admin_v1_create_college.go implements the operator create-college handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
)

// CreateCollege creates one college dictionary entry.
func (c *ControllerV1) CreateCollege(ctx context.Context, req *v1.CreateCollegeReq) (res *v1.CreateCollegeRes, err error) {
	id, err := c.collegeSvc.Create(ctx, &collegesvc.MutateInput{Name: req.Name, Sort: req.Sort})
	if err != nil {
		return nil, err
	}
	return &v1.CreateCollegeRes{Id: id}, nil
}
