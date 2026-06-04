// admin_v1_list_colleges.go implements the operator college list handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
)

// ListColleges returns one DB-side paged college list page.
func (c *ControllerV1) ListColleges(ctx context.Context, req *v1.ListCollegesReq) (res *v1.ListCollegesRes, err error) {
	out, err := c.collegeSvc.List(ctx, &collegesvc.ListInput{
		Keyword:  req.Keyword,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.CollegeItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.CollegeItem{
			Id:        item.Id,
			Name:      item.Name,
			Sort:      item.Sort,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return &v1.ListCollegesRes{List: items, Total: out.Total}, nil
}
