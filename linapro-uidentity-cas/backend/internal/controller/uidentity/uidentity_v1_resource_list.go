package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ResourceList queries one UIdentity resource by page.
func (c *ControllerV1) ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error) {
	out, err := c.uidentitySvc.ListResource(ctx, uidentitysvc.ResourceListInput{
		Resource:    req.Resource,
		PageNum:     req.PageNum,
		PageSize:    req.PageSize,
		Keyword:     req.Keyword,
		AccountId:   req.AccountId,
		AppId:       req.AppId,
		GroupId:     req.GroupId,
		ContainerId: req.ContainerId,
		UnitId:      req.UnitId,
		Status:      req.Status,
		PassLevels:  req.PassLevels,
		GroupIds:    req.GroupIds,
		OrderBy:     req.OrderBy,
		Order:       req.Order,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResourceListRes{List: toAPIRecords(out.List), Total: out.Total}, nil
}
