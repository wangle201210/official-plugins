// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

type IAdminV1 interface {
	CreateCard(ctx context.Context, req *v1.CreateCardReq) (res *v1.CreateCardRes, err error)
	DeleteCard(ctx context.Context, req *v1.DeleteCardReq) (res *v1.DeleteCardRes, err error)
	GetCard(ctx context.Context, req *v1.GetCardReq) (res *v1.GetCardRes, err error)
	ListCard(ctx context.Context, req *v1.ListCardReq) (res *v1.ListCardRes, err error)
	UpdateCard(ctx context.Context, req *v1.UpdateCardReq) (res *v1.UpdateCardRes, err error)
	CreateCollege(ctx context.Context, req *v1.CreateCollegeReq) (res *v1.CreateCollegeRes, err error)
	DeleteCollege(ctx context.Context, req *v1.DeleteCollegeReq) (res *v1.DeleteCollegeRes, err error)
	ListColleges(ctx context.Context, req *v1.ListCollegesReq) (res *v1.ListCollegesRes, err error)
	UpdateCollege(ctx context.Context, req *v1.UpdateCollegeReq) (res *v1.UpdateCollegeRes, err error)
	CreateHonor(ctx context.Context, req *v1.CreateHonorReq) (res *v1.CreateHonorRes, err error)
	DeleteHonor(ctx context.Context, req *v1.DeleteHonorReq) (res *v1.DeleteHonorRes, err error)
	GetHonor(ctx context.Context, req *v1.GetHonorReq) (res *v1.GetHonorRes, err error)
	ListHonor(ctx context.Context, req *v1.ListHonorReq) (res *v1.ListHonorRes, err error)
	UpdateHonor(ctx context.Context, req *v1.UpdateHonorReq) (res *v1.UpdateHonorRes, err error)
	CreateIron(ctx context.Context, req *v1.CreateIronReq) (res *v1.CreateIronRes, err error)
	DeleteIron(ctx context.Context, req *v1.DeleteIronReq) (res *v1.DeleteIronRes, err error)
	ListIron(ctx context.Context, req *v1.ListIronReq) (res *v1.ListIronRes, err error)
	UpdateIronReportingCycle(ctx context.Context, req *v1.UpdateIronReportingCycleReq) (res *v1.UpdateIronReportingCycleRes, err error)
	UpdateIron(ctx context.Context, req *v1.UpdateIronReq) (res *v1.UpdateIronRes, err error)
	GetMiniappConfig(ctx context.Context, req *v1.GetMiniappConfigReq) (res *v1.GetMiniappConfigRes, err error)
	UpdateMiniappConfig(ctx context.Context, req *v1.UpdateMiniappConfigReq) (res *v1.UpdateMiniappConfigRes, err error)
	CreateNiu(ctx context.Context, req *v1.CreateNiuReq) (res *v1.CreateNiuRes, err error)
	DeleteNiu(ctx context.Context, req *v1.DeleteNiuReq) (res *v1.DeleteNiuRes, err error)
	GetNiu(ctx context.Context, req *v1.GetNiuReq) (res *v1.GetNiuRes, err error)
	ImportNiu(ctx context.Context, req *v1.ImportNiuReq) (res *v1.ImportNiuRes, err error)
	ListNiu(ctx context.Context, req *v1.ListNiuReq) (res *v1.ListNiuRes, err error)
	UpdateNiu(ctx context.Context, req *v1.UpdateNiuReq) (res *v1.UpdateNiuRes, err error)
	ListPhotoAudit(ctx context.Context, req *v1.ListPhotoAuditReq) (res *v1.ListPhotoAuditRes, err error)
	AuditPhotoContent(ctx context.Context, req *v1.AuditPhotoContentReq) (res *v1.AuditPhotoContentRes, err error)
	RevokeActivation(ctx context.Context, req *v1.RevokeActivationReq) (res *v1.RevokeActivationRes, err error)
	ListPlayers(ctx context.Context, req *v1.ListPlayersReq) (res *v1.ListPlayersRes, err error)
	CreateQuote(ctx context.Context, req *v1.CreateQuoteReq) (res *v1.CreateQuoteRes, err error)
	DeleteQuote(ctx context.Context, req *v1.DeleteQuoteReq) (res *v1.DeleteQuoteRes, err error)
	ListQuote(ctx context.Context, req *v1.ListQuoteReq) (res *v1.ListQuoteRes, err error)
	UpdateQuote(ctx context.Context, req *v1.UpdateQuoteReq) (res *v1.UpdateQuoteRes, err error)
}
