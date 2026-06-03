// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

type IPlayerV1 interface {
	CollegeOptions(ctx context.Context, req *v1.CollegeOptionsReq) (res *v1.CollegeOptionsRes, err error)
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	BindPhone(ctx context.Context, req *v1.BindPhoneReq) (res *v1.BindPhoneRes, err error)
	GetProfile(ctx context.Context, req *v1.GetProfileReq) (res *v1.GetProfileRes, err error)
	UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error)
}
