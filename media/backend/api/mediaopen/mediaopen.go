// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
)

type IMediaopenV1 interface {
	SetRouteData(ctx context.Context, req *v1.SetRouteDataReq) (res *v1.SetRouteDataRes, err error)
	GetRouteData(ctx context.Context, req *v1.GetRouteDataReq) (res *v1.GetRouteDataRes, err error)
	DelRouteData(ctx context.Context, req *v1.DelRouteDataReq) (res *v1.DelRouteDataRes, err error)
	UserDeviceStrategyByToken(ctx context.Context, req *v1.UserDeviceStrategyByTokenReq) (res *v1.UserDeviceStrategyByTokenRes, err error)
	TenantWhiteIPsByToken(ctx context.Context, req *v1.TenantWhiteIPsByTokenReq) (res *v1.TenantWhiteIPsByTokenRes, err error)
}
