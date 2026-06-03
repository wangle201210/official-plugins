// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

type IAdminV1 interface {
	CreateCollege(ctx context.Context, req *v1.CreateCollegeReq) (res *v1.CreateCollegeRes, err error)
	DeleteCollege(ctx context.Context, req *v1.DeleteCollegeReq) (res *v1.DeleteCollegeRes, err error)
	ListColleges(ctx context.Context, req *v1.ListCollegesReq) (res *v1.ListCollegesRes, err error)
	UpdateCollege(ctx context.Context, req *v1.UpdateCollegeReq) (res *v1.UpdateCollegeRes, err error)
	ListPlayers(ctx context.Context, req *v1.ListPlayersReq) (res *v1.ListPlayersRes, err error)
}
