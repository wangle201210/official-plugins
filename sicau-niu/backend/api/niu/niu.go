// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package niu

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/niu/v1"
)

type INiuV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Ping(ctx context.Context, req *v1.PingReq) (res *v1.PingRes, err error)
}
