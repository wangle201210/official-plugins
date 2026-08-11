// activationphoto_audit.go implements the bounded operator photo evidence list.
package activationphoto

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListAudit returns photo-bearing activations with fixed-count batch assembly.
func (s *serviceImpl) ListAudit(ctx context.Context, in *AuditListInput) (*AuditListOutput, error) {
	pageNum, pageSize := 1, 20
	model := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().PhotoPath + " <> ''")
	if in != nil {
		if in.PageNum > 0 {
			pageNum = in.PageNum
		}
		if in.PageSize > 0 {
			pageSize = in.PageSize
		}
		if pageSize > 100 {
			pageSize = 100
		}
		if in.UserID > 0 {
			model = model.Where(dao.Activation.Columns().UserId, in.UserID)
		}
		if in.NiuID > 0 {
			model = model.Where(dao.Activation.Columns().NiuId, in.NiuID)
		}
	}
	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
	}
	rows := make([]*entitymodel.Activation, 0, pageSize)
	if err = model.Fields(dao.Activation.Columns().Id, dao.Activation.Columns().UserId, dao.Activation.Columns().NiuId, dao.Activation.Columns().PhotoPath, dao.Activation.Columns().ActivatedAt).
		OrderDesc(dao.Activation.Columns().Id).Page(pageNum, pageSize).Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
	}
	userIDs := make([]int64, 0, len(rows))
	niuIDs := make([]int64, 0, len(rows))
	photoIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserId)
		niuIDs = append(niuIDs, row.NiuId)
		photoIDs = append(photoIDs, row.PhotoPath)
	}
	users := make([]*entitymodel.User, 0)
	if len(userIDs) > 0 {
		if err = dao.User.Ctx(ctx).Fields(dao.User.Columns().Id, dao.User.Columns().Nickname).WhereIn(dao.User.Columns().Id, userIDs).Scan(&users); err != nil {
			return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
		}
	}
	niuRows := make([]*entitymodel.Niu, 0)
	if len(niuIDs) > 0 {
		if err = dao.Niu.Ctx(ctx).Fields(dao.Niu.Columns().Id, dao.Niu.Columns().Name, dao.Niu.Columns().Code).WhereIn(dao.Niu.Columns().Id, niuIDs).Scan(&niuRows); err != nil {
			return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
		}
	}
	photos := make([]*entitymodel.ActivationPhoto, 0)
	if len(photoIDs) > 0 {
		if err = dao.ActivationPhoto.Ctx(ctx).Fields(dao.ActivationPhoto.Columns().Token, dao.ActivationPhoto.Columns().ContentType, dao.ActivationPhoto.Columns().SizeBytes).WhereIn(dao.ActivationPhoto.Columns().Token, photoIDs).Scan(&photos); err != nil {
			return nil, bizerr.WrapCode(err, CodePhotoQueryFailed)
		}
	}
	names := make(map[int64]string, len(users))
	for _, row := range users {
		names[row.Id] = row.Nickname
	}
	niuByID := make(map[int64]*entitymodel.Niu, len(niuRows))
	for _, row := range niuRows {
		niuByID[row.Id] = row
	}
	photoByID := make(map[string]*entitymodel.ActivationPhoto, len(photos))
	for _, row := range photos {
		photoByID[row.Token] = row
	}
	list := make([]*AuditItem, 0, len(rows))
	for _, row := range rows {
		item := &AuditItem{ActivationID: row.Id, PhotoID: row.PhotoPath, UserID: row.UserId, Nickname: names[row.UserId], NiuID: row.NiuId, ActivatedAt: row.ActivatedAt}
		if niu := niuByID[row.NiuId]; niu != nil {
			item.NiuName, item.NiuCode = niu.Name, niu.Code
		}
		if photo := photoByID[row.PhotoPath]; photo != nil {
			item.SizeBytes, item.ContentType = photo.SizeBytes, photo.ContentType
		}
		list = append(list, item)
	}
	return &AuditListOutput{List: list, Total: total}, nil
}
