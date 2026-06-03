// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

type IPlayerV1 interface {
	Activate(ctx context.Context, req *v1.ActivateReq) (res *v1.ActivateRes, err error)
	Collection(ctx context.Context, req *v1.CollectionReq) (res *v1.CollectionRes, err error)
	CollegeOptions(ctx context.Context, req *v1.CollegeOptionsReq) (res *v1.CollegeOptionsRes, err error)
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	VisibleNiu(ctx context.Context, req *v1.VisibleNiuReq) (res *v1.VisibleNiuRes, err error)
	BindPhone(ctx context.Context, req *v1.BindPhoneReq) (res *v1.BindPhoneRes, err error)
	Poster(ctx context.Context, req *v1.PosterReq) (res *v1.PosterRes, err error)
	GetProfile(ctx context.Context, req *v1.GetProfileReq) (res *v1.GetProfileRes, err error)
	UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error)
}
