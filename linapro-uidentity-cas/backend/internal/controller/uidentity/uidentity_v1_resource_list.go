package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ResourceList queries one UIdentity resource by page.
func (c *ControllerV1) ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error) {
	out, err := c.uidentitySvc.ListResource(ctx, uidentitysvc.ResourceListInput{
		Resource:    "accounts",
		PageNum:     req.PageNum,
		PageSize:    req.PageSize,
		Filters:     legacyAccountFilters(req),
		ContainerId: req.ContainerId,
		UnitId:      req.UnitId,
		Status:      req.Status,
		PassLevels:  req.PassLevels,
		GroupIds:    req.GroupIds,
		OrderBy:     legacyAccountOrderBy(req),
		Order:       legacyAccountOrder(req),
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResourceListRes{List: toAPIRecords(out.List), Total: out.Total}, nil
}

func legacyAccountFilters(req *v1.ResourceListReq) map[string]any {
	filters := make(map[string]any, 3)
	if req.Number != "" {
		filters["number"] = req.Number
	}
	if req.Name != "" {
		filters["name"] = req.Name
	}
	if req.Phone != "" {
		filters["phone"] = req.Phone
	}
	return filters
}

func legacyAccountOrderBy(req *v1.ResourceListReq) string {
	field, _ := legacyAccountOrderPair(req)
	return field
}

func legacyAccountOrder(req *v1.ResourceListReq) string {
	_, order := legacyAccountOrderPair(req)
	return order
}

func legacyAccountOrderPair(req *v1.ResourceListReq) (string, string) {
	for _, item := range []struct {
		field string
		order string
	}{
		{"id", req.IdOrder},
		{"number", req.NumberOrder},
		{"name", req.NameOrder},
		{"phone", req.PhoneOrder},
		{"effectAt", req.EffectAtOrder},
		{"expireAt", req.ExpireAtOrder},
		{"groupId", req.GroupIdOrder},
		{"passLevel", req.PassLevelOrder},
		{"containerId", req.ContainerIdOrder},
		{"status", req.StatusOrder},
		{"createdAt", req.CreatedAtOrder},
		{"updatedAt", req.UpdatedAtOrder},
		{"deletedAt", req.DeletedAtOrder},
		{"createBy", req.CreateByOrder},
		{"updateBy", req.UpdateByOrder},
	} {
		if item.order != "" {
			return item.field, item.order
		}
	}
	return "", ""
}
