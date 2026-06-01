// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

type IUidentityV1 interface {
	AccountPassword(ctx context.Context, req *v1.AccountPasswordReq) (res *v1.AccountPasswordRes, err error)
	AccountPasswordChallenge(ctx context.Context, req *v1.AccountPasswordChallengeReq) (res *v1.AccountPasswordChallengeRes, err error)
	AccountPasswordPhoneVerify(ctx context.Context, req *v1.AccountPasswordPhoneVerifyReq) (res *v1.AccountPasswordPhoneVerifyRes, err error)
	AccountPasswordSelfReset(ctx context.Context, req *v1.AccountPasswordSelfResetReq) (res *v1.AccountPasswordSelfResetRes, err error)
	CasLogin(ctx context.Context, req *v1.CasLoginReq) (res *v1.CasLoginRes, err error)
	OAuthIssue(ctx context.Context, req *v1.OAuthIssueReq) (res *v1.OAuthIssueRes, err error)
	ResourceCreate(ctx context.Context, req *v1.ResourceCreateReq) (res *v1.ResourceCreateRes, err error)
	ResourceDelete(ctx context.Context, req *v1.ResourceDeleteReq) (res *v1.ResourceDeleteRes, err error)
	ResourceGet(ctx context.Context, req *v1.ResourceGetReq) (res *v1.ResourceGetRes, err error)
	ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error)
	ResourceUpdate(ctx context.Context, req *v1.ResourceUpdateReq) (res *v1.ResourceUpdateRes, err error)
	Stats(ctx context.Context, req *v1.StatsReq) (res *v1.StatsRes, err error)
}
