// admin_v1_update_college.go implements the operator update-college handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
)

// UpdateCollege updates one college dictionary entry.
func (c *ControllerV1) UpdateCollege(ctx context.Context, req *v1.UpdateCollegeReq) (res *v1.UpdateCollegeRes, err error) {
	err = c.collegeSvc.Update(ctx, req.Id, &collegesvc.MutateInput{Name: req.Name, Sort: req.Sort})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateCollegeRes{}, nil
}
