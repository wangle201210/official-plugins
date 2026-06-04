// player_v1_college_options.go implements the player college dropdown handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// CollegeOptions returns the bounded, sort-ordered college dropdown for player
// identity selection.
func (c *ControllerV1) CollegeOptions(ctx context.Context, req *v1.CollegeOptionsReq) (res *v1.CollegeOptionsRes, err error) {
	options, err := c.collegeSvc.Options(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.CollegeOptionItem, 0, len(options))
	for _, option := range options {
		items = append(items, &v1.CollegeOptionItem{Id: option.Id, Name: option.Name})
	}
	return &v1.CollegeOptionsRes{List: items}, nil
}
