// record_handlers.go implements the six activity-record query handlers and their
// service-to-DTO projections. Each handler delegates the DB-side paged query to the
// record service and projects the assembled rows to the response DTO.

package record

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/record/v1"
	recordsvc "lina-plugin-sicau-niu/backend/internal/service/record"
)

// Feedings returns the paged feeding records.
func (c *ControllerV1) Feedings(ctx context.Context, req *v1.FeedingsReq) (res *v1.FeedingsRes, err error) {
	out, err := c.recordSvc.ListFeedings(ctx, &recordsvc.ListFeedingsInput{
		NiuId: req.NiuId, UserId: req.UserId, PageNum: req.PageNum, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.FeedingItem, 0, len(out.List))
	for _, row := range out.List {
		list = append(list, &v1.FeedingItem{
			Id: row.Id, UserId: row.UserId, Nickname: row.Nickname,
			NiuId: row.NiuId, NiuName: row.NiuName, NiuCode: row.NiuCode,
			BaseAmount: row.BaseAmount, CoefficientBasis: row.CoefficientBasis,
			EffectAmount: row.EffectAmount, IsIronBonus: row.IsIronBonus,
			FedAt: row.FedAt, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.FeedingsRes{List: list, Total: out.Total}, nil
}

// Steals returns the paged steal records.
func (c *ControllerV1) Steals(ctx context.Context, req *v1.StealsReq) (res *v1.StealsRes, err error) {
	out, err := c.recordSvc.ListSteals(ctx, &recordsvc.ListStealsInput{
		ActorUserId: req.ActorUserId, TargetUserId: req.TargetUserId, PageNum: req.PageNum, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.StealItem, 0, len(out.List))
	for _, row := range out.List {
		list = append(list, &v1.StealItem{
			Id: row.Id, ActorUserId: row.ActorUserId, ActorNickname: row.ActorNickname,
			TargetUserId: row.TargetUserId, TargetNickname: row.TargetNickname,
			Amount: row.Amount, StealDate: row.StealDate, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.StealsRes{List: list, Total: out.Total}, nil
}

// Gifts returns the paged gift records.
func (c *ControllerV1) Gifts(ctx context.Context, req *v1.GiftsReq) (res *v1.GiftsRes, err error) {
	out, err := c.recordSvc.ListGifts(ctx, &recordsvc.ListGiftsInput{
		FromUserId: req.FromUserId, ToUserId: req.ToUserId, PageNum: req.PageNum, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.GiftItem, 0, len(out.List))
	for _, row := range out.List {
		list = append(list, &v1.GiftItem{
			Id: row.Id, FromUserId: row.FromUserId, FromNickname: row.FromNickname,
			ToUserId: row.ToUserId, ToNickname: row.ToNickname,
			Amount: row.Amount, GiftDate: row.GiftDate, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.GiftsRes{List: list, Total: out.Total}, nil
}

// Checkins returns the paged check-in records.
func (c *ControllerV1) Checkins(ctx context.Context, req *v1.CheckinsReq) (res *v1.CheckinsRes, err error) {
	out, err := c.recordSvc.ListCheckins(ctx, &recordsvc.ListCheckinsInput{
		UserId: req.UserId, PageNum: req.PageNum, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.CheckinItem, 0, len(out.List))
	for _, row := range out.List {
		list = append(list, &v1.CheckinItem{
			Id: row.Id, UserId: row.UserId, Nickname: row.Nickname,
			CheckinDate: row.CheckinDate, Amount: row.Amount, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.CheckinsRes{List: list, Total: out.Total}, nil
}

// Activations returns the paged activation records.
func (c *ControllerV1) Activations(ctx context.Context, req *v1.ActivationsReq) (res *v1.ActivationsRes, err error) {
	out, err := c.recordSvc.ListActivations(ctx, &recordsvc.ListActivationsInput{
		UserId: req.UserId, NiuId: req.NiuId, PageNum: req.PageNum, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ActivationItem, 0, len(out.List))
	for _, row := range out.List {
		list = append(list, &v1.ActivationItem{
			Id: row.Id, UserId: row.UserId, Nickname: row.Nickname,
			NiuId: row.NiuId, NiuName: row.NiuName, NiuCode: row.NiuCode,
			ActivityDate: row.ActivityDate, IsFirst: row.IsFirst, OrderNo: row.OrderNo,
			PhotoPath: row.PhotoPath, ActivatedAt: row.ActivatedAt, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.ActivationsRes{List: list, Total: out.Total}, nil
}

// ActivationAttempts returns the paged photo check-in attempt audit records.
func (c *ControllerV1) ActivationAttempts(ctx context.Context, req *v1.ActivationAttemptsReq) (res *v1.ActivationAttemptsRes, err error) {
	out, err := c.recordSvc.ListActivationAttempts(ctx, &recordsvc.ListActivationAttemptsInput{
		UserId: req.UserId, NiuId: req.NiuId, Result: req.Result, PageNum: req.PageNum, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ActivationAttemptItem, 0, len(out.List))
	for _, row := range out.List {
		list = append(list, &v1.ActivationAttemptItem{
			Id: row.Id, UserId: row.UserId, Nickname: row.Nickname,
			NiuId: row.NiuId, NiuName: row.NiuName, NiuCode: row.NiuCode,
			NearestNiuId: row.NearestNiuId, NearestNiuName: row.NearestNiuName, NearestNiuCode: row.NearestNiuCode,
			Result: row.Result, Lat: row.Lat, Lng: row.Lng,
			DistanceM: row.DistanceM, ThresholdM: row.ThresholdM, PhotoPath: row.PhotoPath,
			AttemptedAt: row.AttemptedAt, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.ActivationAttemptsRes{List: list, Total: out.Total}, nil
}

// GrassTxns returns the paged grass-ledger entries.
func (c *ControllerV1) GrassTxns(ctx context.Context, req *v1.GrassTxnsReq) (res *v1.GrassTxnsRes, err error) {
	out, err := c.recordSvc.ListGrassTxns(ctx, &recordsvc.ListGrassTxnsInput{
		UserId: req.UserId, PageNum: req.PageNum, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.GrassTxnItem, 0, len(out.List))
	for _, row := range out.List {
		list = append(list, &v1.GrassTxnItem{
			Id: row.Id, UserId: row.UserId, Nickname: row.Nickname,
			Delta: row.Delta, TxnType: row.TxnType, RefId: row.RefId, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.GrassTxnsRes{List: list, Total: out.Total}, nil
}
