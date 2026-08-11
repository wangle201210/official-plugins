// admin_v1_list_photo_audit.go implements protected evidence list reads.
package admin

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-plugin-sicau-niu/backend/api/admin/v1"
	activationphotosvc "lina-plugin-sicau-niu/backend/internal/service/activationphoto"
)

func (c *ControllerV1) ListPhotoAudit(ctx context.Context, req *v1.ListPhotoAuditReq) (*v1.ListPhotoAuditRes, error) {
	out, err := c.photoSvc.ListAudit(ctx, &activationphotosvc.AuditListInput{UserID: req.UserId, NiuID: req.NiuId, PageNum: req.PageNum, PageSize: req.PageSize})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.PhotoAuditItem, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, &v1.PhotoAuditItem{
			ActivationId: item.ActivationID, PhotoId: item.PhotoID, UserId: item.UserID, Nickname: item.Nickname,
			NiuId: item.NiuID, NiuName: item.NiuName, NiuCode: item.NiuCode,
			ContentType: item.ContentType, SizeBytes: item.SizeBytes, ActivatedAt: apitime.Milli(item.ActivatedAt),
		})
	}
	return &v1.ListPhotoAuditRes{List: list, Total: out.Total}, nil
}
