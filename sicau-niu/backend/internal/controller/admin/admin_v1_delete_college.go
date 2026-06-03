// admin_v1_delete_college.go implements the operator delete-college handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// DeleteCollege soft-deletes one college after reference protection.
func (c *ControllerV1) DeleteCollege(ctx context.Context, req *v1.DeleteCollegeReq) (res *v1.DeleteCollegeRes, err error) {
	err = c.collegeSvc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteCollegeRes{}, nil
}
