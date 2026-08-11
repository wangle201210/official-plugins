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
	Activities(ctx context.Context, req *v1.ActivitiesReq) (res *v1.ActivitiesRes, err error)
	Collection(ctx context.Context, req *v1.CollectionReq) (res *v1.CollectionRes, err error)
	Certificate(ctx context.Context, req *v1.CertificateReq) (res *v1.CertificateRes, err error)
	Checkin(ctx context.Context, req *v1.CheckinReq) (res *v1.CheckinRes, err error)
	CollegeOptions(ctx context.Context, req *v1.CollegeOptionsReq) (res *v1.CollegeOptionsRes, err error)
	Config(ctx context.Context, req *v1.ConfigReq) (res *v1.ConfigRes, err error)
	Feed(ctx context.Context, req *v1.FeedReq) (res *v1.FeedRes, err error)
	FeedingTrail(ctx context.Context, req *v1.FeedingTrailReq) (res *v1.FeedingTrailRes, err error)
	Gift(ctx context.Context, req *v1.GiftReq) (res *v1.GiftRes, err error)
	GrassAccount(ctx context.Context, req *v1.GrassAccountReq) (res *v1.GrassAccountRes, err error)
	PlayerHonors(ctx context.Context, req *v1.PlayerHonorsReq) (res *v1.PlayerHonorsRes, err error)
	IronTransportState(ctx context.Context, req *v1.IronTransportStateReq) (res *v1.IronTransportStateRes, err error)
	CreateTransportTeam(ctx context.Context, req *v1.CreateTransportTeamReq) (res *v1.CreateTransportTeamRes, err error)
	JoinTransportTeam(ctx context.Context, req *v1.JoinTransportTeamReq) (res *v1.JoinTransportTeamRes, err error)
	LeaveTransportTeam(ctx context.Context, req *v1.LeaveTransportTeamReq) (res *v1.LeaveTransportTeamRes, err error)
	StartTransport(ctx context.Context, req *v1.StartTransportReq) (res *v1.StartTransportRes, err error)
	HeartbeatTransport(ctx context.Context, req *v1.HeartbeatTransportReq) (res *v1.HeartbeatTransportRes, err error)
	EndTransport(ctx context.Context, req *v1.EndTransportReq) (res *v1.EndTransportRes, err error)
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	Messages(ctx context.Context, req *v1.MessagesReq) (res *v1.MessagesRes, err error)
	MarkMessageRead(ctx context.Context, req *v1.MarkMessageReadReq) (res *v1.MarkMessageReadRes, err error)
	VisibleNiu(ctx context.Context, req *v1.VisibleNiuReq) (res *v1.VisibleNiuRes, err error)
	NiuDetail(ctx context.Context, req *v1.NiuDetailReq) (res *v1.NiuDetailRes, err error)
	BindPhone(ctx context.Context, req *v1.BindPhoneReq) (res *v1.BindPhoneRes, err error)
	UploadPhoto(ctx context.Context, req *v1.UploadPhotoReq) (res *v1.UploadPhotoRes, err error)
	PhotoContent(ctx context.Context, req *v1.PhotoContentReq) (res *v1.PhotoContentRes, err error)
	Poster(ctx context.Context, req *v1.PosterReq) (res *v1.PosterRes, err error)
	GetProfile(ctx context.Context, req *v1.GetProfileReq) (res *v1.GetProfileRes, err error)
	UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error)
	FeedRanking(ctx context.Context, req *v1.FeedRankingReq) (res *v1.FeedRankingRes, err error)
	CollegeRanking(ctx context.Context, req *v1.CollegeRankingReq) (res *v1.CollegeRankingRes, err error)
	FriendRanking(ctx context.Context, req *v1.FriendRankingReq) (res *v1.FriendRankingRes, err error)
	StealTargets(ctx context.Context, req *v1.StealTargetsReq) (res *v1.StealTargetsRes, err error)
	Steal(ctx context.Context, req *v1.StealReq) (res *v1.StealRes, err error)
}
