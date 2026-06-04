// This file implements the public mediaopen full node-list endpoint.

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListAllNodes returns all media node configs without pagination.
func (c *ControllerV1) ListAllNodes(ctx context.Context, _ *v1.ListAllNodesReq) (res *v1.ListAllNodesRes, err error) {
	out, err := c.mediaSvc.ListAllNodes(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.NodeInfo, 0, len(out))
	for _, item := range out {
		items = append(items, nodeOutputToPublicInfo(item))
	}
	return &v1.ListAllNodesRes{List: items}, nil
}

// nodeOutputToPublicInfo maps the service node projection into the public DTO.
func nodeOutputToPublicInfo(out *mediasvc.NodeOutput) *v1.NodeInfo {
	if out == nil {
		return nil
	}
	return &v1.NodeInfo{
		Id:         out.Id,
		NodeNum:    out.NodeNum,
		Name:       out.Name,
		QnUrl:      out.QnUrl,
		BasicUrl:   out.BasicUrl,
		DnUrl:      out.DnUrl,
		CreatorId:  out.CreatorId,
		CreateTime: out.CreateTime,
		UpdaterId:  out.UpdaterId,
		UpdateTime: out.UpdateTime,
	}
}
