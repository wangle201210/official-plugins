// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
)

type IMediaopenV1 interface {
	ResolveStrategyByToken(ctx context.Context, req *v1.ResolveStrategyByTokenReq) (res *v1.ResolveStrategyByTokenRes, err error)
	UserDeviceStrategyByToken(ctx context.Context, req *v1.UserDeviceStrategyByTokenReq) (res *v1.UserDeviceStrategyByTokenRes, err error)
	SetRouteData(ctx context.Context, req *v1.SetRouteDataReq) (res *v1.SetRouteDataRes, err error)
	GetRouteData(ctx context.Context, req *v1.GetRouteDataReq) (res *v1.GetRouteDataRes, err error)
	DelRouteData(ctx context.Context, req *v1.DelRouteDataReq) (res *v1.DelRouteDataRes, err error)
}
